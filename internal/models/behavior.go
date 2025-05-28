// internal/models/behavior_models.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// USER SESSION MODELS

type UserSession struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	SessionID    string             `bson:"session_id" json:"session_id"`
	StartTime    time.Time          `bson:"start_time" json:"start_time"`
	EndTime      *time.Time         `bson:"end_time,omitempty" json:"end_time,omitempty"`
	Duration     int64              `bson:"duration" json:"duration"` // seconds
	DeviceInfo   string             `bson:"device_info" json:"device_info"`
	IPAddress    string             `bson:"ip_address" json:"ip_address"`
	UserAgent    string             `bson:"user_agent" json:"user_agent"`
	PagesVisited []PageVisit        `bson:"pages_visited" json:"pages_visited"`
	Actions      []UserAction       `bson:"actions" json:"actions"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

type PageVisit struct {
	URL       string    `bson:"url" json:"url"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	Duration  int64     `bson:"duration" json:"duration"` // milliseconds
	Referrer  string    `bson:"referrer,omitempty" json:"referrer,omitempty"`
}

type UserAction struct {
	Type      string                 `bson:"type" json:"type"`     // click, scroll, hover, etc.
	Target    string                 `bson:"target" json:"target"` // element clicked
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// CONTENT ENGAGEMENT MODELS

type ContentEngagement struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
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

type Interaction struct {
	Type      string    `bson:"type" json:"type"` // like, share, comment, save
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	Value     string    `bson:"value,omitempty" json:"value,omitempty"` // for reactions
}

// USER JOURNEY MODELS

type UserJourney struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	SessionID   string             `bson:"session_id" json:"session_id"`
	Touchpoints []Touchpoint       `bson:"touchpoints" json:"touchpoints"`
	Goal        string             `bson:"goal,omitempty" json:"goal,omitempty"` // registration, post_creation, etc.
	Completed   bool               `bson:"completed" json:"completed"`
	Duration    int64              `bson:"duration" json:"duration"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	CompletedAt *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

type Touchpoint struct {
	Page      string                 `bson:"page" json:"page"`
	Action    string                 `bson:"action" json:"action"`
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// RECOMMENDATION MODELS

type RecommendationEvent struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
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
}

// EXPERIMENT MODELS

type ExperimentEvent struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	ExperimentID string             `bson:"experiment_id" json:"experiment_id"`
	VariantID    string             `bson:"variant_id" json:"variant_id"`
	Event        string             `bson:"event" json:"event"` // exposure, conversion
	Value        float64            `bson:"value,omitempty" json:"value,omitempty"`
	Timestamp    time.Time          `bson:"timestamp" json:"timestamp"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}

// ANALYTICS MODELS

type UserBehaviorAnalytics struct {
	UserID             primitive.ObjectID     `json:"user_id"`
	TimeRange          string                 `json:"time_range"`
	StartTime          time.Time              `json:"start_time"`
	EndTime            time.Time              `json:"end_time"`
	SessionStats       SessionStats           `json:"session_stats"`
	EngagementStats    EngagementStats        `json:"engagement_stats"`
	ContentPreferences map[string]float64     `json:"content_preferences"`
	ActivityPatterns   map[string]interface{} `json:"activity_patterns"`
	GeneratedAt        time.Time              `json:"generated_at"`
}

type SessionStats struct {
	TotalSessions int64   `json:"total_sessions"`
	TotalDuration int64   `json:"total_duration"`
	AvgDuration   float64 `json:"avg_duration"`
	TotalPages    int64   `json:"total_pages"`
	TotalActions  int64   `json:"total_actions"`
}

type EngagementStats struct {
	ByContentType map[string]ContentTypeStats `json:"by_content_type"`
}

type ContentTypeStats struct {
	TotalViews        int64   `json:"total_views"`
	TotalDuration     int64   `json:"total_duration"`
	AvgDuration       float64 `json:"avg_duration"`
	AvgScrollDepth    float64 `json:"avg_scroll_depth"`
	TotalInteractions int64   `json:"total_interactions"`
}

type RecommendationPerformance struct {
	Algorithm      string  `json:"algorithm"`
	TimeRange      string  `json:"time_range"`
	TotalPresented int64   `json:"total_presented"`
	TotalClicked   int64   `json:"total_clicked"`
	TotalConverted int64   `json:"total_converted"`
	CTR            float64 `json:"ctr"` // Click-through rate
	ConversionRate float64 `json:"conversion_rate"`
}

type ContentPopularity struct {
	ContentID          primitive.ObjectID `bson:"_id" json:"content_id"`
	TotalViews         int64              `bson:"total_views" json:"total_views"`
	UniqueViewersCount int64              `bson:"unique_viewers_count" json:"unique_viewers_count"`
	TotalDuration      int64              `bson:"total_duration" json:"total_duration"`
	AvgDuration        float64            `bson:"avg_duration" json:"avg_duration"`
	TotalInteractions  int64              `bson:"total_interactions" json:"total_interactions"`
}

// REQUEST/RESPONSE MODELS

type TrackPageViewRequest struct {
	URL       string `json:"url" binding:"required"`
	Referrer  string `json:"referrer"`
	Duration  int64  `json:"duration"` // time spent on previous page
	SessionID string `json:"session_id" binding:"required"`
}

type TrackUserActionRequest struct {
	Type      string                 `json:"type" binding:"required"`
	Target    string                 `json:"target" binding:"required"`
	SessionID string                 `json:"session_id" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type TrackContentEngagementRequest struct {
	ContentID        string  `json:"content_id" binding:"required"`
	ContentType      string  `json:"content_type" binding:"required"`
	ViewDuration     int64   `json:"view_duration"` // milliseconds
	ScrollDepth      float64 `json:"scroll_depth"`  // 0-100%
	Source           string  `json:"source"`
	InteractionType  string  `json:"interaction_type,omitempty"`
	InteractionValue string  `json:"interaction_value,omitempty"`
}

type TrackRecommendationRequest struct {
	RecommendationType string  `json:"recommendation_type" binding:"required"`
	ItemID             string  `json:"item_id" binding:"required"`
	Algorithm          string  `json:"algorithm" binding:"required"`
	Score              float64 `json:"score"`
	Position           int     `json:"position"`
	Action             string  `json:"action"` // presented, clicked, converted
}

type TrackExperimentRequest struct {
	ExperimentID string  `json:"experiment_id" binding:"required"`
	VariantID    string  `json:"variant_id" binding:"required"`
	Event        string  `json:"event" binding:"required"` // exposure, conversion
	Value        float64 `json:"value,omitempty"`
}

// AUTOMATIC TRACKING EVENTS

type AutoTrackEvent struct {
	UserID    primitive.ObjectID     `json:"user_id"`
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
}

// FEED INTERACTION MODELS (Enhanced from existing)

type FeedInteractionEvent struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	PostID          primitive.ObjectID `bson:"post_id" json:"post_id"`
	InteractionType string             `bson:"interaction_type" json:"interaction_type"`     // view, like, comment, share, save, hide
	Source          string             `bson:"source" json:"source"`                         // feed, profile, search, trending, discover
	TimeSpent       int64              `bson:"time_spent" json:"time_spent"`                 // milliseconds
	Position        int                `bson:"position,omitempty" json:"position,omitempty"` // position in feed
	SessionID       string             `bson:"session_id" json:"session_id"`
	Timestamp       time.Time          `bson:"timestamp" json:"timestamp"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}

// SEARCH BEHAVIOR MODELS (Enhanced from existing)

type SearchEvent struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Query         string             `bson:"query" json:"query"`
	SearchType    string             `bson:"search_type" json:"search_type"` // all, posts, users, hashtags
	ResultsCount  int                `bson:"results_count" json:"results_count"`
	ClickedResult *SearchClickEvent  `bson:"clicked_result,omitempty" json:"clicked_result,omitempty"`
	SessionID     string             `bson:"session_id" json:"session_id"`
	Timestamp     time.Time          `bson:"timestamp" json:"timestamp"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

type SearchClickEvent struct {
	ResultID   primitive.ObjectID `bson:"result_id" json:"result_id"`
	ResultType string             `bson:"result_type" json:"result_type"`
	Position   int                `bson:"position" json:"position"`
	ClickedAt  time.Time          `bson:"clicked_at" json:"clicked_at"`
}

// STORY BEHAVIOR MODELS (Enhanced from existing)

type StoryViewEvent struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	StoryID       primitive.ObjectID `bson:"story_id" json:"story_id"`
	StoryOwnerID  primitive.ObjectID `bson:"story_owner_id" json:"story_owner_id"`
	ViewDuration  int64              `bson:"view_duration" json:"view_duration"` // milliseconds
	CompletedView bool               `bson:"completed_view" json:"completed_view"`
	Source        string             `bson:"source" json:"source"` // feed, profile, search
	Position      int                `bson:"position,omitempty" json:"position,omitempty"`
	SessionID     string             `bson:"session_id" json:"session_id"`
	Timestamp     time.Time          `bson:"timestamp" json:"timestamp"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

// CONVERSION TRACKING MODELS

type ConversionEvent struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	ConversionType string             `bson:"conversion_type" json:"conversion_type"` // registration, first_post, first_follow, etc.
	Value          float64            `bson:"value,omitempty" json:"value,omitempty"` // monetary or engagement value
	Source         string             `bson:"source" json:"source"`                   // how they got to conversion
	SessionID      string             `bson:"session_id" json:"session_id"`
	Timestamp      time.Time          `bson:"timestamp" json:"timestamp"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

// REAL-TIME ANALYTICS MODELS

type RealTimeAnalytics struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ContentID      primitive.ObjectID `bson:"content_id" json:"content_id"`
	ContentType    string             `bson:"content_type" json:"content_type"`
	Date           string             `bson:"date" json:"date"` // YYYY-MM-DD
	Hour           int                `bson:"hour" json:"hour"` // 0-23
	TotalViews     int64              `bson:"total_views" json:"total_views"`
	UniqueViews    int64              `bson:"unique_views" json:"unique_views"`
	TotalDuration  int64              `bson:"total_duration" json:"total_duration"`
	TotalLikes     int64              `bson:"total_likes" json:"total_likes"`
	TotalComments  int64              `bson:"total_comments" json:"total_comments"`
	TotalShares    int64              `bson:"total_shares" json:"total_shares"`
	EngagementRate float64            `bson:"engagement_rate" json:"engagement_rate"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

// COHORT ANALYSIS MODELS

type UserCohort struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CohortDate     string             `bson:"cohort_date" json:"cohort_date"` // YYYY-MM-DD
	UsersCount     int64              `bson:"users_count" json:"users_count"`
	RetentionRates map[string]float64 `bson:"retention_rates" json:"retention_rates"` // day/week/month -> rate
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

// FUNNEL ANALYSIS MODELS

type FunnelStep struct {
	StepName       string  `bson:"step_name" json:"step_name"`
	UsersCount     int64   `bson:"users_count" json:"users_count"`
	ConversionRate float64 `bson:"conversion_rate" json:"conversion_rate"`
}

type FunnelAnalysis struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FunnelName  string             `bson:"funnel_name" json:"funnel_name"`
	TimeRange   string             `bson:"time_range" json:"time_range"`
	Steps       []FunnelStep       `bson:"steps" json:"steps"`
	TotalUsers  int64              `bson:"total_users" json:"total_users"`
	OverallRate float64            `bson:"overall_rate" json:"overall_rate"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}
