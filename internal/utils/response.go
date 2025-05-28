// utils/response.go
package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, message string, error string) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   error,
	})
}

// ValidationErrorResponse sends validation error response
func ValidationErrorResponse(c *gin.Context, err error) {
	var errors []string
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, err.Field()+" validation failed: "+err.Tag())
	}

	c.JSON(400, gin.H{
		"success": false,
		"message": "Validation failed",
		"errors":  errors,
	})
}
