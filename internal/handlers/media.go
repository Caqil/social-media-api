// internal/handlers/media.go
package handlers

import (
	"path/filepath"
	"strconv"
	"strings"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MediaHandler struct {
	mediaService *services.MediaService
	validator    *validator.Validate
}

func NewMediaHandler(mediaService *services.MediaService) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
		validator:    validator.New(),
	}
}

// UploadMedia handles file upload
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequestResponse(c, "No file provided", err)
		return
	}
	defer file.Close()

	// Get media type from form or infer from file extension
	mediaType := c.PostForm("type")
	if mediaType == "" {
		ext := strings.ToLower(filepath.Ext(header.Filename))
		mediaType = utils.InferMediaTypeFromExtension(ext)
	}

	// Validate media type
	if !utils.IsValidMediaType(mediaType) {
		utils.BadRequestResponse(c, "Invalid media type", nil)
		return
	}

	// Create request from form data
	req := models.CreateMediaRequest{
		Type:        mediaType,
		Category:    c.PostForm("category"),
		AltText:     c.PostForm("alt_text"),
		Description: c.PostForm("description"),
		RelatedTo:   c.PostForm("related_to"),
		RelatedID:   c.PostForm("related_id"),
		IsPublic:    c.PostForm("is_public") == "true",
	}

	// Parse expiry date if provided
	if expiryStr := c.PostForm("expires_at"); expiryStr != "" {
		if expiryTime, err := utils.ParseDateTime(expiryStr); err == nil {
			req.ExpiresAt = &expiryTime
		}
	}

	result, err := h.mediaService.UploadMedia(userID.(primitive.ObjectID), file, header, req)
	if err != nil {
		if strings.Contains(err.Error(), "size exceeds") || strings.Contains(err.Error(), "unsupported") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to upload media", err)
		return
	}

	utils.CreatedResponse(c, "Media uploaded successfully", gin.H{
		"media":    result.Media.ToMediaResponse(),
		"url":      result.URL,
		"filename": result.Filename,
	})
}

// GetMedia retrieves media by ID
func (h *MediaHandler) GetMedia(c *gin.Context) {
	mediaIDStr := c.Param("id")
	mediaID, err := primitive.ObjectIDFromHex(mediaIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid media ID format", err)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	media, err := h.mediaService.GetMediaByID(mediaID, currentUserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Media not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get media", err)
		return
	}

	utils.OkResponse(c, "Media retrieved successfully", media.ToMediaResponse())
}

// GetUserMedia retrieves media uploaded by a user
func (h *MediaHandler) GetUserMedia(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(primitive.ObjectID)
		currentUserID = &id
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get media type filter
	mediaType := c.Query("type")

	media, err := h.mediaService.GetUserMedia(userID, currentUserID, mediaType, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user media", err)
		return
	}

	// Convert to response format
	var mediaResponses []models.MediaResponse
	for _, m := range media {
		mediaResponses = append(mediaResponses, m.ToMediaResponse())
	}

	totalCount := int64(len(mediaResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "User media retrieved successfully", mediaResponses, paginationMeta, nil)
}

// UpdateMedia updates media information
func (h *MediaHandler) UpdateMedia(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	mediaIDStr := c.Param("id")
	mediaID, err := primitive.ObjectIDFromHex(mediaIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid media ID format", err)
		return
	}

	var req models.UpdateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate alt text length if provided
	if req.AltText != nil && len(*req.AltText) > utils.MaxAltTextLength {
		utils.BadRequestResponse(c, "Alt text exceeds maximum length", nil)
		return
	}

	// Validate description length if provided
	if req.Description != nil && len(*req.Description) > utils.MaxMediaDescriptionLength {
		utils.BadRequestResponse(c, "Description exceeds maximum length", nil)
		return
	}

	media, err := h.mediaService.UpdateMedia(mediaID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Media not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update media", err)
		return
	}

	utils.OkResponse(c, "Media updated successfully", media.ToMediaResponse())
}

// DeleteMedia deletes media
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	mediaIDStr := c.Param("id")
	mediaID, err := primitive.ObjectIDFromHex(mediaIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid media ID format", err)
		return
	}

	err = h.mediaService.DeleteMedia(mediaID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Media not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to delete media", err)
		return
	}

	utils.OkResponse(c, "Media deleted successfully", nil)
}

// SearchMedia searches for media
func (h *MediaHandler) SearchMedia(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	if len(query) < 2 {
		utils.BadRequestResponse(c, "Search query must be at least 2 characters", nil)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get media type filter
	mediaType := c.Query("type")

	media, err := h.mediaService.SearchMedia(query, mediaType, currentUserID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search media", err)
		return
	}

	// Convert to response format
	var mediaResponses []models.MediaResponse
	for _, m := range media {
		mediaResponses = append(mediaResponses, m.ToMediaResponse())
	}

	totalCount := int64(len(mediaResponses))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Search results retrieved successfully", mediaResponses, paginationMeta, nil)
}

// GetMediaStats retrieves media statistics
func (h *MediaHandler) GetMediaStats(c *gin.Context) {
	// Get user ID if provided for user-specific stats
	var userID *primitive.ObjectID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if uid, err := primitive.ObjectIDFromHex(userIDStr); err == nil {
			userID = &uid
		}
	}

	// If no user_id provided, use current user's ID if authenticated
	if userID == nil {
		if uid, exists := c.Get("user_id"); exists {
			id := uid.(primitive.ObjectID)
			userID = &id
		}
	}

	stats, err := h.mediaService.GetMediaStats(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get media statistics", err)
		return
	}

	utils.OkResponse(c, "Media statistics retrieved successfully", stats)
}

// DownloadMedia handles media download
func (h *MediaHandler) DownloadMedia(c *gin.Context) {
	mediaIDStr := c.Param("id")
	mediaID, err := primitive.ObjectIDFromHex(mediaIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid media ID format", err)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	media, err := h.mediaService.GetMediaByID(mediaID, currentUserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Media not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get media", err)
		return
	}

	// Increment download count
	go h.mediaService.IncrementDownloadCount(mediaID)

	// Set headers for download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+media.OriginalName)
	c.Header("Content-Type", media.MimeType)

	// Serve file
	c.File(media.FilePath)
}

// GetMediaVariant retrieves a specific variant of media (thumbnail, etc.)
func (h *MediaHandler) GetMediaVariant(c *gin.Context) {
	mediaIDStr := c.Param("id")
	mediaID, err := primitive.ObjectIDFromHex(mediaIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid media ID format", err)
		return
	}

	variant := c.Param("variant")
	if variant == "" {
		variant = "original"
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	media, err := h.mediaService.GetMediaByID(mediaID, currentUserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Media not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get media", err)
		return
	}

	// Get variant URL
	url := h.mediaService.GetMediaURL(media, variant)

	utils.OkResponse(c, "Media variant URL retrieved successfully", gin.H{
		"url":     url,
		"variant": variant,
	})
}

// BulkUploadMedia handles multiple file uploads
func (h *MediaHandler) BulkUploadMedia(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		utils.BadRequestResponse(c, "Failed to parse multipart form", err)
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		utils.BadRequestResponse(c, "No files provided", nil)
		return
	}

	// Limit number of files
	if len(files) > utils.MaxBulkUploadFiles {
		utils.BadRequestResponse(c, "Too many files. Maximum allowed: "+strconv.Itoa(utils.MaxBulkUploadFiles), nil)
		return
	}

	var results []gin.H
	var errors []string

	// Get common parameters
	mediaType := c.PostForm("type")
	category := c.PostForm("category")
	isPublic := c.PostForm("is_public") == "true"

	// Process each file
	for i, header := range files {
		file, err := header.Open()
		if err != nil {
			errors = append(errors, "Failed to open file "+header.Filename+": "+err.Error())
			continue
		}

		// Infer media type if not provided
		currentMediaType := mediaType
		if currentMediaType == "" {
			ext := strings.ToLower(filepath.Ext(header.Filename))
			currentMediaType = utils.InferMediaTypeFromExtension(ext)
		}

		req := models.CreateMediaRequest{
			Type:     currentMediaType,
			Category: category,
			IsPublic: isPublic,
		}

		result, err := h.mediaService.UploadMedia(userID.(primitive.ObjectID), file, header, req)
		file.Close()

		if err != nil {
			errors = append(errors, "Failed to upload "+header.Filename+": "+err.Error())
			continue
		}

		results = append(results, gin.H{
			"index":    i,
			"filename": header.Filename,
			"media":    result.Media.ToMediaResponse(),
			"url":      result.URL,
		})
	}

	response := gin.H{
		"success_count": len(results),
		"error_count":   len(errors),
		"results":       results,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	if len(results) > 0 {
		utils.CreatedResponse(c, "Bulk upload completed", response)
	} else {
		utils.BadRequestResponse(c, "All uploads failed", response)
	}
}

// GetMediaMetadata retrieves detailed metadata for media
func (h *MediaHandler) GetMediaMetadata(c *gin.Context) {
	mediaIDStr := c.Param("id")
	mediaID, err := primitive.ObjectIDFromHex(mediaIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid media ID format", err)
		return
	}

	// Get current user ID if authenticated
	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(primitive.ObjectID)
		currentUserID = &uid
	}

	media, err := h.mediaService.GetMediaByID(mediaID, currentUserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			utils.NotFoundResponse(c, "Media not found or access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get media", err)
		return
	}

	// Return detailed metadata
	metadata := gin.H{
		"id":                media.ID.Hex(),
		"original_name":     media.OriginalName,
		"file_name":         media.FileName,
		"file_size":         media.FileSize,
		"mime_type":         media.MimeType,
		"file_extension":    media.FileExtension,
		"type":              media.Type,
		"width":             media.Width,
		"height":            media.Height,
		"duration":          media.Duration,
		"thumbnails":        media.Thumbnails,
		"variants":          media.Variants,
		"storage_provider":  media.StorageProvider,
		"storage_key":       media.StorageKey,
		"view_count":        media.ViewCount,
		"download_count":    media.DownloadCount,
		"is_processed":      media.IsProcessed,
		"processing_status": media.ProcessingStatus,
		"created_at":        media.CreatedAt,
		"updated_at":        media.UpdatedAt,
	}

	utils.OkResponse(c, "Media metadata retrieved successfully", metadata)
}
