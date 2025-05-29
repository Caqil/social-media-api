// internal/middleware/admin_middleware.go
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"social-media-api/internal/models"
	"social-media-api/internal/utils"
)

// AdminMiddleware handles admin authentication and authorization
type AdminMiddleware struct {
	db             *mongo.Database
	authMiddleware *AuthMiddleware
}

// NewAdminMiddleware creates a new admin middleware instance
func NewAdminMiddleware(db *mongo.Database, authMiddleware *AuthMiddleware) *AdminMiddleware {
	return &AdminMiddleware{
		db:             db,
		authMiddleware: authMiddleware,
	}
}

// RequireAdmin middleware that requires admin role or higher
func (am *AdminMiddleware) RequireAdmin() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First ensure user is authenticated
		am.authMiddleware.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// Check if user has admin role
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusForbidden, "Role information not found", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		if role != models.RoleAdmin && role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Admin access required", nil)
			c.Abort()
			return
		}

		// Log admin access
		am.logAdminAccess(c, "admin_access")

		c.Next()
	})
}

// RequireSuperAdmin middleware that requires super admin role
func (am *AdminMiddleware) RequireSuperAdmin() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First ensure user is authenticated
		am.authMiddleware.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// Check if user has super admin role
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusForbidden, "Role information not found", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		if role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Super admin access required", nil)
			c.Abort()
			return
		}

		// Log super admin access
		am.logAdminAccess(c, "super_admin_access")

		c.Next()
	})
}

// RequireModerator middleware that requires moderator role or higher
func (am *AdminMiddleware) RequireModerator() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First ensure user is authenticated
		am.authMiddleware.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// Check if user has moderator, admin, or super admin role
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusForbidden, "Role information not found", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		if role != models.RoleModerator && role != models.RoleAdmin && role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Moderator access required", nil)
			c.Abort()
			return
		}

		// Log moderator access
		am.logAdminAccess(c, "moderator_access")

		c.Next()
	})
}

// RequireStaff middleware that requires any staff role (moderator, admin, or super admin)
func (am *AdminMiddleware) RequireStaff() gin.HandlerFunc {
	return am.RequireModerator()
}

// AdminOnly middleware for admin-only routes with additional checks
func (am *AdminMiddleware) AdminOnly() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First ensure user is authenticated
		am.authMiddleware.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		user, exists := c.Get("user")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		u := user.(*models.User)

		// Check if user account is suspended
		if u.IsSuspended {
			utils.ErrorResponse(c, http.StatusForbidden, "Account is suspended", nil)
			c.Abort()
			return
		}

		// Check if user account is active
		if !u.IsActive {
			utils.ErrorResponse(c, http.StatusForbidden, "Account is inactive", nil)
			c.Abort()
			return
		}

		// Check if user has admin or super admin role
		if u.Role != models.RoleAdmin && u.Role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Admin access required", nil)
			c.Abort()
			return
		}

		// Check if user email is verified (additional security for admin)
		if !u.EmailVerified {
			utils.ErrorResponse(c, http.StatusForbidden, "Email verification required for admin access", nil)
			c.Abort()
			return
		}

		// Add admin-specific context
		c.Set("is_admin", true)
		c.Set("can_access_admin_panel", true)
		c.Set("admin_permissions", am.getAdminPermissions(u.Role))

		// Log admin panel access
		am.logAdminAccess(c, "admin_panel_access")

		c.Next()
	})
}

// CheckResourceOwnership middleware for checking resource ownership with admin override
func (am *AdminMiddleware) CheckResourceOwnership() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		currentUser, exists := c.Get("user")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		user := currentUser.(*models.User)

		// Get resource owner ID from URL parameter
		resourceOwnerID := c.Param("user_id")
		if resourceOwnerID == "" {
			resourceOwnerID = c.Param("id")
		}

		// If no resource owner specified, continue
		if resourceOwnerID == "" {
			c.Next()
			return
		}

		// Check if current user is the resource owner
		isOwner := user.ID.Hex() == resourceOwnerID

		// Check if current user is admin/moderator
		isStaff := user.Role == models.RoleAdmin ||
			user.Role == models.RoleSuperAdmin ||
			user.Role == models.RoleModerator

		// Set context variables
		c.Set("is_resource_owner", isOwner)
		c.Set("can_modify_resource", isOwner || isStaff)
		c.Set("has_admin_override", isStaff)

		c.Next()
	})
}

// RequireResourceOwnership middleware that requires user to be owner or staff
func (am *AdminMiddleware) RequireResourceOwnership() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		am.CheckResourceOwnership()(c)
		if c.IsAborted() {
			return
		}

		canModify, exists := c.Get("can_modify_resource")
		if !exists || !canModify.(bool) {
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// AdminIPWhitelist middleware for IP-based admin access control
func (am *AdminMiddleware) AdminIPWhitelist(allowedIPs []string) gin.HandlerFunc {
	ipMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		ipMap[ip] = true
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Skip IP check if no whitelist is configured
		if len(allowedIPs) == 0 {
			c.Next()
			return
		}

		// Check if IP is whitelisted
		if !ipMap[clientIP] {
			// Log unauthorized access attempt
			am.logSecurityEvent(c, "ADMIN_IP_BLOCKED", "Access denied from non-whitelisted IP: "+clientIP)

			utils.ErrorResponse(c, http.StatusForbidden, "Access denied from this IP", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// AdminSessionTimeout middleware for admin session management
func (am *AdminMiddleware) AdminSessionTimeout(timeout time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.Next()
			return
		}

		u := user.(*models.User)

		// Check if last activity is within timeout period
		if time.Since(*u.LastActiveAt) > timeout {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Admin session expired", nil)
			c.Abort()
			return
		}

		// Update last active time
		go am.updateLastActive(u.ID)

		// Set session timeout header
		c.Header("X-Admin-Session-Timeout", timeout.String())
		c.Header("X-Admin-Session", "active")

		c.Next()
	})
}

// AdminTwoFactorAuth middleware that requires 2FA for admin operations
func (am *AdminMiddleware) AdminTwoFactorAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		u := user.(*models.User)

		// Check if 2FA is enabled and verified for this session
		if u.TwoFactorEnabled {
			sessionVerified, _ := c.Get("2fa_verified")
			if sessionVerified == nil || !sessionVerified.(bool) {
				utils.ErrorResponse(c, http.StatusForbidden, "Two-factor authentication required", nil)
				c.Abort()
				return
			}
		}

		c.Next()
	})
}

// AdminRateLimit creates a rate limiter specifically for admin operations
func (am *AdminMiddleware) AdminRateLimit(rate int, window time.Duration) gin.HandlerFunc {
	return RateLimit(RateLimitConfig{
		Rate:   rate,
		Window: window,
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
		Skip: func(c *gin.Context) bool {
			// Skip rate limiting for super admins in emergency situations
			if userRole, exists := c.Get("user_role"); exists {
				role := userRole.(models.UserRole)
				emergencyMode := c.GetHeader("X-Emergency-Mode")
				return role == models.RoleSuperAdmin && emergencyMode == "true"
			}
			return false
		},
	})
}

// AdminAuditLog middleware that logs all admin actions
func (am *AdminMiddleware) AdminAuditLog() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		startTime := time.Now()

		// Capture request details
		requestBody := am.captureRequestBody(c)

		c.Next()

		// Log after request completion
		go am.logAuditEvent(c, startTime, requestBody)
	})
}

// AdminPermissionCheck checks specific permissions for fine-grained access control
func (am *AdminMiddleware) AdminPermissionCheck(requiredPermissions ...string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusForbidden, "Role information not found", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		userPermissions := am.getAdminPermissions(role)

		// Check if user has all required permissions
		for _, permission := range requiredPermissions {
			if !am.hasPermission(userPermissions, permission) {
				utils.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions: "+permission, nil)
				c.Abort()
				return
			}
		}

		c.Next()
	})
}

// AdminMaintenanceMode middleware that blocks admin access during maintenance
func (am *AdminMiddleware) AdminMaintenanceMode() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check if maintenance mode is enabled
		isMaintenanceMode := am.isMaintenanceModeEnabled()

		if isMaintenanceMode {
			// Allow super admins to access during maintenance
			userRole, exists := c.Get("user_role")
			if exists && userRole.(models.UserRole) == models.RoleSuperAdmin {
				c.Header("X-Maintenance-Mode", "true")
				c.Header("X-Admin-Override", "true")
				c.Next()
				return
			}

			utils.ErrorResponse(c, http.StatusServiceUnavailable, "System is under maintenance", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// AdminSecurityHeaders middleware that adds security headers for admin panel
func (am *AdminMiddleware) AdminSecurityHeaders() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Add security headers specific to admin panel
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-inline'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-Admin-Panel", "true")

		c.Next()
	})
}

// AdminCORS middleware with restricted CORS policy for admin
func (am *AdminMiddleware) AdminCORS(allowedOrigins []string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed || len(allowedOrigins) == 0 {
			if len(allowedOrigins) == 0 {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
			}
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Admin-Token")
			c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// AdminCSRFProtection middleware for CSRF protection in admin panel
func (am *AdminMiddleware) AdminCSRFProtection() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Check CSRF token
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			csrfToken = c.PostForm("_csrf_token")
		}

		if csrfToken == "" {
			utils.ErrorResponse(c, http.StatusForbidden, "CSRF token required", nil)
			c.Abort()
			return
		}

		// Validate CSRF token (implement actual validation)
		if !am.validateCSRFToken(csrfToken, c) {
			am.logSecurityEvent(c, "CSRF_TOKEN_INVALID", "Invalid CSRF token provided")
			utils.ErrorResponse(c, http.StatusForbidden, "Invalid CSRF token", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// AdminActivityTracker tracks admin activities for compliance
func (am *AdminMiddleware) AdminActivityTracker() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		// Track activity after request completion
		go am.trackAdminActivity(c, startTime)
	})
}

// AdminBruteForceProtection protects against brute force attacks
func (am *AdminMiddleware) AdminBruteForceProtection() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Check if IP is currently blocked
		if am.isIPBlocked(clientIP) {
			am.logSecurityEvent(c, "BRUTE_FORCE_BLOCKED", "Blocked IP attempted access: "+clientIP)
			utils.ErrorResponse(c, http.StatusTooManyRequests, "Too many failed attempts", nil)
			c.Abort()
			return
		}

		c.Next()

		// Check for failed authentication
		if c.Writer.Status() == 401 || c.Writer.Status() == 403 {
			go am.recordFailedAttempt(clientIP)
		}
	})
}

// AdminDeviceTrust middleware for device-based security
func (am *AdminMiddleware) AdminDeviceTrust() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		deviceFingerprint := c.GetHeader("X-Device-Fingerprint")
		userAgent := c.GetHeader("User-Agent")

		user, exists := c.Get("user")
		if !exists {
			c.Next()
			return
		}

		u := user.(*models.User)

		// Check if device is trusted
		if !am.isDeviceTrusted(u.ID, deviceFingerprint, userAgent) {
			// Log suspicious device access
			am.logSecurityEvent(c, "UNTRUSTED_DEVICE", "Admin access from untrusted device")

			// Could require additional verification or block access
			c.Header("X-Device-Trust", "false")
			c.Header("X-Additional-Verification-Required", "true")
		} else {
			c.Header("X-Device-Trust", "true")
		}

		c.Next()
	})
}

// AdminGeoRestriction middleware for geographical access control
func (am *AdminMiddleware) AdminGeoRestriction(allowedCountries []string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if len(allowedCountries) == 0 {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		country := am.getCountryFromIP(clientIP) // Implement IP geolocation

		allowed := false
		for _, allowedCountry := range allowedCountries {
			if country == allowedCountry {
				allowed = true
				break
			}
		}

		if !allowed {
			am.logSecurityEvent(c, "GEO_RESTRICTION", "Admin access denied from country: "+country)
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied from this location", nil)
			c.Abort()
			return
		}

		c.Header("X-Admin-Country", country)
		c.Next()
	})
}

// AdminTimeRestriction middleware for time-based access control
func (am *AdminMiddleware) AdminTimeRestriction(allowedHours []int) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if len(allowedHours) == 0 {
			c.Next()
			return
		}

		currentHour := time.Now().Hour()

		allowed := false
		for _, allowedHour := range allowedHours {
			if currentHour == allowedHour {
				allowed = true
				break
			}
		}

		if !allowed {
			// Allow super admins during emergency
			userRole, exists := c.Get("user_role")
			if exists && userRole.(models.UserRole) == models.RoleSuperAdmin {
				emergencyMode := c.GetHeader("X-Emergency-Mode")
				if emergencyMode == "true" {
					c.Header("X-Emergency-Access", "true")
					c.Next()
					return
				}
			}

			am.logSecurityEvent(c, "TIME_RESTRICTION", "Admin access denied outside allowed hours")
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied outside allowed hours", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// Helper methods

func (am *AdminMiddleware) logAdminAccess(c *gin.Context, accessType string) {
	userID, _ := c.Get("user_id")

	logEntry := bson.M{
		"type":       accessType,
		"user_id":    userID,
		"ip_address": c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"timestamp":  time.Now(),
		"created_at": time.Now(),
	}

	go am.db.Collection("admin_access_logs").InsertOne(context.Background(), logEntry)
}

func (am *AdminMiddleware) logSecurityEvent(c *gin.Context, eventType, description string) {
	userID, _ := c.Get("user_id")

	event := bson.M{
		"event_type":  eventType,
		"description": description,
		"user_id":     userID,
		"ip_address":  c.ClientIP(),
		"user_agent":  c.GetHeader("User-Agent"),
		"path":        c.Request.URL.Path,
		"method":      c.Request.Method,
		"timestamp":   time.Now(),
		"created_at":  time.Now(),
	}

	go am.db.Collection("security_events").InsertOne(context.Background(), event)
}

func (am *AdminMiddleware) logAuditEvent(c *gin.Context, startTime time.Time, requestBody string) {
	userID, _ := c.Get("user_id")

	auditLog := bson.M{
		"user_id":         userID,
		"action":          c.Request.Method + " " + c.Request.URL.Path,
		"request_body":    requestBody,
		"response_status": c.Writer.Status(),
		"ip_address":      c.ClientIP(),
		"user_agent":      c.GetHeader("User-Agent"),
		"duration_ms":     time.Since(startTime).Milliseconds(),
		"timestamp":       time.Now(),
		"created_at":      time.Now(),
	}

	go am.db.Collection("admin_audit_logs").InsertOne(context.Background(), auditLog)
}

func (am *AdminMiddleware) trackAdminActivity(c *gin.Context, startTime time.Time) {
	userID, _ := c.Get("user_id")

	activity := bson.M{
		"user_id":         userID,
		"action":          c.Request.Method + " " + c.Request.URL.Path,
		"response_status": c.Writer.Status(),
		"ip_address":      c.ClientIP(),
		"user_agent":      c.GetHeader("User-Agent"),
		"duration_ms":     time.Since(startTime).Milliseconds(),
		"timestamp":       time.Now(),
		"created_at":      time.Now(),
	}

	go am.db.Collection("admin_activities").InsertOne(context.Background(), activity)
}

func (am *AdminMiddleware) updateLastActive(userID primitive.ObjectID) {
	update := bson.M{
		"$set": bson.M{
			"last_active_at": time.Now(),
			"updated_at":     time.Now(),
		},
	}

	am.db.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		update,
	)
}

func (am *AdminMiddleware) getAdminPermissions(role models.UserRole) []string {
	switch role {
	case models.RoleSuperAdmin:
		return []string{
			"users.read", "users.write", "users.delete",
			"posts.read", "posts.write", "posts.delete",
			"comments.read", "comments.write", "comments.delete",
			"reports.read", "reports.write", "reports.resolve",
			"system.read", "system.write", "system.maintenance",
			"config.read", "config.write",
			"analytics.read", "analytics.export",
			"admin.create", "admin.delete",
		}
	case models.RoleAdmin:
		return []string{
			"users.read", "users.write",
			"posts.read", "posts.write", "posts.delete",
			"comments.read", "comments.write", "comments.delete",
			"reports.read", "reports.write", "reports.resolve",
			"system.read",
			"config.read",
			"analytics.read",
		}
	case models.RoleModerator:
		return []string{
			"users.read",
			"posts.read", "posts.write",
			"comments.read", "comments.write",
			"reports.read", "reports.write",
		}
	default:
		return []string{}
	}
}

func (am *AdminMiddleware) hasPermission(userPermissions []string, requiredPermission string) bool {
	for _, permission := range userPermissions {
		if permission == requiredPermission {
			return true
		}
	}
	return false
}

func (am *AdminMiddleware) captureRequestBody(c *gin.Context) string {
	// This would capture and sanitize request body for audit logging
	// Be careful with sensitive data
	return ""
}

func (am *AdminMiddleware) isMaintenanceModeEnabled() bool {
	// Check if maintenance mode is enabled (from config/database)
	var config bson.M
	err := am.db.Collection("system_config").FindOne(
		context.Background(),
		bson.M{"key": "maintenance_mode"},
	).Decode(&config)

	if err != nil {
		return false
	}

	enabled, ok := config["value"].(bool)
	return ok && enabled
}

func (am *AdminMiddleware) validateCSRFToken(token string, c *gin.Context) bool {
	// Implement CSRF token validation
	// This is a simplified version - use a proper CSRF implementation
	sessionToken, exists := c.Get("csrf_token")
	if !exists {
		return false
	}
	return token == sessionToken.(string)
}

func (am *AdminMiddleware) isIPBlocked(ip string) bool {
	// Check if IP is in blocked list
	var blockedIP bson.M
	err := am.db.Collection("blocked_ips").FindOne(
		context.Background(),
		bson.M{
			"ip":         ip,
			"expires_at": bson.M{"$gt": time.Now()},
		},
	).Decode(&blockedIP)

	return err == nil
}

func (am *AdminMiddleware) recordFailedAttempt(ip string) {
	// Record failed attempt and block IP if necessary
	failedAttempt := bson.M{
		"ip":         ip,
		"timestamp":  time.Now(),
		"created_at": time.Now(),
	}

	am.db.Collection("failed_attempts").InsertOne(context.Background(), failedAttempt)

	// Check if IP should be blocked (e.g., 5 failed attempts in 1 hour)
	count, _ := am.db.Collection("failed_attempts").CountDocuments(
		context.Background(),
		bson.M{
			"ip":        ip,
			"timestamp": bson.M{"$gte": time.Now().Add(-1 * time.Hour)},
		},
	)

	if count >= 5 {
		// Block IP for 1 hour
		blockedIP := bson.M{
			"ip":         ip,
			"reason":     "brute_force",
			"expires_at": time.Now().Add(1 * time.Hour),
			"created_at": time.Now(),
		}

		am.db.Collection("blocked_ips").InsertOne(context.Background(), blockedIP)
	}
}

func (am *AdminMiddleware) isDeviceTrusted(userID primitive.ObjectID, fingerprint, userAgent string) bool {
	// Check if device is in trusted devices list
	var trustedDevice bson.M
	err := am.db.Collection("trusted_devices").FindOne(
		context.Background(),
		bson.M{
			"user_id":     userID,
			"fingerprint": fingerprint,
			"user_agent":  userAgent,
			"is_active":   true,
		},
	).Decode(&trustedDevice)

	return err == nil
}

func (am *AdminMiddleware) getCountryFromIP(ip string) string {
	// Implement IP geolocation
	// This would use a service like MaxMind GeoIP2 or similar
	// For now, return a default
	return "US"
}

// WebSocketUpgradeMiddleware handles WebSocket upgrades for admin panel
func (am *AdminMiddleware) WebSocketUpgradeMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Verify WebSocket upgrade request
		if c.GetHeader("Upgrade") != "websocket" {
			utils.ErrorResponse(c, http.StatusBadRequest, "WebSocket upgrade required", nil)
			c.Abort()
			return
		}

		// Additional WebSocket security checks for admin
		origin := c.GetHeader("Origin")
		if origin == "" {
			utils.ErrorResponse(c, http.StatusForbidden, "Origin header required", nil)
			c.Abort()
			return
		}

		// Validate origin (should be from admin panel)
		// This would check against allowed admin panel origins

		c.Next()
	})
}

// AdminMetrics middleware for collecting admin panel metrics
func (am *AdminMiddleware) AdminMetrics() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		// Collect metrics
		go am.collectMetrics(c, startTime)
	})
}

func (am *AdminMiddleware) collectMetrics(c *gin.Context, startTime time.Time) {
	userID, _ := c.Get("user_id")

	metric := bson.M{
		"user_id":         userID,
		"endpoint":        c.Request.URL.Path,
		"method":          c.Request.Method,
		"response_status": c.Writer.Status(),
		"response_time":   time.Since(startTime).Milliseconds(),
		"timestamp":       time.Now(),
		"created_at":      time.Now(),
	}

	am.db.Collection("admin_metrics").InsertOne(context.Background(), metric)
}

// AdminErrorHandler handles admin-specific errors
func (am *AdminMiddleware) AdminErrorHandler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Handle admin-specific errors
		if len(c.Errors) > 0 {
			// Log admin errors
			for _, err := range c.Errors {
				am.logAdminError(c, err.Error())
			}
		}
	})
}

func (am *AdminMiddleware) logAdminError(c *gin.Context, errorMsg string) {
	userID, _ := c.Get("user_id")

	errorLog := bson.M{
		"user_id":    userID,
		"error":      errorMsg,
		"endpoint":   c.Request.URL.Path,
		"method":     c.Request.Method,
		"ip_address": c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
		"timestamp":  time.Now(),
		"created_at": time.Now(),
	}

	go am.db.Collection("admin_errors").InsertOne(context.Background(), errorLog)
}

// Convenience functions for commonly used middleware combinations

// StandardAdminMiddleware returns a standard set of admin middleware
func (am *AdminMiddleware) StandardAdminMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		am.AdminSecurityHeaders(),
		am.AdminBruteForceProtection(),
		am.authMiddleware.RequireAuth(),
		am.RequireAdmin(),
		am.AdminAuditLog(),
		am.AdminActivityTracker(),
		am.AdminMetrics(),
		am.AdminErrorHandler(),
	}
}

// HighSecurityAdminMiddleware returns admin middleware with enhanced security
func (am *AdminMiddleware) HighSecurityAdminMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		am.AdminSecurityHeaders(),
		am.AdminBruteForceProtection(),
		am.AdminCSRFProtection(),
		am.AdminDeviceTrust(),
		am.authMiddleware.RequireAuth(),
		am.RequireAdmin(),
		am.AdminTwoFactorAuth(),
		am.AdminSessionTimeout(30 * time.Minute),
		am.AdminAuditLog(),
		am.AdminActivityTracker(),
		am.AdminMetrics(),
		am.AdminErrorHandler(),
	}
}

// SuperAdminMiddleware returns middleware for super admin operations
func (am *AdminMiddleware) SuperAdminMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		am.AdminSecurityHeaders(),
		am.AdminBruteForceProtection(),
		am.AdminCSRFProtection(),
		am.AdminDeviceTrust(),
		am.authMiddleware.RequireAuth(),
		am.RequireSuperAdmin(),
		am.AdminTwoFactorAuth(),
		am.AdminSessionTimeout(15 * time.Minute), // Shorter timeout for super admin
		am.AdminAuditLog(),
		am.AdminActivityTracker(),
		am.AdminMetrics(),
		am.AdminErrorHandler(),
	}
}
