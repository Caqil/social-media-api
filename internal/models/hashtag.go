// models/hashtag.go
package models

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Hashtag represents a hashtag in the system
type Hashtag struct {
	BaseModel `bson:",inline"`

	// Basic Information
	Tag           string `json:"tag" bson:"tag" validate:"required,min=1,max=100"` // Without the # symbol
	NormalizedTag string `json:"normalized_tag" bson:"normalized_tag"`             // Lowercase, no special chars
	DisplayTag    string `json:"display_tag" bson:"display_tag"`                   // Original casing for display

	// Usage Statistics
	PostsCount   int64 `json:"posts_count" bson:"posts_count"`
	StoriesCount int64 `json:"stories_count" bson:"stories_count"`
	TotalUsage   int64 `json:"total_usage" bson:"total_usage"`

	// Trending Information
	IsTrending       bool       `json:"is_trending" bson:"is_trending"`
	TrendingScore    float64    `json:"trending_score" bson:"trending_score"`
	TrendingRank     int        `json:"trending_rank,omitempty" bson:"trending_rank,omitempty"`
	LastTrendingAt   *time.Time `json:"last_trending_at,omitempty" bson:"last_trending_at,omitempty"`
	TrendingDuration int64      `json:"trending_duration" bson:"trending_duration"` // in hours

	// Usage Analytics
	DailyUsage     []HashtagUsageData `json:"daily_usage,omitempty" bson:"daily_usage,omitempty"`
	WeeklyUsage    []HashtagUsageData `json:"weekly_usage,omitempty" bson:"weekly_usage,omitempty"`
	MonthlyUsage   []HashtagUsageData `json:"monthly_usage,omitempty" bson:"monthly_usage,omitempty"`
	PeakUsageDate  *time.Time         `json:"peak_usage_date,omitempty" bson:"peak_usage_date,omitempty"`
	PeakUsageCount int64              `json:"peak_usage_count" bson:"peak_usage_count"`

	// Geographic and Demographic Data
	TopCountries       []CountryUsage   `json:"top_countries,omitempty" bson:"top_countries,omitempty"`
	AgeDistribution    map[string]int64 `json:"age_distribution,omitempty" bson:"age_distribution,omitempty"`
	GenderDistribution map[string]int64 `json:"gender_distribution,omitempty" bson:"gender_distribution,omitempty"`

	// Related Hashtags
	RelatedTags        []string              `json:"related_tags,omitempty" bson:"related_tags,omitempty"`
	FrequentlyUsedWith []HashtagCooccurrence `json:"frequently_used_with,omitempty" bson:"frequently_used_with,omitempty"`

	// Content Classification
	Category    string   `json:"category,omitempty" bson:"category,omitempty"` // entertainment, sports, news, etc.
	Language    string   `json:"language,omitempty" bson:"language,omitempty"`
	ContentType []string `json:"content_type,omitempty" bson:"content_type,omitempty"` // text, image, video
	Sentiment   string   `json:"sentiment,omitempty" bson:"sentiment,omitempty"`       // positive, negative, neutral

	// Moderation
	IsBlocked     bool                `json:"is_blocked" bson:"is_blocked"`
	IsSensitive   bool                `json:"is_sensitive" bson:"is_sensitive"`
	BlockedReason string              `json:"blocked_reason,omitempty" bson:"blocked_reason,omitempty"`
	BlockedAt     *time.Time          `json:"blocked_at,omitempty" bson:"blocked_at,omitempty"`
	BlockedBy     *primitive.ObjectID `json:"blocked_by,omitempty" bson:"blocked_by,omitempty"`
	IsReported    bool                `json:"is_reported" bson:"is_reported"`
	ReportsCount  int64               `json:"reports_count" bson:"reports_count"`

	// First Usage
	FirstUsedBy   primitive.ObjectID `json:"first_used_by" bson:"first_used_by"`
	FirstUsedAt   time.Time          `json:"first_used_at" bson:"first_used_at"`
	FirstUsedIn   string             `json:"first_used_in" bson:"first_used_in"` // post, story, comment
	FirstUsedInID primitive.ObjectID `json:"first_used_in_id" bson:"first_used_in_id"`

	// Search and Discovery
	SearchCount     int64   `json:"search_count" bson:"search_count"`
	ClickCount      int64   `json:"click_count" bson:"click_count"`
	SearchClickRate float64 `json:"search_click_rate" bson:"search_click_rate"`

	// External Data
	ExternalReferences []ExternalReference `json:"external_references,omitempty" bson:"external_references,omitempty"`
	WikipediaURL       string              `json:"wikipedia_url,omitempty" bson:"wikipedia_url,omitempty"`
	Description        string              `json:"description,omitempty" bson:"description,omitempty"`

	// Admin Features
	IsFeatured bool   `json:"is_featured" bson:"is_featured"`
	IsPromoted bool   `json:"is_promoted" bson:"is_promoted"`
	AdminNotes string `json:"admin_notes,omitempty" bson:"-"` // Not exposed to users
}

// HashtagUsageData represents usage data for a specific time period
type HashtagUsageData struct {
	Date        time.Time `json:"date" bson:"date"`
	Count       int64     `json:"count" bson:"count"`
	UniqueUsers int64     `json:"unique_users" bson:"unique_users"`
}

// CountryUsage represents hashtag usage by country
type CountryUsage struct {
	CountryCode string `json:"country_code" bson:"country_code"`
	CountryName string `json:"country_name" bson:"country_name"`
	Count       int64  `json:"count" bson:"count"`
}

// HashtagCooccurrence represents hashtags frequently used together
type HashtagCooccurrence struct {
	Tag   string  `json:"tag" bson:"tag"`
	Score float64 `json:"score" bson:"score"` // Correlation score
	Count int64   `json:"count" bson:"count"` // Number of times used together
}

// ExternalReference represents external references to the hashtag
type ExternalReference struct {
	Platform string `json:"platform" bson:"platform"` // twitter, instagram, tiktok
	URL      string `json:"url" bson:"url"`
	Count    int64  `json:"count" bson:"count"`
}

// HashtagResponse represents hashtag data returned in API responses
type HashtagResponse struct {
	ID            string         `json:"id"`
	Tag           string         `json:"tag"`
	DisplayTag    string         `json:"display_tag"`
	PostsCount    int64          `json:"posts_count"`
	StoriesCount  int64          `json:"stories_count"`
	TotalUsage    int64          `json:"total_usage"`
	IsTrending    bool           `json:"is_trending"`
	TrendingScore float64        `json:"trending_score"`
	TrendingRank  int            `json:"trending_rank,omitempty"`
	Category      string         `json:"category,omitempty"`
	Language      string         `json:"language,omitempty"`
	Sentiment     string         `json:"sentiment,omitempty"`
	RelatedTags   []string       `json:"related_tags,omitempty"`
	TopCountries  []CountryUsage `json:"top_countries,omitempty"`
	Description   string         `json:"description,omitempty"`
	IsFeatured    bool           `json:"is_featured"`
	IsPromoted    bool           `json:"is_promoted"`
	FirstUsedAt   time.Time      `json:"first_used_at"`
	CreatedAt     time.Time      `json:"created_at"`

	// Analytics (only for admins/analytics)
	DailyUsage      []HashtagUsageData `json:"daily_usage,omitempty"`
	WeeklyUsage     []HashtagUsageData `json:"weekly_usage,omitempty"`
	SearchCount     int64              `json:"search_count,omitempty"`
	ClickCount      int64              `json:"click_count,omitempty"`
	SearchClickRate float64            `json:"search_click_rate,omitempty"`
}

// HashtagTrendingResponse represents trending hashtag data
type HashtagTrendingResponse struct {
	HashtagResponse `json:",inline"`
	Rank            int     `json:"rank"`
	GrowthRate      float64 `json:"growth_rate"`
	PreviousRank    int     `json:"previous_rank,omitempty"`
	RankChange      string  `json:"rank_change"` // up, down, new, same
}

// CreateHashtagRequest represents request to create/track a hashtag
type CreateHashtagRequest struct {
	Tag         string `json:"tag" validate:"required,min=1,max=100"`
	Category    string `json:"category,omitempty"`
	Language    string `json:"language,omitempty"`
	Description string `json:"description,omitempty"`
}


// Methods for Hashtag model

// BeforeCreate sets default values before creating hashtag
func (h *Hashtag) BeforeCreate() {
	h.BaseModel.BeforeCreate()

	// Normalize tag
	h.NormalizeTag()

	// Set default values
	h.PostsCount = 0
	h.StoriesCount = 0
	h.TotalUsage = 0
	h.IsTrending = false
	h.TrendingScore = 0.0
	h.TrendingDuration = 0
	h.PeakUsageCount = 0
	h.IsBlocked = false
	h.IsSensitive = false
	h.IsReported = false
	h.ReportsCount = 0
	h.IsFeatured = false
	h.IsPromoted = false
	h.SearchCount = 0
	h.ClickCount = 0
	h.SearchClickRate = 0.0

	// Set first usage info
	h.FirstUsedAt = h.CreatedAt

	// Initialize maps
	if h.AgeDistribution == nil {
		h.AgeDistribution = make(map[string]int64)
	}
	if h.GenderDistribution == nil {
		h.GenderDistribution = make(map[string]int64)
	}
}

// NormalizeTag normalizes the hashtag for consistent storage and searching
func (h *Hashtag) NormalizeTag() {
	// Remove # if present
	tag := strings.TrimPrefix(h.Tag, "#")

	// Store display version
	if h.DisplayTag == "" {
		h.DisplayTag = tag
	}

	// Create normalized version (lowercase, remove special chars except numbers and letters)
	normalized := strings.ToLower(tag)
	// Additional normalization logic would go here
	h.NormalizedTag = normalized

	// Store clean tag without #
	h.Tag = tag
}

// ToHashtagResponse converts Hashtag to HashtagResponse
func (h *Hashtag) ToHashtagResponse() HashtagResponse {
	return HashtagResponse{
		ID:            h.ID.Hex(),
		Tag:           h.Tag,
		DisplayTag:    h.DisplayTag,
		PostsCount:    h.PostsCount,
		StoriesCount:  h.StoriesCount,
		TotalUsage:    h.TotalUsage,
		IsTrending:    h.IsTrending,
		TrendingScore: h.TrendingScore,
		TrendingRank:  h.TrendingRank,
		Category:      h.Category,
		Language:      h.Language,
		Sentiment:     h.Sentiment,
		RelatedTags:   h.RelatedTags,
		TopCountries:  h.TopCountries,
		Description:   h.Description,
		IsFeatured:    h.IsFeatured,
		IsPromoted:    h.IsPromoted,
		FirstUsedAt:   h.FirstUsedAt,
		CreatedAt:     h.CreatedAt,
	}
}

// IncrementPostsCount increments the posts count
func (h *Hashtag) IncrementPostsCount() {
	h.PostsCount++
	h.TotalUsage++
	h.UpdateTrendingScore()
	h.BeforeUpdate()
}

// IncrementStoriesCount increments the stories count
func (h *Hashtag) IncrementStoriesCount() {
	h.StoriesCount++
	h.TotalUsage++
	h.UpdateTrendingScore()
	h.BeforeUpdate()
}

// UpdateTrendingScore calculates and updates the trending score
func (h *Hashtag) UpdateTrendingScore() {
	// Simple trending score calculation
	// In practice, this would be more sophisticated
	now := time.Now()
	hoursSinceCreation := now.Sub(h.CreatedAt).Hours()

	if hoursSinceCreation > 0 {
		h.TrendingScore = float64(h.TotalUsage) / hoursSinceCreation
	} else {
		h.TrendingScore = float64(h.TotalUsage)
	}
}

// IncrementSearchCount increments the search count
func (h *Hashtag) IncrementSearchCount() {
	h.SearchCount++
	h.UpdateSearchClickRate()
	h.BeforeUpdate()
}

// IncrementClickCount increments the click count
func (h *Hashtag) IncrementClickCount() {
	h.ClickCount++
	h.UpdateSearchClickRate()
	h.BeforeUpdate()
}

// UpdateSearchClickRate calculates the click-through rate
func (h *Hashtag) UpdateSearchClickRate() {
	if h.SearchCount > 0 {
		h.SearchClickRate = (float64(h.ClickCount) / float64(h.SearchCount)) * 100
	}
}

// Block blocks the hashtag
func (h *Hashtag) Block(reason string, blockedBy primitive.ObjectID) {
	h.IsBlocked = true
	h.BlockedReason = reason
	h.BlockedBy = &blockedBy
	now := time.Now()
	h.BlockedAt = &now
	h.BeforeUpdate()
}

// Unblock unblocks the hashtag
func (h *Hashtag) Unblock() {
	h.IsBlocked = false
	h.BlockedReason = ""
	h.BlockedBy = nil
	h.BlockedAt = nil
	h.BeforeUpdate()
}


// IncrementViewCount increments the view count
func (m *Mention) IncrementViewCount() {
	m.ViewCount++
	m.UpdateInteractionScore()
	m.BeforeUpdate()
}

// IncrementClickCount increments the click count
func (m *Mention) IncrementClickCount() {
	m.ClickCount++
	m.UpdateInteractionScore()
	m.BeforeUpdate()
}

// UpdateInteractionScore calculates interaction score
func (m *Mention) UpdateInteractionScore() {
	if m.ViewCount > 0 {
		m.InteractionScore = (float64(m.ClickCount) / float64(m.ViewCount)) * 100
	}
}

// Block blocks this mention from being visible to the mentioned user
func (m *Mention) Block() {
	m.IsBlocked = true
	m.IsVisible = false
	m.BeforeUpdate()
}

// Unblock unblocks this mention
func (m *Mention) Unblock() {
	m.IsBlocked = false
	m.IsVisible = true
	m.BeforeUpdate()
}


// ExtractHashtagsFromText extracts hashtags from text content
func ExtractHashtagsFromText(text string) []string {
	// Simple hashtag extraction - in practice would use regex
	var hashtags []string
	words := strings.Fields(text)

	for _, word := range words {
		if strings.HasPrefix(word, "#") && len(word) > 1 {
			// Clean the hashtag
			tag := strings.TrimPrefix(word, "#")
			// Remove trailing punctuation
			tag = strings.TrimRight(tag, ".,!?;:")
			if len(tag) > 0 {
				hashtags = append(hashtags, tag)
			}
		}
	}

	return hashtags
}


// IsValidHashtag checks if a hashtag string is valid
func IsValidHashtag(tag string) bool {
	// Remove # if present
	tag = strings.TrimPrefix(tag, "#")

	// Check length
	if len(tag) < 1 || len(tag) > 100 {
		return false
	}

	// Check for invalid characters (simplified)
	// In practice, would have more sophisticated validation
	return len(strings.TrimSpace(tag)) > 0
}


// GetHashtagCategories returns available hashtag categories
func GetHashtagCategories() []string {
	return []string{
		"general",
		"entertainment",
		"sports",
		"news",
		"technology",
		"business",
		"lifestyle",
		"travel",
		"food",
		"fashion",
		"art",
		"music",
		"health",
		"education",
		"politics",
		"science",
		"nature",
		"photography",
		"fitness",
		"gaming",
	}
}

// IsValidHashtagCategory checks if a category is valid
func IsValidHashtagCategory(category string) bool {
	validCategories := GetHashtagCategories()
	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}
