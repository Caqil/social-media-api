// Enhanced User Behavior Tracking for Social Media API

package handlers

import (
	"social-media-api/internal/services"
	"social-media-api/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserBehaviorHandler handles advanced user behavior tracking
type UserBehaviorHandler struct {
	behaviorService  *services.UserBehaviorService
	analyticsService *services.AnalyticsService
}

// 1. SESSION TRACKING
type UserSession struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserID       primitive.ObjectID `bson:"user_id"`
	SessionID    string             `bson:"session_id"`
	StartTime    time.Time          `bson:"start_time"`
	EndTime      *time.Time         `bson:"end_time,omitempty"`
	Duration     int64              `bson:"duration"` // seconds
	DeviceInfo   string             `bson:"device_info"`
	IPAddress    string             `bson:"ip_address"`
	UserAgent    string             `bson:"user_agent"`
	PagesVisited []PageVisit        `bson:"pages_visited"`
	Actions      []UserAction       `bson:"actions"`
}

type PageVisit struct {
	URL       string    `bson:"url"`
	Timestamp time.Time `bson:"timestamp"`
	Duration  int64     `bson:"duration"` // milliseconds
}

type UserAction struct {
	Type      string                 `bson:"type"`   // click, scroll, hover, etc.
	Target    string                 `bson:"target"` // element clicked
	Timestamp time.Time              `bson:"timestamp"`
	Metadata  map[string]interface{} `bson:"metadata"`
}

// 2. CONTENT ENGAGEMENT TRACKING
type ContentEngagement struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty"`
	UserID       primitive.ObjectID     `bson:"user_id"`
	ContentID    primitive.ObjectID     `bson:"content_id"`
	ContentType  string                 `bson:"content_type"` // post, story, comment
	ViewTime     time.Time              `bson:"view_time"`
	ViewDuration int64                  `bson:"view_duration"` // milliseconds
	ScrollDepth  float64                `bson:"scroll_depth"`  // percentage
	Interactions []Interaction          `bson:"interactions"`
	Source       string                 `bson:"source"` // feed, profile, search
	Context      map[string]interface{} `bson:"context"`
}

type Interaction struct {
	Type      string    `bson:"type"` // like, share, comment, save
	Timestamp time.Time `bson:"timestamp"`
	Value     string    `bson:"value,omitempty"` // for reactions
}

// 3. USER JOURNEY TRACKING
type UserJourney struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id"`
	SessionID   string             `bson:"session_id"`
	Touchpoints []Touchpoint       `bson:"touchpoints"`
	Goal        string             `bson:"goal"` // registration, post_creation, etc.
	Completed   bool               `bson:"completed"`
	Duration    int64              `bson:"duration"`
}

type Touchpoint struct {
	Page      string                 `bson:"page"`
	Action    string                 `bson:"action"`
	Timestamp time.Time              `bson:"timestamp"`
	Metadata  map[string]interface{} `bson:"metadata"`
}

// 4. RECOMMENDATION TRACKING
type RecommendationEvent struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	UserID             primitive.ObjectID `bson:"user_id"`
	RecommendationType string             `bson:"recommendation_type"` // content, user, group
	ItemID             primitive.ObjectID `bson:"item_id"`
	Algorithm          string             `bson:"algorithm"`
	Score              float64            `bson:"score"`
	Position           int                `bson:"position"`
	Presented          time.Time          `bson:"presented"`
	Clicked            *time.Time         `bson:"clicked,omitempty"`
	Converted          *time.Time         `bson:"converted,omitempty"`
	Feedback           string             `bson:"feedback,omitempty"` // liked, hidden, reported
}

// HANDLER METHODS

// TrackPageView tracks when user visits a page
func (h *UserBehaviorHandler) TrackPageView(c *gin.Context) {
	userID, _ := c.Get("user_id")

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

	pageVisit := PageVisit{
		URL:       req.URL,
		Timestamp: time.Now(),
		Duration:  req.Duration,
	}

	err := h.behaviorService.RecordPageVisit(userID.(primitive.ObjectID), req.SessionID, pageVisit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to track page view", err)
		return
	}

	utils.OkResponse(c, "Page view tracked", nil)
}

// TrackUserAction tracks user interactions (clicks, scrolls, etc.)
func (h *UserBehaviorHandler) TrackUserAction(c *gin.Context) {
	userID, _ := c.Get("user_id")

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

	action := UserAction{
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

	utils.OkResponse(c, "User action tracked", nil)
}

// TrackContentEngagement tracks detailed content interaction
func (h *UserBehaviorHandler) TrackContentEngagement(c *gin.Context) {
	userID, _ := c.Get("user_id")

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

	engagement := ContentEngagement{
		UserID:       userID.(primitive.ObjectID),
		ContentID:    contentID,
		ContentType:  req.ContentType,
		ViewTime:     time.Now(),
		ViewDuration: req.ViewDuration,
		ScrollDepth:  req.ScrollDepth,
		Source:       req.Source,
	}

	if req.InteractionType != "" {
		engagement.Interactions = []Interaction{{
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

	utils.OkResponse(c, "Content engagement tracked", nil)
}

// GetUserBehaviorAnalytics returns comprehensive user behavior analytics
func (h *UserBehaviorHandler) GetUserBehaviorAnalytics(c *gin.Context) {
	userID, _ := c.Get("user_id")

	timeRange := c.DefaultQuery("time_range", "week") // day, week, month

	analytics, err := h.analyticsService.GetUserBehaviorAnalytics(
		userID.(primitive.ObjectID),
		timeRange,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get analytics", err)
		return
	}

	utils.OkResponse(c, "Analytics retrieved", analytics)
}

// GetRecommendationMetrics tracks recommendation performance
func (h *UserBehaviorHandler) TrackRecommendation(c *gin.Context) {
	userID, _ := c.Get("user_id")

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

	err = h.behaviorService.TrackRecommendation(userID.(primitive.ObjectID), RecommendationEvent{
		UserID:             userID.(primitive.ObjectID),
		RecommendationType: req.RecommendationType,
		ItemID:             itemID,
		Algorithm:          req.Algorithm,
		Score:              req.Score,
		Position:           req.Position,
		Presented:          time.Now(),
	}, req.Action)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to track recommendation", err)
		return
	}

	utils.OkResponse(c, "Recommendation tracked", nil)
}

// A/B Testing Support
func (h *UserBehaviorHandler) TrackExperiment(c *gin.Context) {
	userID, _ := c.Get("user_id")

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

	err := h.behaviorService.TrackExperiment(userID.(primitive.ObjectID), req.ExperimentID, req.VariantID, req.Event, req.Value)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to track experiment", err)
		return
	}

	utils.OkResponse(c, "Experiment tracked", nil)
}
