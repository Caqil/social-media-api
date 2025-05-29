// internal/middleware/admin.go
package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"social-media-api/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminAuthService interface for admin authentication
type AdminAuthService interface {
	ValidateAdminToken(token string) (*models.User, error)
	GetUserRole(userID primitive.ObjectID) (models.UserRole, error)
	LogAdminActivity(userID primitive.ObjectID, action, description string, metadata map[string]interface{}, c *gin.Context)
}

// AdminMiddleware struct holds the admin authentication service
type AdminMiddleware struct {
	authService AdminAuthService
}

// NewAdminMiddleware creates a new admin middleware instance
func NewAdminMiddleware(authService AdminAuthService) *AdminMiddleware {
	return &AdminMiddleware{
		authService: authService,
	}
}

// AuthRequired middleware ensures user is authenticated
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authentication token required",
				"code":    "AUTH_TOKEN_REQUIRED",
			})
			c.Abort()
			return
		}

		// This would typically validate the JWT token
		// For now, we'll simulate token validation
		user, err := validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid or expired token",
				"code":    "AUTH_TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Set("username", user.Username)
		c.Set("is_verified", user.IsVerified)

		c.Next()
	}
}

// AdminRequired middleware ensures user has admin privileges
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Admin access required",
				"code":    "ACCESS_DENIED",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(models.UserRole)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Invalid role information",
				"code":    "INVALID_ROLE",
			})
			c.Abort()
			return
		}

		if !isAdminRole(role) {
			// Log unauthorized admin access attempt
			logUnauthorizedAccess(c, "admin", role)

			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Forbidden",
				"message":       "Admin privileges required",
				"code":          "INSUFFICIENT_PRIVILEGES",
				"required_role": "admin",
				"current_role":  role,
			})
			c.Abort()
			return
		}

		// Log admin access
		logAdminAccess(c, "admin_access")

		c.Next()
	}
}

// ModeratorRequired middleware ensures user has moderator or higher privileges
func ModeratorRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Moderator access required",
				"code":    "ACCESS_DENIED",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(models.UserRole)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Invalid role information",
				"code":    "INVALID_ROLE",
			})
			c.Abort()
			return
		}

		if !isModeratorOrHigher(role) {
			// Log unauthorized moderator access attempt
			logUnauthorizedAccess(c, "moderator", role)

			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Forbidden",
				"message":       "Moderator privileges required",
				"code":          "INSUFFICIENT_PRIVILEGES",
				"required_role": "moderator",
				"current_role":  role,
			})
			c.Abort()
			return
		}

		// Log moderator access
		logAdminAccess(c, "moderator_access")

		c.Next()
	}
}

// SuperAdminRequired middleware ensures user has super admin privileges
func SuperAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Super admin access required",
				"code":    "ACCESS_DENIED",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(models.UserRole)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Invalid role information",
				"code":    "INVALID_ROLE",
			})
			c.Abort()
			return
		}

		if role != models.RoleSuperAdmin {
			// Log unauthorized super admin access attempt
			logUnauthorizedAccess(c, "super_admin", role)

			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Forbidden",
				"message":       "Super admin privileges required",
				"code":          "INSUFFICIENT_PRIVILEGES",
				"required_role": "super_admin",
				"current_role":  role,
			})
			c.Abort()
			return
		}

		// Log super admin access
		logAdminAccess(c, "super_admin_access")

		c.Next()
	}
}

// AdminActivityLogger middleware logs all admin activities
func AdminActivityLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for health checks and non-destructive operations
		if shouldSkipLogging(c) {
			c.Next()
			return
		}

		// Capture request details
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		userID := getUserIDFromContext(c)

		// Continue to next handler
		c.Next()

		// Log after request completion
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		activity := map[string]interface{}{
			"method":      method,
			"path":        path,
			"status_code": statusCode,
			"duration_ms": duration.Milliseconds(),
			"ip_address":  c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"request_id":  c.GetString("request_id"),
		}

		// Add query parameters for GET requests
		if method == "GET" && len(c.Request.URL.RawQuery) > 0 {
			activity["query_params"] = c.Request.URL.RawQuery
		}

		// Add error information if request failed
		if statusCode >= 400 {
			if errorMsg, exists := c.Get("error_message"); exists {
				activity["error"] = errorMsg
			}
		}

		// Log the activity
		logActivity(userID, "admin_api_access", generateDescription(method, path, statusCode), activity)
	}
}

// RateLimitAdmin middleware for admin endpoint rate limiting
func RateLimitAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserIDFromContext(c)
		if userID == "" {
			c.Next()
			return
		}

		// Implement rate limiting logic based on user role
		userRole, exists := c.Get("user_role")
		if !exists {
			c.Next()
			return
		}

		role := userRole.(models.UserRole)
		limit := getAdminRateLimit(role)

		// Check rate limit
		if isRateLimited(userID, limit) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate Limit Exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d per minute", limit),
				"code":        "RATE_LIMIT_EXCEEDED",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeaders middleware adds security headers for admin endpoints
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Remove server information
		c.Header("Server", "")

		c.Next()
	}
}

// AuditTrail middleware creates audit logs for sensitive operations
func AuditTrail() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only audit destructive operations
		if !isDestructiveOperation(c) {
			c.Next()
			return
		}

		userID := getUserIDFromContext(c)
		action := determineAction(c)

		// Capture request body for audit
		var requestBody map[string]interface{}
		if c.Request.Method != "GET" && c.Request.Method != "DELETE" {
			if err := c.ShouldBindJSON(&requestBody); err == nil {
				// Request body captured for audit
			}
		}

		// Continue to handler
		c.Next()

		// Create audit log after request completion
		statusCode := c.Writer.Status()
		success := statusCode >= 200 && statusCode < 300

		auditData := map[string]interface{}{
			"action":       action,
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"status_code":  statusCode,
			"success":      success,
			"ip_address":   c.ClientIP(),
			"user_agent":   c.Request.UserAgent(),
			"request_body": requestBody,
		}

		// Add response data for successful operations
		if success {
			if responseData, exists := c.Get("response_data"); exists {
				auditData["response_data"] = responseData
			}
		}

		// Log audit trail
		logAuditTrail(userID, action, auditData, c)
	}
}

// IPWhitelist middleware for admin endpoints (optional)
func IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(allowedIPs) == 0 {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		if !isIPAllowed(clientIP, allowedIPs) {
			// Log IP access attempt
			logSecurityIncident("ip_whitelist_violation", map[string]interface{}{
				"client_ip":  clientIP,
				"path":       c.Request.URL.Path,
				"user_agent": c.Request.UserAgent(),
			})

			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Access denied from this IP address",
				"code":    "IP_NOT_ALLOWED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TimeBasedAccess middleware restricts admin access to certain hours
func TimeBasedAccess(allowedHours []int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(allowedHours) == 0 {
			c.Next()
			return
		}

		currentHour := time.Now().Hour()
		if !contains(allowedHours, currentHour) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Forbidden",
				"message":       "Admin access not allowed at this time",
				"code":          "TIME_RESTRICTED_ACCESS",
				"allowed_hours": allowedHours,
				"current_hour":  currentHour,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ==================== HELPER FUNCTIONS ====================

// extractToken extracts the bearer token from request
func extractToken(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}

	// Also check for token in query parameter (for websockets, etc.)
	if token := c.Query("token"); token != "" {
		return token
	}

	return ""
}

// validateToken validates the JWT token and returns user information
func validateToken(token string) (*models.User, error) {
	// This is a placeholder - implement actual JWT validation
	// For now, simulate a valid user
	user := &models.User{
		ID:         primitive.NewObjectID(),
		Username:   "admin_user",
		Role:       models.RoleAdmin,
		IsVerified: true,
	}
	return user, nil
}

// isAdminRole checks if the role has admin privileges
func isAdminRole(role models.UserRole) bool {
	return role == models.RoleAdmin || role == models.RoleSuperAdmin
}

// isModeratorOrHigher checks if the role has moderator or higher privileges
func isModeratorOrHigher(role models.UserRole) bool {
	return role == models.RoleModerator || role == models.RoleAdmin || role == models.RoleSuperAdmin
}

// getUserIDFromContext extracts user ID from gin context
func getUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(primitive.ObjectID); ok {
			return id.Hex()
		}
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// shouldSkipLogging determines if the request should be skipped from logging
func shouldSkipLogging(c *gin.Context) bool {
	skipPaths := []string{
		"/api/v1/admin/system/health",
		"/api/v1/admin/dashboard/real-time",
		"/api/v1/admin/monitor",
	}

	path := c.Request.URL.Path
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// generateDescription creates a description for the activity log
func generateDescription(method, path string, statusCode int) string {
	action := "accessed"
	if method == "POST" {
		action = "created"
	} else if method == "PUT" {
		action = "updated"
	} else if method == "DELETE" {
		action = "deleted"
	}

	status := "successfully"
	if statusCode >= 400 {
		status = "unsuccessfully"
	}

	return fmt.Sprintf("Admin %s %s %s", action, path, status)
}

// getAdminRateLimit returns rate limit based on user role
func getAdminRateLimit(role models.UserRole) int {
	switch role {
	case models.RoleSuperAdmin:
		return 1000 // Higher limit for super admins
	case models.RoleAdmin:
		return 500 // High limit for admins
	case models.RoleModerator:
		return 200 // Moderate limit for moderators
	default:
		return 60 // Default limit
	}
}

// isRateLimited checks if user has exceeded rate limit
func isRateLimited(userID string, limit int) bool {
	// This is a placeholder - implement actual rate limiting
	// Using Redis or in-memory cache with sliding window or token bucket
	return false
}

// isDestructiveOperation checks if the operation is destructive
func isDestructiveOperation(c *gin.Context) bool {
	destructiveMethods := []string{"POST", "PUT", "DELETE", "PATCH"}
	method := c.Request.Method

	for _, destructiveMethod := range destructiveMethods {
		if method == destructiveMethod {
			return true
		}
	}

	// Also check for destructive GET operations (like bulk exports)
	destructivePaths := []string{
		"/export",
		"/delete",
		"/suspend",
		"/ban",
		"/bulk-action",
	}

	path := c.Request.URL.Path
	for _, destructivePath := range destructivePaths {
		if strings.Contains(path, destructivePath) {
			return true
		}
	}

	return false
}

// determineAction determines the action being performed
func determineAction(c *gin.Context) string {
	path := c.Request.URL.Path
	method := c.Request.Method

	// Extract action from path
	if strings.Contains(path, "suspend") {
		return "user_suspend"
	} else if strings.Contains(path, "delete") {
		return "content_delete"
	} else if strings.Contains(path, "bulk-action") {
		return "bulk_operation"
	} else if strings.Contains(path, "config") {
		return "config_update"
	} else if strings.Contains(path, "export") {
		return "data_export"
	}

	// Default action based on method
	switch method {
	case "POST":
		return "create"
	case "PUT":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "access"
	}
}

// isIPAllowed checks if IP is in the whitelist
func isIPAllowed(clientIP string, allowedIPs []string) bool {
	for _, allowedIP := range allowedIPs {
		if clientIP == allowedIP {
			return true
		}
		// Could also implement CIDR matching here
	}
	return false
}

// contains checks if slice contains a value
func contains(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// ==================== LOGGING FUNCTIONS ====================

// logUnauthorizedAccess logs unauthorized access attempts
func logUnauthorizedAccess(c *gin.Context, requiredRole string, currentRole models.UserRole) {
	userID := getUserIDFromContext(c)

	logData := map[string]interface{}{
		"required_role": requiredRole,
		"current_role":  currentRole,
		"path":          c.Request.URL.Path,
		"method":        c.Request.Method,
		"ip_address":    c.ClientIP(),
		"user_agent":    c.Request.UserAgent(),
	}

	logSecurityIncident("unauthorized_admin_access", logData)
	logActivity(userID, "unauthorized_access", fmt.Sprintf("Attempted to access %s requiring %s role", c.Request.URL.Path, requiredRole), logData)
}

// logAdminAccess logs successful admin access
func logAdminAccess(c *gin.Context, accessType string) {
	userID := getUserIDFromContext(c)

	logData := map[string]interface{}{
		"access_type": accessType,
		"path":        c.Request.URL.Path,
		"method":      c.Request.Method,
		"ip_address":  c.ClientIP(),
		"user_agent":  c.Request.UserAgent(),
	}

	logActivity(userID, accessType, fmt.Sprintf("Accessed admin endpoint: %s", c.Request.URL.Path), logData)
}

// logActivity logs general activity
func logActivity(userID, action, description string, metadata map[string]interface{}) {
	// This would integrate with your logging service
	// For now, it's a placeholder
	fmt.Printf("[ACTIVITY] User: %s, Action: %s, Description: %s\n", userID, action, description)
}

// logAuditTrail logs audit trail information
func logAuditTrail(userID, action string, auditData map[string]interface{}, c *gin.Context) {
	// This would integrate with your audit logging service
	// For now, it's a placeholder
	fmt.Printf("[AUDIT] User: %s, Action: %s, Data: %+v\n", userID, action, auditData)
}

// logSecurityIncident logs security-related incidents
func logSecurityIncident(incidentType string, data map[string]interface{}) {
	// This would integrate with your security monitoring service
	// For now, it's a placeholder
	fmt.Printf("[SECURITY] Incident: %s, Data: %+v\n", incidentType, data)
}

// ==================== MIDDLEWARE CHAIN HELPERS ====================

// AdminMiddlewareChain returns a complete middleware chain for admin endpoints
func AdminMiddlewareChain() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		SecurityHeaders(),
		RateLimitAdmin(),
		AuthRequired(),
		AdminRequired(),
		AdminActivityLogger(),
		AuditTrail(),
	}
}

// ModeratorMiddlewareChain returns a complete middleware chain for moderator endpoints
func ModeratorMiddlewareChain() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		SecurityHeaders(),
		RateLimitAdmin(),
		AuthRequired(),
		ModeratorRequired(),
		AdminActivityLogger(),
		AuditTrail(),
	}
}

// SuperAdminMiddlewareChain returns a complete middleware chain for super admin endpoints
func SuperAdminMiddlewareChain() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		SecurityHeaders(),
		AuthRequired(),
		SuperAdminRequired(),
		AdminActivityLogger(),
		AuditTrail(),
	}
}

// WithIPWhitelist adds IP whitelist middleware to the chain
func WithIPWhitelist(chain []gin.HandlerFunc, allowedIPs []string) []gin.HandlerFunc {
	return append([]gin.HandlerFunc{IPWhitelist(allowedIPs)}, chain...)
}

// WithTimeRestriction adds time-based access restriction to the chain
func WithTimeRestriction(chain []gin.HandlerFunc, allowedHours []int) []gin.HandlerFunc {
	return append([]gin.HandlerFunc{TimeBasedAccess(allowedHours)}, chain...)
}
