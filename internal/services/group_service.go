// internal/services/group.go
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"social-media-api/internal/models"
	"social-media-api/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GroupService struct {
	db                  *mongo.Database
	groupsColl          *mongo.Collection
	membersColl         *mongo.Collection
	invitesColl         *mongo.Collection
	usersColl           *mongo.Collection
	postsColl           *mongo.Collection
	notificationService *NotificationService
}

func NewGroupService(db *mongo.Database, notificationService *NotificationService) *GroupService {
	return &GroupService{
		db:                  db,
		groupsColl:          db.Collection("groups"),
		membersColl:         db.Collection("group_members"),
		invitesColl:         db.Collection("group_invites"),
		usersColl:           db.Collection("users"),
		postsColl:           db.Collection("posts"),
		notificationService: notificationService,
	}
}

// CreateGroup creates a new group
func (s *GroupService) CreateGroup(creatorID primitive.ObjectID, req models.CreateGroupRequest) (*models.Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate category
	if !models.IsValidGroupCategory(req.Category) {
		return nil, errors.New("invalid group category")
	}

	// Check if group name already exists (case-insensitive)
	existingCount, err := s.groupsColl.CountDocuments(ctx, bson.M{
		"name":       bson.M{"$regex": "^" + req.Name + "$", "$options": "i"},
		"deleted_at": bson.M{"$exists": false},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check group name uniqueness: %w", err)
	}
	if existingCount > 0 {
		return nil, errors.New("group name already exists")
	}

	// Create group
	group := models.Group{
		Name:                   req.Name,
		Description:            req.Description,
		Privacy:                req.Privacy,
		Category:               req.Category,
		Tags:                   req.Tags,
		Location:               req.Location,
		Website:                req.Website,
		Rules:                  req.Rules,
		CreatedBy:              creatorID,
		PostApprovalRequired:   req.PostApprovalRequired,
		MemberApprovalRequired: req.MemberApprovalRequired,
		AllowMemberInvites:     req.AllowMemberInvites,
		AllowExternalSharing:   req.AllowExternalSharing,
		AllowPolls:             req.AllowPolls,
		AllowEvents:            req.AllowEvents,
		AllowDiscussions:       req.AllowDiscussions,
	}

	group.BeforeCreate()

	// Insert group
	result, err := s.groupsColl.InsertOne(ctx, group)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	group.ID = result.InsertedID.(primitive.ObjectID)

	// Add creator as owner/admin
	member := models.GroupMember{
		GroupID: group.ID,
		UserID:  creatorID,
		Role:    models.GroupRoleOwner,
		Status:  "active",
	}
	member.BeforeCreate()

	_, err = s.membersColl.InsertOne(ctx, member)
	if err != nil {
		// Rollback group creation
		s.groupsColl.DeleteOne(ctx, bson.M{"_id": group.ID})
		return nil, fmt.Errorf("failed to add creator as owner: %w", err)
	}

	return &group, nil
}

// GetGroupByID retrieves a group by ID
func (s *GroupService) GetGroupByID(groupID, currentUserID primitive.ObjectID) (*models.Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get group
	var group models.Group
	err := s.groupsColl.FindOne(ctx, bson.M{
		"_id":        groupID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&group)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("group not found")
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Check if user can view group
	memberStatus, _ := s.GetMemberStatus(groupID, currentUserID)
	if !group.CanViewGroup(currentUserID, memberStatus) {
		return nil, errors.New("access denied")
	}

	// Populate creator info
	var creator models.User
	err = s.usersColl.FindOne(ctx, bson.M{"_id": group.CreatedBy}).Decode(&creator)
	if err == nil {
		group.Creator = creator.ToUserResponse()
	}

	return &group, nil
}

// GetGroupBySlug retrieves a group by slug
func (s *GroupService) GetGroupBySlug(slug string, currentUserID primitive.ObjectID) (*models.Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var group models.Group
	err := s.groupsColl.FindOne(ctx, bson.M{
		"slug":       slug,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&group)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("group not found")
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Check if user can view group
	memberStatus, _ := s.GetMemberStatus(group.ID, currentUserID)
	if !group.CanViewGroup(currentUserID, memberStatus) {
		return nil, errors.New("access denied")
	}

	return &group, nil
}

// UpdateGroup updates a group
func (s *GroupService) UpdateGroup(groupID, userID primitive.ObjectID, req models.UpdateGroupRequest) (*models.Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user is admin/owner
	member, err := s.GetGroupMember(groupID, userID)
	if err != nil {
		return nil, errors.New("access denied")
	}

	if member.Role != models.GroupRoleAdmin && member.Role != models.GroupRoleOwner {
		return nil, errors.New("admin privileges required")
	}

	// Build update document
	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	if req.Name != nil {
		// Check name uniqueness
		existingCount, err := s.groupsColl.CountDocuments(ctx, bson.M{
			"name":       bson.M{"$regex": "^" + *req.Name + "$", "$options": "i"},
			"_id":        bson.M{"$ne": groupID},
			"deleted_at": bson.M{"$exists": false},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
		}
		if existingCount > 0 {
			return nil, errors.New("group name already exists")
		}
		update["$set"].(bson.M)["name"] = *req.Name
		update["$set"].(bson.M)["slug"] = utils.GenerateSlug(*req.Name)
	}

	if req.Description != nil {
		update["$set"].(bson.M)["description"] = *req.Description
	}

	if req.Privacy != nil {
		update["$set"].(bson.M)["privacy"] = *req.Privacy
	}

	if req.Category != nil {
		if !models.IsValidGroupCategory(*req.Category) {
			return nil, errors.New("invalid group category")
		}
		update["$set"].(bson.M)["category"] = *req.Category
	}

	if req.Tags != nil {
		update["$set"].(bson.M)["tags"] = req.Tags
	}

	if req.Location != nil {
		update["$set"].(bson.M)["location"] = req.Location
	}

	if req.Website != nil {
		update["$set"].(bson.M)["website"] = *req.Website
	}

	if req.Color != nil {
		update["$set"].(bson.M)["color"] = *req.Color
	}

	if req.Rules != nil {
		update["$set"].(bson.M)["rules"] = req.Rules
	}

	if req.PostApprovalRequired != nil {
		update["$set"].(bson.M)["post_approval_required"] = *req.PostApprovalRequired
	}

	if req.MemberApprovalRequired != nil {
		update["$set"].(bson.M)["member_approval_required"] = *req.MemberApprovalRequired
	}

	if req.AllowMemberInvites != nil {
		update["$set"].(bson.M)["allow_member_invites"] = *req.AllowMemberInvites
	}

	if req.AllowExternalSharing != nil {
		update["$set"].(bson.M)["allow_external_sharing"] = *req.AllowExternalSharing
	}

	if req.AllowPolls != nil {
		update["$set"].(bson.M)["allow_polls"] = *req.AllowPolls
	}

	if req.AllowEvents != nil {
		update["$set"].(bson.M)["allow_events"] = *req.AllowEvents
	}

	if req.AllowDiscussions != nil {
		update["$set"].(bson.M)["allow_discussions"] = *req.AllowDiscussions
	}

	// Update group
	_, err = s.groupsColl.UpdateOne(ctx, bson.M{"_id": groupID}, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	// Return updated group
	return s.GetGroupByID(groupID, userID)
}

// DeleteGroup soft deletes a group
func (s *GroupService) DeleteGroup(groupID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user is owner
	member, err := s.GetGroupMember(groupID, userID)
	if err != nil {
		return errors.New("access denied")
	}

	if member.Role != models.GroupRoleOwner {
		return errors.New("owner privileges required")
	}

	// Soft delete group
	_, err = s.groupsColl.UpdateOne(ctx, bson.M{"_id": groupID}, bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return nil
}

// JoinGroup allows a user to join a group
func (s *GroupService) JoinGroup(groupID, userID primitive.ObjectID, req models.JoinGroupRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get group
	var group models.Group
	err := s.groupsColl.FindOne(ctx, bson.M{"_id": groupID}).Decode(&group)
	if err != nil {
		return errors.New("group not found")
	}

	// Check if user can join
	memberStatus, _ := s.GetMemberStatus(groupID, userID)
	if !group.CanJoinGroup(userID, memberStatus) {
		return errors.New("cannot join this group")
	}

	// Check if already a member
	if memberStatus == "member" {
		return errors.New("already a member of this group")
	}

	// Check if already has pending request
	if memberStatus == "pending" {
		return errors.New("join request already pending")
	}

	// Create membership request
	member := models.GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    models.GroupRoleMember,
	}

	if group.MemberApprovalRequired && group.Privacy != models.GroupPublic {
		member.Status = "pending"
	} else {
		member.Status = "active"
		// Increment member count
		s.groupsColl.UpdateOne(ctx, bson.M{"_id": groupID}, bson.M{
			"$inc": bson.M{"members_count": 1},
			"$set": bson.M{"updated_at": time.Now()},
		})
	}

	member.BeforeCreate()

	_, err = s.membersColl.InsertOne(ctx, member)
	if err != nil {
		return fmt.Errorf("failed to join group: %w", err)
	}

	// Send notification to group admins if approval required
	if member.Status == "pending" {
		go s.notifyGroupAdmins(groupID, userID, "join_request")
	}

	return nil
}

// LeaveGroup allows a user to leave a group
func (s *GroupService) LeaveGroup(groupID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get member
	member, err := s.GetGroupMember(groupID, userID)
	if err != nil {
		return errors.New("not a member of this group")
	}

	// Check if user is owner and is the last admin
	if member.Role == models.GroupRoleOwner {
		adminCount, err := s.membersColl.CountDocuments(ctx, bson.M{
			"group_id": groupID,
			"role":     bson.M{"$in": []models.GroupRole{models.GroupRoleOwner, models.GroupRoleAdmin}},
			"status":   "active",
		})
		if err != nil {
			return fmt.Errorf("failed to check admin count: %w", err)
		}

		if adminCount <= 1 {
			return errors.New("cannot leave as the only admin. Transfer ownership first")
		}
	}

	// Remove member
	_, err = s.membersColl.DeleteOne(ctx, bson.M{
		"group_id": groupID,
		"user_id":  userID,
	})
	if err != nil {
		return fmt.Errorf("failed to leave group: %w", err)
	}

	// Update member count if was active member
	if member.Status == "active" {
		s.groupsColl.UpdateOne(ctx, bson.M{"_id": groupID}, bson.M{
			"$inc": bson.M{"members_count": -1},
			"$set": bson.M{"updated_at": time.Now()},
		})
	}

	return nil
}

// InviteToGroup invites users to a group
func (s *GroupService) InviteToGroup(groupID, inviterID primitive.ObjectID, req models.InviteToGroupRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if inviter has permission
	group, err := s.GetGroupByID(groupID, inviterID)
	if err != nil {
		return err
	}

	member, err := s.GetGroupMember(groupID, inviterID)
	if err != nil {
		return errors.New("access denied")
	}

	if !group.CanInviteToGroup(member.Role) {
		return errors.New("insufficient permissions to invite")
	}

	// Process each invitation
	for _, userIDStr := range req.UserIDs {
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			continue // Skip invalid IDs
		}

		// Check if user exists
		userCount, err := s.usersColl.CountDocuments(ctx, bson.M{"_id": userID})
		if err != nil || userCount == 0 {
			continue // Skip non-existent users
		}

		// Check if already member or invited
		memberStatus, _ := s.GetMemberStatus(groupID, userID)
		if memberStatus == "member" || memberStatus == "invited" {
			continue // Skip already members or invited users
		}

		// Create invitation
		invite := models.GroupInvite{
			GroupID:   groupID,
			InviterID: inviterID,
			InviteeID: userID,
			Message:   req.Message,
			Status:    "pending",
		}
		invite.BeforeCreate()

		_, err = s.invitesColl.InsertOne(ctx, invite)
		if err != nil {
			continue // Skip on error
		}

		// Send notification
		if s.notificationService != nil {
			go s.notificationService.NotifyGroupInvite(inviterID, userID, groupID)
		}
	}

	return nil
}

// AcceptGroupInvite accepts a group invitation
func (s *GroupService) AcceptGroupInvite(inviteID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get invitation
	var invite models.GroupInvite
	err := s.invitesColl.FindOne(ctx, bson.M{
		"_id":        inviteID,
		"invitee_id": userID,
		"status":     "pending",
	}).Decode(&invite)
	if err != nil {
		return errors.New("invitation not found or already processed")
	}

	// Check if invitation is expired
	if invite.IsExpired() {
		return errors.New("invitation has expired")
	}

	// Accept invitation
	invite.Accept()
	_, err = s.invitesColl.UpdateOne(ctx, bson.M{"_id": inviteID}, bson.M{
		"$set": bson.M{
			"status":      invite.Status,
			"accepted_at": invite.AcceptedAt,
			"updated_at":  invite.UpdatedAt,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Add user as member
	member := models.GroupMember{
		GroupID: invite.GroupID,
		UserID:  userID,
		Role:    models.GroupRoleMember,
		Status:  "active",
	}
	member.BeforeCreate()

	_, err = s.membersColl.InsertOne(ctx, member)
	if err != nil {
		return fmt.Errorf("failed to add as member: %w", err)
	}

	// Increment member count
	s.groupsColl.UpdateOne(ctx, bson.M{"_id": invite.GroupID}, bson.M{
		"$inc": bson.M{"members_count": 1},
		"$set": bson.M{"updated_at": time.Now()},
	})

	return nil
}

// RejectGroupInvite rejects a group invitation
func (s *GroupService) RejectGroupInvite(inviteID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get invitation
	var invite models.GroupInvite
	err := s.invitesColl.FindOne(ctx, bson.M{
		"_id":        inviteID,
		"invitee_id": userID,
		"status":     "pending",
	}).Decode(&invite)
	if err != nil {
		return errors.New("invitation not found or already processed")
	}

	// Reject invitation
	invite.Decline()
	_, err = s.invitesColl.UpdateOne(ctx, bson.M{"_id": inviteID}, bson.M{
		"$set": bson.M{
			"status":      invite.Status,
			"declined_at": invite.DeclinedAt,
			"updated_at":  invite.UpdatedAt,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	return nil
}

// GetGroupMembers retrieves group members
func (s *GroupService) GetGroupMembers(groupID, currentUserID primitive.ObjectID, limit, offset int) ([]models.GroupMemberResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user can view members
	memberStatus, _ := s.GetMemberStatus(groupID, currentUserID)
	if memberStatus != "member" {
		// Check if group is public
		var group models.Group
		err := s.groupsColl.FindOne(ctx, bson.M{"_id": groupID}).Decode(&group)
		if err != nil || group.Privacy != models.GroupPublic {
			return nil, errors.New("access denied")
		}
	}

	// Get members with user info
	pipeline := []bson.M{
		{"$match": bson.M{
			"group_id": groupID,
			"status":   "active",
		}},
		{"$lookup": bson.M{
			"from":         "users",
			"localField":   "user_id",
			"foreignField": "_id",
			"as":           "user",
		}},
		{"$unwind": "$user"},
		{"$sort": bson.M{"created_at": -1}},
		{"$skip": offset},
		{"$limit": limit},
	}

	cursor, err := s.membersColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	defer cursor.Close(ctx)

	var members []models.GroupMemberResponse
	for cursor.Next(ctx) {
		var result struct {
			models.GroupMember `bson:",inline"`
			User               models.User `bson:"user"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		memberResponse := result.GroupMember.ToGroupMemberResponse()
		memberResponse.User = result.User.ToUserResponse()
		members = append(members, memberResponse)
	}

	return members, nil
}

// GetUserGroups retrieves groups that a user is a member of
func (s *GroupService) GetUserGroups(userID primitive.ObjectID, limit, offset int) ([]models.GroupResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user's group memberships
	pipeline := []bson.M{
		{"$match": bson.M{
			"user_id": userID,
			"status":  "active",
		}},
		{"$lookup": bson.M{
			"from":         "groups",
			"localField":   "group_id",
			"foreignField": "_id",
			"as":           "group",
		}},
		{"$unwind": "$group"},
		{"$match": bson.M{
			"group.deleted_at": bson.M{"$exists": false},
		}},
		{"$sort": bson.M{"created_at": -1}},
		{"$skip": offset},
		{"$limit": limit},
	}

	cursor, err := s.membersColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}
	defer cursor.Close(ctx)

	var groups []models.GroupResponse
	for cursor.Next(ctx) {
		var result struct {
			models.GroupMember `bson:",inline"`
			Group              models.Group `bson:"group"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		groupResponse := result.Group.ToGroupResponse()
		groupResponse.UserRole = result.Role
		groupResponse.UserStatus = result.Status
		groupResponse.IsMember = true
		groupResponse.IsAdmin = result.Role == models.GroupRoleAdmin || result.Role == models.GroupRoleOwner
		groupResponse.IsModerator = result.Role == models.GroupRoleModerator
		groupResponse.JoinedAt = &result.JoinedAt

		groups = append(groups, groupResponse)
	}

	return groups, nil
}

// SearchGroups searches for groups
func (s *GroupService) SearchGroups(query string, currentUserID *primitive.ObjectID, limit, offset int) ([]models.GroupResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build search filter
	searchFilter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$in": []string{query}}},
		},
		"deleted_at": bson.M{"$exists": false},
		"is_active":  true,
	}

	// Only show public groups to non-members
	if currentUserID == nil {
		searchFilter["privacy"] = models.GroupPublic
	}

	cursor, err := s.groupsColl.Find(ctx, searchFilter, &options.FindOptions{
		Sort:  bson.D{{Key: "members_count", Value: -1}},
		Skip:  func() *int64 { skip := int64(offset); return &skip }(),
		Limit: func() *int64 { limit := int64(limit); return &limit }(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search groups: %w", err)
	}
	defer cursor.Close(ctx)

	var groups []models.GroupResponse
	for cursor.Next(ctx) {
		var group models.Group
		if err := cursor.Decode(&group); err != nil {
			continue
		}

		groupResponse := group.ToGroupResponse()

		// Add user context if authenticated
		if currentUserID != nil {
			memberStatus, role := s.GetMemberStatus(group.ID, *currentUserID)
			groupResponse.UserStatus = memberStatus
			groupResponse.UserRole = role
			groupResponse.IsMember = memberStatus == "member"
			groupResponse.IsAdmin = role == models.GroupRoleAdmin || role == models.GroupRoleOwner
			groupResponse.IsModerator = role == models.GroupRoleModerator
		}

		groups = append(groups, groupResponse)
	}

	return groups, nil
}

// UpdateMemberRole updates a member's role
func (s *GroupService) UpdateMemberRole(groupID, adminID, memberID primitive.ObjectID, req models.UpdateMemberRoleRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check admin permissions
	admin, err := s.GetGroupMember(groupID, adminID)
	if err != nil {
		return errors.New("access denied")
	}

	// Get target member
	member, err := s.GetGroupMember(groupID, memberID)
	if err != nil {
		return errors.New("member not found")
	}

	// Check if admin can update this member's role
	if !member.CanUpdateRole(admin.Role) {
		return errors.New("insufficient permissions")
	}

	// Update role
	_, err = s.membersColl.UpdateOne(ctx, bson.M{
		"group_id": groupID,
		"user_id":  memberID,
	}, bson.M{
		"$set": bson.M{
			"role":       req.Role,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	// Update group admin/mod counts
	go s.updateGroupRoleCounts(groupID)

	return nil
}

// RemoveGroupMember removes a member from a group
func (s *GroupService) RemoveGroupMember(groupID, adminID, memberID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check admin permissions
	admin, err := s.GetGroupMember(groupID, adminID)
	if err != nil {
		return errors.New("access denied")
	}

	// Get target member
	member, err := s.GetGroupMember(groupID, memberID)
	if err != nil {
		return errors.New("member not found")
	}

	// Check if admin can remove this member
	if !member.CanRemoveMember(admin.Role) {
		return errors.New("insufficient permissions")
	}

	// Remove member
	_, err = s.membersColl.DeleteOne(ctx, bson.M{
		"group_id": groupID,
		"user_id":  memberID,
	})
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Update member count
	s.groupsColl.UpdateOne(ctx, bson.M{"_id": groupID}, bson.M{
		"$inc": bson.M{"members_count": -1},
		"$set": bson.M{"updated_at": time.Now()},
	})

	return nil
}

// GetGroupStats retrieves group statistics
func (s *GroupService) GetGroupStats(groupID primitive.ObjectID, userID primitive.ObjectID) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user can view stats
	member, err := s.GetGroupMember(groupID, userID)
	if err != nil {
		return nil, errors.New("access denied")
	}

	if member.Role != models.GroupRoleAdmin && member.Role != models.GroupRoleOwner {
		return nil, errors.New("admin privileges required")
	}

	// Get basic stats from group document
	var group models.Group
	err = s.groupsColl.FindOne(ctx, bson.M{"_id": groupID}).Decode(&group)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Get additional stats
	stats := map[string]interface{}{
		"total_members":    group.MembersCount,
		"total_posts":      group.PostsCount,
		"total_events":     group.EventsCount,
		"admins_count":     group.AdminsCount,
		"moderators_count": group.ModsCount,
		"created_at":       group.CreatedAt,
		"last_activity":    group.LastActivityAt,
		"engagement_score": group.EngagementScore,
		"activity_score":   group.ActivityScore,
		"weekly_growth":    group.WeeklyGrowthRate,
	}

	// Get recent activity stats (last 30 days)
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)

	// New members in last 30 days
	newMembers, _ := s.membersColl.CountDocuments(ctx, bson.M{
		"group_id":   groupID,
		"status":     "active",
		"created_at": bson.M{"$gte": thirtyDaysAgo},
	})
	stats["new_members_30d"] = newMembers

	// Posts in last 30 days
	newPosts, _ := s.postsColl.CountDocuments(ctx, bson.M{
		"group_id":   groupID,
		"created_at": bson.M{"$gte": thirtyDaysAgo},
		"deleted_at": bson.M{"$exists": false},
	})
	stats["new_posts_30d"] = newPosts

	return stats, nil
}

// Helper methods

// GetGroupMember retrieves a group member
func (s *GroupService) GetGroupMember(groupID, userID primitive.ObjectID) (*models.GroupMember, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var member models.GroupMember
	err := s.membersColl.FindOne(ctx, bson.M{
		"group_id": groupID,
		"user_id":  userID,
	}).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("not a member of this group")
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	return &member, nil
}

// GetMemberStatus returns the membership status and role of a user in a group
func (s *GroupService) GetMemberStatus(groupID, userID primitive.ObjectID) (string, models.GroupRole) {
	member, err := s.GetGroupMember(groupID, userID)
	if err != nil {
		// Check for pending invitation
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		inviteCount, err := s.invitesColl.CountDocuments(ctx, bson.M{
			"group_id":   groupID,
			"invitee_id": userID,
			"status":     "pending",
		})
		if err == nil && inviteCount > 0 {
			return "invited", ""
		}

		return "not_member", ""
	}

	return member.Status, member.Role
}

// notifyGroupAdmins sends notification to group admins
func (s *GroupService) notifyGroupAdmins(groupID, actorID primitive.ObjectID, notificationType string) {
	if s.notificationService == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get group admins
	cursor, err := s.membersColl.Find(ctx, bson.M{
		"group_id": groupID,
		"role": bson.M{"$in": []models.GroupRole{
			models.GroupRoleOwner,
			models.GroupRoleAdmin,
			models.GroupRoleModerator,
		}},
		"status": "active",
	})
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var member models.GroupMember
		if cursor.Decode(&member) == nil {
			// Don't notify the actor
			if member.UserID != actorID {
				switch notificationType {
				case "join_request":
					s.notificationService.NotifyGroupJoinRequest(actorID, member.UserID, groupID)
				}
			}
		}
	}
}

// updateGroupRoleCounts updates the admin and moderator counts for a group
func (s *GroupService) updateGroupRoleCounts(groupID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Count admins
	adminCount, _ := s.membersColl.CountDocuments(ctx, bson.M{
		"group_id": groupID,
		"role": bson.M{"$in": []models.GroupRole{
			models.GroupRoleOwner,
			models.GroupRoleAdmin,
		}},
		"status": "active",
	})

	// Count moderators
	modCount, _ := s.membersColl.CountDocuments(ctx, bson.M{
		"group_id": groupID,
		"role":     models.GroupRoleModerator,
		"status":   "active",
	})

	// Update group document
	s.groupsColl.UpdateOne(ctx, bson.M{"_id": groupID}, bson.M{
		"$set": bson.M{
			"admins_count": adminCount,
			"mods_count":   modCount,
			"updated_at":   time.Now(),
		},
	})
}

// GetPublicGroups retrieves public groups for discovery
func (s *GroupService) GetPublicGroups(limit, offset int) ([]models.GroupResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.groupsColl.Find(ctx, bson.M{
		"privacy":    models.GroupPublic,
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}, &options.FindOptions{
		Sort:  bson.D{{Key: "members_count", Value: -1}},
		Skip:  func() *int64 { skip := int64(offset); return &skip }(),
		Limit: func() *int64 { limit := int64(limit); return &limit }(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get public groups: %w", err)
	}
	defer cursor.Close(ctx)

	var groups []models.GroupResponse
	for cursor.Next(ctx) {
		var group models.Group
		if err := cursor.Decode(&group); err != nil {
			continue
		}

		groups = append(groups, group.ToGroupResponse())
	}

	return groups, nil
}

// GetTrendingGroups retrieves trending groups
func (s *GroupService) GetTrendingGroups(limit, offset int, timeRange string) ([]models.GroupResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Calculate time range for trending
	var since time.Time
	switch timeRange {
	case "week":
		since = time.Now().Add(-7 * 24 * time.Hour)
	case "month":
		since = time.Now().Add(-30 * 24 * time.Hour)
	default:
		since = time.Now().Add(-24 * time.Hour) // day
	}

	// For now, sort by activity score and member count
	// In a full implementation, you'd calculate trending based on recent activity
	cursor, err := s.groupsColl.Find(ctx, bson.M{
		"privacy":          models.GroupPublic,
		"is_active":        true,
		"deleted_at":       bson.M{"$exists": false},
		"last_activity_at": bson.M{"$gte": since},
	}, &options.FindOptions{
		Sort: bson.D{
			{Key: "activity_score", Value: -1},
			{Key: "members_count", Value: -1},
		},
		Skip:  func() *int64 { skip := int64(offset); return &skip }(),
		Limit: func() *int64 { limit := int64(limit); return &limit }(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get trending groups: %w", err)
	}
	defer cursor.Close(ctx)

	var groups []models.GroupResponse
	for cursor.Next(ctx) {
		var group models.Group
		if err := cursor.Decode(&group); err != nil {
			continue
		}

		groups = append(groups, group.ToGroupResponse())
	}

	return groups, nil
}
