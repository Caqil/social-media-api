// middleware/rate_limit.go
package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"social-media-api/internal/models"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RateLimiter represents a rate limiter
type RateLimiter struct {
	requests        map[string]*ClientInfo
	mutex           sync.RWMutex
	rate            int           // requests per window
	window          time.Duration // time window
	cleanupInterval time.Duration // cleanup interval
}

// ClientInfo stores information about a client's requests
type ClientInfo struct {
	requests  []time.Time
	lastSeen  time.Time
	blocked   bool
	blockTime time.Time
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Rate    int           // requests per window
	Window  time.Duration // time window
	KeyFunc func(*gin.Context) string
	Message string
	Headers bool // whether to add rate limit headers
	Skip    func(*gin.Context) bool
	OnLimit func(*gin.Context) // callback when rate limit is exceeded
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration, cleanupInterval ...time.Duration) *RateLimiter {
	// Default cleanup interval to 1 minute if not provided
	cleanup := time.Minute
	if len(cleanupInterval) > 0 {
		cleanup = cleanupInterval[0]
	}

	rl := &RateLimiter{
		requests:        make(map[string]*ClientInfo),
		rate:            rate,
		window:          window,
		cleanupInterval: cleanup,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(config.Rate, config.Window) // Now works with optional parameter

	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip rate limiting if configured
		if config.Skip != nil && config.Skip(c) {
			c.Next()
			return
		}

		// Get client key
		key := ""
		if config.KeyFunc != nil {
			key = config.KeyFunc(c)
		}
		if key == "" {
			key = c.ClientIP()
		}

		// Check rate limit
		allowed, remaining, resetTime := limiter.isAllowed(key)

		// Add headers if configured
		if config.Headers {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.Rate))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
			c.Header("X-RateLimit-Window", config.Window.String())
		}

		if !allowed {
			// Call limit callback if configured
			if config.OnLimit != nil {
				config.OnLimit(c)
			}

			message := config.Message
			if message == "" {
				message = "Rate limit exceeded"
			}

			utils.ErrorResponse(c, http.StatusTooManyRequests, message, nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// IPRateLimit creates an IP-based rate limiter
func IPRateLimit(rate int, window time.Duration) gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Rate:   rate,
		Window: window,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
		Headers: true,
		Message: "Too many requests from this IP address",
	})
}

// UserRateLimit creates a user-based rate limiter
func UserRateLimit(rate int, window time.Duration) gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Rate:   rate,
		Window: window,
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("user_id"); exists {
				if objID, ok := userID.(primitive.ObjectID); ok {
					return objID.Hex()
				}
			}
			return c.ClientIP() // fallback to IP
		},
		Headers: true,
		Message: "Too many requests from this user",
		Skip: func(c *gin.Context) bool {
			// Skip for unauthenticated users (they'll be limited by IP)
			_, exists := c.Get("user_id")
			return !exists
		},
	})
}

// GlobalRateLimit creates a global rate limiter
func GlobalRateLimit(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, window) // Now works with optional parameter

	return gin.HandlerFunc(func(c *gin.Context) {
		allowed, remaining, resetTime := limiter.isAllowed("global")

		c.Header("X-Global-RateLimit-Limit", fmt.Sprintf("%d", rate))
		c.Header("X-Global-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-Global-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allowed {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "Global rate limit exceeded", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// LoginRateLimit creates a rate limiter for login attempts
func LoginRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Rate:   5,                // 5 attempts
		Window: time.Minute * 15, // per 15 minutes
		KeyFunc: func(c *gin.Context) string {
			return "login_" + c.ClientIP()
		},
		Headers: true,
		Message: "Too many login attempts",
		OnLimit: func(c *gin.Context) {
			// Log failed login attempts
			SetAuthEvent(c, "LOGIN_RATE_LIMIT_EXCEEDED")
		},
	})
}

// RegisterRateLimit creates a rate limiter for registration
func RegisterRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Rate:   3,             // 3 registrations
		Window: time.Hour * 1, // per hour
		KeyFunc: func(c *gin.Context) string {
			return "register_" + c.ClientIP()
		},
		Headers: true,
		Message: "Too many registration attempts",
	})
}

// PostRateLimit creates a rate limiter for creating posts
func PostRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Rate:   10,              // 10 posts
		Window: time.Minute * 5, // per 5 minutes
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("user_id"); exists {
				if objID, ok := userID.(primitive.ObjectID); ok {
					return "post_" + objID.Hex()
				}
			}
			return "post_" + c.ClientIP()
		},
		Headers: true,
		Message: "Too many posts created",
		Skip: func(c *gin.Context) bool {
			// Skip for moderators and admins
			if userRole, exists := c.Get("user_role"); exists {
				role := userRole.(models.UserRole)
				return role == models.RoleModerator || role == models.RoleAdmin || role == models.RoleSuperAdmin
			}
			return false
		},
	})
}

// CommentRateLimit creates a rate limiter for comments
func CommentRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Rate:   20,              // 20 comments
		Window: time.Minute * 5, // per 5 minutes
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("user_id"); exists {
				if objID, ok := userID.(primitive.ObjectID); ok {
					return "comment_" + objID.Hex()
				}
			}
			return "comment_" + c.ClientIP()
		},
		Headers: true,
		Message: "Too many comments posted",
	})
}

// MessageRateLimit creates a rate limiter for messages
func MessageRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Rate:   50,              // 50 messages
		Window: time.Minute * 5, // per 5 minutes
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("user_id"); exists {
				if objID, ok := userID.(primitive.ObjectID); ok {
					return "message_" + objID.Hex()
				}
			}
			return "message_" + c.ClientIP()
		},
		Headers: true,
		Message: "Too many messages sent",
	})
}

// FollowRateLimit creates a rate limiter for follow actions
func FollowRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Rate:   30,               // 30 follow actions
		Window: time.Minute * 10, // per 10 minutes
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("user_id"); exists {
				if objID, ok := userID.(primitive.ObjectID); ok {
					return "follow_" + objID.Hex()
				}
			}
			return "follow_" + c.ClientIP()
		},
		Headers: true,
		Message: "Too many follow/unfollow actions",
	})
}

// LikeRateLimit creates a rate limiter for like actions
func LikeRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Rate:   100,             // 100 likes
		Window: time.Minute * 5, // per 5 minutes
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("user_id"); exists {
				if objID, ok := userID.(primitive.ObjectID); ok {
					return "like_" + objID.Hex()
				}
			}
			return "like_" + c.ClientIP()
		},
		Headers: true,
		Message: "Too many like actions",
	})
}

// AdminRateLimit creates a less restrictive rate limiter for admins
func AdminRateLimit() gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Rate:   1000,            // 1000 requests
		Window: time.Minute * 5, // per 5 minutes
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("user_id"); exists {
				if objID, ok := userID.(primitive.ObjectID); ok {
					return "admin_" + objID.Hex()
				}
			}
			return "admin_" + c.ClientIP()
		},
		Headers: true,
		Message: "Admin rate limit exceeded",
	})
}

// Methods for RateLimiter

func (rl *RateLimiter) isAllowed(key string) (bool, int, time.Time) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Get or create client info
	client, exists := rl.requests[key]
	if !exists {
		client = &ClientInfo{
			requests: make([]time.Time, 0),
			lastSeen: now,
		}
		rl.requests[key] = client
	}

	// Update last seen
	client.lastSeen = now

	// Check if client is blocked
	if client.blocked && now.Before(client.blockTime.Add(rl.window)) {
		return false, 0, client.blockTime.Add(rl.window)
	}

	// Remove old requests outside the window
	cutoff := now.Add(-rl.window)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	client.requests = validRequests

	// Check if limit exceeded
	if len(client.requests) >= rl.rate {
		client.blocked = true
		client.blockTime = now
		return false, 0, now.Add(rl.window)
	}

	// Add current request
	client.requests = append(client.requests, now)
	client.blocked = false

	remaining := rl.rate - len(client.requests)
	var resetTime time.Time
	if len(client.requests) > 0 {
		resetTime = client.requests[0].Add(rl.window)
	} else {
		resetTime = now.Add(rl.window)
	}

	return true, remaining, resetTime
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval) // use the renamed field
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window * 2) // Keep data for 2 windows

		for key, client := range rl.requests {
			if client.lastSeen.Before(cutoff) {
				delete(rl.requests, key)
			}
		}
		rl.mutex.Unlock()
	}
}

// GetRateLimitInfo returns current rate limit information for a key
func (rl *RateLimiter) GetRateLimitInfo(key string) (remaining int, resetTime time.Time, blocked bool) {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	client, exists := rl.requests[key]
	if !exists {
		return rl.rate, time.Now().Add(rl.window), false
	}

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Count valid requests
	validCount := 0
	var oldestRequest time.Time
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validCount++
			if oldestRequest.IsZero() || reqTime.Before(oldestRequest) {
				oldestRequest = reqTime
			}
		}
	}

	remaining = rl.rate - validCount
	if remaining < 0 {
		remaining = 0
	}

	if !oldestRequest.IsZero() {
		resetTime = oldestRequest.Add(rl.window)
	} else {
		resetTime = now.Add(rl.window)
	}

	blocked = client.blocked && now.Before(client.blockTime.Add(rl.window))

	return remaining, resetTime, blocked
}
