// internal/handlers/feed.go
package handlers

import (
	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FeedHandler struct {
	feedService *services.FeedService
	validator   *validator.Validate
}

func NewFeedHandler(feedService *services.FeedService) *FeedHandler {
	return &FeedHandler{
		feedService: feedService,
		validator:   validator.New(),
	}
}

// GetPersonalizedFeed retrieves user's personalized feed
func (h *FeedHandler) GetPersonalizedFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get feed type parameter
	feedType := c.DefaultQuery("type", "home")
	if !h.isValidFeedType(feedType) {
		feedType = "home"
	}

	// Get refresh parameter
	refresh := c.Query("refresh") == "true"

	feedItems, err := h.feedService.GetUserFeed(
		userID.(primitive.ObjectID),
		feedType,
		params.Limit,
		params.Offset,
		refresh,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get personalized feed", err)
		return
	}

	totalCount := int64(len(feedItems))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Personalized feed retrieved successfully", feedItems, paginationMeta, gin.H{
		"feed_type": feedType,
		"refreshed": refresh,
	})
}

// GetFollowingFeed retrieves feed from followed users only
func (h *FeedHandler) GetFollowingFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get refresh parameter
	refresh := c.Query("refresh") == "true"

	feedItems, err := h.feedService.GetUserFeed(
		userID.(primitive.ObjectID),
		"following",
		params.Limit,
		params.Offset,
		refresh,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get following feed", err)
		return
	}

	totalCount := int64(len(feedItems))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Following feed retrieved successfully", feedItems, paginationMeta, gin.H{
		"feed_type": "following",
		"refreshed": refresh,
	})
}

// GetTrendingFeed retrieves trending content feed
func (h *FeedHandler) GetTrendingFeed(c *gin.Context) {
	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get refresh parameter
	refresh := c.Query("refresh") == "true"

	// Get current user ID if authenticated
	var userID primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(primitive.ObjectID)
	}

	feedItems, err := h.feedService.GetUserFeed(
		userID,
		"trending",
		params.Limit,
		params.Offset,
		refresh,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get trending feed", err)
		return
	}

	totalCount := int64(len(feedItems))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Trending feed retrieved successfully", feedItems, paginationMeta, gin.H{
		"feed_type": "trending",
		"refreshed": refresh,
	})
}

// GetDiscoverFeed retrieves discover/explore feed
func (h *FeedHandler) GetDiscoverFeed(c *gin.Context) {
	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get refresh parameter
	refresh := c.Query("refresh") == "true"

	// Get current user ID if authenticated
	var userID primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(primitive.ObjectID)
	}

	feedItems, err := h.feedService.GetUserFeed(
		userID,
		"discover",
		params.Limit,
		params.Offset,
		refresh,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get discover feed", err)
		return
	}

	totalCount := int64(len(feedItems))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Discover feed retrieved successfully", feedItems, paginationMeta, gin.H{
		"feed_type": "discover",
		"refreshed": refresh,
	})
}

// RecordInteraction records user interaction with feed content
func (h *FeedHandler) RecordInteraction(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		PostID          string `json:"post_id" binding:"required"`
		InteractionType string `json:"interaction_type" binding:"required"`
		Source          string `json:"source" binding:"required"`
		TimeSpent       int64  `json:"time_spent"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate interaction type
	if !h.isValidInteractionType(req.InteractionType) {
		utils.BadRequestResponse(c, "Invalid interaction type. Must be one of: view, like, comment, share, save", nil)
		return
	}

	// Validate source
	if !h.isValidInteractionSource(req.Source) {
		utils.BadRequestResponse(c, "Invalid source. Must be one of: feed, profile, search, trending, discover", nil)
		return
	}

	postID, err := primitive.ObjectIDFromHex(req.PostID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	err = h.feedService.RecordInteraction(
		userID.(primitive.ObjectID),
		postID,
		req.InteractionType,
		req.Source,
		req.TimeSpent,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to record interaction", err)
		return
	}

	utils.OkResponse(c, "Interaction recorded successfully", gin.H{
		"post_id":          req.PostID,
		"interaction_type": req.InteractionType,
		"source":           req.Source,
		"time_spent":       req.TimeSpent,
	})
}

// RefreshFeed forces refresh of user's cached feed
func (h *FeedHandler) RefreshFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get feed type parameter
	feedType := c.DefaultQuery("type", "home")
	if !h.isValidFeedType(feedType) {
		feedType = "home"
	}

	err := h.feedService.RefreshUserFeed(userID.(primitive.ObjectID), feedType)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to refresh feed", err)
		return
	}

	utils.OkResponse(c, "Feed refreshed successfully", gin.H{
		"feed_type": feedType,
		"status":    "refreshed",
	})
}

// GetFeedPreferences retrieves user's feed preferences
func (h *FeedHandler) GetFeedPreferences(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// This would be implemented in the feed service
	// For now, return default preferences
	preferences := gin.H{
		"algorithm_type":       "personalized",
		"show_trending":        true,
		"show_following_only":  false,
		"content_types":        []string{"text", "image", "video"},
		"language_preferences": []string{"en"},
		"nsfw_filter":          true,
		"minimum_score":        0.0,
	}

	utils.OkResponse(c, "Feed preferences retrieved successfully", preferences)
}

// UpdateFeedPreferences updates user's feed preferences
func (h *FeedHandler) UpdateFeedPreferences(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		AlgorithmType       string   `json:"algorithm_type"`
		ShowTrending        bool     `json:"show_trending"`
		ShowFollowingOnly   bool     `json:"show_following_only"`
		ContentTypes        []string `json:"content_types"`
		LanguagePreferences []string `json:"language_preferences"`
		NSFWFilter          bool     `json:"nsfw_filter"`
		MinimumScore        float64  `json:"minimum_score"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate algorithm type
	if req.AlgorithmType != "" && !h.isValidAlgorithmType(req.AlgorithmType) {
		utils.BadRequestResponse(c, "Invalid algorithm type. Must be one of: chronological, personalized, trending", nil)
		return
	}

	// Validate content types
	for _, contentType := range req.ContentTypes {
		if !h.isValidContentType(contentType) {
			utils.BadRequestResponse(c, "Invalid content type: "+contentType, nil)
			return
		}
	}

	// This would be implemented in the feed service
	// For now, return success response
	utils.OkResponse(c, "Feed preferences updated successfully", req)
}

// GetFeedAnalytics retrieves feed analytics for user
func (h *FeedHandler) GetFeedAnalytics(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get time range parameter
	timeRange := c.DefaultQuery("time_range", "week")
	if !h.isValidTimeRange(timeRange) {
		timeRange = "week"
	}

	// This would be implemented in the feed service
	// For now, return mock analytics data
	analytics := gin.H{
		"time_range": timeRange,
		"total_interactions": gin.H{
			"views":    150,
			"likes":    45,
			"comments": 12,
			"shares":   8,
			"saves":    23,
		},
		"engagement_rate": 18.5,
		"top_content_types": []gin.H{
			{"type": "image", "percentage": 45.2},
			{"type": "text", "percentage": 32.1},
			{"type": "video", "percentage": 22.7},
		},
		"activity_timeline": []gin.H{
			{"date": "2024-01-01", "interactions": 25},
			{"date": "2024-01-02", "interactions": 32},
			{"date": "2024-01-03", "interactions": 18},
		},
		"feed_performance": gin.H{
			"home":      gin.H{"views": 89, "engagement": 15.2},
			"following": gin.H{"views": 34, "engagement": 22.1},
			"trending":  gin.H{"views": 27, "engagement": 18.9},
		},
	}

	utils.OkResponse(c, "Feed analytics retrieved successfully", analytics)
}

// HidePost hides a post from user's feed
func (h *FeedHandler) HidePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	postIDStr := c.Param("postId")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	// Record negative interaction
	err = h.feedService.RecordInteraction(
		userID.(primitive.ObjectID),
		postID,
		"hide",
		"feed",
		0,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to hide post", err)
		return
	}

	utils.OkResponse(c, "Post hidden from feed", gin.H{
		"post_id": postIDStr,
		"hidden":  true,
	})
}

// ReportFeedIssue allows users to report feed algorithm issues
func (h *FeedHandler) ReportFeedIssue(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		IssueType   string `json:"issue_type" binding:"required"`
		Description string `json:"description" binding:"required"`
		PostID      string `json:"post_id"`
		FeedType    string `json:"feed_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate issue type
	if !h.isValidIssueType(req.IssueType) {
		utils.BadRequestResponse(c, "Invalid issue type", nil)
		return
	}

	// This would be implemented to log the issue for review
	// For now, return success response
	utils.OkResponse(c, "Feed issue reported successfully", gin.H{
		"issue_type": req.IssueType,
		"reported":   true,
	})
}

// Helper methods for validation

func (h *FeedHandler) isValidFeedType(feedType string) bool {
	validTypes := []string{"home", "personal", "following", "trending", "discover"}
	for _, t := range validTypes {
		if feedType == t {
			return true
		}
	}
	return false
}

func (h *FeedHandler) isValidInteractionType(interactionType string) bool {
	validTypes := []string{"view", "like", "comment", "share", "save", "hide", "report"}
	for _, t := range validTypes {
		if interactionType == t {
			return true
		}
	}
	return false
}

func (h *FeedHandler) isValidInteractionSource(source string) bool {
	validSources := []string{"feed", "profile", "search", "trending", "discover", "notification"}
	for _, s := range validSources {
		if source == s {
			return true
		}
	}
	return false
}

func (h *FeedHandler) isValidAlgorithmType(algorithmType string) bool {
	validTypes := []string{"chronological", "personalized", "trending"}
	for _, t := range validTypes {
		if algorithmType == t {
			return true
		}
	}
	return false
}

func (h *FeedHandler) isValidContentType(contentType string) bool {
	validTypes := []string{"text", "image", "video", "audio", "link"}
	for _, t := range validTypes {
		if contentType == t {
			return true
		}
	}
	return false
}

func (h *FeedHandler) isValidTimeRange(timeRange string) bool {
	validRanges := []string{"day", "week", "month", "year"}
	for _, r := range validRanges {
		if timeRange == r {
			return true
		}
	}
	return false
}

func (h *FeedHandler) isValidIssueType(issueType string) bool {
	validTypes := []string{"inappropriate_content", "spam", "low_quality", "irrelevant", "bug", "other"}
	for _, t := range validTypes {
		if issueType == t {
			return true
		}
	}
	return false
}
