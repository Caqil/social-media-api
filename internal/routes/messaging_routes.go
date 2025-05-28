// internal/routes/messaging_routes.go
package routes

import (
	"social-media-api/internal/handlers"
	"social-media-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupMessagingRoutes sets up messaging and conversation routes
func SetupMessagingRoutes(router *gin.Engine, messageHandler *handlers.MessageHandler, conversationHandler *handlers.ConversationHandler, authMiddleware *middleware.AuthMiddleware) {
	// All messaging routes require authentication
	messaging := router.Group("/api/v1/messaging")
	messaging.Use(authMiddleware.RequireAuth())
	{
		// Conversation management
		conversations := messaging.Group("/conversations")
		{
			// Conversation CRUD
			conversations.POST("/", conversationHandler.CreateConversation)
			conversations.GET("/", conversationHandler.GetUserConversations)
			conversations.GET("/search", conversationHandler.SearchConversations)
			conversations.GET("/unread-counts", conversationHandler.GetUnreadCounts)
			conversations.GET("/:id", conversationHandler.GetConversation)
			conversations.PUT("/:id", conversationHandler.UpdateConversation)
			conversations.DELETE("/:id/leave", conversationHandler.LeaveConversation)

			// Conversation statistics
			conversations.GET("/:id/stats", conversationHandler.GetConversationStats)

			// Participant management
			conversations.POST("/:id/participants", conversationHandler.AddParticipants)
			conversations.DELETE("/:id/participants/:participantId", conversationHandler.RemoveParticipant)
			conversations.PUT("/:id/participants/:participantId/role", conversationHandler.UpdateParticipantRole)

			// Conversation settings
			conversations.PUT("/:id/mute", conversationHandler.MuteConversation)
			conversations.PUT("/:id/archive", conversationHandler.ArchiveConversation)

			// Messages within conversations - RESTRUCTURED to avoid conflicts
			conversations.GET("/:id/messages", conversationHandler.GetConversationMessages)
			conversations.POST("/:id/messages", middleware.MessageRateLimit(), conversationHandler.SendMessage)
			conversations.POST("/:id/mark-read", conversationHandler.MarkAsRead)
		}

		// Individual message management - FIXED: Removed conflicting routes
		messages := messaging.Group("/messages")
		{
			// Message CRUD operations on individual messages
			messages.GET("/:id", messageHandler.GetMessages)            // Get single message
			messages.PUT("/:id", messageHandler.UpdateMessage)         // Update single message
			messages.DELETE("/:id", messageHandler.DeleteMessage)      // Delete single message
			messages.POST("/:id/react", messageHandler.ReactToMessage) // React to single message

			// Global message operations (not conversation-specific)
			messages.GET("/search", messageHandler.SearchMessages) // Search across all messages
			messages.GET("/stats", messageHandler.GetMessageStats) // User's message statistics
		}

		// REMOVED CONFLICTING ROUTES:
		// ❌ messages.POST("/:conversationId/mark-read", messageHandler.MarkMessagesAsRead)
		// ❌ messages.GET("/:conversationId", messageHandler.GetMessages)
		// ❌ messages.POST("/", middleware.MessageRateLimit(), messageHandler.SendMessage)
		//
		// These are now handled under /conversations/:id/ endpoints above

		// Legacy conversation routes (for backward compatibility) - UPDATED
		messaging.POST("/", messageHandler.CreateConversation)
		messaging.GET("/", messageHandler.GetConversations)
		messaging.GET("/:id", messageHandler.GetConversation)
		messaging.PUT("/:id", messageHandler.UpdateConversation)
		messaging.DELETE("/:id/leave", messageHandler.LeaveConversation)
		messaging.POST("/:id/participants", messageHandler.AddParticipants)
		messaging.DELETE("/:id/participants/:participantId", messageHandler.RemoveParticipant)
	}
}
