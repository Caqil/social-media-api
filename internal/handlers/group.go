// internal/handlers/group.go
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

type GroupHandler struct {
	groupService *services.GroupService
	validator    *validator.Validate
}

func NewGroupHandler(groupService *services.GroupService) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		validator:    validator.New(),
	}
}

// CreateGroup creates a new group
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate name length
	if len(req.Name) < 3 || len(req.Name) > utils.MaxGroupNameLength {
		utils.BadRequestResponse(c, "Group name must be between 3 and 100 characters", nil)
		return
	}

	// Validate description length if provided
	if len(req.Description) > 2000 {
		utils.BadRequestResponse(c, "Description is too long (max 2000 characters)", nil)
		return
	}

	group, err := h.groupService.CreateGroup(userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			utils.ConflictResponse(c, err.Error(), err)
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to create group", err)
		return
	}

	utils.CreatedResponse(c, "Group created successfully", group.ToGroupResponse())
}

// GetGroup retrieves a group by ID or slug
func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupIdentifier := c.Param("id") // Can be ID or slug
	if groupIdentifier == "" {
		utils.BadRequestResponse(c, "Group identifier is required", nil)
		return
	}

	var currentUserID primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		currentUserID = userID.(primitive.ObjectID)
	}

	var group *models.Group
	var err error

	// Try to parse as ObjectID first
	if groupID, parseErr := primitive.ObjectIDFromHex(groupIdentifier); parseErr == nil {
		group, err = h.groupService.GetGroupByID(groupID, currentUserID)
	} else {
		// Treat as slug
		group, err = h.groupService.GetGroupBySlug(groupIdentifier, currentUserID)
	}

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "Group not found")
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to retrieve group", err)
		return
	}

	groupResponse := group.ToGroupResponse()

	// Add user context if authenticated
	if !currentUserID.IsZero() {
		memberStatus, role := h.groupService.GetMemberStatus(group.ID, currentUserID)
		groupResponse.UserStatus = memberStatus
		groupResponse.UserRole = role
		groupResponse.IsMember = memberStatus == "member"
		groupResponse.IsAdmin = role == models.GroupRoleAdmin || role == models.GroupRoleOwner
		groupResponse.IsModerator = role == models.GroupRoleModerator
		groupResponse.CanPost = group.CanPostInGroup(role, memberStatus)
		groupResponse.CanInvite = group.CanInviteToGroup(role)
		groupResponse.CanModerate = group.CanModerateGroup(role)
	}

	utils.OkResponse(c, "Group retrieved successfully", groupResponse)
}

// UpdateGroup updates a group
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", err)
		return
	}

	var req models.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate name if provided
	if req.Name != nil && (len(*req.Name) < 3 || len(*req.Name) > utils.MaxGroupNameLength) {
		utils.BadRequestResponse(c, "Group name must be between 3 and 100 characters", nil)
		return
	}

	// Validate description if provided
	if req.Description != nil && len(*req.Description) > 2000 {
		utils.BadRequestResponse(c, "Description is too long (max 2000 characters)", nil)
		return
	}

	group, err := h.groupService.UpdateGroup(groupID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "access denied") || strings.Contains(err.Error(), "privileges required") {
			utils.ForbiddenResponse(c, err.Error())
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			utils.ConflictResponse(c, err.Error(), err)
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update group", err)
		return
	}

	utils.OkResponse(c, "Group updated successfully", group.ToGroupResponse())
}

// DeleteGroup soft deletes a group
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", err)
		return
	}

	err = h.groupService.DeleteGroup(groupID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "access denied") || strings.Contains(err.Error(), "privileges required") {
			utils.ForbiddenResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to delete group", err)
		return
	}

	utils.OkResponse(c, "Group deleted successfully", nil)
}

// JoinGroup allows a user to join a group
func (h *GroupHandler) JoinGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", err)
		return
	}

	var req models.JoinGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body for simple join requests
		req = models.JoinGroupRequest{}
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err = h.groupService.JoinGroup(groupID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "cannot join") || strings.Contains(err.Error(), "already") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to join group", err)
		return
	}

	utils.OkResponse(c, "Join request processed successfully", gin.H{
		"status":  "success",
		"message": "You have successfully joined the group or your request is pending approval",
	})
}

// LeaveGroup allows a user to leave a group
func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", err)
		return
	}

	err = h.groupService.LeaveGroup(groupID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not a member") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		if strings.Contains(err.Error(), "cannot leave") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to leave group", err)
		return
	}

	utils.OkResponse(c, "Left group successfully", nil)
}

// InviteToGroup invites users to a group
func (h *GroupHandler) InviteToGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", err)
		return
	}

	var req models.InviteToGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate message length
	if len(req.Message) > 500 {
		utils.BadRequestResponse(c, "Invitation message is too long (max 500 characters)", nil)
		return
	}

	err = h.groupService.InviteToGroup(groupID, userID.(primitive.ObjectID), req)
	if err != nil {
		if strings.Contains(err.Error(), "access denied") || strings.Contains(err.Error(), "insufficient permissions") {
			utils.ForbiddenResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to send invitations", err)
		return
	}

	utils.OkResponse(c, "Invitations sent successfully", gin.H{
		"invited_count": len(req.UserIDs),
	})
}

// AcceptGroupInvite accepts a group invitation
func (h *GroupHandler) AcceptGroupInvite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	inviteID, err := primitive.ObjectIDFromHex(c.Param("invite_id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid invitation ID", err)
		return
	}

	err = h.groupService.AcceptGroupInvite(inviteID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "expired") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to accept invitation", err)
		return
	}

	utils.OkResponse(c, "Invitation accepted successfully", nil)
}

// RejectGroupInvite rejects a group invitation
func (h *GroupHandler) RejectGroupInvite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	inviteID, err := primitive.ObjectIDFromHex(c.Param("invite_id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid invitation ID", err)
		return
	}

	err = h.groupService.RejectGroupInvite(inviteID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to reject invitation", err)
		return
	}

	utils.OkResponse(c, "Invitation rejected successfully", nil)
}

// GetGroupMembers retrieves group members
func (h *GroupHandler) GetGroupMembers(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", err)
		return
	}

	var currentUserID primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		currentUserID = userID.(primitive.ObjectID)
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	members, err := h.groupService.GetGroupMembers(groupID, currentUserID, params.Limit, params.Offset)
	if err != nil {
		if strings.Contains(err.Error(), "access denied") {
			utils.ForbiddenResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get group members", err)
		return
	}

	totalCount := int64(len(members))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Group members retrieved successfully", members, paginationMeta, nil)
}

// GetUserGroups retrieves groups that the user is a member of
func (h *GroupHandler) GetUserGroups(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	groups, err := h.groupService.GetUserGroups(userID.(primitive.ObjectID), params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get user groups", err)
		return
	}

	totalCount := int64(len(groups))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "User groups retrieved successfully", groups, paginationMeta, nil)
}

// SearchGroups searches for groups
func (h *GroupHandler) SearchGroups(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	if len(query) < utils.MinSearchLength {
		utils.BadRequestResponse(c, "Search query is too short", nil)
		return
	}

	if len(query) > utils.MaxSearchLength {
		utils.BadRequestResponse(c, "Search query is too long", nil)
		return
	}

	var currentUserID *primitive.ObjectID
	if userID, exists := c.Get("user_id"); exists {
		id := userID.(primitive.ObjectID)
		currentUserID = &id
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	groups, err := h.groupService.SearchGroups(query, currentUserID, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search groups", err)
		return
	}

	totalCount := int64(len(groups))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Groups found", groups, paginationMeta, nil)
}

// UpdateMemberRole updates a member's role in the group
func (h *GroupHandler) UpdateMemberRole(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", err)
		return
	}

	memberID, err := primitive.ObjectIDFromHex(c.Param("member_id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid member ID", err)
		return
	}

	var req models.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err = h.groupService.UpdateMemberRole(groupID, userID.(primitive.ObjectID), memberID, req)
	if err != nil {
		if strings.Contains(err.Error(), "access denied") || strings.Contains(err.Error(), "insufficient permissions") {
			utils.ForbiddenResponse(c, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to update member role", err)
		return
	}

	utils.OkResponse(c, "Member role updated successfully", gin.H{
		"member_id": memberID.Hex(),
		"new_role":  req.Role,
	})
}

// RemoveGroupMember removes a member from the group
func (h *GroupHandler) RemoveGroupMember(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", err)
		return
	}

	memberID, err := primitive.ObjectIDFromHex(c.Param("member_id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid member ID", err)
		return
	}

	err = h.groupService.RemoveGroupMember(groupID, userID.(primitive.ObjectID), memberID)
	if err != nil {
		if strings.Contains(err.Error(), "access denied") || strings.Contains(err.Error(), "insufficient permissions") {
			utils.ForbiddenResponse(c, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to remove member", err)
		return
	}

	utils.OkResponse(c, "Member removed successfully", nil)
}

// GetGroupStats retrieves group statistics (admin only)
func (h *GroupHandler) GetGroupStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", err)
		return
	}

	stats, err := h.groupService.GetGroupStats(groupID, userID.(primitive.ObjectID))
	if err != nil {
		if strings.Contains(err.Error(), "access denied") || strings.Contains(err.Error(), "privileges required") {
			utils.ForbiddenResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get group statistics", err)
		return
	}

	utils.OkResponse(c, "Group statistics retrieved successfully", stats)
}

// GetPublicGroups retrieves public groups for discovery
func (h *GroupHandler) GetPublicGroups(c *gin.Context) {
	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	groups, err := h.groupService.GetPublicGroups(params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get public groups", err)
		return
	}

	totalCount := int64(len(groups))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Public groups retrieved successfully", groups, paginationMeta, nil)
}

// GetTrendingGroups retrieves trending groups
func (h *GroupHandler) GetTrendingGroups(c *gin.Context) {
	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get time range parameter
	timeRange := c.DefaultQuery("time_range", "day") // day, week, month

	groups, err := h.groupService.GetTrendingGroups(params.Limit, params.Offset, timeRange)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get trending groups", err)
		return
	}

	totalCount := int64(len(groups))
	paginationMeta := utils.CreatePaginationMeta(params, totalCount)

	utils.PaginatedSuccessResponse(c, "Trending groups retrieved successfully", groups, paginationMeta, nil)
}

// GetGroupCategories retrieves available group categories
func (h *GroupHandler) GetGroupCategories(c *gin.Context) {
	categories := models.GetGroupCategories()
	utils.OkResponse(c, "Group categories retrieved successfully", gin.H{
		"categories": categories,
	})
}

// GetUserGroupInvites retrieves group invitations for the current user
func (h *GroupHandler) GetUserGroupInvites(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// This would require implementing in the service
	// For now, return empty array
	utils.PaginatedSuccessResponse(c, "Group invitations retrieved successfully", []interface{}{}, utils.CreatePaginationMeta(params, 0), nil)
}

// Bulk operations

// BulkRemoveMembers removes multiple members from a group
func (h *GroupHandler) BulkRemoveMembers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", err)
		return
	}

	var req struct {
		MemberIDs []string `json:"member_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	if len(req.MemberIDs) == 0 {
		utils.BadRequestResponse(c, "At least one member ID is required", nil)
		return
	}

	if len(req.MemberIDs) > 50 { // Reasonable limit
		utils.BadRequestResponse(c, "Too many members to remove at once (max 50)", nil)
		return
	}

	successCount := 0
	for _, memberIDStr := range req.MemberIDs {
		memberID, err := primitive.ObjectIDFromHex(memberIDStr)
		if err != nil {
			continue // Skip invalid IDs
		}

		err = h.groupService.RemoveGroupMember(groupID, userID.(primitive.ObjectID), memberID)
		if err == nil {
			successCount++
		}
	}

	utils.OkResponse(c, "Bulk member removal completed", gin.H{
		"total_requested": len(req.MemberIDs),
		"success_count":   successCount,
		"failed_count":    len(req.MemberIDs) - successCount,
	})
}

// Helper methods for common validations

// validateGroupID validates and parses group ID parameter
func (h *GroupHandler) validateGroupID(c *gin.Context) (primitive.ObjectID, error) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid group ID", err)
		return primitive.NilObjectID, err
	}
	return groupID, nil
}

// validateMemberID validates and parses member ID parameter
func (h *GroupHandler) validateMemberID(c *gin.Context) (primitive.ObjectID, error) {
	memberID, err := primitive.ObjectIDFromHex(c.Param("member_id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid member ID", err)
		return primitive.NilObjectID, err
	}
	return memberID, nil
}

// getCurrentUserID gets the current user ID from context
func (h *GroupHandler) getCurrentUserID(c *gin.Context) (primitive.ObjectID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return primitive.NilObjectID, false
	}
	return userID.(primitive.ObjectID), true
}
