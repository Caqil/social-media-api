// middleware/admin_auth.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"social-media-api/internal/models"
	"social-media-api/internal/utils"
)

// RequireAdmin middleware that requires admin role or higher
func RequireAdmin() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First check if user is authenticated
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)

		// Check if user has admin or super admin role
		if role != models.RoleAdmin && role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Admin access required", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireSuperAdmin middleware that requires super admin role
func RequireSuperAdmin() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First check if user is authenticated
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)

		// Check if user has super admin role
		if role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Super admin access required", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireModerator middleware that requires moderator role or higher
func RequireModerator() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First check if user is authenticated
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)

		// Check if user has moderator, admin, or super admin role
		if role != models.RoleModerator && role != models.RoleAdmin && role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Moderator access required", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireStaff middleware that requires any staff role (moderator, admin, or super admin)
func RequireStaff() gin.HandlerFunc {
	return RequireModerator()
}

// AdminOnly middleware for admin-only routes
func AdminOnly() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
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

		// Check if user has admin or super admin role
		if u.Role != models.RoleAdmin && u.Role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Admin access required", nil)
			c.Abort()
			return
		}

		// Add admin-specific context
		c.Set("is_admin", true)
		c.Set("can_access_admin_panel", true)

		c.Next()
	})
}

// CheckResourceOwnership middleware for checking resource ownership
func CheckResourceOwnership() gin.HandlerFunc {
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

		c.Next()
	})
}

// RequireResourceOwnership middleware that requires user to be owner or staff
func RequireResourceOwnership() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		CheckResourceOwnership()(c)
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
func AdminIPWhitelist(allowedIPs []string) gin.HandlerFunc {
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
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied from this IP", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// AdminSessionTimeout middleware for admin session management
func AdminSessionTimeout() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// This could implement session timeout logic specific to admin users
		// For now, we'll just add a header indicating admin session
		c.Header("X-Admin-Session", "active")
		c.Next()
	})
}