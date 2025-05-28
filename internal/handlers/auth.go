// internal/handlers/auth.go
package handlers

import (
	"strings"

	"social-media-api/internal/models"
	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthHandler struct {
	authService *services.AuthService
	userService *services.UserService
	validator   *validator.Validate
}

func NewAuthHandler(authService *services.AuthService, userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		validator:   validator.New(),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Additional validation
	if req.Password != req.ConfirmPassword {
		utils.BadRequestResponse(c, "Passwords do not match", nil)
		return
	}

	// Validate username format
	if !utils.IsValidUsername(req.Username) {
		utils.BadRequestResponse(c, "Invalid username format", nil)
		return
	}

	// Register user
	response, err := h.authService.Register(req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			utils.ConflictResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to register user", err)
		return
	}

	utils.CreatedResponse(c, "User registered successfully", response)
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Set device info and IP address
	req.DeviceInfo = c.GetHeader("User-Agent")
	if req.DeviceInfo == "" {
		req.DeviceInfo = "Unknown Device"
	}

	response, err := h.authService.Login(req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid credentials") {
			utils.UnauthorizedResponse(c, "Invalid email/username or password")
			return
		}
		if strings.Contains(err.Error(), "suspended") {
			utils.ForbiddenResponse(c, "Account is suspended")
			return
		}
		utils.InternalServerErrorResponse(c, "Login failed", err)
		return
	}

	utils.LoginSuccessResponse(c, response.User, gin.H{
		"access_token":  response.AccessToken,
		"refresh_token": response.RefreshToken,
		"expires_in":    response.ExpiresIn,
		"token_type":    response.TokenType,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	response, err := h.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.UnauthorizedResponse(c, "Invalid or expired refresh token")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to refresh token", err)
		return
	}

	utils.OkResponse(c, "Tokens refreshed successfully", response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get session ID from context (set by auth middleware)
	sessionID, exists := c.Get("session_id")
	if !exists {
		utils.UnauthorizedResponse(c, "No active session")
		return
	}

	err := h.authService.Logout(sessionID.(string))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to logout", err)
		return
	}

	utils.LogoutSuccessResponse(c)
}

// LogoutAll handles logout from all devices
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	err := h.authService.LogoutAll(userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to logout from all devices", err)
		return
	}

	utils.OkResponse(c, "Logged out from all devices successfully", nil)
}

// ForgotPassword handles password reset request
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := h.authService.ForgotPassword(req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to process password reset", err)
		return
	}

	utils.PasswordResetEmailSentResponse(c)
}

// ResetPassword handles password reset with token
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := h.authService.ResetPassword(req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		if strings.Contains(err.Error(), "do not match") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to reset password", err)
		return
	}

	utils.OkResponse(c, "Password reset successfully", nil)
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.BadRequestResponse(c, "Verification token is required", nil)
		return
	}

	err := h.authService.VerifyEmail(token)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			utils.BadRequestResponse(c, "Invalid verification token", err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to verify email", err)
		return
	}

	utils.EmailVerificationSuccessResponse(c)
}

// GetProfile returns current user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	user, err := h.userService.GetUserByID(userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user profile", err)
		return
	}

	utils.OkResponse(c, "Profile retrieved successfully", user.ToUserResponse())
}

// UpdateProfile updates user profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	user, err := h.userService.UpdateUser(userID.(primitive.ObjectID), req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update profile", err)
		return
	}

	utils.ProfileUpdateSuccessResponse(c, user.ToUserResponse())
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := h.userService.ChangePassword(userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "incorrect") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		if strings.Contains(err.Error(), "do not match") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to change password", err)
		return
	}

	utils.PasswordChangeSuccessResponse(c)
}

// GetSessions returns user's active sessions
func (h *AuthHandler) GetSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	sessions, err := h.authService.GetUserSessions(userID.(primitive.ObjectID))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get sessions", err)
		return
	}

	utils.OkResponse(c, "Sessions retrieved successfully", sessions)
}

// RevokeSession revokes a specific session
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		utils.BadRequestResponse(c, "Session ID is required", nil)
		return
	}

	err := h.authService.RevokeSession(sessionID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to revoke session", err)
		return
	}

	utils.OkResponse(c, "Session revoked successfully", nil)
}
