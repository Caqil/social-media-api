// internal/middleware/behavior_tracking.go
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"social-media-api/internal/models"
	"social-media-api/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BehaviorTrackingMiddleware struct {
	behaviorService *services.UserBehaviorService
}

func NewBehaviorTrackingMiddleware(behaviorService *services.UserBehaviorService) *BehaviorTrackingMiddleware {
	return &BehaviorTrackingMiddleware{
		behaviorService: behaviorService,
	}
}

// AutoTrackBehavior automatically tracks user behavior for all requests
func (m *BehaviorTrackingMiddleware) AutoTrackBehavior() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		startTime := time.Now()

		// Generate or get session ID
		sessionID := m.getOrCreateSessionID(c)
		c.Set("session_id", sessionID)

		// Get user info if authenticated
		userID, exists := c.Get("user_id")
		if exists {
			// Start session tracking if new session
			if m.isNewSession(c, sessionID) {
				go m.startSessionTracking(userID.(primitive.ObjectID), sessionID, c)
			}

			// Track page visit
			go m.trackPageVisit(userID.(primitive.ObjectID), sessionID, c)
		}

		// Continue to next handler
		c.Next()

		// Track response and duration
		if exists {
			duration := time.Since(startTime)
			go m.trackRequestCompletion(userID.(primitive.ObjectID), sessionID, c, duration)
		}
	})
}

// TrackContentInteraction specifically tracks content-related interactions
func (m *BehaviorTrackingMiddleware) TrackContentInteraction() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next() // Execute the main handler first

		// Track after successful request
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			userID, exists := c.Get("user_id")
			if exists {
				go m.trackContentInteraction(userID.(primitive.ObjectID), c)
			}
		}
	})
}

// TrackAPIUsage tracks API endpoint usage patterns
func (m *BehaviorTrackingMiddleware) TrackAPIUsage() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		// Track API usage
		userID, exists := c.Get("user_id")
		if exists {
			duration := time.Since(startTime)
			go m.trackAPIUsage(userID.(primitive.ObjectID), c, duration)
		}
	})
}

// PRIVATE METHODS

func (m *BehaviorTrackingMiddleware) getOrCreateSessionID(c *gin.Context) string {
	// Try to get from header first
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID != "" {
		return sessionID
	}

	// Try to get from cookie
	if cookie, err := c.Request.Cookie("session_id"); err == nil {
		return cookie.Value
	}

	// Generate new session ID
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (m *BehaviorTrackingMiddleware) isNewSession(c *gin.Context, sessionID string) bool {
	// Check if session exists in header or is marked as new
	return c.GetHeader("X-New-Session") == "true" || c.GetHeader("X-Session-ID") == ""
}

func (m *BehaviorTrackingMiddleware) startSessionTracking(userID primitive.ObjectID, sessionID string, c *gin.Context) {
	deviceInfo := c.GetHeader("User-Agent")
	if deviceInfo == "" {
		deviceInfo = "Unknown"
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	m.behaviorService.StartSession(userID, sessionID, deviceInfo, ipAddress, userAgent)
}

func (m *BehaviorTrackingMiddleware) trackPageVisit(userID primitive.ObjectID, sessionID string, c *gin.Context) {
	pageVisit := models.PageVisit{
		URL:       c.Request.URL.Path,
		Timestamp: time.Now(),
		Referrer:  c.GetHeader("Referer"),
	}

	m.behaviorService.RecordPageVisit(userID, sessionID, pageVisit)
}

func (m *BehaviorTrackingMiddleware) trackRequestCompletion(userID primitive.ObjectID, sessionID string, c *gin.Context, duration time.Duration) {
	action := models.UserAction{
		Type:      "api_request",
		Target:    fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status_code":   c.Writer.Status(),
			"duration_ms":   duration.Milliseconds(),
			"response_size": c.Writer.Size(),
			"user_agent":    c.GetHeader("User-Agent"),
		},
	}

	m.behaviorService.RecordUserAction(userID, sessionID, action)
}

func (m *BehaviorTrackingMiddleware) trackContentInteraction(userID primitive.ObjectID, c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method

	// Parse content interactions based on URL patterns
	if strings.Contains(path, "/posts/") {
		m.trackPostInteraction(userID, path, method, c)
	} else if strings.Contains(path, "/stories/") {
		m.trackStoryInteraction(userID, path, method, c)
	} else if strings.Contains(path, "/comments/") {
		m.trackCommentInteraction(userID, path, method, c)
	} else if strings.Contains(path, "/search") {
		m.trackSearchInteraction(userID, c)
	} else if strings.Contains(path, "/feed") {
		m.trackFeedInteraction(userID, c)
	}
}

func (m *BehaviorTrackingMiddleware) trackPostInteraction(userID primitive.ObjectID, path, method string, c *gin.Context) {
	// Extract post ID from path
	parts := strings.Split(path, "/")
	var postID primitive.ObjectID
	var interactionType string

	for i, part := range parts {
		if part == "posts" && i+1 < len(parts) {
			if id, err := primitive.ObjectIDFromHex(parts[i+1]); err == nil {
				postID = id
				break
			}
		}
	}

	if postID.IsZero() {
		return
	}

	// Determine interaction type
	switch {
	case method == "GET":
		interactionType = "view"
	case method == "POST" && strings.Contains(path, "/like"):
		interactionType = "like"
	case method == "DELETE" && strings.Contains(path, "/like"):
		interactionType = "unlike"
	case method == "POST" && strings.Contains(path, "/report"):
		interactionType = "report"
	case method == "POST" && strings.Contains(path, "/pin"):
		interactionType = "pin"
	case method == "PUT":
		interactionType = "edit"
	case method == "DELETE":
		interactionType = "delete"
	default:
		interactionType = "unknown"
	}

	source := m.getSourceFromReferer(c)
	m.behaviorService.AutoTrackPostInteraction(userID, postID, interactionType, source)
}

func (m *BehaviorTrackingMiddleware) trackStoryInteraction(userID primitive.ObjectID, path, method string, c *gin.Context) {
	// Extract story ID from path
	parts := strings.Split(path, "/")
	var storyID primitive.ObjectID
	var interactionType string

	for i, part := range parts {
		if part == "stories" && i+1 < len(parts) {
			if id, err := primitive.ObjectIDFromHex(parts[i+1]); err == nil {
				storyID = id
				break
			}
		}
	}

	if storyID.IsZero() {
		return
	}

	// Determine interaction type
	switch {
	case method == "GET":
		interactionType = "view"
	case method == "POST" && strings.Contains(path, "/view"):
		interactionType = "view"
	case method == "POST" && strings.Contains(path, "/react"):
		interactionType = "react"
	case method == "DELETE" && strings.Contains(path, "/react"):
		interactionType = "unreact"
	default:
		interactionType = "unknown"
	}

	source := m.getSourceFromReferer(c)

	// For story views, estimate duration (placeholder)
	duration := int64(5000) // 5 seconds default
	if interactionType == "view" {
		m.behaviorService.AutoTrackStoryView(userID, storyID, source, duration)
	}
}

func (m *BehaviorTrackingMiddleware) trackCommentInteraction(userID primitive.ObjectID, path, method string, c *gin.Context) {
	// Extract comment/post ID from path
	parts := strings.Split(path, "/")
	var contentID primitive.ObjectID
	var interactionType string

	for i, part := range parts {
		if (part == "comments" || part == "posts") && i+1 < len(parts) {
			if id, err := primitive.ObjectIDFromHex(parts[i+1]); err == nil {
				contentID = id
				break
			}
		}
	}

	if contentID.IsZero() {
		return
	}

	// Determine interaction type
	switch {
	case method == "POST" && !strings.Contains(path, "/like") && !strings.Contains(path, "/report"):
		interactionType = "create_comment"
	case method == "POST" && strings.Contains(path, "/like"):
		interactionType = "like_comment"
	case method == "DELETE" && strings.Contains(path, "/like"):
		interactionType = "unlike_comment"
	case method == "PUT":
		interactionType = "edit_comment"
	case method == "DELETE":
		interactionType = "delete_comment"
	default:
		interactionType = "view_comment"
	}

	source := m.getSourceFromReferer(c)
	metadata := map[string]interface{}{
		"auto_tracked":     true,
		"timestamp":        time.Now(),
		"interaction_type": interactionType,
	}

	m.behaviorService.RecordInteraction(userID, contentID, "comment", interactionType, source, metadata)
}

func (m *BehaviorTrackingMiddleware) trackSearchInteraction(userID primitive.ObjectID, c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		return
	}

	searchType := c.DefaultQuery("type", "all")

	// Estimate results count (would need actual results)
	resultsCount := 0 // This would be set from actual search results

	m.behaviorService.AutoTrackSearch(userID, query, searchType, resultsCount)
}

func (m *BehaviorTrackingMiddleware) trackFeedInteraction(userID primitive.ObjectID, c *gin.Context) {
	feedType := c.DefaultQuery("type", "personalized")

	action := models.UserAction{
		Type:      "feed_view",
		Target:    feedType,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"feed_type":    feedType,
			"auto_tracked": true,
		},
	}

	sessionID, _ := c.Get("session_id")
	if sessionID != nil {
		m.behaviorService.RecordUserAction(userID, sessionID.(string), action)
	}
}

func (m *BehaviorTrackingMiddleware) trackAPIUsage(userID primitive.ObjectID, c *gin.Context, duration time.Duration) {
	endpoint := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)

	action := models.UserAction{
		Type:      "api_usage",
		Target:    endpoint,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"status_code":  c.Writer.Status(),
			"duration_ms":  duration.Milliseconds(),
			"success":      c.Writer.Status() >= 200 && c.Writer.Status() < 300,
			"auto_tracked": true,
		},
	}

	sessionID, _ := c.Get("session_id")
	if sessionID != nil {
		m.behaviorService.RecordUserAction(userID, sessionID.(string), action)
	}
}

func (m *BehaviorTrackingMiddleware) getSourceFromReferer(c *gin.Context) string {
	referer := c.GetHeader("Referer")
	if referer == "" {
		return "direct"
	}

	// Parse source from referer
	if strings.Contains(referer, "/feed") {
		return "feed"
	} else if strings.Contains(referer, "/profile") {
		return "profile"
	} else if strings.Contains(referer, "/search") {
		return "search"
	} else if strings.Contains(referer, "/trending") {
		return "trending"
	} else if strings.Contains(referer, "/discover") {
		return "discover"
	}

	return "unknown"
}

// SPECIALIZED TRACKING MIDDLEWARE

// TrackRecommendations tracks recommendation performance
func (m *BehaviorTrackingMiddleware) TrackRecommendations() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Track recommendation events
		userID, exists := c.Get("user_id")
		if exists && c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// Check if this is a recommended content view
			if recommendationData := c.GetHeader("X-Recommendation-Data"); recommendationData != "" {
				go m.trackRecommendationEvent(userID.(primitive.ObjectID), recommendationData, c)
			}
		}
	})
}

func (m *BehaviorTrackingMiddleware) trackRecommendationEvent(userID primitive.ObjectID, recommendationData string, c *gin.Context) {
	// Parse recommendation data (JSON string with algorithm, score, etc.)
	// This would be populated by recommendation engine
	// For now, create a basic tracking event

	action := models.UserAction{
		Type:      "recommendation_interaction",
		Target:    c.Request.URL.Path,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"recommendation_data": recommendationData,
			"auto_tracked":        true,
		},
	}

	sessionID, _ := c.Get("session_id")
	if sessionID != nil {
		m.behaviorService.RecordUserAction(userID, sessionID.(string), action)
	}
}

// TrackConversions tracks conversion events
func (m *BehaviorTrackingMiddleware) TrackConversions() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			userID, exists := c.Get("user_id")
			if exists {
				go m.trackConversionEvents(userID.(primitive.ObjectID), c)
			}
		}
	})
}

func (m *BehaviorTrackingMiddleware) trackConversionEvents(userID primitive.ObjectID, c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method

	var conversionType string

	// Define conversion events
	switch {
	case method == "POST" && strings.Contains(path, "/auth/register"):
		conversionType = "registration"
	case method == "POST" && strings.Contains(path, "/posts") && !strings.Contains(path, "/like"):
		conversionType = "first_post"
	case method == "POST" && strings.Contains(path, "/follow"):
		conversionType = "first_follow"
	case method == "POST" && strings.Contains(path, "/comments"):
		conversionType = "first_comment"
	case method == "POST" && strings.Contains(path, "/stories"):
		conversionType = "first_story"
	case method == "POST" && strings.Contains(path, "/groups") && !strings.Contains(path, "/join"):
		conversionType = "group_creation"
	case method == "POST" && strings.Contains(path, "/groups") && strings.Contains(path, "/join"):
		conversionType = "group_join"
	}

	if conversionType != "" {
		action := models.UserAction{
			Type:      "conversion",
			Target:    conversionType,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"conversion_type": conversionType,
				"path":            path,
				"auto_tracked":    true,
			},
		}

		sessionID, _ := c.Get("session_id")
		if sessionID != nil {
			m.behaviorService.RecordUserAction(userID, sessionID.(string), action)
		}
	}
}

// TrackErrors tracks error patterns
func (m *BehaviorTrackingMiddleware) TrackErrors() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Track errors
		if c.Writer.Status() >= 400 {
			userID, exists := c.Get("user_id")
			if exists {
				go m.trackErrorEvent(userID.(primitive.ObjectID), c)
			}
		}
	})
}

func (m *BehaviorTrackingMiddleware) trackErrorEvent(userID primitive.ObjectID, c *gin.Context) {
	action := models.UserAction{
		Type:      "error",
		Target:    fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"status_code":  c.Writer.Status(),
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"user_agent":   c.GetHeader("User-Agent"),
			"auto_tracked": true,
		},
	}

	sessionID, _ := c.Get("session_id")
	if sessionID != nil {
		m.behaviorService.RecordUserAction(userID, sessionID.(string), action)
	}
}

// SessionCleanup handles session cleanup when user logs out or session expires
func (m *BehaviorTrackingMiddleware) SessionCleanup() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check if this is a logout endpoint
		if strings.Contains(c.Request.URL.Path, "/logout") {
			sessionID, exists := c.Get("session_id")
			if exists {
				// End session tracking
				go m.behaviorService.EndSession(sessionID.(string))
			}
		}

		c.Next()
	})
}

// BehaviorBasedCaching provides behavior-based cache strategies
func (m *BehaviorTrackingMiddleware) BehaviorBasedCaching() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// Get user behavior patterns for cache optimization
		// This is a simplified implementation
		path := c.Request.URL.Path

		// Set cache headers based on user behavior
		if strings.Contains(path, "/feed") {
			// Feed content - short cache for active users
			c.Header("Cache-Control", "private, max-age=300") // 5 minutes
		} else if strings.Contains(path, "/posts") && c.Request.Method == "GET" {
			// Post content - longer cache
			c.Header("Cache-Control", "private, max-age=1800") // 30 minutes
		}

		// Track cache performance
		go func() {
			metadata := map[string]interface{}{
				"cache_strategy": "behavior_based",
				"path":           path,
				"user_id":        userID.(primitive.ObjectID).Hex(),
			}

			sessionID, _ := c.Get("session_id")
			if sessionID != nil {
				action := models.UserAction{
					Type:      "cache_access",
					Target:    path,
					Timestamp: time.Now(),
					Metadata:  metadata,
				}
				m.behaviorService.RecordUserAction(userID.(primitive.ObjectID), sessionID.(string), action)
			}
		}()

		c.Next()
	})
}
