// internal/routes/group_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupGroupRoutes sets up group-related routes
func SetupGroupRoutes(router *gin.Engine, groupHandler *handlers.GroupHandler, authMiddleware *middleware.AuthMiddleware) {
	// Public group routes
	groups := router.Group("/api/v1/groups")
	{
		// Group discovery (public/optional auth)
		groups.GET("/", authMiddleware.OptionalAuth(), groupHandler.GetPublicGroups)
		groups.GET("/search", authMiddleware.OptionalAuth(), groupHandler.SearchGroups)
		groups.GET("/trending", authMiddleware.OptionalAuth(), groupHandler.GetTrendingGroups)
		groups.GET("/categories", groupHandler.GetGroupCategories)
		groups.GET("/:id", authMiddleware.OptionalAuth(), groupHandler.GetGroup)
		groups.GET("/:id/members", authMiddleware.OptionalAuth(), groupHandler.GetGroupMembers)
	}

	// Protected group routes
	groupsProtected := router.Group("/api/v1/groups")
	groupsProtected.Use(authMiddleware.RequireAuth())
	{
		// Group creation and management
		groupsProtected.POST("/", groupHandler.CreateGroup)
		groupsProtected.PUT("/:id", groupHandler.UpdateGroup)
		groupsProtected.DELETE("/:id", groupHandler.DeleteGroup)

		// Group membership
		groupsProtected.POST("/:id/join", groupHandler.JoinGroup)
		groupsProtected.POST("/:id/leave", groupHandler.LeaveGroup)
		groupsProtected.POST("/:id/invite", groupHandler.InviteToGroup)

		// Member management (admin/moderator only)
		groupsProtected.PUT("/:id/members/:member_id/role", groupHandler.UpdateMemberRole)
		groupsProtected.DELETE("/:id/members/:member_id", groupHandler.RemoveGroupMember)
		groupsProtected.POST("/:id/members/bulk-remove", groupHandler.BulkRemoveMembers)

		// Group statistics (admin/moderator only)
		groupsProtected.GET("/:id/stats", groupHandler.GetGroupStats)

		// User-specific group endpoints
		groupsProtected.GET("/my-groups", groupHandler.GetUserGroups)
		groupsProtected.GET("/invites", groupHandler.GetUserGroupInvites)
	}

	// Group invitations
	groupInvites := router.Group("/api/v1/group-invites")
	groupInvites.Use(authMiddleware.RequireAuth())
	{
		groupInvites.POST("/:invite_id/accept", groupHandler.AcceptGroupInvite)
		groupInvites.POST("/:invite_id/reject", groupHandler.RejectGroupInvite)
	}
}
