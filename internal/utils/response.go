// utils/response.go
package utils

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Response represents the standard API response structure
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      interface{} `json:"meta,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorInfo represents detailed error information
type ErrorInfo struct {
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Field   string      `json:"field,omitempty"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidationError represents validation error details
type ValidationError struct {
	Field   string      `json:"field"`
	Tag     string      `json:"tag"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// PaginatedResponse represents paginated data response
type PaginatedResponse struct {
	Success    bool             `json:"success"`
	Message    string           `json:"message"`
	Data       interface{}      `json:"data"`
	Pagination PaginationMeta   `json:"pagination"`
	Links      *PaginationLinks `json:"links,omitempty"`
	Timestamp  int64            `json:"timestamp"`
}

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	response := Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: getCurrentTimestamp(),
	}
	c.JSON(statusCode, response)
}

// SuccessResponseWithMeta sends a success response with metadata
func SuccessResponseWithMeta(c *gin.Context, statusCode int, message string, data interface{}, meta interface{}) {
	response := Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Meta:      meta,
		Timestamp: getCurrentTimestamp(),
	}
	c.JSON(statusCode, response)
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	var errorInfo *ErrorInfo

	if err != nil {
		errorInfo = &ErrorInfo{
			Message: err.Error(),
		}
	}

	response := Response{
		Success:   false,
		Message:   message,
		Error:     errorInfo,
		Timestamp: getCurrentTimestamp(),
	}
	c.JSON(statusCode, response)
}

// ErrorResponseWithCode sends an error response with error code
func ErrorResponseWithCode(c *gin.Context, statusCode int, message string, errorCode string, err error) {
	var errorInfo *ErrorInfo

	if err != nil {
		errorInfo = &ErrorInfo{
			Code:    errorCode,
			Message: err.Error(),
		}
	} else {
		errorInfo = &ErrorInfo{
			Code:    errorCode,
			Message: message,
		}
	}

	response := Response{
		Success:   false,
		Message:   message,
		Error:     errorInfo,
		Timestamp: getCurrentTimestamp(),
	}
	c.JSON(statusCode, response)
}

// ErrorResponseWithDetails sends an error response with detailed error information
func ErrorResponseWithDetails(c *gin.Context, statusCode int, message string, errorCode string, details interface{}) {
	errorInfo := &ErrorInfo{
		Code:    errorCode,
		Message: message,
		Details: details,
	}

	response := Response{
		Success:   false,
		Message:   message,
		Error:     errorInfo,
		Timestamp: getCurrentTimestamp(),
	}
	c.JSON(statusCode, response)
}

// ValidationErrorResponse sends validation error response
func ValidationErrorResponse(c *gin.Context, err error) {
	var validationErrors []ValidationError

	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErr {
			validationErrors = append(validationErrors, ValidationError{
				Field:   getJSONFieldName(fieldErr),
				Tag:     fieldErr.Tag(),
				Message: getValidationErrorMessage(fieldErr),
				Value:   fieldErr.Value(),
			})
		}
	} else {
		// Handle other types of validation errors
		validationErrors = append(validationErrors, ValidationError{
			Message: err.Error(),
		})
	}

	errorInfo := &ErrorInfo{
		Code:    "VALIDATION_ERROR",
		Message: "Validation failed",
		Details: validationErrors,
	}

	response := Response{
		Success:   false,
		Message:   "Validation failed",
		Error:     errorInfo,
		Timestamp: getCurrentTimestamp(),
	}

	c.JSON(http.StatusBadRequest, response)
}

// PaginatedSuccessResponse sends a paginated success response
func PaginatedSuccessResponse(c *gin.Context, message string, data interface{}, pagination PaginationMeta, links *PaginationLinks) {
	response := PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
		Links:      links,
		Timestamp:  getCurrentTimestamp(),
	}
	c.JSON(http.StatusOK, response)
}

// NotFoundResponse sends a 404 not found response
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponseWithCode(c, http.StatusNotFound, message, "NOT_FOUND", nil)
}

// UnauthorizedResponse sends a 401 unauthorized response
func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponseWithCode(c, http.StatusUnauthorized, message, "UNAUTHORIZED", nil)
}

// ForbiddenResponse sends a 403 forbidden response
func ForbiddenResponse(c *gin.Context, message string) {
	ErrorResponseWithCode(c, http.StatusForbidden, message, "FORBIDDEN", nil)
}

// BadRequestResponse sends a 400 bad request response
func BadRequestResponse(c *gin.Context, message string, err error) {
	ErrorResponseWithCode(c, http.StatusBadRequest, message, "BAD_REQUEST", err)
}

// InternalServerErrorResponse sends a 500 internal server error response
func InternalServerErrorResponse(c *gin.Context, message string, err error) {
	ErrorResponseWithCode(c, http.StatusInternalServerError, message, "INTERNAL_ERROR", err)
}

// ConflictResponse sends a 409 conflict response
func ConflictResponse(c *gin.Context, message string, err error) {
	ErrorResponseWithCode(c, http.StatusConflict, message, "CONFLICT", err)
}

// TooManyRequestsResponse sends a 429 too many requests response
func TooManyRequestsResponse(c *gin.Context, message string) {
	ErrorResponseWithCode(c, http.StatusTooManyRequests, message, "RATE_LIMIT_EXCEEDED", nil)
}

// ServiceUnavailableResponse sends a 503 service unavailable response
func ServiceUnavailableResponse(c *gin.Context, message string) {
	ErrorResponseWithCode(c, http.StatusServiceUnavailable, message, "SERVICE_UNAVAILABLE", nil)
}

// CreatedResponse sends a 201 created response
func CreatedResponse(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusCreated, message, data)
}

// AcceptedResponse sends a 202 accepted response
func AcceptedResponse(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusAccepted, message, data)
}

// NoContentResponse sends a 204 no content response
func NoContentResponse(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// OkResponse sends a 200 OK response
func OkResponse(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusOK, message, data)
}

// Custom response functions for specific use cases

// LoginSuccessResponse sends a successful login response
func LoginSuccessResponse(c *gin.Context, user interface{}, tokens interface{}) {
	data := gin.H{
		"user":   user,
		"tokens": tokens,
	}
	SuccessResponse(c, http.StatusOK, "Login successful", data)
}

// LogoutSuccessResponse sends a successful logout response
func LogoutSuccessResponse(c *gin.Context) {
	SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}

// RegistrationSuccessResponse sends a successful registration response
func RegistrationSuccessResponse(c *gin.Context, user interface{}) {
	SuccessResponse(c, http.StatusCreated, "User registered successfully", user)
}

// ProfileUpdateSuccessResponse sends a successful profile update response
func ProfileUpdateSuccessResponse(c *gin.Context, user interface{}) {
	SuccessResponse(c, http.StatusOK, "Profile updated successfully", user)
}

// PasswordChangeSuccessResponse sends a successful password change response
func PasswordChangeSuccessResponse(c *gin.Context) {
	SuccessResponse(c, http.StatusOK, "Password changed successfully", nil)
}

// EmailVerificationSuccessResponse sends a successful email verification response
func EmailVerificationSuccessResponse(c *gin.Context) {
	SuccessResponse(c, http.StatusOK, "Email verified successfully", nil)
}

// PasswordResetEmailSentResponse sends password reset email sent response
func PasswordResetEmailSentResponse(c *gin.Context) {
	SuccessResponse(c, http.StatusOK, "Password reset link sent to email", nil)
}

// Helper functions

// getCurrentTimestamp returns current unix timestamp
func getCurrentTimestamp() int64 {
	return getCurrentTime().Unix()
}

// getJSONFieldName extracts JSON field name from validation error
func getJSONFieldName(fe validator.FieldError) string {
	field := fe.Field()

	// Convert struct field name to JSON field name (snake_case)
	var result strings.Builder
	for i, r := range field {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// getValidationErrorMessage returns user-friendly validation error message
func getValidationErrorMessage(fe validator.FieldError) string {
	field := getJSONFieldName(fe)

	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + fe.Param() + " characters long"
	case "max":
		return field + " must be at most " + fe.Param() + " characters long"
	case "len":
		return field + " must be exactly " + fe.Param() + " characters long"
	case "numeric":
		return field + " must be a number"
	case "alpha":
		return field + " must contain only letters"
	case "alphanum":
		return field + " must contain only letters and numbers"
	case "url":
		return field + " must be a valid URL"
	case "uri":
		return field + " must be a valid URI"
	case "uuid":
		return field + " must be a valid UUID"
	case "oneof":
		return field + " must be one of: " + fe.Param()
	case "gte":
		return field + " must be greater than or equal to " + fe.Param()
	case "lte":
		return field + " must be less than or equal to " + fe.Param()
	case "gt":
		return field + " must be greater than " + fe.Param()
	case "lt":
		return field + " must be less than " + fe.Param()
	case "eqfield":
		return field + " must be equal to " + fe.Param()
	case "nefield":
		return field + " must not be equal to " + fe.Param()
	case "unique":
		return field + " must be unique"
	default:
		return field + " is invalid"
	}
}

// ApiResponse represents a generic API response wrapper
type ApiResponse struct {
	StatusCode int         `json:"status_code"`
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Error      *ErrorInfo  `json:"error,omitempty"`
	Meta       interface{} `json:"meta,omitempty"`
	Timestamp  int64       `json:"timestamp"`
}

// NewApiResponse creates a new API response
func NewApiResponse(statusCode int, success bool, message string, data interface{}, errorInfo *ErrorInfo, meta interface{}) *ApiResponse {
	return &ApiResponse{
		StatusCode: statusCode,
		Success:    success,
		Message:    message,
		Data:       data,
		Error:      errorInfo,
		Meta:       meta,
		Timestamp:  getCurrentTimestamp(),
	}
}

// Send sends the API response
func (ar *ApiResponse) Send(c *gin.Context) {
	c.JSON(ar.StatusCode, ar)
}

// JSONResponse sends a custom JSON response
func JSONResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// XMLResponse sends an XML response
func XMLResponse(c *gin.Context, statusCode int, data interface{}) {
	c.XML(statusCode, data)
}

// YAMLResponse sends a YAML response
func YAMLResponse(c *gin.Context, statusCode int, data interface{}) {
	c.YAML(statusCode, data)
}

// FileResponse sends a file response
func FileResponse(c *gin.Context, filepath string) {
	c.File(filepath)
}

// AttachmentResponse sends a file as attachment
func AttachmentResponse(c *gin.Context, filepath, filename string) {
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.File(filepath)
}

// RedirectResponse sends a redirect response
func RedirectResponse(c *gin.Context, statusCode int, location string) {
	c.Redirect(statusCode, location)
}

// HealthCheckResponse sends a health check response
func HealthCheckResponse(c *gin.Context, status string) {
	data := gin.H{
		"status":    status,
		"timestamp": getCurrentTimestamp(),
		"uptime":    "healthy", // You can calculate actual uptime here
	}
	SuccessResponse(c, http.StatusOK, "Service is healthy", data)
}

// getCurrentTime returns current time (can be mocked for testing)
var getCurrentTime = func() time.Time {
	return time.Unix(1640995200, 0) // Mock timestamp for testing
}

// SetTimeProvider sets a custom time provider (useful for testing)
func SetTimeProvider(provider func() time.Time) {
	getCurrentTime = provider
}
