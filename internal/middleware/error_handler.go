// middleware/error_handler.go
package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Error     string      `json:"error,omitempty"`
	Code      string      `json:"code,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Path      string      `json:"path"`
	Method    string      `json:"method"`
	RequestID string      `json:"request_id,omitempty"`
}

// GlobalErrorHandler handles all unhandled errors and panics
func GlobalErrorHandler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				log.Printf("PANIC RECOVERED: %v\n%s", err, debug.Stack())

				// Create structured error response
				errorResponse := ErrorResponse{
					Success:   false,
					Message:   "Internal server error",
					Error:     "An unexpected error occurred",
					Code:      "INTERNAL_ERROR",
					Timestamp: time.Now(),
					Path:      c.Request.URL.Path,
					Method:    c.Request.Method,
					RequestID: getRequestID(c),
				}

				// Set appropriate status code
				c.JSON(http.StatusInternalServerError, errorResponse)
				c.Abort()
			}
		}()

		// Continue with next middleware/handler
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			handleGinErrors(c)
		}
	})
}

// DatabaseErrorHandler handles MongoDB-specific errors
func DatabaseErrorHandler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Check for database errors in the context
		if dbErr, exists := c.Get("db_error"); exists {
			handleDatabaseError(c, dbErr.(error))
		}
	})
}

// ValidationErrorHandler handles validation errors
func ValidationErrorHandler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Check for validation errors
		if validationErr, exists := c.Get("validation_error"); exists {
			handleValidationError(c, validationErr.(error))
		}
	})
}

// NotFoundHandler handles 404 errors
func NotFoundHandler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		errorResponse := ErrorResponse{
			Success:   false,
			Message:   "Route not found",
			Error:     fmt.Sprintf("The requested endpoint %s %s was not found", c.Request.Method, c.Request.URL.Path),
			Code:      "ROUTE_NOT_FOUND",
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
			Method:    c.Request.Method,
			RequestID: getRequestID(c),
		}

		c.JSON(http.StatusNotFound, errorResponse)
	})
}

// MethodNotAllowedHandler handles 405 errors
func MethodNotAllowedHandler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		errorResponse := ErrorResponse{
			Success:   false,
			Message:   "Method not allowed",
			Error:     fmt.Sprintf("The %s method is not allowed for this endpoint", c.Request.Method),
			Code:      "METHOD_NOT_ALLOWED",
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
			Method:    c.Request.Method,
			RequestID: getRequestID(c),
		}

		c.JSON(http.StatusMethodNotAllowed, errorResponse)
	})
}

// handleGinErrors processes Gin framework errors
func handleGinErrors(c *gin.Context) {
	ginError := c.Errors.Last()
	if ginError == nil {
		return
	}

	var statusCode int
	var message string
	var errorCode string

	// Determine error type and appropriate response
	switch ginError.Type {
	case gin.ErrorTypeBind:
		statusCode = http.StatusBadRequest
		message = "Invalid request data"
		errorCode = "INVALID_REQUEST"
	case gin.ErrorTypePublic:
		statusCode = http.StatusBadRequest
		message = ginError.Error()
		errorCode = "BAD_REQUEST"
	case gin.ErrorTypePrivate:
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
		errorCode = "INTERNAL_ERROR"
		// Log private errors
		log.Printf("Private error: %v", ginError.Error())
	default:
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
		errorCode = "INTERNAL_ERROR"
	}

	errorResponse := ErrorResponse{
		Success:   false,
		Message:   message,
		Error:     ginError.Error(),
		Code:      errorCode,
		Timestamp: time.Now(),
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		RequestID: getRequestID(c),
	}

	c.JSON(statusCode, errorResponse)
}

// handleDatabaseError processes MongoDB errors
func handleDatabaseError(c *gin.Context, err error) {
	var statusCode int
	var message string
	var errorCode string

	switch {
	case mongo.IsDuplicateKeyError(err):
		statusCode = http.StatusConflict
		message = "Resource already exists"
		errorCode = "DUPLICATE_RESOURCE"

		// Extract field name from duplicate key error
		if strings.Contains(err.Error(), "username") {
			message = "Username already exists"
		} else if strings.Contains(err.Error(), "email") {
			message = "Email already exists"
		}

	case err == mongo.ErrNoDocuments:
		statusCode = http.StatusNotFound
		message = "Resource not found"
		errorCode = "RESOURCE_NOT_FOUND"

	case mongo.IsTimeout(err):
		statusCode = http.StatusRequestTimeout
		message = "Database operation timed out"
		errorCode = "DATABASE_TIMEOUT"

	case mongo.IsNetworkError(err):
		statusCode = http.StatusServiceUnavailable
		message = "Database connection error"
		errorCode = "DATABASE_CONNECTION_ERROR"

	default:
		statusCode = http.StatusInternalServerError
		message = "Database error"
		errorCode = "DATABASE_ERROR"
		// Log unexpected database errors
		log.Printf("Unexpected database error: %v", err)
	}

	errorResponse := ErrorResponse{
		Success:   false,
		Message:   message,
		Error:     err.Error(),
		Code:      errorCode,
		Timestamp: time.Now(),
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		RequestID: getRequestID(c),
	}

	c.JSON(statusCode, errorResponse)
}

// handleValidationError processes validation errors
func handleValidationError(c *gin.Context, err error) {
	errorResponse := ErrorResponse{
		Success:   false,
		Message:   "Validation failed",
		Error:     err.Error(),
		Code:      "VALIDATION_ERROR",
		Timestamp: time.Now(),
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		RequestID: getRequestID(c),
	}

	c.JSON(http.StatusBadRequest, errorResponse)
}

// SetDBError sets a database error in the context
func SetDBError(c *gin.Context, err error) {
	c.Set("db_error", err)
}

// SetValidationError sets a validation error in the context
func SetValidationError(c *gin.Context, err error) {
	c.Set("validation_error", err)
}

// CustomError creates a custom error response
func CustomError(c *gin.Context, statusCode int, message, errorCode string, details interface{}) {
	errorResponse := ErrorResponse{
		Success:   false,
		Message:   message,
		Code:      errorCode,
		Details:   details,
		Timestamp: time.Now(),
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		RequestID: getRequestID(c),
	}

	c.JSON(statusCode, errorResponse)
	c.Abort()
}

// getRequestID extracts or generates a request ID
func getRequestID(c *gin.Context) string {
	// Try to get request ID from header
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Try to get from context
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}

	return ""
}

// LogError logs an error with context information
func LogError(c *gin.Context, err error, level string) {
	requestID := getRequestID(c)
	userID := ""

	if user, exists := c.Get("user_id"); exists {
		userID = user.(string)
	}

	log.Printf("[%s] Error: %v | Path: %s | Method: %s | UserID: %s | RequestID: %s",
		level, err, c.Request.URL.Path, c.Request.Method, userID, requestID)
}

// ErrorLogger middleware for logging errors
func ErrorLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Log any errors that occurred
		if len(c.Errors) > 0 {
			for _, ginError := range c.Errors {
				LogError(c, ginError, "ERROR")
			}
		}
	})
}
