// internal/handlers/user_behavior.go
package handlers

import (
	"fmt"
	"strconv"
	"time"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserBehaviorHandler handles advanced user behavior tracking
type UserBehaviorHandler struct {
	behaviorService  *services.UserBehaviorService
	analyticsService *services.AnalyticsService
	validator        *validator.Validate
}

func NewUserBehaviorHandler(behaviorService *services.UserBehaviorService, analyticsService *services.AnalyticsService) *UserBehaviorHandler {
	return &UserBehaviorHandler{
		behaviorService:  behaviorService,
		analyticsService: analyticsService,
		validator:        validator.New(),
	}
}

// TrackPageView tracks when user visits a page
func (h *UserBehaviorHandler) TrackPageView(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		URL       string `json:"url" binding:"required"`
		Referrer  string `json:"referrer"`
		Duration  int64  `json:"duration"` // time spent on previous page
		SessionID string `json:"session_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request", err)
		return
	}

	pageVisit := models.PageVisit{
		URL:       req.URL,
		Timestamp: time.Now(),
		Duration:  req.Duration,
		Referrer:  req.Referrer,
	}

	err := h.behaviorService.RecordPageVisit(userID.(primitive.ObjectID), req.SessionID, pageVisit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to track page view", err)
		return
	}

	utils.OkResponse(c, "Page view tracked successfully", gin.H{
		"url":        req.URL,
		"session_id": req.SessionID,
		"timestamp":  pageVisit.Timestamp,
	})
}

// TrackUserAction tracks user interactions (clicks, scrolls, etc.)
func (h *UserBehaviorHandler) TrackUserAction(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		Type      string                 `json:"type" binding:"required"`
		Target    string                 `json:"target" binding:"required"`
		SessionID string                 `json:"session_id" binding:"required"`
		Metadata  map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request", err)
		return
	}

	action := models.UserAction{
		Type:      req.Type,
		Target:    req.Target,
		Timestamp: time.Now(),
		Metadata:  req.Metadata,
	}

	err := h.behaviorService.RecordUserAction(userID.(primitive.ObjectID), req.SessionID, action)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to track action", err)
		return
	}

	utils.OkResponse(c, "User action tracked successfully", gin.H{
		"type":      req.Type,
		"target":    req.Target,
		"timestamp": action.Timestamp,
	})
}

// TrackContentEngagement tracks detailed content interaction
func (h *UserBehaviorHandler) TrackContentEngagement(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		ContentID        string  `json:"content_id" binding:"required"`
		ContentType      string  `json:"content_type" binding:"required"`
		ViewDuration     int64   `json:"view_duration"` // milliseconds
		ScrollDepth      float64 `json:"scroll_depth"`  // 0-100%
		Source           string  `json:"source"`
		InteractionType  string  `json:"interaction_type,omitempty"`
		InteractionValue string  `json:"interaction_value,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request", err)
		return
	}

	contentID, err := primitive.ObjectIDFromHex(req.ContentID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid content ID", err)
		return
	}

	engagement := models.ContentEngagement{
		UserID:       userID.(primitive.ObjectID),
		ContentID:    contentID,
		ContentType:  req.ContentType,
		ViewTime:     time.Now(),
		ViewDuration: req.ViewDuration,
		ScrollDepth:  req.ScrollDepth,
		Source:       req.Source,
	}

	if req.InteractionType != "" {
		engagement.Interactions = []models.Interaction{{
			Type:      req.InteractionType,
			Timestamp: time.Now(),
			Value:     req.InteractionValue,
		}}
	}

	err = h.behaviorService.RecordContentEngagement(engagement)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to track engagement", err)
		return
	}

	utils.OkResponse(c, "Content engagement tracked successfully", gin.H{
		"content_id":    req.ContentID,
		"content_type":  req.ContentType,
		"view_duration": req.ViewDuration,
		"tracked_at":    engagement.ViewTime,
	})
}

// StartSession starts a new user session
func (h *UserBehaviorHandler) StartSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		SessionID  string `json:"session_id" binding:"required"`
		DeviceInfo string `json:"device_info"`
		UserAgent  string `json:"user_agent"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request", err)
		return
	}

	ipAddress := c.ClientIP()
	userAgent := req.UserAgent
	if userAgent == "" {
		userAgent = c.GetHeader("User-Agent")
	}

	err := h.behaviorService.StartSession(
		userID.(primitive.ObjectID),
		req.SessionID,
		req.DeviceInfo,
		ipAddress,
		userAgent,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to start session", err)
		return
	}

	utils.OkResponse(c, "Session started successfully", gin.H{
		"session_id": req.SessionID,
		"started_at": time.Now(),
		"ip_address": ipAddress,
	})
}

// EndSession ends a user session
func (h *UserBehaviorHandler) EndSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		SessionID string `json:"session_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request", err)
		return
	}

	err := h.behaviorService.EndSession(req.SessionID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to end session", err)
		return
	}

	utils.OkResponse(c, "Session ended successfully", gin.H{
		"session_id": req.SessionID,
		"ended_at":   time.Now(),
		"user_id":    userID.(primitive.ObjectID).Hex(),
	})
}

// GetUserBehaviorAnalytics returns comprehensive user behavior analytics
func (h *UserBehaviorHandler) GetUserBehaviorAnalytics(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	timeRange := c.DefaultQuery("time_range", "week") // day, week, month

	analytics, err := h.behaviorService.GetUserBehaviorAnalytics(
		userID.(primitive.ObjectID),
		timeRange,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get analytics", err)
		return
	}

	utils.OkResponse(c, "Analytics retrieved successfully", gin.H{
		"user_id":      userID.(primitive.ObjectID).Hex(),
		"time_range":   timeRange,
		"analytics":    analytics,
		"generated_at": time.Now(),
	})
}

// GetUserContentPreferences returns user's content preferences
func (h *UserBehaviorHandler) GetUserContentPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	preferences, err := h.behaviorService.GetUserContentPreferences(userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get content preferences", err)
		return
	}

	utils.OkResponse(c, "Content preferences retrieved successfully", gin.H{
		"user_id":      userID.(primitive.ObjectID).Hex(),
		"preferences":  preferences,
		"retrieved_at": time.Now(),
	})
}

// GetSimilarUsers returns users with similar behavior patterns
func (h *UserBehaviorHandler) GetSimilarUsers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	similarUsers, err := h.behaviorService.GetSimilarUsers(userID.(primitive.ObjectID), limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get similar users", err)
		return
	}

	// Convert ObjectIDs to strings for response
	var userIDs []string
	for _, id := range similarUsers {
		userIDs = append(userIDs, id.Hex())
	}

	utils.OkResponse(c, "Similar users retrieved successfully", gin.H{
		"user_id":       userID.(primitive.ObjectID).Hex(),
		"similar_users": userIDs,
		"count":         len(userIDs),
		"limit":         limit,
	})
}

// TrackRecommendation tracks recommendation performance
func (h *UserBehaviorHandler) TrackRecommendation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		RecommendationType string  `json:"recommendation_type" binding:"required"`
		ItemID             string  `json:"item_id" binding:"required"`
		Algorithm          string  `json:"algorithm" binding:"required"`
		Score              float64 `json:"score"`
		Position           int     `json:"position"`
		Action             string  `json:"action"` // presented, clicked, converted
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request", err)
		return
	}

	itemID, err := primitive.ObjectIDFromHex(req.ItemID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid item ID", err)
		return
	}

	event := models.RecommendationEvent{
		UserID:             userID.(primitive.ObjectID),
		RecommendationType: req.RecommendationType,
		ItemID:             itemID,
		Algorithm:          req.Algorithm,
		Score:              req.Score,
		Position:           req.Position,
		Presented:          time.Now(),
	}

	err = h.behaviorService.TrackRecommendation(userID.(primitive.ObjectID), event, req.Action)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to track recommendation", err)
		return
	}

	utils.OkResponse(c, "Recommendation tracked successfully", gin.H{
		"recommendation_type": req.RecommendationType,
		"item_id":             req.ItemID,
		"action":              req.Action,
		"tracked_at":          time.Now(),
	})
}

// TrackExperiment tracks A/B testing experiments
func (h *UserBehaviorHandler) TrackExperiment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		ExperimentID string  `json:"experiment_id" binding:"required"`
		VariantID    string  `json:"variant_id" binding:"required"`
		Event        string  `json:"event" binding:"required"` // exposure, conversion
		Value        float64 `json:"value,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request", err)
		return
	}

	err := h.behaviorService.TrackExperiment(
		userID.(primitive.ObjectID),
		req.ExperimentID,
		req.VariantID,
		req.Event,
		req.Value,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to track experiment", err)
		return
	}

	utils.OkResponse(c, "Experiment tracked successfully", gin.H{
		"experiment_id": req.ExperimentID,
		"variant_id":    req.VariantID,
		"event":         req.Event,
		"value":         req.Value,
		"tracked_at":    time.Now(),
	})
}

// GetInterestScore gets user's interest score for specific content
func (h *UserBehaviorHandler) GetInterestScore(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	contentIDStr := c.Param("contentId")
	contentID, err := primitive.ObjectIDFromHex(contentIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid content ID", err)
		return
	}

	score, err := h.behaviorService.GetUserInterestScore(userID.(primitive.ObjectID), contentID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get interest score", err)
		return
	}

	utils.OkResponse(c, "Interest score retrieved successfully", gin.H{
		"user_id":       userID.(primitive.ObjectID).Hex(),
		"content_id":    contentIDStr,
		"score":         score,
		"calculated_at": time.Now(),
	})
}

// RecordInteraction records generic user interaction for behavior learning
func (h *UserBehaviorHandler) RecordInteraction(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		ContentID       string                 `json:"content_id" binding:"required"`
		ContentType     string                 `json:"content_type" binding:"required"`
		InteractionType string                 `json:"interaction_type" binding:"required"`
		Source          string                 `json:"source" binding:"required"`
		Metadata        map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request", err)
		return
	}

	contentID, err := primitive.ObjectIDFromHex(req.ContentID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid content ID", err)
		return
	}

	err = h.behaviorService.RecordInteraction(
		userID.(primitive.ObjectID),
		contentID,
		req.ContentType,
		req.InteractionType,
		req.Source,
		req.Metadata,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to record interaction", err)
		return
	}

	utils.OkResponse(c, "Interaction recorded successfully", gin.H{
		"content_id":       req.ContentID,
		"content_type":     req.ContentType,
		"interaction_type": req.InteractionType,
		"source":           req.Source,
		"recorded_at":      time.Now(),
	})
}

// GetBehaviorInsights provides actionable insights based on user behavior
func (h *UserBehaviorHandler) GetBehaviorInsights(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
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

// Helper method to generate actionable insights
func (h *UserBehaviorHandler) generateBehaviorInsights(analytics map[string]interface{}, preferences map[string]float64) []map[string]interface{} {
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
				"description":    fmt.Sprintf("You engage most with %s content (score: %.2f)", topContentType, maxScore),
				"score":          maxScore,
				"recommendation": fmt.Sprintf("We'll show you more %s content in your feed", topContentType),
			})
		}
	}

	// Analyze session patterns
	if sessions, ok := analytics["sessions"].(map[string]interface{}); ok {
		if totalSessions, ok := sessions["total_sessions"].(int); ok && totalSessions > 0 {
			if avgDuration, ok := sessions["avg_duration"].(float64); ok {
				insight := map[string]interface{}{
					"type":  "session_pattern",
					"title": "Session Pattern Insight",
					"score": avgDuration / 60, // Convert to minutes
				}

				if avgDuration > 1800000 { // 30 minutes in milliseconds
					insight["description"] = "You have long, engaged sessions"
					insight["recommendation"] = "Consider exploring more diverse content during your sessions"
				} else if avgDuration < 300000 { // 5 minutes
					insight["description"] = "You have quick, focused sessions"
					insight["recommendation"] = "We'll prioritize high-quality content in your feed"
				} else {
					insight["description"] = "You have balanced session lengths"
					insight["recommendation"] = "Your current engagement pattern looks healthy"
				}

				insights = append(insights, insight)
			}
		}
	}

	// Add engagement insights
	if engagement, ok := analytics["engagement"].(map[string]interface{}); ok {
		if byContentType, ok := engagement["by_content_type"].([]map[string]interface{}); ok && len(byContentType) > 0 {
			insights = append(insights, map[string]interface{}{
				"type":           "engagement_pattern",
				"title":          "Engagement Pattern",
				"description":    fmt.Sprintf("You actively engage with %d different content types", len(byContentType)),
				"score":          float64(len(byContentType)),
				"recommendation": "Your diverse engagement helps us personalize your experience better",
			})
		}
	}

	return insights
}
