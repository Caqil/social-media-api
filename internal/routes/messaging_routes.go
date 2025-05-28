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
			conversations.DELETE("/:id/participants/:participant_id", conversationHandler.RemoveParticipant)
			conversations.PUT("/:id/participants/:participant_id/role", conversationHandler.UpdateParticipantRole)

			// Conversation settings
			conversations.PUT("/:id/mute", conversationHandler.MuteConversation)
			conversations.PUT("/:id/archive", conversationHandler.ArchiveConversation)

			// Messages within conversations
			conversations.GET("/:id/messages", conversationHandler.GetConversationMessages)
			conversations.POST("/:id/messages", middleware.MessageRateLimit(), conversationHandler.SendMessage)
			conversations.POST("/:id/mark-read", conversationHandler.MarkAsRead)
		}

		// Message management
		messages := messaging.Group("/messages")
		{
			// Message CRUD
			messages.POST("/", middleware.MessageRateLimit(), messageHandler.SendMessage)
			messages.GET("/:conversationId", messageHandler.GetMessages)
			messages.PUT("/:id", messageHandler.UpdateMessage)
			messages.DELETE("/:id", messageHandler.DeleteMessage)

			// Message interactions
			messages.POST("/:id/react", messageHandler.ReactToMessage)
			messages.POST("/:conversationId/mark-read", messageHandler.MarkMessagesAsRead)

			// Message search and statistics
			messages.GET("/search", messageHandler.SearchMessages)
			messages.GET("/stats", messageHandler.GetMessageStats)
		}

		// Legacy conversation routes (for backward compatibility)
		messaging.POST("/", messageHandler.CreateConversation)
		messaging.GET("/", messageHandler.GetConversations)
		messaging.GET("/:id", messageHandler.GetConversation)
		messaging.PUT("/:id", messageHandler.UpdateConversation)
		messaging.DELETE("/:id/leave", messageHandler.LeaveConversation)
		messaging.POST("/:id/participants", messageHandler.AddParticipants)
		messaging.DELETE("/:id/participants/:participantId", messageHandler.RemoveParticipant)
	}
}
