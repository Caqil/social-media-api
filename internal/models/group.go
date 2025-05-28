// models/group.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Group represents a community/group where users can join and share content
type Group struct {
	BaseModel `bson:",inline"`

	// Basic Information
	Name        string `json:"name" bson:"name" validate:"required,min=3,max=100"`
	Slug        string `json:"slug" bson:"slug" validate:"required,min=3,max=100"` // URL-friendly name
	Description string `json:"description" bson:"description" validate:"max=2000"`

	// Visual Identity
	ProfilePic string `json:"profile_pic" bson:"profile_pic"`
	CoverPic   string `json:"cover_pic" bson:"cover_pic"`
	Color      string `json:"color,omitempty" bson:"color,omitempty"` // Brand color

	// Group Settings
	Privacy  GroupPrivacy `json:"privacy" bson:"privacy" validate:"required"`
	Category string       `json:"category" bson:"category" validate:"required,max=50"`
	Tags     []string     `json:"tags,omitempty" bson:"tags,omitempty"`
	Location *Location    `json:"location,omitempty" bson:"location,omitempty"`
	Website  string       `json:"website,omitempty" bson:"website,omitempty" validate:"omitempty,url"`

	// Membership
	CreatedBy    primitive.ObjectID `json:"created_by" bson:"created_by" validate:"required"`
	Creator      UserResponse       `json:"creator,omitempty" bson:"-"` // Populated when querying
	MembersCount int64              `json:"members_count" bson:"members_count"`
	AdminsCount  int64              `json:"admins_count" bson:"admins_count"`
	ModsCount    int64              `json:"mods_count" bson:"mods_count"`

	// Content Statistics
	PostsCount  int64 `json:"posts_count" bson:"posts_count"`
	EventsCount int64 `json:"events_count" bson:"events_count"`

	// Group Rules and Settings
	Rules                  []GroupRule `json:"rules,omitempty" bson:"rules,omitempty"`
	PostApprovalRequired   bool        `json:"post_approval_required" bson:"post_approval_required"`
	MemberApprovalRequired bool        `json:"member_approval_required" bson:"member_approval_required"`
	AllowMemberInvites     bool        `json:"allow_member_invites" bson:"allow_member_invites"`
	AllowExternalSharing   bool        `json:"allow_external_sharing" bson:"allow_external_sharing"`
	AllowPolls             bool        `json:"allow_polls" bson:"allow_polls"`
	AllowEvents            bool        `json:"allow_events" bson:"allow_events"`
	AllowDiscussions       bool        `json:"allow_discussions" bson:"allow_discussions"`

	// Activity and Engagement
	LastActivityAt   *time.Time `json:"last_activity_at,omitempty" bson:"last_activity_at,omitempty"`
	LastPostAt       *time.Time `json:"last_post_at,omitempty" bson:"last_post_at,omitempty"`
	WeeklyGrowthRate float64    `json:"weekly_growth_rate" bson:"weekly_growth_rate"`
	EngagementScore  float64    `json:"engagement_score" bson:"engagement_score"`
	ActivityScore    float64    `json:"activity_score" bson:"activity_score"`

	// Moderation
	IsVerified       bool       `json:"is_verified" bson:"is_verified"`
	IsActive         bool       `json:"is_active" bson:"is_active"`
	IsSuspended      bool       `json:"is_suspended" bson:"is_suspended"`
	SuspendedAt      *time.Time `json:"suspended_at,omitempty" bson:"suspended_at,omitempty"`
	SuspensionReason string     `json:"suspension_reason,omitempty" bson:"suspension_reason,omitempty"`

	// Premium Features
	IsPremium       bool     `json:"is_premium" bson:"is_premium"`
	PremiumFeatures []string `json:"premium_features,omitempty" bson:"premium_features,omitempty"`

	// Custom Fields
	CustomFields map[string]interface{} `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`
}

// GroupRule represents a rule for the group
type GroupRule struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title" validate:"required,max=100"`
	Description string             `json:"description" bson:"description" validate:"required,max=500"`
	Order       int                `json:"order" bson:"order"`
	IsActive    bool               `json:"is_active" bson:"is_active"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// GroupMember represents a user's membership in a group
type GroupMember struct {
	BaseModel `bson:",inline"`

	GroupID primitive.ObjectID `json:"group_id" bson:"group_id" validate:"required"`
	UserID  primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`

	// User and Group info (populated when querying)
	User  UserResponse  `json:"user,omitempty" bson:"-"`
	Group GroupResponse `json:"group,omitempty" bson:"-"`

	// Membership details
	Role       GroupRole           `json:"role" bson:"role" validate:"required"`
	Status     string              `json:"status" bson:"status"` // active, pending, invited, banned
	JoinedAt   time.Time           `json:"joined_at" bson:"joined_at"`
	InvitedBy  *primitive.ObjectID `json:"invited_by,omitempty" bson:"invited_by,omitempty"`
	InvitedAt  *time.Time          `json:"invited_at,omitempty" bson:"invited_at,omitempty"`
	ApprovedAt *time.Time          `json:"approved_at,omitempty" bson:"approved_at,omitempty"`
	ApprovedBy *primitive.ObjectID `json:"approved_by,omitempty" bson:"approved_by,omitempty"`

	// Member activity
	PostsCount    int64      `json:"posts_count" bson:"posts_count"`
	CommentsCount int64      `json:"comments_count" bson:"comments_count"`
	LastActiveAt  *time.Time `json:"last_active_at,omitempty" bson:"last_active_at,omitempty"`

	// Member settings
	NotificationsEnabled bool       `json:"notifications_enabled" bson:"notifications_enabled"`
	IsMuted              bool       `json:"is_muted" bson:"is_muted"`
	MutedUntil           *time.Time `json:"muted_until,omitempty" bson:"muted_until,omitempty"`

	// Custom member data
	Nickname     string                 `json:"nickname,omitempty" bson:"nickname,omitempty"`
	Bio          string                 `json:"bio,omitempty" bson:"bio,omitempty" validate:"max=200"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`
}

// GroupInvite represents an invitation to join a group
type GroupInvite struct {
	BaseModel `bson:",inline"`

	GroupID   primitive.ObjectID `json:"group_id" bson:"group_id" validate:"required"`
	InviterID primitive.ObjectID `json:"inviter_id" bson:"inviter_id" validate:"required"`
	InviteeID primitive.ObjectID `json:"invitee_id" bson:"invitee_id" validate:"required"`

	// Invite details
	Message   string    `json:"message,omitempty" bson:"message,omitempty" validate:"max=500"`
	Status    string    `json:"status" bson:"status"` // pending, accepted, declined, expired
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`

	// Response tracking
	AcceptedAt *time.Time `json:"accepted_at,omitempty" bson:"accepted_at,omitempty"`
	DeclinedAt *time.Time `json:"declined_at,omitempty" bson:"declined_at,omitempty"`

	// Populated when querying
	Group   GroupResponse `json:"group,omitempty" bson:"-"`
	Inviter UserResponse  `json:"inviter,omitempty" bson:"-"`
	Invitee UserResponse  `json:"invitee,omitempty" bson:"-"`
}

// Response Models

// GroupResponse represents the group data returned in API responses
type GroupResponse struct {
	ID                     string       `json:"id"`
	Name                   string       `json:"name"`
	Slug                   string       `json:"slug"`
	Description            string       `json:"description"`
	ProfilePic             string       `json:"profile_pic"`
	CoverPic               string       `json:"cover_pic"`
	Color                  string       `json:"color,omitempty"`
	Privacy                GroupPrivacy `json:"privacy"`
	Category               string       `json:"category"`
	Tags                   []string     `json:"tags,omitempty"`
	Location               *Location    `json:"location,omitempty"`
	Website                string       `json:"website,omitempty"`
	CreatedBy              string       `json:"created_by"`
	Creator                UserResponse `json:"creator,omitempty"`
	MembersCount           int64        `json:"members_count"`
	AdminsCount            int64        `json:"admins_count"`
	ModsCount              int64        `json:"mods_count"`
	PostsCount             int64        `json:"posts_count"`
	EventsCount            int64        `json:"events_count"`
	Rules                  []GroupRule  `json:"rules,omitempty"`
	PostApprovalRequired   bool         `json:"post_approval_required"`
	MemberApprovalRequired bool         `json:"member_approval_required"`
	AllowMemberInvites     bool         `json:"allow_member_invites"`
	AllowExternalSharing   bool         `json:"allow_external_sharing"`
	AllowPolls             bool         `json:"allow_polls"`
	AllowEvents            bool         `json:"allow_events"`
	AllowDiscussions       bool         `json:"allow_discussions"`
	LastActivityAt         *time.Time   `json:"last_activity_at,omitempty"`
	LastPostAt             *time.Time   `json:"last_post_at,omitempty"`
	WeeklyGrowthRate       float64      `json:"weekly_growth_rate"`
	EngagementScore        float64      `json:"engagement_score"`
	ActivityScore          float64      `json:"activity_score"`
	IsVerified             bool         `json:"is_verified"`
	IsActive               bool         `json:"is_active"`
	IsPremium              bool         `json:"is_premium"`
	CreatedAt              time.Time    `json:"created_at"`
	UpdatedAt              time.Time    `json:"updated_at"`

	// User-specific context
	UserRole    GroupRole  `json:"user_role,omitempty"`
	UserStatus  string     `json:"user_status,omitempty"` // member, pending, invited, not_member
	IsMember    bool       `json:"is_member,omitempty"`
	IsAdmin     bool       `json:"is_admin,omitempty"`
	IsModerator bool       `json:"is_moderator,omitempty"`
	CanPost     bool       `json:"can_post,omitempty"`
	CanInvite   bool       `json:"can_invite,omitempty"`
	CanModerate bool       `json:"can_moderate,omitempty"`
	JoinedAt    *time.Time `json:"joined_at,omitempty"`
}

// GroupMemberResponse represents group member data
type GroupMemberResponse struct {
	ID                   string                 `json:"id"`
	GroupID              string                 `json:"group_id"`
	UserID               string                 `json:"user_id"`
	User                 UserResponse           `json:"user"`
	Role                 GroupRole              `json:"role"`
	Status               string                 `json:"status"`
	JoinedAt             time.Time              `json:"joined_at"`
	PostsCount           int64                  `json:"posts_count"`
	CommentsCount        int64                  `json:"comments_count"`
	LastActiveAt         *time.Time             `json:"last_active_at,omitempty"`
	NotificationsEnabled bool                   `json:"notifications_enabled"`
	IsMuted              bool                   `json:"is_muted"`
	Nickname             string                 `json:"nickname,omitempty"`
	Bio                  string                 `json:"bio,omitempty"`
	CustomFields         map[string]interface{} `json:"custom_fields,omitempty"`
	TimeAgo              string                 `json:"time_ago,omitempty"`
}

// GroupInviteResponse represents group invite data
type GroupInviteResponse struct {
	ID        string        `json:"id"`
	GroupID   string        `json:"group_id"`
	Group     GroupResponse `json:"group"`
	InviterID string        `json:"inviter_id"`
	Inviter   UserResponse  `json:"inviter"`
	InviteeID string        `json:"invitee_id"`
	Invitee   UserResponse  `json:"invitee,omitempty"`
	Message   string        `json:"message,omitempty"`
	Status    string        `json:"status"`
	ExpiresAt time.Time     `json:"expires_at"`
	CreatedAt time.Time     `json:"created_at"`
	TimeAgo   string        `json:"time_ago,omitempty"`
}

// Request Models

// CreateGroupRequest represents the request to create a group
type CreateGroupRequest struct {
	Name                   string       `json:"name" validate:"required,min=3,max=100"`
	Description            string       `json:"description" validate:"max=2000"`
	Privacy                GroupPrivacy `json:"privacy" validate:"required,oneof=public private secret"`
	Category               string       `json:"category" validate:"required,max=50"`
	Tags                   []string     `json:"tags,omitempty"`
	Location               *Location    `json:"location,omitempty"`
	Website                string       `json:"website,omitempty" validate:"omitempty,url"`
	Rules                  []GroupRule  `json:"rules,omitempty"`
	PostApprovalRequired   bool         `json:"post_approval_required"`
	MemberApprovalRequired bool         `json:"member_approval_required"`
	AllowMemberInvites     bool         `json:"allow_member_invites"`
	AllowExternalSharing   bool         `json:"allow_external_sharing"`
	AllowPolls             bool         `json:"allow_polls"`
	AllowEvents            bool         `json:"allow_events"`
	AllowDiscussions       bool         `json:"allow_discussions"`
}

// UpdateGroupRequest represents the request to update a group
type UpdateGroupRequest struct {
	Name                   *string       `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description            *string       `json:"description,omitempty" validate:"omitempty,max=2000"`
	Privacy                *GroupPrivacy `json:"privacy,omitempty" validate:"omitempty,oneof=public private secret"`
	Category               *string       `json:"category,omitempty" validate:"omitempty,max=50"`
	Tags                   []string      `json:"tags,omitempty"`
	Location               *Location     `json:"location,omitempty"`
	Website                *string       `json:"website,omitempty" validate:"omitempty,url"`
	Color                  *string       `json:"color,omitempty"`
	Rules                  []GroupRule   `json:"rules,omitempty"`
	PostApprovalRequired   *bool         `json:"post_approval_required,omitempty"`
	MemberApprovalRequired *bool         `json:"member_approval_required,omitempty"`
	AllowMemberInvites     *bool         `json:"allow_member_invites,omitempty"`
	AllowExternalSharing   *bool         `json:"allow_external_sharing,omitempty"`
	AllowPolls             *bool         `json:"allow_polls,omitempty"`
	AllowEvents            *bool         `json:"allow_events,omitempty"`
	AllowDiscussions       *bool         `json:"allow_discussions,omitempty"`
}

// JoinGroupRequest represents the request to join a group
type JoinGroupRequest struct {
	Message string `json:"message,omitempty" validate:"max=500"`
}

// InviteToGroupRequest represents the request to invite users to a group
type InviteToGroupRequest struct {
	UserIDs []string `json:"user_ids" validate:"required,min=1,max=20"`
	Message string   `json:"message,omitempty" validate:"max=500"`
}

// UpdateMemberRoleRequest represents the request to update a member's role
type UpdateMemberRoleRequest struct {
	Role GroupRole `json:"role" validate:"required,oneof=member moderator admin"`
}

// Methods for Group model

// BeforeCreate sets default values before creating group
func (g *Group) BeforeCreate() {
	g.BaseModel.BeforeCreate()

	// Set default values
	g.MembersCount = 1 // Creator is first member
	g.AdminsCount = 1  // Creator is first admin
	g.ModsCount = 0
	g.PostsCount = 0
	g.EventsCount = 0
	g.WeeklyGrowthRate = 0.0
	g.EngagementScore = 0.0
	g.ActivityScore = 0.0
	g.IsVerified = false
	g.IsActive = true
	g.IsSuspended = false
	g.IsPremium = false

	// Set default settings
	g.PostApprovalRequired = false
	g.MemberApprovalRequired = false
	g.AllowMemberInvites = true
	g.AllowExternalSharing = true
	g.AllowPolls = true
	g.AllowEvents = true
	g.AllowDiscussions = true

	// Generate slug from name if not provided
	if g.Slug == "" {
		g.Slug = generateSlug(g.Name)
	}

	// Set last activity
	now := time.Now()
	g.LastActivityAt = &now
}

// generateSlug generates URL-friendly slug from name
func generateSlug(name string) string {
	// This is a simplified slug generation
	// In practice, you'd use a proper slug library
	return name // Placeholder implementation
}

// ToGroupResponse converts Group model to GroupResponse
func (g *Group) ToGroupResponse() GroupResponse {
	return GroupResponse{
		ID:                     g.ID.Hex(),
		Name:                   g.Name,
		Slug:                   g.Slug,
		Description:            g.Description,
		ProfilePic:             g.ProfilePic,
		CoverPic:               g.CoverPic,
		Color:                  g.Color,
		Privacy:                g.Privacy,
		Category:               g.Category,
		Tags:                   g.Tags,
		Location:               g.Location,
		Website:                g.Website,
		CreatedBy:              g.CreatedBy.Hex(),
		MembersCount:           g.MembersCount,
		AdminsCount:            g.AdminsCount,
		ModsCount:              g.ModsCount,
		PostsCount:             g.PostsCount,
		EventsCount:            g.EventsCount,
		Rules:                  g.Rules,
		PostApprovalRequired:   g.PostApprovalRequired,
		MemberApprovalRequired: g.MemberApprovalRequired,
		AllowMemberInvites:     g.AllowMemberInvites,
		AllowExternalSharing:   g.AllowExternalSharing,
		AllowPolls:             g.AllowPolls,
		AllowEvents:            g.AllowEvents,
		AllowDiscussions:       g.AllowDiscussions,
		LastActivityAt:         g.LastActivityAt,
		LastPostAt:             g.LastPostAt,
		WeeklyGrowthRate:       g.WeeklyGrowthRate,
		EngagementScore:        g.EngagementScore,
		ActivityScore:          g.ActivityScore,
		IsVerified:             g.IsVerified,
		IsActive:               g.IsActive,
		IsPremium:              g.IsPremium,
		CreatedAt:              g.CreatedAt,
		UpdatedAt:              g.UpdatedAt,
	}
}

// IncrementMembersCount increments the members count
func (g *Group) IncrementMembersCount() {
	g.MembersCount++
	g.UpdateActivity()
	g.BeforeUpdate()
}

// DecrementMembersCount decrements the members count
func (g *Group) DecrementMembersCount() {
	if g.MembersCount > 0 {
		g.MembersCount--
	}
	g.BeforeUpdate()
}

// IncrementPostsCount increments the posts count
func (g *Group) IncrementPostsCount() {
	g.PostsCount++
	now := time.Now()
	g.LastPostAt = &now
	g.UpdateActivity()
	g.BeforeUpdate()
}

// DecrementPostsCount decrements the posts count
func (g *Group) DecrementPostsCount() {
	if g.PostsCount > 0 {
		g.PostsCount--
	}
	g.BeforeUpdate()
}

// UpdateActivity updates the last activity timestamp
func (g *Group) UpdateActivity() {
	now := time.Now()
	g.LastActivityAt = &now
}

// CanViewGroup checks if a user can view this group
func (g *Group) CanViewGroup(currentUserID primitive.ObjectID, memberStatus string) bool {
	// Check if group is active
	if !g.IsActive || g.IsSuspended {
		return false
	}

	// Check privacy settings
	switch g.Privacy {
	case GroupPublic:
		return true
	case GroupPrivate:
		return memberStatus == "member" || memberStatus == "pending"
	case GroupSecret:
		return memberStatus == "member"
	default:
		return false
	}
}

// CanJoinGroup checks if a user can join this group
func (g *Group) CanJoinGroup(currentUserID primitive.ObjectID, memberStatus string) bool {
	// Check if already a member
	if memberStatus == "member" {
		return false
	}

	// Check if group is active
	if !g.IsActive || g.IsSuspended {
		return false
	}

	// Check privacy settings
	switch g.Privacy {
	case GroupPublic:
		return true
	case GroupPrivate:
		return true // Can request to join
	case GroupSecret:
		return false // Must be invited
	default:
		return false
	}
}

// CanPostInGroup checks if a user can post in this group
func (g *Group) CanPostInGroup(userRole GroupRole, memberStatus string) bool {
	// Must be a member
	if memberStatus != "member" {
		return false
	}

	// Check if group allows discussions
	if !g.AllowDiscussions {
		return userRole == GroupRoleAdmin || userRole == GroupRoleModerator
	}

	return true
}

// CanInviteToGroup checks if a user can invite others to this group
func (g *Group) CanInviteToGroup(userRole GroupRole) bool {
	if !g.AllowMemberInvites {
		return userRole == GroupRoleAdmin || userRole == GroupRoleModerator
	}

	return userRole != ""
}

// CanModerateGroup checks if a user can moderate this group
func (g *Group) CanModerateGroup(userRole GroupRole) bool {
	return userRole == GroupRoleAdmin || userRole == GroupRoleModerator || userRole == GroupRoleOwner
}

// Methods for GroupMember model

// BeforeCreate sets default values before creating group member
func (gm *GroupMember) BeforeCreate() {
	gm.BaseModel.BeforeCreate()

	// Set default values
	gm.Status = "active"
	gm.JoinedAt = gm.CreatedAt
	gm.PostsCount = 0
	gm.CommentsCount = 0
	gm.NotificationsEnabled = true
	gm.IsMuted = false

	// Set default role
	if gm.Role == "" {
		gm.Role = GroupRoleMember
	}
}

// ToGroupMemberResponse converts GroupMember to GroupMemberResponse
func (gm *GroupMember) ToGroupMemberResponse() GroupMemberResponse {
	return GroupMemberResponse{
		ID:                   gm.ID.Hex(),
		GroupID:              gm.GroupID.Hex(),
		UserID:               gm.UserID.Hex(),
		Role:                 gm.Role,
		Status:               gm.Status,
		JoinedAt:             gm.JoinedAt,
		PostsCount:           gm.PostsCount,
		CommentsCount:        gm.CommentsCount,
		LastActiveAt:         gm.LastActiveAt,
		NotificationsEnabled: gm.NotificationsEnabled,
		IsMuted:              gm.IsMuted,
		Nickname:             gm.Nickname,
		Bio:                  gm.Bio,
		CustomFields:         gm.CustomFields,
	}
}

// CanUpdateRole checks if the member's role can be updated
func (gm *GroupMember) CanUpdateRole(updaterRole GroupRole) bool {
	// Owner can update anyone
	if updaterRole == GroupRoleOwner {
		return true
	}

	// Admin can update members and moderators (but not other admins or owner)
	if updaterRole == GroupRoleAdmin {
		return gm.Role == GroupRoleMember || gm.Role == GroupRoleModerator
	}

	return false
}

// CanRemoveMember checks if the member can be removed
func (gm *GroupMember) CanRemoveMember(removerRole GroupRole) bool {
	// Owner can remove anyone except themselves
	if removerRole == GroupRoleOwner {
		return gm.Role != GroupRoleOwner
	}

	// Admin can remove members and moderators
	if removerRole == GroupRoleAdmin {
		return gm.Role == GroupRoleMember || gm.Role == GroupRoleModerator
	}

	// Moderator can remove members
	if removerRole == GroupRoleModerator {
		return gm.Role == GroupRoleMember
	}

	return false
}

// Methods for GroupInvite model

// BeforeCreate sets default values before creating group invite
func (gi *GroupInvite) BeforeCreate() {
	gi.BaseModel.BeforeCreate()

	// Set default values
	gi.Status = "pending"
	gi.ExpiresAt = gi.CreatedAt.Add(7 * 24 * time.Hour) // Expires in 7 days
}

// ToGroupInviteResponse converts GroupInvite to GroupInviteResponse
func (gi *GroupInvite) ToGroupInviteResponse() GroupInviteResponse {
	return GroupInviteResponse{
		ID:        gi.ID.Hex(),
		GroupID:   gi.GroupID.Hex(),
		InviterID: gi.InviterID.Hex(),
		InviteeID: gi.InviteeID.Hex(),
		Message:   gi.Message,
		Status:    gi.Status,
		ExpiresAt: gi.ExpiresAt,
		CreatedAt: gi.CreatedAt,
	}
}

// Accept accepts the group invitation
func (gi *GroupInvite) Accept() {
	gi.Status = "accepted"
	now := time.Now()
	gi.AcceptedAt = &now
	gi.BeforeUpdate()
}

// Decline declines the group invitation
func (gi *GroupInvite) Decline() {
	gi.Status = "declined"
	now := time.Now()
	gi.DeclinedAt = &now
	gi.BeforeUpdate()
}

// IsExpired checks if the invitation has expired
func (gi *GroupInvite) IsExpired() bool {
	return time.Now().After(gi.ExpiresAt)
}

// Utility functions

// GetGroupCategories returns available group categories
func GetGroupCategories() []string {
	return []string{
		"general",
		"technology",
		"business",
		"health",
		"education",
		"entertainment",
		"sports",
		"travel",
		"food",
		"art",
		"music",
		"books",
		"gaming",
		"fitness",
		"parenting",
		"pets",
		"hobbies",
		"local",
		"support",
		"professional",
	}
}

// IsValidGroupCategory checks if a category is valid
func IsValidGroupCategory(category string) bool {
	validCategories := GetGroupCategories()
	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}
