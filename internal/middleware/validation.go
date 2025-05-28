// middleware/validation.go
package middleware

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CustomValidator wraps the validator instance
type CustomValidator struct {
	validator *validator.Validate
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents the validation error response
type ValidationErrorResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors"`
}

// Global validator instance
var customValidator *CustomValidator

// InitValidator initializes the custom validator with custom rules
func InitValidator() {
	customValidator = &CustomValidator{
		validator: validator.New(),
	}

	// Register custom validation tags
	registerCustomValidations()

	// Register custom tag name function
	customValidator.validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// GetValidator returns the validator instance
func GetValidator() *validator.Validate {
	if customValidator == nil {
		InitValidator()
	}
	return customValidator.validator
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) []ValidationError {
	validate := GetValidator()
	err := validate.Struct(s)

	if err == nil {
		return nil
	}

	var validationErrors []ValidationError

	// Type assert to validator.ValidationErrors with proper error handling
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range ve {
			validationError := ValidationError{
				Field:   fieldError.Field(),
				Tag:     fieldError.Tag(),
				Value:   fmt.Sprintf("%v", fieldError.Value()),
				Message: getValidationMessage(fieldError),
			}
			validationErrors = append(validationErrors, validationError)
		}
	} else {
		// Handle non-validation errors (like JSON parsing errors)
		validationErrors = append(validationErrors, ValidationError{
			Field:   "general",
			Tag:     "error",
			Value:   "",
			Message: err.Error(),
		})
	}

	return validationErrors
}

// ValidateJSON middleware validates JSON request body
func ValidateJSON(model interface{}) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Create a new instance of the model type
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}

		newModel := reflect.New(modelType).Interface()

		// Bind JSON to model
		if err := c.ShouldBindJSON(newModel); err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format", nil)
			c.Abort()
			return
		}

		// Validate the model
		if validationErrors := ValidateStruct(newModel); validationErrors != nil {
			c.JSON(http.StatusBadRequest, ValidationErrorResponse{
				Success: false,
				Message: "Validation failed",
				Errors:  validationErrors,
			})
			c.Abort()
			return
		}

		// Set the validated model in context
		c.Set("validated_data", newModel)
		c.Next()
	})
}

// ValidateQuery validates query parameters
func ValidateQuery(rules map[string]string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var errors []ValidationError

		for param, rule := range rules {
			value := c.Query(param)

			if err := validateField(param, value, rule); err != nil {
				errors = append(errors, *err)
			}
		}

		if len(errors) > 0 {
			c.JSON(http.StatusBadRequest, ValidationErrorResponse{
				Success: false,
				Message: "Query parameter validation failed",
				Errors:  errors,
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// ValidateParams validates URL parameters
func ValidateParams(rules map[string]string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var errors []ValidationError

		for param, rule := range rules {
			value := c.Param(param)

			if err := validateField(param, value, rule); err != nil {
				errors = append(errors, *err)
			}
		}

		if len(errors) > 0 {
			c.JSON(http.StatusBadRequest, ValidationErrorResponse{
				Success: false,
				Message: "URL parameter validation failed",
				Errors:  errors,
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// ValidateObjectID validates MongoDB ObjectID parameters
func ValidateObjectID(paramNames ...string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var errors []ValidationError

		for _, paramName := range paramNames {
			value := c.Param(paramName)
			if value == "" {
				continue
			}

			if !primitive.IsValidObjectID(value) {
				errors = append(errors, ValidationError{
					Field:   paramName,
					Tag:     "objectid",
					Value:   value,
					Message: fmt.Sprintf("%s must be a valid ObjectID", paramName),
				})
			}
		}

		if len(errors) > 0 {
			c.JSON(http.StatusBadRequest, ValidationErrorResponse{
				Success: false,
				Message: "Invalid ObjectID format",
				Errors:  errors,
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// ValidatePagination validates pagination parameters
func ValidatePagination() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var errors []ValidationError

		// Validate page parameter
		if pageStr := c.Query("page"); pageStr != "" {
			if page, err := strconv.Atoi(pageStr); err != nil || page < 1 {
				errors = append(errors, ValidationError{
					Field:   "page",
					Tag:     "min",
					Value:   pageStr,
					Message: "page must be a positive integer",
				})
			}
		}

		// Validate limit parameter
		if limitStr := c.Query("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err != nil || limit < 1 || limit > 100 {
				errors = append(errors, ValidationError{
					Field:   "limit",
					Tag:     "range",
					Value:   limitStr,
					Message: "limit must be between 1 and 100",
				})
			}
		}

		if len(errors) > 0 {
			c.JSON(http.StatusBadRequest, ValidationErrorResponse{
				Success: false,
				Message: "Pagination validation failed",
				Errors:  errors,
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// ValidateFileUpload validates file upload
func ValidateFileUpload(maxSize int64, allowedTypes []string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "No file uploaded", nil)
			c.Abort()
			return
		}
		defer file.Close()

		var errors []ValidationError

		// Validate file size
		if header.Size > maxSize {
			errors = append(errors, ValidationError{
				Field:   "file",
				Tag:     "max_size",
				Value:   fmt.Sprintf("%d", header.Size),
				Message: fmt.Sprintf("File size must be less than %d bytes", maxSize),
			})
		}

		// Validate file type
		if len(allowedTypes) > 0 {
			contentType := header.Header.Get("Content-Type")
			allowed := false
			for _, allowedType := range allowedTypes {
				if strings.HasPrefix(contentType, allowedType) {
					allowed = true
					break
				}
			}
			if !allowed {
				errors = append(errors, ValidationError{
					Field:   "file",
					Tag:     "file_type",
					Value:   contentType,
					Message: fmt.Sprintf("File type must be one of: %s", strings.Join(allowedTypes, ", ")),
				})
			}
		}

		if len(errors) > 0 {
			c.JSON(http.StatusBadRequest, ValidationErrorResponse{
				Success: false,
				Message: "File validation failed",
				Errors:  errors,
			})
			c.Abort()
			return
		}

		// Set file info in context
		c.Set("uploaded_file", file)
		c.Set("file_header", header)
		c.Next()
	})
}

// Helper functions

func registerCustomValidations() {
	validator := GetValidator()

	// Register custom validation for username
	validator.RegisterValidation("username", validateUsername)

	// Register custom validation for ObjectID
	validator.RegisterValidation("objectid", validateObjectIDTag)

	// Register custom validation for privacy level
	validator.RegisterValidation("privacy_level", validatePrivacyLevel)

	// Register custom validation for user role
	validator.RegisterValidation("user_role", validateUserRole)

	// Register custom validation for reaction type
	validator.RegisterValidation("reaction_type", validateReactionType)

	// Register custom validation for content type
	validator.RegisterValidation("content_type", validateContentType)

	// Register custom validation for notification type
	validator.RegisterValidation("notification_type", validateNotificationType)
}

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	return utils.IsValidUsername(username)
}

func validateObjectIDTag(fl validator.FieldLevel) bool {
	objectID := fl.Field().String()
	return primitive.IsValidObjectID(objectID)
}

func validatePrivacyLevel(fl validator.FieldLevel) bool {
	level := fl.Field().String()
	validLevels := []string{"public", "friends", "private"}

	for _, validLevel := range validLevels {
		if level == validLevel {
			return true
		}
	}
	return false
}

func validateUserRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	validRoles := []string{"user", "moderator", "admin", "super_admin"}

	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

func validateReactionType(fl validator.FieldLevel) bool {
	reaction := fl.Field().String()
	validReactions := []string{"like", "love", "haha", "wow", "sad", "angry", "support"}

	for _, validReaction := range validReactions {
		if reaction == validReaction {
			return true
		}
	}
	return false
}

func validateContentType(fl validator.FieldLevel) bool {
	contentType := fl.Field().String()
	validTypes := []string{"text", "image", "video", "audio", "file", "link", "gif", "poll"}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

func validateNotificationType(fl validator.FieldLevel) bool {
	notifType := fl.Field().String()
	validTypes := []string{
		"like", "love", "comment", "follow", "message", "mention",
		"group_invite", "event_invite", "friend_request", "post_share",
		"story_view", "group_post", "event_reminder",
	}

	for _, validType := range validTypes {
		if notifType == validType {
			return true
		}
	}
	return false
}

func validateField(fieldName, value, rule string) *ValidationError {
	rules := strings.Split(rule, "|")

	for _, r := range rules {
		parts := strings.Split(r, ":")
		tag := parts[0]
		var param string
		if len(parts) > 1 {
			param = parts[1]
		}

		switch tag {
		case "required":
			if value == "" {
				return &ValidationError{
					Field:   fieldName,
					Tag:     tag,
					Value:   value,
					Message: fmt.Sprintf("%s is required", fieldName),
				}
			}
		case "min":
			if minLen, err := strconv.Atoi(param); err == nil {
				if len(value) < minLen {
					return &ValidationError{
						Field:   fieldName,
						Tag:     tag,
						Value:   value,
						Message: fmt.Sprintf("%s must be at least %d characters", fieldName, minLen),
					}
				}
			}
		case "max":
			if maxLen, err := strconv.Atoi(param); err == nil {
				if len(value) > maxLen {
					return &ValidationError{
						Field:   fieldName,
						Tag:     tag,
						Value:   value,
						Message: fmt.Sprintf("%s must be at most %d characters", fieldName, maxLen),
					}
				}
			}
		case "numeric":
			if _, err := strconv.Atoi(value); err != nil {
				return &ValidationError{
					Field:   fieldName,
					Tag:     tag,
					Value:   value,
					Message: fmt.Sprintf("%s must be numeric", fieldName),
				}
			}
		case "objectid":
			if !primitive.IsValidObjectID(value) {
				return &ValidationError{
					Field:   fieldName,
					Tag:     tag,
					Value:   value,
					Message: fmt.Sprintf("%s must be a valid ObjectID", fieldName),
				}
			}
		}
	}

	return nil
}

func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", fe.Field(), fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "username":
		return fmt.Sprintf("%s must be a valid username (3-50 characters, alphanumeric and underscores)", fe.Field())
	case "objectid":
		return fmt.Sprintf("%s must be a valid ObjectID", fe.Field())
	case "privacy_level":
		return fmt.Sprintf("%s must be one of: public, friends, private", fe.Field())
	case "user_role":
		return fmt.Sprintf("%s must be one of: user, moderator, admin, super_admin", fe.Field())
	case "reaction_type":
		return fmt.Sprintf("%s must be one of: like, love, haha, wow, sad, angry, support", fe.Field())
	case "content_type":
		return fmt.Sprintf("%s must be one of: text, image, video, audio, file, link, gif, poll", fe.Field())
	case "notification_type":
		return fmt.Sprintf("%s must be a valid notification type", fe.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fe.Field(), fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}
