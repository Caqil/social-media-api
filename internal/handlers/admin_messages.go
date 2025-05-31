// internal/handlers/admin_messages.go - Fixed Messages & Conversations Handler
package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"social-media-api/internal/utils"
)

// GetAllMessages with simplified and fixed implementation
func (h *AdminHandler) GetAllMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	skip := (page - 1) * limit

	ctx := c.Request.Context()

	// Build basic match filter
	matchFilter := bson.M{
		"deleted_at": bson.M{"$exists": false},
	}

	// Add search filter
	if search := c.Query("search"); search != "" {
		matchFilter["$or"] = []bson.M{
			{"content": bson.M{"$regex": search, "$options": "i"}},
			{"file_name": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	// Add conversation_id filter with proper validation
	if conversationID := c.Query("conversation_id"); conversationID != "" && conversationID != "all" {
		if objID, err := primitive.ObjectIDFromHex(conversationID); err == nil {
			matchFilter["conversation_id"] = objID
		} else {
			utils.BadRequestResponse(c, "Invalid conversation ID format", nil)
			return
		}
	}

	// Add content_type filter
	if contentType := c.Query("content_type"); contentType != "" && contentType != "all" {
		matchFilter["content_type"] = contentType
	}

	// Add is_read filter
	if isRead := c.Query("is_read"); isRead != "" && isRead != "all" {
		matchFilter["is_read"] = isRead == "true"
	}

	// Add date range filters
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
			matchFilter["created_at"] = bson.M{"$gte": parsedDate}
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateTo); err == nil {
			if existingDateFilter, exists := matchFilter["created_at"]; exists {
				if dateFilter, ok := existingDateFilter.(bson.M); ok {
					dateFilter["$lte"] = parsedDate.Add(24 * time.Hour)
				}
			} else {
				matchFilter["created_at"] = bson.M{"$lte": parsedDate.Add(24 * time.Hour)}
			}
		}
	}

	// Simplified aggregation pipeline
	pipeline := []bson.M{
		{"$match": matchFilter},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "sender_id",
				"foreignField": "_id",
				"as":           "sender_data",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "conversations",
				"localField":   "conversation_id",
				"foreignField": "_id",
				"as":           "conversation_data",
			},
		},
		{
			"$addFields": bson.M{
				"id":              bson.M{"$toString": "$_id"},
				"conversation_id": bson.M{"$toString": "$conversation_id"},
				"sender_id":       bson.M{"$toString": "$sender_id"},
				"sender": bson.M{
					"$cond": bson.M{
						"if": bson.M{"$gt": []interface{}{bson.M{"$size": "$sender_data"}, 0}},
						"then": bson.M{
							"id":              bson.M{"$toString": bson.M{"$arrayElemAt": []interface{}{"$sender_data._id", 0}}},
							"username":        bson.M{"$arrayElemAt": []interface{}{"$sender_data.username", 0}},
							"email":           bson.M{"$arrayElemAt": []interface{}{"$sender_data.email", 0}},
							"first_name":      bson.M{"$arrayElemAt": []interface{}{"$sender_data.first_name", 0}},
							"last_name":       bson.M{"$arrayElemAt": []interface{}{"$sender_data.last_name", 0}},
							"profile_picture": bson.M{"$arrayElemAt": []interface{}{"$sender_data.profile_picture", 0}},
							"is_verified":     bson.M{"$arrayElemAt": []interface{}{"$sender_data.is_verified", 0}},
						},
						"else": nil,
					},
				},
				"conversation": bson.M{
					"$cond": bson.M{
						"if": bson.M{"$gt": []interface{}{bson.M{"$size": "$conversation_data"}, 0}},
						"then": bson.M{
							"id":    bson.M{"$toString": bson.M{"$arrayElemAt": []interface{}{"$conversation_data._id", 0}}},
							"type":  bson.M{"$arrayElemAt": []interface{}{"$conversation_data.type", 0}},
							"title": bson.M{"$arrayElemAt": []interface{}{"$conversation_data.title", 0}},
						},
						"else": nil,
					},
				},
			},
		},
		{
			"$project": bson.M{
				"_id":             0,
				"id":              1,
				"conversation_id": 1,
				"sender_id":       1,
				"content":         1,
				"content_type":    1,
				"media_url":       1,
				"file_name":       1,
				"file_size":       1,
				"is_read":         1,
				"read_at":         1,
				"is_edited":       1,
				"edited_at":       1,
				"created_at":      1,
				"updated_at":      1,
				"sender":          1,
				"conversation":    1,
			},
		},
		{"$sort": bson.M{"created_at": -1}},
		{"$skip": skip},
		{"$limit": limit},
	}

	cursor, err := h.db.Collection("messages").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get messages", err)
		return
	}
	defer cursor.Close(ctx)

	var messages []bson.M
	if err := cursor.All(ctx, &messages); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode messages", err)
		return
	}

	// Get total count
	totalCount, err := h.db.Collection("messages").CountDocuments(ctx, matchFilter)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to count messages", err)
		return
	}

	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       totalCount,
		TotalPages:  int((totalCount + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(page*limit) < totalCount,
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Messages retrieved successfully", messages, *pagination, links)
}

// GetAllConversations with simplified implementation
func (h *AdminHandler) GetAllConversations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	skip := (page - 1) * limit

	ctx := c.Request.Context()

	// Build match filter
	matchFilter := bson.M{
		"deleted_at": bson.M{"$exists": false},
	}

	// Add search filter
	if search := c.Query("search"); search != "" {
		matchFilter["$or"] = []bson.M{
			{"title": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	// Add type filter
	if convType := c.Query("type"); convType != "" && convType != "all" {
		matchFilter["type"] = convType
	}

	// Add archived filter
	if isArchived := c.Query("is_archived"); isArchived != "" && isArchived != "all" {
		matchFilter["is_archived"] = isArchived == "true"
	}

	// Add muted filter
	if isMuted := c.Query("is_muted"); isMuted != "" && isMuted != "all" {
		matchFilter["is_muted"] = isMuted == "true"
	}

	// Simplified pipeline
	pipeline := []bson.M{
		{"$match": matchFilter},
		{
			"$lookup": bson.M{
				"from": "users",
				"let":  bson.M{"participant_ids": "$participant_ids"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{"$in": []interface{}{"$_id", "$$participant_ids"}},
						},
					},
					{
						"$project": bson.M{
							"id":              bson.M{"$toString": "$_id"},
							"username":        1,
							"first_name":      1,
							"last_name":       1,
							"profile_picture": 1,
						},
					},
				},
				"as": "participants",
			},
		},
		{
			"$addFields": bson.M{
				"id": bson.M{"$toString": "$_id"},
				"participant_ids": bson.M{
					"$map": bson.M{
						"input": "$participant_ids",
						"as":    "pid",
						"in":    bson.M{"$toString": "$$pid"},
					},
				},
			},
		},
		{
			"$project": bson.M{
				"_id":             0,
				"id":              1,
				"type":            1,
				"title":           1,
				"participant_ids": 1,
				"last_message_at": 1,
				"is_archived":     1,
				"is_muted":        1,
				"unread_count":    1,
				"created_at":      1,
				"participants":    1,
			},
		},
		{"$sort": bson.M{"last_message_at": -1}},
		{"$skip": skip},
		{"$limit": limit},
	}

	cursor, err := h.db.Collection("conversations").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get conversations", err)
		return
	}
	defer cursor.Close(ctx)

	var conversations []bson.M
	if err := cursor.All(ctx, &conversations); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to decode conversations", err)
		return
	}

	// Get total count
	totalCount, err := h.db.Collection("conversations").CountDocuments(ctx, matchFilter)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to count conversations", err)
		return
	}

	pagination := &utils.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       totalCount,
		TotalPages:  int((totalCount + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(page*limit) < totalCount,
		HasPrevious: page > 1,
	}

	links := h.createPaginationLinks(c, pagination)
	utils.PaginatedSuccessResponse(c, "Conversations retrieved successfully", conversations, *pagination, links)
}

// Fixed GetMessage with proper error handling
func (h *AdminHandler) GetMessage(c *gin.Context) {
	messageID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid message ID", nil)
		return
	}

	ctx := c.Request.Context()

	// Simple find with lookup
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id":        objID,
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "sender_id",
				"foreignField": "_id",
				"as":           "sender",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "conversations",
				"localField":   "conversation_id",
				"foreignField": "_id",
				"as":           "conversation",
			},
		},
		{
			"$addFields": bson.M{
				"id":           bson.M{"$toString": "$_id"},
				"sender":       bson.M{"$arrayElemAt": []interface{}{"$sender", 0}},
				"conversation": bson.M{"$arrayElemAt": []interface{}{"$conversation", 0}},
			},
		},
	}

	cursor, err := h.db.Collection("messages").Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get message", err)
		return
	}
	defer cursor.Close(ctx)

	var message bson.M
	if cursor.Next(ctx) {
		if err := cursor.Decode(&message); err != nil {
			utils.InternalServerErrorResponse(c, "Failed to decode message", err)
			return
		}
	} else {
		utils.NotFoundResponse(c, "Message not found")
		return
	}

	utils.OkResponse(c, "Message retrieved successfully", message)
}

// Fixed DeleteMessage
func (h *AdminHandler) DeleteMessage(c *gin.Context) {
	messageID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid message ID", nil)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, use default reason
		req.Reason = "Deleted by admin"
	}

	ctx := c.Request.Context()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	result, err := h.db.Collection("messages").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete message", err)
		return
	}

	if result.MatchedCount == 0 {
		utils.NotFoundResponse(c, "Message not found")
		return
	}

	h.logAdminActivity(c, "message_deletion", "Deleted message ID: "+messageID+" Reason: "+req.Reason)
	utils.OkResponse(c, "Message deleted successfully", gin.H{
		"message_id": messageID,
		"reason":     req.Reason,
	})
}

// Fixed BulkMessageAction
func (h *AdminHandler) BulkMessageAction(c *gin.Context) {
	var req struct {
		MessageIDs []string `json:"message_ids" binding:"required"`
		Action     string   `json:"action" binding:"required"`
		Reason     string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if len(req.MessageIDs) > 100 {
		utils.BadRequestResponse(c, "Maximum 100 messages allowed per bulk operation", nil)
		return
	}

	ctx := c.Request.Context()
	successCount := 0
	failureCount := 0
	var errors []string

	for _, messageID := range req.MessageIDs {
		objID, err := primitive.ObjectIDFromHex(messageID)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Message %s: invalid ID", messageID))
			continue
		}

		var update bson.M
		switch req.Action {
		case "delete":
			update = bson.M{
				"$set": bson.M{
					"deleted_at": time.Now(),
					"updated_at": time.Now(),
				},
			}
		case "mark_read":
			update = bson.M{
				"$set": bson.M{
					"is_read":    true,
					"read_at":    time.Now(),
					"updated_at": time.Now(),
				},
			}
		case "mark_unread":
			update = bson.M{
				"$set": bson.M{
					"is_read":    false,
					"updated_at": time.Now(),
				},
				"$unset": bson.M{
					"read_at": "",
				},
			}
		default:
			failureCount++
			errors = append(errors, fmt.Sprintf("Message %s: invalid action", messageID))
			continue
		}

		_, err = h.db.Collection("messages").UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			failureCount++
			errors = append(errors, fmt.Sprintf("Message %s: %v", messageID, err))
		} else {
			successCount++
		}
	}

	h.logAdminActivity(c, "bulk_message_action", fmt.Sprintf("Bulk %s action: %d succeeded, %d failed", req.Action, successCount, failureCount))

	utils.OkResponse(c, "Bulk message action completed", gin.H{
		"action":        req.Action,
		"total":         len(req.MessageIDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"errors":        errors,
	})
}
