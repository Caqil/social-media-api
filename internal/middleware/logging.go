// middleware/logging.go
package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LogLevel represents different log levels
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	Level        LogLevel  `json:"level"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int64     `json:"response_time_ms"`
	ClientIP     string    `json:"client_ip"`
	UserAgent    string    `json:"user_agent"`
	UserID       string    `json:"user_id,omitempty"`
	RequestID    string    `json:"request_id,omitempty"`
	Referer      string    `json:"referer,omitempty"`
	RequestSize  int64     `json:"request_size,omitempty"`
	ResponseSize int       `json:"response_size,omitempty"`
	Error        string    `json:"error,omitempty"`
	Details      string    `json:"details,omitempty"`
}

// CustomLogWriter wraps the response writer to capture response data
type CustomLogWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *CustomLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *CustomLogWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// getUserID safely extracts user ID from context
func getUserID(c *gin.Context) interface{} {
	if userID, exists := c.Get("user_id"); exists {
		if objID, ok := userID.(primitive.ObjectID); ok {
			return objID.Hex()
		}
		return userID
	}
	return "anonymous"
}

// Logger creates a custom logging middleware
func Logger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Header("X-Request-ID", requestID)
		}
		c.Set("request_id", requestID)

		// Get request size
		var requestSize int64
		if c.Request.ContentLength > 0 {
			requestSize = c.Request.ContentLength
		}

		// Wrap the response writer to capture response data
		customWriter := &CustomLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
			statusCode:     200,
		}
		c.Writer = customWriter

		// Process request
		c.Next()

		// Calculate response time
		responseTime := time.Since(start).Milliseconds()

		// Get user ID if authenticated
		var userID string
		if user, exists := c.Get("user_id"); exists {
			if objectID, ok := user.(primitive.ObjectID); ok {
				userID = objectID.Hex()
			} else if strID, ok := user.(string); ok {
				userID = strID
			}
		}

		// Create log entry
		entry := LogEntry{
			Timestamp:    start,
			Level:        getLogLevel(customWriter.statusCode),
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			StatusCode:   customWriter.statusCode,
			ResponseTime: responseTime,
			ClientIP:     c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			UserID:       userID,
			RequestID:    requestID,
			Referer:      c.Request.Referer(),
			RequestSize:  requestSize,
			ResponseSize: customWriter.body.Len(),
		}

		// Add query parameters if present
		if c.Request.URL.RawQuery != "" {
			entry.Path += "?" + c.Request.URL.RawQuery
		}

		// Add error information if present
		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
		}

		// Log the entry
		logEntry(entry)
	})
}

// DetailedLogger creates a more detailed logging middleware for debugging
func DetailedLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// Log request details
		logRequestDetails(c)

		// Process request
		c.Next()

		// Log response details
		logResponseDetails(c, start)
	})
}

// APILogger creates an API-specific logger with custom formatting
func APILogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// Skip logging for health checks and static files
		if shouldSkipLogging(c.Request.URL.Path) {
			c.Next()
			return
		}

		c.Next()

		// Create structured log
		duration := time.Since(start)

		logData := map[string]interface{}{
			"timestamp":     start.Format(time.RFC3339),
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status":        c.Writer.Status(),
			"duration_ms":   duration.Milliseconds(),
			"client_ip":     c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"response_size": c.Writer.Size(),
		}

		// Add user context if available
		if userID, exists := c.Get("user_id"); exists {
			logData["user_id"] = userID
		}

		// Add request ID if available
		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			logData["request_id"] = requestID
		}

		// Add error information if present
		if len(c.Errors) > 0 {
			logData["errors"] = c.Errors.String()
		}

		// Convert to JSON and log
		jsonData, _ := json.Marshal(logData)
		log.Printf("API_LOG: %s", string(jsonData))
	})
}

// DatabaseLogger logs database operations
func DatabaseLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Log database operations if any
		if operations, exists := c.Get("db_operations"); exists {
			logDatabaseOperations(c, operations)
		}
	})
}

// SecurityLogger logs security-related events
func SecurityLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Log authentication attempts
		if authEvent, exists := c.Get("auth_event"); exists {
			logSecurityEvent(c, authEvent.(string))
		}

		// Log failed authorization attempts
		if c.Writer.Status() == 403 {
			logSecurityEvent(c, "AUTHORIZATION_FAILED")
		}

		// Log suspicious activity
		if c.Writer.Status() == 429 {
			logSecurityEvent(c, "RATE_LIMIT_EXCEEDED")
		}
	})
}

// PerformanceLogger logs performance metrics
func PerformanceLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		// Log slow requests (>1 second)
		if duration > time.Second {
			log.Printf("SLOW_REQUEST: %s %s took %v | UserID: %v | IP: %s",
				c.Request.Method,
				c.Request.URL.Path,
				duration,
				getUserID(c),
				c.ClientIP(),
			)
		}

		// Log performance metrics
		if metrics, exists := c.Get("performance_metrics"); exists {
			logPerformanceMetrics(c, metrics, duration)
		}
	})
}

// Helper functions

func generateRequestID() string {
	return primitive.NewObjectID().Hex()
}

func getLogLevel(statusCode int) LogLevel {
	switch {
	case statusCode >= 500:
		return ERROR
	case statusCode >= 400:
		return WARN
	case statusCode >= 300:
		return INFO
	default:
		return INFO
	}
}

func logEntry(entry LogEntry) {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	// Write to stdout (can be configured to write to files)
	fmt.Fprintf(os.Stdout, "%s\n", string(jsonData))
}

func logRequestDetails(c *gin.Context) {
	log.Printf("REQUEST: %s %s | IP: %s | UserAgent: %s | ContentLength: %d",
		c.Request.Method,
		c.Request.URL.Path,
		c.ClientIP(),
		c.Request.UserAgent(),
		c.Request.ContentLength,
	)

	// Log headers in development mode
	if gin.Mode() == gin.DebugMode {
		log.Printf("REQUEST_HEADERS: %v", c.Request.Header)
	}
}

func logResponseDetails(c *gin.Context, start time.Time) {
	duration := time.Since(start)

	log.Printf("RESPONSE: %s %s | Status: %d | Duration: %v | Size: %d",
		c.Request.Method,
		c.Request.URL.Path,
		c.Writer.Status(),
		duration,
		c.Writer.Size(),
	)
}

func shouldSkipLogging(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/favicon.ico",
		"/static/",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

func logDatabaseOperations(c *gin.Context, operations interface{}) {
	log.Printf("DB_OPERATIONS: %s %s | Operations: %v | UserID: %v",
		c.Request.Method,
		c.Request.URL.Path,
		operations,
		getUserID(c),
	)
}

func logSecurityEvent(c *gin.Context, event string) {
	log.Printf("SECURITY_EVENT: %s | %s %s | IP: %s | UserID: %v | UserAgent: %s",
		event,
		c.Request.Method,
		c.Request.URL.Path,
		c.ClientIP(),
		getUserID(c),
		c.Request.UserAgent(),
	)
}

func logPerformanceMetrics(c *gin.Context, metrics interface{}, duration time.Duration) {
	log.Printf("PERFORMANCE: %s %s | Duration: %v | Metrics: %v",
		c.Request.Method,
		c.Request.URL.Path,
		duration,
		metrics,
	)
}

// SetDBOperations sets database operations in context for logging
func SetDBOperations(c *gin.Context, operations []string) {
	c.Set("db_operations", operations)
}

// SetAuthEvent sets authentication event in context for logging
func SetAuthEvent(c *gin.Context, eventType string) {
	// Store authentication event in context for logging
	c.Set("auth_event", eventType)
	c.Set("auth_event_time", time.Now())

	// You can also log it immediately if preferred
	userAgent := c.Request.UserAgent()
	clientIP := c.ClientIP()

	// Log the authentication event
	fmt.Printf("AUTH_EVENT: %s | IP: %s | UserAgent: %s | Time: %v\n",
		eventType, clientIP, userAgent, time.Now())
}

// SetPerformanceMetrics sets performance metrics in context for logging
func SetPerformanceMetrics(c *gin.Context, metrics map[string]interface{}) {
	c.Set("performance_metrics", metrics)
}
