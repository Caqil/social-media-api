// internal/handlers/report.go
package handlers

import (
	"strings"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReportHandler struct {
	reportService *services.ReportService
	validator     *validator.Validate
}

func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		validator:     validator.New(),
	}
}

// CreateReport creates a new report
func (h *ReportHandler) CreateReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.TargetID) == "" {
		utils.BadRequestResponse(c, "Target ID is required", nil)
		return
	}

	if req.Reason == "" {
		utils.BadRequestResponse(c, "Report reason is required", nil)
		return
	}

	// Validate target type
	validTargetTypes := []string{"user", "post", "comment", "story", "group", "event", "message"}
	if !h.isValidTargetType(req.TargetType, validTargetTypes) {
		utils.BadRequestResponse(c, "Invalid target type", nil)
		return
	}

	// Validate description length if provided
	if len(req.Description) > 1000 {
		utils.BadRequestResponse(c, "Description exceeds maximum length of 1000 characters", nil)
		return
	}

	report, err := h.reportService.CreateReport(userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "already reported") {
			utils.ConflictResponse(c, "You have already reported this content", err)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Target content not found")
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to create report", err)
		return
	}

	utils.CreatedResponse(c, "Report created successfully", report.ToReportResponse())
}

// GetReports retrieves reports with filtering and pagination
func (h *ReportHandler) GetReports(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user is moderator/admin
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Build filter from query parameters
	filter := models.ReportFilter{
		Status:     c.Query("status"),
		TargetType: c.Query("target_type"),
		Priority:   c.Query("priority"),
		ReporterID: c.Query("reporter_id"),
		AssignedTo: c.Query("assigned_to"),
		SortBy:     c.DefaultQuery("sort_by", "newest"),
	}

	// Handle reason filter
	if reason := c.Query("reason"); reason != "" {
		filter.Reason = models.ReportReason(reason)
	}

	// Handle boolean filters
	if autoDetected := c.Query("auto_detected"); autoDetected != "" {
		if autoDetected == "true" {
			val := true
			filter.AutoDetected = &val
		} else if autoDetected == "false" {
			val := false
			filter.AutoDetected = &val
		}
	}

	if requiresFollowUp := c.Query("requires_follow_up"); requiresFollowUp != "" {
		if requiresFollowUp == "true" {
			val := true
			filter.RequiresFollowUp = &val
		} else if requiresFollowUp == "false" {
			val := false
			filter.RequiresFollowUp = &val
		}
	}

	// Handle date filters
	if createdAfter := c.Query("created_after"); createdAfter != "" {
		if date, err := utils.ParseDateTime(createdAfter); err == nil {
			filter.CreatedAfter = date
		}
	}

	if createdBefore := c.Query("created_before"); createdBefore != "" {
		if date, err := utils.ParseDateTime(createdBefore); err == nil {
			filter.CreatedBefore = date
		}
	}

	reports, totalCount, err := h.reportService.GetReports(filter, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get reports", err)
		return
	}

	// Convert to response format
	var reportResponses []models.ReportResponse
	for _, report := range reports {
		reportResponses = append(reportResponses, report.ToReportResponse())
	}

	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Reports retrieved successfully", reportResponses, paginationMeta, nil)
}

// GetReport retrieves a specific report by ID
func (h *ReportHandler) GetReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	reportIDStr := c.Param("id")
	reportID, err := primitive.ObjectIDFromHex(reportIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid report ID format", err)
		return
	}

	report, err := h.reportService.GetReportByID(reportID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Report not found")
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get report", err)
		return
	}

	utils.OkResponse(c, "Report retrieved successfully", report.ToReportResponse())
}

// UpdateReport updates a report (moderator only)
func (h *ReportHandler) UpdateReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user is moderator/admin
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	reportIDStr := c.Param("id")
	reportID, err := primitive.ObjectIDFromHex(reportIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid report ID format", err)
		return
	}

	var req models.UpdateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := []string{"pending", "reviewing", "resolved", "rejected"}
		if !h.isValidStatus(string(*req.Status), validStatuses) {
			utils.BadRequestResponse(c, "Invalid status", nil)
			return
		}
	}

	// Validate priority if provided
	if req.Priority != nil {
		validPriorities := []string{"low", "medium", "high", "urgent"}
		if !h.isValidPriority(*req.Priority, validPriorities) {
			utils.BadRequestResponse(c, "Invalid priority", nil)
			return
		}
	}

	// Validate resolution note length if provided
	if req.ResolutionNote != nil && len(*req.ResolutionNote) > 2000 {
		utils.BadRequestResponse(c, "Resolution note exceeds maximum length", nil)
		return
	}

	report, err := h.reportService.UpdateReport(reportID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Report not found")
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update report", err)
		return
	}

	utils.OkResponse(c, "Report updated successfully", report.ToReportResponse())
}

// AssignReport assigns a report to a moderator
func (h *ReportHandler) AssignReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user is moderator/admin
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	reportIDStr := c.Param("id")
	reportID, err := primitive.ObjectIDFromHex(reportIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid report ID format", err)
		return
	}

	var req models.ReportAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	assignedTo, err := primitive.ObjectIDFromHex(req.AssignedTo)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid assigned_to ID format", err)
		return
	}

	err = h.reportService.AssignReport(reportID, userID.(primitive.ObjectID), assignedTo)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Report or assignee not found")
			return
		}
		if strings.Contains(err.Error(), "only assign to moderators") {
			utils.BadRequestResponse(c, "Can only assign reports to moderators or admins", err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to assign report", err)
		return
	}

	utils.OkResponse(c, "Report assigned successfully", gin.H{
		"report_id":   reportIDStr,
		"assigned_to": req.AssignedTo,
	})
}

// ResolveReport resolves a report
func (h *ReportHandler) ResolveReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user is moderator/admin
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	reportIDStr := c.Param("id")
	reportID, err := primitive.ObjectIDFromHex(reportIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid report ID format", err)
		return
	}

	var req models.ReportResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate resolution length
	if len(req.Resolution) > 500 {
		utils.BadRequestResponse(c, "Resolution exceeds maximum length", nil)
		return
	}

	// Validate note length
	if len(req.Note) > 2000 {
		utils.BadRequestResponse(c, "Note exceeds maximum length", nil)
		return
	}

	err = h.reportService.ResolveReport(reportID, userID.(primitive.ObjectID), req.Resolution, req.Note)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Report not found")
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to resolve report", err)
		return
	}

	utils.OkResponse(c, "Report resolved successfully", gin.H{
		"report_id":  reportIDStr,
		"status":     "resolved",
		"resolution": req.Resolution,
	})
}

// RejectReport rejects a report
func (h *ReportHandler) RejectReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user is moderator/admin
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	reportIDStr := c.Param("id")
	reportID, err := primitive.ObjectIDFromHex(reportIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid report ID format", err)
		return
	}

	var req models.ReportRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate note length
	if len(req.Note) > 2000 {
		utils.BadRequestResponse(c, "Note exceeds maximum length", nil)
		return
	}

	if strings.TrimSpace(req.Note) == "" {
		utils.BadRequestResponse(c, "Rejection note is required", nil)
		return
	}

	err = h.reportService.RejectReport(reportID, userID.(primitive.ObjectID), req.Note)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Report not found")
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to reject report", err)
		return
	}

	utils.OkResponse(c, "Report rejected successfully", gin.H{
		"report_id": reportIDStr,
		"status":    "rejected",
		"note":      req.Note,
	})
}

// GetReportStats retrieves report statistics
func (h *ReportHandler) GetReportStats(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user is moderator/admin
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	stats, err := h.reportService.GetReportStats()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get report statistics", err)
		return
	}

	utils.OkResponse(c, "Report statistics retrieved successfully", stats)
}

// GetReportSummary retrieves report summary by reason
func (h *ReportHandler) GetReportSummary(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user is moderator/admin
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	summary, err := h.reportService.GetReportSummary()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get report summary", err)
		return
	}

	utils.OkResponse(c, "Report summary retrieved successfully", summary)
}

// GetUserReports retrieves reports made by a specific user
func (h *ReportHandler) GetUserReports(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userIDStr := c.Param("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	// Users can only see their own reports unless they're moderators
	if userID != currentUserID.(primitive.ObjectID) && !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	reports, err := h.reportService.GetUserReports(userID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user reports", err)
		return
	}

	// Convert to response format
	var reportResponses []models.ReportResponse
	for _, report := range reports {
		reportResponses = append(reportResponses, report.ToReportResponse())
	}

	totalCount := int64(len(reportResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "User reports retrieved successfully", reportResponses, paginationMeta, nil)
}

// GetReportsByTarget retrieves reports for a specific target
func (h *ReportHandler) GetReportsByTarget(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user is moderator/admin
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	targetType := c.Param("type")
	targetIDStr := c.Param("id")

	targetID, err := primitive.ObjectIDFromHex(targetIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid target ID format", err)
		return
	}

	// Validate target type
	validTargetTypes := []string{"user", "post", "comment", "story", "group", "event", "message"}
	if !h.isValidTargetType(targetType, validTargetTypes) {
		utils.BadRequestResponse(c, "Invalid target type", nil)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	reports, err := h.reportService.GetReportsByTarget(targetType, targetID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get target reports", err)
		return
	}

	// Convert to response format
	var reportResponses []models.ReportResponse
	for _, report := range reports {
		reportResponses = append(reportResponses, report.ToReportResponse())
	}

	totalCount := int64(len(reportResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	responseData := gin.H{
		"target_type": targetType,
		"target_id":   targetIDStr,
		"reports":     reportResponses,
	}

	utils.PaginatedSuccessResponse(c, "Target reports retrieved successfully", responseData, paginationMeta, nil)
}

// BulkUpdateReports updates multiple reports at once
func (h *ReportHandler) BulkUpdateReports(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user is moderator/admin
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	var req models.BulkReportUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if len(req.ReportIDs) == 0 {
		utils.BadRequestResponse(c, "At least one report ID is required", nil)
		return
	}

	if len(req.ReportIDs) > 50 {
		utils.BadRequestResponse(c, "Cannot update more than 50 reports at once", nil)
		return
	}

	// Validate action
	validActions := []string{"assign", "resolve", "reject", "update_priority", "update_status"}
	if !h.isValidAction(req.Action, validActions) {
		utils.BadRequestResponse(c, "Invalid action", nil)
		return
	}

	successCount := 0
	failedReports := make([]string, 0)

	for _, reportIDStr := range req.ReportIDs {
		reportID, err := primitive.ObjectIDFromHex(reportIDStr)
		if err != nil {
			failedReports = append(failedReports, reportIDStr+": invalid ID format")
			continue
		}

		_, err = h.reportService.UpdateReport(reportID, userID.(primitive.ObjectID), req.Data)
		if err != nil {
			failedReports = append(failedReports, reportIDStr+": "+err.Error())
		} else {
			successCount++
		}
	}

	responseData := gin.H{
		"total_reports":    len(req.ReportIDs),
		"successful_count": successCount,
		"failed_count":     len(failedReports),
		"action":           req.Action,
	}

	if len(failedReports) > 0 {
		responseData["failed_reports"] = failedReports
	}

	if successCount > 0 {
		utils.OkResponse(c, "Bulk update completed", responseData)
	} else {
		utils.BadRequestResponse(c, "All updates failed", nil)
	}
}

// GetMyReports retrieves reports made by the current user
func (h *ReportHandler) GetMyReports(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	reports, err := h.reportService.GetUserReports(userID.(primitive.ObjectID), params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get your reports", err)
		return
	}

	// Convert to response format
	var reportResponses []models.ReportResponse
	for _, report := range reports {
		reportResponses = append(reportResponses, report.ToReportResponse())
	}

	totalCount := int64(len(reportResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Your reports retrieved successfully", reportResponses, paginationMeta, nil)
}

// GetPendingReports retrieves pending reports assigned to current user
func (h *ReportHandler) GetPendingReports(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user is moderator/admin
	if !h.isModerator(c) {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Build filter for pending reports assigned to current user
	filter := models.ReportFilter{
		AssignedTo: userID.(primitive.ObjectID).Hex(),
		Status:     "pending",
		SortBy:     "priority",
	}

	reports, totalCount, err := h.reportService.GetReports(filter, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get pending reports", err)
		return
	}

	// Convert to response format
	var reportResponses []models.ReportResponse
	for _, report := range reports {
		reportResponses = append(reportResponses, report.ToReportResponse())
	}

	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Pending reports retrieved successfully", reportResponses, paginationMeta, nil)
}

// Helper methods

func (h *ReportHandler) isModerator(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}

	role, ok := userRole.(models.UserRole)
	if !ok {
		return false
	}

	return role == models.RoleModerator || role == models.RoleAdmin || role == models.RoleSuperAdmin
}

func (h *ReportHandler) isValidTargetType(targetType string, validTypes []string) bool {
	for _, validType := range validTypes {
		if targetType == validType {
			return true
		}
	}
	return false
}

func (h *ReportHandler) isValidStatus(status string, validStatuses []string) bool {
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

func (h *ReportHandler) isValidPriority(priority string, validPriorities []string) bool {
	for _, validPriority := range validPriorities {
		if priority == validPriority {
			return true
		}
	}
	return false
}

func (h *ReportHandler) isValidAction(action string, validActions []string) bool {
	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

// Additional utility methods for report reasons

// GetReportReasons retrieves all available report reasons
func (h *ReportHandler) GetReportReasons(c *gin.Context) {
	reasons := []gin.H{
		{"value": "spam", "label": "Spam", "description": "Repetitive, unwanted, or promotional content"},
		{"value": "harassment", "label": "Harassment or Bullying", "description": "Content that targets individuals with harmful intent"},
		{"value": "hate_speech", "label": "Hate Speech", "description": "Content that promotes hatred against individuals or groups"},
		{"value": "violence", "label": "Violence or Threats", "description": "Content depicting or threatening violence"},
		{"value": "nudity", "label": "Nudity or Sexual Content", "description": "Inappropriate sexual or nude content"},
		{"value": "fake_news", "label": "False Information", "description": "Misleading or false information"},
		{"value": "copyright", "label": "Copyright Violation", "description": "Unauthorized use of copyrighted material"},
		{"value": "other", "label": "Other", "description": "Other violations not covered above"},
	}

	utils.OkResponse(c, "Report reasons retrieved successfully", gin.H{
		"reasons": reasons,
	})
}

// GetReportCategories retrieves report categories for filtering
func (h *ReportHandler) GetReportCategories(c *gin.Context) {
	categories := gin.H{
		"target_types": []string{"user", "post", "comment", "story", "group", "event", "message"},
		"statuses":     []string{"pending", "reviewing", "resolved", "rejected"},
		"priorities":   []string{"low", "medium", "high", "urgent"},
		"reasons":      []string{"spam", "harassment", "hate_speech", "violence", "nudity", "fake_news", "copyright", "other"},
	}

	utils.OkResponse(c, "Report categories retrieved successfully", categories)
}
