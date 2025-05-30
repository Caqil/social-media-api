// internal/middleware/role.go
package middleware

import (
	"net/http"

	"social-media-api/internal/models"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// RequireAdmin checks if user has admin or super admin role
func RequireAdmin() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		if role != models.RoleAdmin && role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Admin access required", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireSuperAdmin checks if user has super admin role
func RequireSuperAdmin() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		if role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Super admin access required", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireModerator checks if user has moderator role or higher
func RequireModerator() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		if role != models.RoleModerator && role != models.RoleAdmin && role != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "Moderator access required", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireStaff checks if user has any staff role (moderator, admin, or super admin)
func RequireStaff() gin.HandlerFunc {
	return RequireModerator()
}

// RequireRole checks if user has any of the specified roles
func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		currentRole := userRole.(models.UserRole)

		// Check if user has any of the required roles
		for _, role := range roles {
			if currentRole == role {
				c.Next()
				return
			}
		}

		utils.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", nil)
		c.Abort()
	})
}

// IsAdmin checks if current user is admin or super admin (helper function)
func IsAdmin(c *gin.Context) bool {
	return HasAnyRole(c, models.RoleAdmin, models.RoleSuperAdmin)
}

// IsModerator checks if current user is moderator or higher (helper function)
func IsModerator(c *gin.Context) bool {
	return HasAnyRole(c, models.RoleModerator, models.RoleAdmin, models.RoleSuperAdmin)
}
