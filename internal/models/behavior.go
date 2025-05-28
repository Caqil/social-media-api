// internal/models/user_behavior.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserSession tracks user session information
type UserSession struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	SessionID    string             `bson:"session_id" json:"session_id"`
	StartTime    time.Time          `bson:"start_time" json:"start_time"`
	EndTime      *time.Time         `bson:"end_time,omitempty" json:"end_time,omitempty"`
	Duration     int64              `bson:"duration" json:"duration"` // milliseconds
	DeviceInfo   string             `bson:"device_info" json:"device_info"`
	IPAddress    string             `bson:"ip_address" json:"ip_address"`
	UserAgent    string             `bson:"user_agent" json:"user_agent"`
	PagesVisited []PageVisit        `bson:"pages_visited" json:"pages_visited"`
	Actions      []UserAction       `bson:"actions" json:"actions"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// PageVisit tracks individual page visits within a session
type PageVisit struct {
	URL       string    `bson:"url" json:"url"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	Duration  int64     `bson:"duration" json:"duration"` // milliseconds
	Referrer  string    `bson:"referrer,omitempty" json:"referrer,omitempty"`
}

// UserAction tracks specific user actions
type UserAction struct {
	Type      string                 `bson:"type" json:"type"`     // click, scroll, hover, etc.
	Target    string                 `bson:"target" json:"target"` // element clicked
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// ContentEngagement tracks detailed content interaction
type ContentEngagement struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	UserID       primitive.ObjectID     `bson:"user_id" json:"user_id"`
	ContentID    primitive.ObjectID     `bson:"content_id" json:"content_id"`
	ContentType  string                 `bson:"content_type" json:"content_type"` // post, story, comment
	ViewTime     time.Time              `bson:"view_time" json:"view_time"`
	ViewDuration int64                  `bson:"view_duration" json:"view_duration"` // milliseconds
	ScrollDepth  float64                `bson:"scroll_depth" json:"scroll_depth"`   // percentage
	Interactions []Interaction          `bson:"interactions" json:"interactions"`
	Source       string                 `bson:"source" json:"source"` // feed, profile, search
	Context      map[string]interface{} `bson:"context,omitempty" json:"context,omitempty"`
	CreatedAt    time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time              `bson:"updated_at" json:"updated_at"`
}

// Interaction represents specific interactions with content
type Interaction struct {
	Type      string    `bson:"type" json:"type"` // like, share, comment, save
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	Value     string    `bson:"value,omitempty" json:"value,omitempty"` // for reactions
}

// UserJourney tracks user journey through the application
type UserJourney struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	SessionID   string             `bson:"session_id" json:"session_id"`
	Touchpoints []Touchpoint       `bson:"touchpoints" json:"touchpoints"`
	Goal        string             `bson:"goal" json:"goal"` // registration, post_creation, etc.
	Completed   bool               `bson:"completed" json:"completed"`
	Duration    int64              `bson:"duration" json:"duration"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// Touchpoint represents a point in the user journey
type Touchpoint struct {
	Page      string                 `bson:"page" json:"page"`
	Action    string                 `bson:"action" json:"action"`
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// RecommendationEvent tracks recommendation performance
type RecommendationEvent struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID             primitive.ObjectID `bson:"user_id" json:"user_id"`
	RecommendationType string             `bson:"recommendation_type" json:"recommendation_type"` // content, user, group
	ItemID             primitive.ObjectID `bson:"item_id" json:"item_id"`
	Algorithm          string             `bson:"algorithm" json:"algorithm"`
	Score              float64            `bson:"score" json:"score"`
	Position           int                `bson:"position" json:"position"`
	Presented          time.Time          `bson:"presented" json:"presented"`
	Clicked            *time.Time         `bson:"clicked,omitempty" json:"clicked,omitempty"`
	Converted          *time.Time         `bson:"converted,omitempty" json:"converted,omitempty"`
	Feedback           string             `bson:"feedback,omitempty" json:"feedback,omitempty"` // liked, hidden, reported
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

// UserBehaviorProfile represents a user's behavior profile
type UserBehaviorProfile struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID              primitive.ObjectID `bson:"user_id" json:"user_id"`
	ContentPreferences  map[string]float64 `bson:"content_preferences" json:"content_preferences"`
	InteractionPatterns map[string]int     `bson:"interaction_patterns" json:"interaction_patterns"`
	PreferredSources    []string           `bson:"preferred_sources" json:"preferred_sources"`
	ActiveHours         []int              `bson:"active_hours" json:"active_hours"` // hours 0-23
	EngagementScore     float64            `bson:"engagement_score" json:"engagement_score"`
	LastUpdated         time.Time          `bson:"last_updated" json:"last_updated"`
	CreatedAt           time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time          `bson:"updated_at" json:"updated_at"`
}

// BehaviorInsight represents insights derived from user behavior
type BehaviorInsight struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      primitive.ObjectID     `bson:"user_id" json:"user_id"`
	Type        string                 `bson:"type" json:"type"` // preference, trend, anomaly
	Title       string                 `bson:"title" json:"title"`
	Description string                 `bson:"description" json:"description"`
	Score       float64                `bson:"score" json:"score"`
	Data        map[string]interface{} `bson:"data" json:"data"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
}

// Response structures for API
type UserBehaviorAnalytics struct {
	UserID              string              `json:"user_id"`
	TimeRange           string              `json:"time_range"`
	Sessions            SessionStats        `json:"sessions"`
	Engagement          EngagementStats     `json:"engagement"`
	ContentPreferences  map[string]float64  `json:"content_preferences"`
	InteractionPatterns map[string]int      `json:"interaction_patterns"`
	RecommendationStats RecommendationStats `json:"recommendation_stats"`
	TopSources          []SourceStat        `json:"top_sources"`
	ActivityPatterns    []ActivityPattern   `json:"activity_patterns"`
}

type SessionStats struct {
	TotalSessions  int     `json:"total_sessions"`
	AvgDuration    float64 `json:"avg_duration_minutes"`
	TotalPageViews int     `json:"total_page_views"`
	TotalActions   int     `json:"total_actions"`
	UniquePages    int     `json:"unique_pages"`
}

type EngagementStats struct {
	TotalViews        int                `json:"total_views"`
	AvgViewDuration   float64            `json:"avg_view_duration_seconds"`
	TotalInteractions int                `json:"total_interactions"`
	EngagementRate    float64            `json:"engagement_rate"`
	ByContentType     []ContentTypeStats `json:"by_content_type"`
}

type ContentTypeStats struct {
	ContentType    string  `json:"content_type"`
	Views          int     `json:"views"`
	AvgDuration    float64 `json:"avg_duration"`
	Interactions   int     `json:"interactions"`
	EngagementRate float64 `json:"engagement_rate"`
}

type RecommendationStats struct {
	TotalRecommendations int     `json:"total_recommendations"`
	ClickThroughRate     float64 `json:"click_through_rate"`
	ConversionRate       float64 `json:"conversion_rate"`
	AvgScore             float64 `json:"avg_score"`
}

type SourceStat struct {
	Source          string  `json:"source"`
	EngagementCount int     `json:"engagement_count"`
	AvgDuration     float64 `json:"avg_duration"`
	Percentage      float64 `json:"percentage"`
}

type ActivityPattern struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// Helper methods for UserSession
func (us *UserSession) BeforeCreate() {
	now := time.Now()
	us.CreatedAt = now
	us.UpdatedAt = now
	if us.ID.IsZero() {
		us.ID = primitive.NewObjectID()
	}
}

// Helper methods for ContentEngagement
func (ce *ContentEngagement) BeforeCreate() {
	now := time.Now()
	ce.CreatedAt = now
	ce.UpdatedAt = now
	if ce.ID.IsZero() {
		ce.ID = primitive.NewObjectID()
	}
}

// Helper methods for UserJourney
func (uj *UserJourney) BeforeCreate() {
	now := time.Now()
	uj.CreatedAt = now
	uj.UpdatedAt = now
	if uj.ID.IsZero() {
		uj.ID = primitive.NewObjectID()
	}
}

// Helper methods for RecommendationEvent
func (re *RecommendationEvent) BeforeCreate() {
	now := time.Now()
	re.CreatedAt = now
	re.UpdatedAt = now
	if re.ID.IsZero() {
		re.ID = primitive.NewObjectID()
	}
}

// Helper methods for UserBehaviorProfile
func (ubp *UserBehaviorProfile) BeforeCreate() {
	now := time.Now()
	ubp.CreatedAt = now
	ubp.UpdatedAt = now
	ubp.LastUpdated = now
	if ubp.ID.IsZero() {
		ubp.ID = primitive.NewObjectID()
	}
}

func (ubp *UserBehaviorProfile) BeforeUpdate() {
	ubp.UpdatedAt = time.Now()
	ubp.LastUpdated = time.Now()
}
