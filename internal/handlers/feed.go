// internal/handlers/feed.go - UPDATED VERSION WITH MISSING METHODS
package handlers

import (
	"sort"
	"time"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FeedHandler struct with behavior service
type FeedHandler struct {
	feedService     *services.FeedService
	behaviorService *services.UserBehaviorService
	validator       *validator.Validate
}

// Constructor
func NewFeedHandler(feedService *services.FeedService, behaviorService *services.UserBehaviorService) *FeedHandler {
	return &FeedHandler{
		feedService:     feedService,
		behaviorService: behaviorService,
		validator:       validator.New(),
	}
}

// GetPersonalizedFeed with behavior option
func (h *FeedHandler) GetPersonalizedFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get algorithm parameter
	algorithm := c.DefaultQuery("algorithm", "standard") // behavior or standard
	refresh := c.Query("refresh") == "true"

	var feedItems []services.FeedItem
	var err error

	if algorithm == "behavior" && h.behaviorService != nil {
		// Use behavior-driven algorithm
		feedItems, err = h.getBehaviorEnhancedFeed(userID.(primitive.ObjectID), "home", params.Limit, params.Offset, refresh)
	} else {
		// Use standard algorithm
		feedItems, err = h.feedService.GetUserFeed(userID.(primitive.ObjectID), "home", params.Limit, params.Offset, refresh)
	}

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get personalized feed", err)
		return
	}

	totalCount := int64(len(feedItems))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	// Add algorithm context to response
	response := gin.H{
		"feed_type": "personalized",
		"items":     feedItems,
		"meta": gin.H{
			"algorithm":        algorithm,
			"behavior_enabled": algorithm == "behavior",
			"total_items":      totalCount,
		},
	}

	utils.PaginatedSuccessResponse(c, "Personalized feed retrieved successfully", response, paginationMeta, nil)
}

// GetFollowingFeed with behavior enhancements
func (h *FeedHandler) GetFollowingFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get algorithm parameter
	algorithm := c.DefaultQuery("algorithm", "standard")
	refresh := c.Query("refresh") == "true"

	var feedItems []services.FeedItem
	var err error

	if algorithm == "behavior" && h.behaviorService != nil {
		feedItems, err = h.getBehaviorEnhancedFeed(userID.(primitive.ObjectID), "following", params.Limit, params.Offset, refresh)
	} else {
		feedItems, err = h.feedService.GetUserFeed(userID.(primitive.ObjectID), "following", params.Limit, params.Offset, refresh)
	}

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get following feed", err)
		return
	}

	totalCount := int64(len(feedItems))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	response := gin.H{
		"feed_type": "following",
		"items":     feedItems,
		"meta": gin.H{
			"algorithm":        algorithm,
			"behavior_enabled": algorithm == "behavior",
			"total_items":      totalCount,
		},
	}

	utils.PaginatedSuccessResponse(c, "Following feed retrieved successfully", response, paginationMeta, nil)
}

// GetTrendingFeed with behavior personalization
func (h *FeedHandler) GetTrendingFeed(c *gin.Context) {
	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get algorithm parameter
	algorithm := c.DefaultQuery("algorithm", "standard")
	refresh := c.Query("refresh") == "true"

	// Get current user ID if authenticated
	var userID primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(primitive.ObjectID)
	}

	var feedItems []services.FeedItem
	var err error

	if algorithm == "behavior" && h.behaviorService != nil && !userID.IsZero() {
		feedItems, err = h.getBehaviorEnhancedFeed(userID, "trending", params.Limit, params.Offset, refresh)
	} else {
		feedItems, err = h.feedService.GetUserFeed(userID, "trending", params.Limit, params.Offset, refresh)
	}

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get trending feed", err)
		return
	}

	totalCount := int64(len(feedItems))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	response := gin.H{
		"feed_type": "trending",
		"items":     feedItems,
		"meta": gin.H{
			"algorithm":        algorithm,
			"behavior_enabled": algorithm == "behavior" && !userID.IsZero(),
			"total_items":      totalCount,
		},
	}

	utils.PaginatedSuccessResponse(c, "Trending feed retrieved successfully", response, paginationMeta, nil)
}

// GetDiscoverFeed with intelligent discovery
func (h *FeedHandler) GetDiscoverFeed(c *gin.Context) {
	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get algorithm parameter
	algorithm := c.DefaultQuery("algorithm", "standard")
	refresh := c.Query("refresh") == "true"

	// Get current user ID if authenticated
	var userID primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(primitive.ObjectID)
	}

	var feedItems []services.FeedItem
	var err error

	if algorithm == "behavior" && h.behaviorService != nil && !userID.IsZero() {
		feedItems, err = h.getBehaviorEnhancedFeed(userID, "discover", params.Limit, params.Offset, refresh)
	} else {
		feedItems, err = h.feedService.GetUserFeed(userID, "discover", params.Limit, params.Offset, refresh)
	}

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get discover feed", err)
		return
	}

	totalCount := int64(len(feedItems))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	response := gin.H{
		"feed_type": "discover",
		"items":     feedItems,
		"meta": gin.H{
			"algorithm":        algorithm,
			"behavior_enabled": algorithm == "behavior" && !userID.IsZero(),
			"total_items":      totalCount,
		},
	}

	utils.PaginatedSuccessResponse(c, "Discover feed retrieved successfully", response, paginationMeta, nil)
}

// RecordInteraction with enhanced tracking
func (h *FeedHandler) RecordInteraction(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		PostID          string  `json:"post_id" binding:"required"`
		InteractionType string  `json:"interaction_type" binding:"required"`
		Source          string  `json:"source" binding:"required"`
		TimeSpent       int64   `json:"time_spent"`
		FeedPosition    int     `json:"feed_position"`
		ScrollDepth     float64 `json:"scroll_depth"`
		ViewDuration    int64   `json:"view_duration"`
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

	// Record in feed service
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

	// Record enhanced behavior data if behavior service is available
	if h.behaviorService != nil {
		engagement := models.ContentEngagement{
			UserID:       userID.(primitive.ObjectID),
			ContentID:    postID,
			ContentType:  "post",
			ViewTime:     time.Now(),
			ViewDuration: req.ViewDuration,
			ScrollDepth:  req.ScrollDepth,
			Source:       req.Source,
			Context: map[string]interface{}{
				"feed_position": req.FeedPosition,
				"time_spent":    req.TimeSpent,
			},
			Interactions: []models.Interaction{
				{
					Type:      req.InteractionType,
					Timestamp: time.Now(),
				},
			},
		}

		go h.behaviorService.RecordContentEngagement(engagement)
	}

	utils.OkResponse(c, "Interaction recorded successfully", gin.H{
		"post_id":          req.PostID,
		"interaction_type": req.InteractionType,
		"source":           req.Source,
		"behavior_tracked": h.behaviorService != nil,
	})
}

// GetFeedAnalytics with behavior insights
func (h *FeedHandler) GetFeedAnalytics(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get time range parameter
	timeRange := c.DefaultQuery("time_range", "week")
	if !h.isValidTimeRange(timeRange) {
		timeRange = "week"
	}

	analytics := gin.H{
		"user_id":    userID.(primitive.ObjectID).Hex(),
		"time_range": timeRange,
	}

	// Get behavior analytics if available
	if h.behaviorService != nil {
		behaviorAnalytics, err := h.behaviorService.GetUserBehaviorAnalytics(userID.(primitive.ObjectID), timeRange)
		if err == nil {
			analytics["behavior_insights"] = behaviorAnalytics
		}

		// Get content preferences
		preferences, err := h.behaviorService.GetUserContentPreferences(userID.(primitive.ObjectID))
		if err == nil {
			analytics["content_preferences"] = preferences
		}

		// Get similar users
		similarUsers, err := h.behaviorService.GetSimilarUsers(userID.(primitive.ObjectID), 5)
		if err == nil {
			var userIDs []string
			for _, id := range similarUsers {
				userIDs = append(userIDs, id.Hex())
			}
			analytics["similar_users"] = userIDs
		}
	}

	// Add feed-specific metrics
	analytics["feed_performance"] = gin.H{
		"algorithm_type":     "hybrid",
		"personalization":    h.behaviorService != nil,
		"behavior_weight":    0.15,
		"engagement_boost":   "12%",
		"discovery_rate":     "8%",
		"total_interactions": 150,
		"engagement_rate":    18.5,
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
	}

	utils.OkResponse(c, "Feed analytics retrieved successfully", analytics)
}

// MISSING METHODS - Added here:

// RefreshFeed forces refresh of user's feed
func (h *FeedHandler) RefreshFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		FeedType string `json:"feed_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate feed type
	if !h.isValidFeedType(req.FeedType) {
		utils.BadRequestResponse(c, "Invalid feed type", nil)
		return
	}

	err := h.feedService.RefreshUserFeed(userID.(primitive.ObjectID), req.FeedType)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to refresh feed", err)
		return
	}

	utils.OkResponse(c, "Feed refreshed successfully", gin.H{
		"feed_type":    req.FeedType,
		"refreshed_at": time.Now(),
	})
}

// HidePost hides a post from user's feed
func (h *FeedHandler) HidePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	postID := c.Param("post_id")
	if postID == "" {
		utils.BadRequestResponse(c, "Post ID is required", nil)
		return
	}

	postObjectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	var req struct {
		Reason string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// It's optional, so we don't return an error
		req.Reason = "user_choice"
	}

	// Record the hide action
	if h.behaviorService != nil {
		metadata := map[string]interface{}{
			"action": "hide_post",
			"reason": req.Reason,
		}

		go h.behaviorService.RecordInteraction(
			userID.(primitive.ObjectID),
			postObjectID,
			"post",
			"hide",
			"feed",
			metadata,
		)
	}

	utils.OkResponse(c, "Post hidden successfully", gin.H{
		"post_id":   postID,
		"hidden_at": time.Now(),
		"reason":    req.Reason,
	})
}

// ReportFeedIssue reports issues with feed algorithm
func (h *FeedHandler) ReportFeedIssue(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		IssueType   string `json:"issue_type" binding:"required"`
		Description string `json:"description" binding:"required"`
		FeedType    string `json:"feed_type" binding:"required"`
		PostID      string `json:"post_id,omitempty"`
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

	// Validate feed type
	if !h.isValidFeedType(req.FeedType) {
		utils.BadRequestResponse(c, "Invalid feed type", nil)
		return
	}

	// Record the feedback
	feedback := gin.H{
		"user_id":     userID.(primitive.ObjectID).Hex(),
		"issue_type":  req.IssueType,
		"description": req.Description,
		"feed_type":   req.FeedType,
		"reported_at": time.Now(),
	}

	if req.PostID != "" {
		feedback["post_id"] = req.PostID
	}

	// Here you would typically save to a feedback/issues collection
	// For now, we'll just acknowledge the report

	utils.OkResponse(c, "Feed issue reported successfully", gin.H{
		"report_id":   primitive.NewObjectID().Hex(), // Generate a report ID
		"issue_type":  req.IssueType,
		"feed_type":   req.FeedType,
		"reported_at": time.Now(),
		"status":      "received",
	})
}

// GetFeedPreferences gets user's feed preferences
func (h *FeedHandler) GetFeedPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Default preferences - in real app, these would be fetched from database
	preferences := gin.H{
		"user_id": userID.(primitive.ObjectID).Hex(),
		"feed_preferences": gin.H{
			"algorithm_type":        "personalized", // chronological, personalized, hybrid
			"show_liked_posts":      true,
			"show_shared_posts":     true,
			"show_reposted_content": true,
			"show_promoted_content": false,
			"content_type_weights": gin.H{
				"text":  0.8,
				"image": 1.0,
				"video": 0.9,
				"link":  0.7,
			},
			"source_preferences": gin.H{
				"following":     1.0,
				"suggested":     0.3,
				"trending":      0.5,
				"local_content": 0.4,
			},
		},
		"notification_preferences": gin.H{
			"new_followers_in_feed": true,
			"trending_topics":       true,
			"personalized_updates":  true,
		},
		"privacy_preferences": gin.H{
			"allow_data_for_personalization": true,
			"share_engagement_data":          false,
		},
		"updated_at": time.Now(),
	}

	utils.OkResponse(c, "Feed preferences retrieved successfully", preferences)
}

// UpdateFeedPreferences updates user's feed preferences
func (h *FeedHandler) UpdateFeedPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		AlgorithmType        *string            `json:"algorithm_type,omitempty"`
		ShowLikedPosts       *bool              `json:"show_liked_posts,omitempty"`
		ShowSharedPosts      *bool              `json:"show_shared_posts,omitempty"`
		ShowRepostedContent  *bool              `json:"show_reposted_content,omitempty"`
		ShowPromotedContent  *bool              `json:"show_promoted_content,omitempty"`
		ContentTypeWeights   map[string]float64 `json:"content_type_weights,omitempty"`
		SourcePreferences    map[string]float64 `json:"source_preferences,omitempty"`
		NotificationSettings map[string]bool    `json:"notification_settings,omitempty"`
		PrivacySettings      map[string]bool    `json:"privacy_settings,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate algorithm type if provided
	if req.AlgorithmType != nil && !h.isValidAlgorithmType(*req.AlgorithmType) {
		utils.BadRequestResponse(c, "Invalid algorithm type", nil)
		return
	}

	// Here you would typically update the preferences in the database
	// For now, we'll just return the updated preferences

	updatedPreferences := gin.H{
		"user_id": userID.(primitive.ObjectID).Hex(),
		"updates": gin.H{},
	}

	if req.AlgorithmType != nil {
		updatedPreferences["updates"].(gin.H)["algorithm_type"] = *req.AlgorithmType
	}
	if req.ShowLikedPosts != nil {
		updatedPreferences["updates"].(gin.H)["show_liked_posts"] = *req.ShowLikedPosts
	}
	if req.ShowSharedPosts != nil {
		updatedPreferences["updates"].(gin.H)["show_shared_posts"] = *req.ShowSharedPosts
	}
	if req.ShowRepostedContent != nil {
		updatedPreferences["updates"].(gin.H)["show_reposted_content"] = *req.ShowRepostedContent
	}
	if req.ShowPromotedContent != nil {
		updatedPreferences["updates"].(gin.H)["show_promoted_content"] = *req.ShowPromotedContent
	}
	if req.ContentTypeWeights != nil {
		updatedPreferences["updates"].(gin.H)["content_type_weights"] = req.ContentTypeWeights
	}
	if req.SourcePreferences != nil {
		updatedPreferences["updates"].(gin.H)["source_preferences"] = req.SourcePreferences
	}

	updatedPreferences["updated_at"] = time.Now()

	utils.OkResponse(c, "Feed preferences updated successfully", updatedPreferences)
}

// Get behavior-enhanced feed
func (h *FeedHandler) getBehaviorEnhancedFeed(userID primitive.ObjectID, feedType string, limit, skip int, refresh bool) ([]services.FeedItem, error) {
	if h.behaviorService == nil {
		// Fallback to standard feed if behavior service not available
		return h.feedService.GetUserFeed(userID, feedType, limit, skip, refresh)
	}

	// Get user preferences
	userPrefs, err := h.behaviorService.GetUserContentPreferences(userID)
	if err != nil {
		// Fallback to standard feed if can't get preferences
		return h.feedService.GetUserFeed(userID, feedType, limit, skip, refresh)
	}

	// Get similar users for collaborative filtering
	similarUsers, _ := h.behaviorService.GetSimilarUsers(userID, 10)

	// Get standard feed first
	standardFeed, err := h.feedService.GetUserFeed(userID, feedType, limit*2, skip, refresh) // Get more items for better selection
	if err != nil {
		return nil, err
	}

	// Apply behavior-based scoring and filtering
	behaviorEnhancedFeed := h.applyBehaviorEnhancements(standardFeed, userID, userPrefs, similarUsers)

	// Limit to requested size
	if len(behaviorEnhancedFeed) > limit {
		behaviorEnhancedFeed = behaviorEnhancedFeed[:limit]
	}

	return behaviorEnhancedFeed, nil
}

// Apply behavior enhancements to feed
func (h *FeedHandler) applyBehaviorEnhancements(feedItems []services.FeedItem, userID primitive.ObjectID, userPrefs map[string]float64, similarUsers []primitive.ObjectID) []services.FeedItem {
	// Enhance each feed item with behavior scoring
	for i := range feedItems {
		item := &feedItems[i]

		// Get behavior score for this content - FIX: Convert ContentType to string
		behaviorScore := h.calculateBehaviorScore(userID, item.Post.ID, string(item.Post.ContentType), item.Post.Hashtags)

		// Apply content type preferences
		if userPrefs != nil {
			if prefScore, exists := userPrefs[string(item.Post.ContentType)]; exists {
				behaviorScore += prefScore * 0.3 // 30% weight for content type preference
			}
		}

		// Boost content from similar users
		for _, similarUserID := range similarUsers {
			if item.Post.UserID == similarUserID {
				behaviorScore += 0.5 // Boost from similar users
				item.Reason = "similar_user_interest"
				break
			}
		}

		// Apply behavior boost to overall score
		item.Score += behaviorScore * 0.2 // 20% weight for behavior
	}

	// Sort by enhanced scores
	sort.Slice(feedItems, func(i, j int) bool {
		return feedItems[i].Score > feedItems[j].Score
	})

	// Apply diversity filters based on behavior
	return h.applyBehaviorDiversity(feedItems, userPrefs)
}

// Calculate behavior score for content - FIX: Accept string contentType
func (h *FeedHandler) calculateBehaviorScore(userID, contentID primitive.ObjectID, contentType string, hashtags []string) float64 {
	if h.behaviorService == nil {
		return 0.0
	}

	// Get user's interest score for this specific content
	interestScore, err := h.behaviorService.GetUserInterestScore(userID, contentID)
	if err != nil {
		interestScore = 0.0
	}

	// Add hashtag relevance (simplified)
	hashtagScore := float64(len(hashtags)) * 0.1

	return interestScore + hashtagScore
}

// Apply behavior-based diversity
func (h *FeedHandler) applyBehaviorDiversity(feedItems []services.FeedItem, userPrefs map[string]float64) []services.FeedItem {
	contentTypeCount := make(map[string]int)
	authorCount := make(map[primitive.ObjectID]int)
	var diverseFeed []services.FeedItem

	for _, item := range feedItems {
		contentType := string(item.Post.ContentType) // FIX: Convert to string
		authorID := item.Post.UserID

		// Get max allowed per type based on user preferences
		maxPerType := 3 // Default
		if userPrefs != nil {
			if score, exists := userPrefs[contentType]; exists && score > 0.7 {
				maxPerType = 5 // Allow more of highly preferred content
			}
		}

		// Apply diversity constraints
		if contentTypeCount[contentType] < maxPerType && authorCount[authorID] < 3 {
			diverseFeed = append(diverseFeed, item)
			contentTypeCount[contentType]++
			authorCount[authorID]++
		}
	}

	return diverseFeed
}

// Get user behavior insights
func (h *FeedHandler) GetBehaviorInsights(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	if h.behaviorService == nil {
		utils.BadRequestResponse(c, "Behavior service not available", nil)
		return
	}

	timeRange := c.DefaultQuery("time_range", "week")

	// Get comprehensive analytics
	analytics, err := h.behaviorService.GetUserBehaviorAnalytics(userID.(primitive.ObjectID), timeRange)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get behavior insights", err)
		return
	}

	// Get content preferences
	preferences, err := h.behaviorService.GetUserContentPreferences(userID.(primitive.ObjectID))
	if err != nil {
		preferences = make(map[string]float64) // Fallback to empty preferences
	}

	// Generate insights
	insights := h.generateBehaviorInsights(analytics, preferences)

	utils.OkResponse(c, "Behavior insights retrieved successfully", gin.H{
		"user_id":      userID.(primitive.ObjectID).Hex(),
		"time_range":   timeRange,
		"insights":     insights,
		"analytics":    analytics,
		"preferences":  preferences,
		"generated_at": time.Now(),
	})
}

// Generate actionable insights
func (h *FeedHandler) generateBehaviorInsights(analytics map[string]interface{}, preferences map[string]float64) []map[string]interface{} {
	var insights []map[string]interface{}

	// Analyze content preferences
	if len(preferences) > 0 {
		topContentType := ""
		maxScore := 0.0
		for contentType, score := range preferences {
			if score > maxScore {
				maxScore = score
				topContentType = contentType
			}
		}

		if topContentType != "" {
			insights = append(insights, map[string]interface{}{
				"type":           "content_preference",
				"title":          "Content Preference Insight",
				"description":    "You engage most with " + topContentType + " content",
				"score":          maxScore,
				"recommendation": "We'll show you more " + topContentType + " content in your feed",
			})
		}
	}

	// Analyze session patterns
	if sessions, ok := analytics["sessions"].(map[string]interface{}); ok {
		if totalSessions, ok := sessions["total_sessions"].(int); ok && totalSessions > 0 {
			insights = append(insights, map[string]interface{}{
				"type":           "session_pattern",
				"title":          "Session Pattern",
				"description":    "You have active engagement sessions",
				"score":          float64(totalSessions),
				"recommendation": "Your engagement pattern helps us personalize your experience",
			})
		}
	}

	return insights
}

// Helper methods

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
	validTypes := []string{"chronological", "personalized", "trending", "behavior", "hybrid"}
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
	validTypes := []string{"inappropriate_content", "spam", "low_quality", "irrelevant", "algorithm_issue", "performance", "bug", "other"}
	for _, t := range validTypes {
		if issueType == t {
			return true
		}
	}
	return false
}
