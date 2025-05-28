// internal/handlers/user.go
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

type UserHandler struct {
	userService *services.UserService
	validator   *validator.Validate
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator.New(),
	}
}

// GetUserProfile retrieves user profile by ID
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	// Get current user ID if authenticated
	var currentUserID primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		currentUserID = uid.(primitive.ObjectID)
	}

	profile, err := h.userService.GetUserProfile(userID, currentUserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get user profile", err)
		return
	}

	utils.OkResponse(c, "User profile retrieved successfully", profile)
}

// GetUserByUsername retrieves user profile by username
func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		utils.BadRequestResponse(c, "Username is required", nil)
		return
	}

	user, err := h.userService.GetUserByUsername(username)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get user", err)
		return
	}

	// Get current user ID if authenticated for context
	var currentUserID primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		currentUserID = uid.(primitive.ObjectID)
	}

	// Get full profile with context
	profile, err := h.userService.GetUserProfile(user.ID, currentUserID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user profile", err)
		return
	}

	utils.OkResponse(c, "User profile retrieved successfully", profile)
}

// UpdateProfile updates user profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate bio length if provided
	if req.Bio != nil && len(*req.Bio) > utils.MaxBioLength {
		utils.BadRequestResponse(c, "Bio exceeds maximum length", nil)
		return
	}

	// Validate display name length if provided
	if req.DisplayName != nil && len(*req.DisplayName) > utils.MaxDisplayNameLength {
		utils.BadRequestResponse(c, "Display name exceeds maximum length", nil)
		return
	}

	user, err := h.userService.UpdateUser(userID.(primitive.ObjectID), req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update profile", err)
		return
	}

	utils.ProfileUpdateSuccessResponse(c, user.ToUserResponse())
}

// UpdatePrivacySettings updates user privacy settings
func (h *UserHandler) UpdatePrivacySettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.PrivacySettings
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	err := h.userService.UpdateUserPrivacySettings(userID.(primitive.ObjectID), req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update privacy settings", err)
		return
	}

	utils.OkResponse(c, "Privacy settings updated successfully", req)
}

// UpdateNotificationSettings updates user notification settings
func (h *UserHandler) UpdateNotificationSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.NotificationSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	err := h.userService.UpdateNotificationSettings(userID.(primitive.ObjectID), req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update notification settings", err)
		return
	}

	utils.OkResponse(c, "Notification settings updated successfully", req)
}

// SearchUsers searches for users
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	if len(query) < 2 {
		utils.BadRequestResponse(c, "Search query must be at least 2 characters", nil)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	users, err := h.userService.SearchUsers(query, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search users", err)
		return
	}

	// Convert to response format
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToUserResponse())
	}

	totalCount := int64(len(userResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Search results retrieved successfully", userResponses, paginationMeta, nil)
}

// GetUserStats retrieves user statistics
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	stats, err := h.userService.GetUserStats(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get user statistics", err)
		return
	}

	utils.OkResponse(c, "User statistics retrieved successfully", stats)
}

// GetSuggestedUsers gets suggested users for current user
func (h *UserHandler) GetSuggestedUsers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := utils.StringToInt(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	users, err := h.userService.GetSuggestedUsers(userID.(primitive.ObjectID), limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get suggested users", err)
		return
	}

	// Convert to response format
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToUserResponse())
	}

	utils.OkResponse(c, "Suggested users retrieved successfully", userResponses)
}

// BlockUser blocks a user
func (h *UserHandler) BlockUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	blockedUserIDStr := c.Param("id")
	blockedUserID, err := primitive.ObjectIDFromHex(blockedUserIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	// Check if user is trying to block themselves
	if userID.(primitive.ObjectID) == blockedUserID {
		utils.BadRequestResponse(c, "Cannot block yourself", nil)
		return
	}

	err = h.userService.BlockUser(userID.(primitive.ObjectID), blockedUserID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to block user", err)
		return
	}

	utils.OkResponse(c, "User blocked successfully", gin.H{
		"blocked_user_id": blockedUserIDStr,
		"blocked":         true,
	})
}

// UnblockUser unblocks a user
func (h *UserHandler) UnblockUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	blockedUserIDStr := c.Param("id")
	blockedUserID, err := primitive.ObjectIDFromHex(blockedUserIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	err = h.userService.UnblockUser(userID.(primitive.ObjectID), blockedUserID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to unblock user", err)
		return
	}

	utils.OkResponse(c, "User unblocked successfully", gin.H{
		"blocked_user_id": blockedUserIDStr,
		"blocked":         false,
	})
}

// GetBlockedUsers retrieves list of blocked users
func (h *UserHandler) GetBlockedUsers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	users, err := h.userService.GetBlockedUsers(userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get blocked users", err)
		return
	}

	// Convert to response format
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToUserResponse())
	}

	utils.OkResponse(c, "Blocked users retrieved successfully", userResponses)
}

// UpdateUserActivity updates user's activity status
func (h *UserHandler) UpdateUserActivity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate status
	if req.Status != "online" && req.Status != "away" && req.Status != "busy" && req.Status != "invisible" {
		utils.BadRequestResponse(c, "Invalid status. Must be one of: online, away, busy, invisible", nil)
		return
	}

	err := h.userService.UpdateUserActivity(userID.(primitive.ObjectID), req.Status)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update user activity", err)
		return
	}

	utils.OkResponse(c, "User activity updated successfully", gin.H{
		"status": req.Status,
	})
}

// VerifyUser verifies a user account (admin only)
func (h *UserHandler) VerifyUser(c *gin.Context) {
	// Check if current user is admin
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	currentUser, err := h.userService.GetUserByID(currentUserID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get current user", err)
		return
	}

	// FIXED: Use correct role constants
	if currentUser.Role != models.RoleAdmin && currentUser.Role != models.RoleSuperAdmin {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	userIDStr := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	err = h.userService.VerifyUser(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to verify user", err)
		return
	}

	utils.OkResponse(c, "User verified successfully", gin.H{
		"user_id":  userIDStr,
		"verified": true,
	})
}

// SuspendUser suspends a user account (admin only)
func (h *UserHandler) SuspendUser(c *gin.Context) {
	// Check if current user is admin
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	currentUser, err := h.userService.GetUserByID(currentUserID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get current user", err)
		return
	}

	// FIXED: Use correct role constants
	if currentUser.Role != models.RoleAdmin && currentUser.Role != models.RoleSuperAdmin {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	userIDStr := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	err = h.userService.SuspendUser(userID, req.Reason)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to suspend user", err)
		return
	}

	utils.OkResponse(c, "User suspended successfully", gin.H{
		"user_id":   userIDStr,
		"suspended": true,
		"reason":    req.Reason,
	})
}

// UnsuspendUser unsuspends a user account (admin only)
func (h *UserHandler) UnsuspendUser(c *gin.Context) {
	// Check if current user is admin
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	currentUser, err := h.userService.GetUserByID(currentUserID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get current user", err)
		return
	}

	// FIXED: Use correct role constants
	if currentUser.Role != models.RoleAdmin && currentUser.Role != models.RoleSuperAdmin {
		utils.ForbiddenResponse(c, "Insufficient permissions")
		return
	}

	userIDStr := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	err = h.userService.UnsuspendUser(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to unsuspend user", err)
		return
	}

	utils.OkResponse(c, "User unsuspended successfully", gin.H{
		"user_id":   userIDStr,
		"suspended": false,
	})
}

// DeactivateAccount allows user to deactivate their own account
func (h *UserHandler) DeactivateAccount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
		Reason   string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Verify password before deactivation
	user, err := h.userService.GetUserByID(userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user", err)
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		utils.BadRequestResponse(c, "Invalid password", nil)
		return
	}

	err = h.userService.SoftDeleteUser(userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to deactivate account", err)
		return
	}

	utils.OkResponse(c, "Account deactivated successfully", gin.H{
		"deactivated": true,
		"reason":      req.Reason,
	})
}
