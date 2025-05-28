// user.go - Auto-generated placeholder
// models/user.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the social media platform
type User struct {
	BaseModel `bson:",inline"`

	// Basic Information
	Username    string `json:"username" bson:"username" validate:"required,min=3,max=50"`
	Email       string `json:"email" bson:"email" validate:"required,email"`
	Password    string `json:"-" bson:"password" validate:"required,min=8"`
	FirstName   string `json:"first_name" bson:"first_name" validate:"required,min=2,max=50"`
	LastName    string `json:"last_name" bson:"last_name" validate:"required,min=2,max=50"`
	DisplayName string `json:"display_name" bson:"display_name" validate:"max=100"`

	// Profile Information
	Bio         string     `json:"bio" bson:"bio" validate:"max=500"`
	ProfilePic  string     `json:"profile_pic" bson:"profile_pic"`
	CoverPic    string     `json:"cover_pic" bson:"cover_pic"`
	Website     string     `json:"website,omitempty" bson:"website,omitempty" validate:"omitempty,url"`
	Location    string     `json:"location,omitempty" bson:"location,omitempty" validate:"max=100"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty" bson:"date_of_birth,omitempty"`
	Gender      string     `json:"gender,omitempty" bson:"gender,omitempty" validate:"omitempty,oneof=male female other prefer_not_to_say"`

	// Contact Information
	Phone          string `json:"phone,omitempty" bson:"phone,omitempty"`
	AlternateEmail string `json:"alternate_email,omitempty" bson:"alternate_email,omitempty" validate:"omitempty,email"`

	// Account Status
	IsVerified  bool     `json:"is_verified" bson:"is_verified"`
	IsActive    bool     `json:"is_active" bson:"is_active"`
	IsPrivate   bool     `json:"is_private" bson:"is_private"`
	IsSuspended bool     `json:"is_suspended" bson:"is_suspended"`
	Role        UserRole `json:"role" bson:"role"`

	// Social Statistics
	FollowersCount int64 `json:"followers_count" bson:"followers_count"`
	FollowingCount int64 `json:"following_count" bson:"following_count"`
	PostsCount     int64 `json:"posts_count" bson:"posts_count"`
	FriendsCount   int64 `json:"friends_count" bson:"friends_count"`

	// Activity Tracking
	LastLoginAt  *time.Time `json:"last_login_at,omitempty" bson:"last_login_at,omitempty"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty" bson:"last_active_at,omitempty"`
	OnlineStatus string     `json:"online_status" bson:"online_status"` // online, offline, away

	// Settings
	PrivacySettings      PrivacySettings      `json:"privacy_settings" bson:"privacy_settings"`
	NotificationSettings NotificationSettings `json:"notification_settings" bson:"notification_settings"`

	// Two-Factor Authentication
	TwoFactorEnabled bool     `json:"two_factor_enabled" bson:"two_factor_enabled"`
	TwoFactorSecret  string   `json:"-" bson:"two_factor_secret,omitempty"`
	BackupCodes      []string `json:"-" bson:"backup_codes,omitempty"`

	// Account Recovery
	PasswordResetToken  string     `json:"-" bson:"password_reset_token,omitempty"`
	PasswordResetExpiry *time.Time `json:"-" bson:"password_reset_expiry,omitempty"`
	EmailVerifyToken    string     `json:"-" bson:"email_verify_token,omitempty"`
	EmailVerified       bool       `json:"email_verified" bson:"email_verified"`
	EmailVerifiedAt     *time.Time `json:"email_verified_at,omitempty" bson:"email_verified_at,omitempty"`

	// Blocked/Reported Users
	BlockedUsers    []primitive.ObjectID `json:"-" bson:"blocked_users,omitempty"`
	ReportedByCount int64                `json:"-" bson:"reported_by_count"`

	// Device and Session Info
	LastDeviceInfo string               `json:"-" bson:"last_device_info,omitempty"`
	FCMTokens      []string             `json:"-" bson:"fcm_tokens,omitempty"` // For push notifications
	ActiveSessions []primitive.ObjectID `json:"-" bson:"active_sessions,omitempty"`

	// Preferences
	Language string `json:"language" bson:"language"`
	Timezone string `json:"timezone" bson:"timezone"`
	Theme    string `json:"theme" bson:"theme"` // light, dark, auto

	// Social Links
	SocialLinks map[string]string `json:"social_links,omitempty" bson:"social_links,omitempty"`

	// Account Metrics (for analytics)
	TotalLikesReceived    int64 `json:"total_likes_received" bson:"total_likes_received"`
	TotalCommentsReceived int64 `json:"total_comments_received" bson:"total_comments_received"`
	TotalSharesReceived   int64 `json:"total_shares_received" bson:"total_shares_received"`
	ProfileViews          int64 `json:"profile_views" bson:"profile_views"`

	// Subscription/Premium Features
	IsPremium        bool       `json:"is_premium" bson:"is_premium"`
	PremiumExpiry    *time.Time `json:"premium_expiry,omitempty" bson:"premium_expiry,omitempty"`
	SubscriptionPlan string     `json:"subscription_plan,omitempty" bson:"subscription_plan,omitempty"`
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID             string            `json:"id"`
	Username       string            `json:"username"`
	Email          string            `json:"email,omitempty"` // Controlled by privacy settings
	FirstName      string            `json:"first_name"`
	LastName       string            `json:"last_name"`
	DisplayName    string            `json:"display_name"`
	Bio            string            `json:"bio"`
	ProfilePic     string            `json:"profile_pic"`
	CoverPic       string            `json:"cover_pic"`
	Website        string            `json:"website,omitempty"`
	Location       string            `json:"location,omitempty"`
	IsVerified     bool              `json:"is_verified"`
	IsPrivate      bool              `json:"is_private"`
	FollowersCount int64             `json:"followers_count"`
	FollowingCount int64             `json:"following_count"`
	PostsCount     int64             `json:"posts_count"`
	FriendsCount   int64             `json:"friends_count"`
	OnlineStatus   string            `json:"online_status,omitempty"`
	LastActiveAt   *time.Time        `json:"last_active_at,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	IsFollowing    bool              `json:"is_following,omitempty"`   // Set based on current user context
	IsFollowedBy   bool              `json:"is_followed_by,omitempty"` // Set based on current user context
	IsFriend       bool              `json:"is_friend,omitempty"`      // Set based on current user context
	IsBlocked      bool              `json:"is_blocked,omitempty"`     // Set based on current user context
	MutualFriends  int64             `json:"mutual_friends,omitempty"` // Set based on current user context
	SocialLinks    map[string]string `json:"social_links,omitempty"`
	IsPremium      bool              `json:"is_premium"`
}

// ProfileResponse represents detailed profile information
type ProfileResponse struct {
	UserResponse          `json:",inline"`
	TotalLikesReceived    int64          `json:"total_likes_received,omitempty"`
	TotalCommentsReceived int64          `json:"total_comments_received,omitempty"`
	ProfileViews          int64          `json:"profile_views,omitempty"`
	RecentPosts           []PostResponse `json:"recent_posts,omitempty"`
	MutualConnections     []UserResponse `json:"mutual_connections,omitempty"`
}

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	Username    string     `json:"username" validate:"required,min=3,max=50"`
	Email       string     `json:"email" validate:"required,email"`
	Password    string     `json:"password" validate:"required,min=8"`
	FirstName   string     `json:"first_name" validate:"required,min=2,max=50"`
	LastName    string     `json:"last_name" validate:"required,min=2,max=50"`
	DisplayName string     `json:"display_name,omitempty" validate:"max=100"`
	Bio         string     `json:"bio,omitempty" validate:"max=500"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Gender      string     `json:"gender,omitempty" validate:"omitempty,oneof=male female other prefer_not_to_say"`
	Phone       string     `json:"phone,omitempty"`
}

// LoginRequest represents the user login request
type LoginRequest struct {
	EmailOrUsername string `json:"email_or_username" validate:"required"`
	Password        string `json:"password" validate:"required"`
	RememberMe      bool   `json:"remember_me"`
	DeviceInfo      string `json:"device_info,omitempty"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	FirstName   *string           `json:"first_name,omitempty" validate:"omitempty,min=2,max=50"`
	LastName    *string           `json:"last_name,omitempty" validate:"omitempty,min=2,max=50"`
	DisplayName *string           `json:"display_name,omitempty" validate:"omitempty,max=100"`
	Bio         *string           `json:"bio,omitempty" validate:"omitempty,max=500"`
	Website     *string           `json:"website,omitempty" validate:"omitempty,url"`
	Location    *string           `json:"location,omitempty" validate:"omitempty,max=100"`
	DateOfBirth *time.Time        `json:"date_of_birth,omitempty"`
	Gender      *string           `json:"gender,omitempty" validate:"omitempty,oneof=male female other prefer_not_to_say"`
	Phone       *string           `json:"phone,omitempty"`
	SocialLinks map[string]string `json:"social_links,omitempty"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents reset password request
type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

// UpdatePrivacySettingsRequest represents privacy settings update request
type UpdatePrivacySettingsRequest struct {
	PrivacySettings PrivacySettings `json:"privacy_settings" validate:"required"`
}

// UpdateNotificationSettingsRequest represents notification settings update request
type UpdateNotificationSettingsRequest struct {
	NotificationSettings NotificationSettings `json:"notification_settings" validate:"required"`
}

// UserSearchResponse represents user search results
type UserSearchResponse struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	DisplayName   string `json:"display_name"`
	ProfilePic    string `json:"profile_pic"`
	IsVerified    bool   `json:"is_verified"`
	MutualFriends int64  `json:"mutual_friends,omitempty"`
	IsFollowing   bool   `json:"is_following,omitempty"`
}

type UserStatus string

const (
	UserStatusActive     UserStatus = "active"
	UserStatusInactive   UserStatus = "inactive"
	UserStatusSuspended  UserStatus = "suspended"
	UserStatusBanned     UserStatus = "banned"
	UserStatusDeleted    UserStatus = "deleted"
	UserStatusPending    UserStatus = "pending"
	UserStatusRestricted UserStatus = "restricted"
	UserStatusVerifying  UserStatus = "verifying"
)

// BeforeCreate sets default values before creating user
func (u *User) BeforeCreate() {
	u.BaseModel.BeforeCreate()
	u.IsVerified = false
	u.IsActive = true
	u.IsPrivate = false
	u.IsSuspended = false
	u.Role = RoleUser
	u.OnlineStatus = "offline"
	u.EmailVerified = false
	u.TwoFactorEnabled = false
	u.Language = "en"
	u.Timezone = "UTC"
	u.Theme = "light"
	u.PrivacySettings = DefaultPrivacySettings()
	u.NotificationSettings = DefaultNotificationSettings()

	// Initialize counts
	u.FollowersCount = 0
	u.FollowingCount = 0
	u.PostsCount = 0
	u.FriendsCount = 0
	u.TotalLikesReceived = 0
	u.TotalCommentsReceived = 0
	u.TotalSharesReceived = 0
	u.ProfileViews = 0
	u.ReportedByCount = 0

	// Set display name if not provided
	if u.DisplayName == "" {
		u.DisplayName = u.FirstName + " " + u.LastName
	}
}

// ToUserResponse converts User model to UserResponse
func (u *User) ToUserResponse() UserResponse {
	return UserResponse{
		ID:             u.ID.Hex(),
		Username:       u.Username,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		DisplayName:    u.DisplayName,
		Bio:            u.Bio,
		ProfilePic:     u.ProfilePic,
		CoverPic:       u.CoverPic,
		Website:        u.Website,
		Location:       u.Location,
		IsVerified:     u.IsVerified,
		IsPrivate:      u.IsPrivate,
		FollowersCount: u.FollowersCount,
		FollowingCount: u.FollowingCount,
		PostsCount:     u.PostsCount,
		FriendsCount:   u.FriendsCount,
		CreatedAt:      u.CreatedAt,
		SocialLinks:    u.SocialLinks,
		IsPremium:      u.IsPremium,
	}
}

// ToUserResponseWithContext converts User model to UserResponse with relationship context
func (u *User) ToUserResponseWithContext(currentUserID primitive.ObjectID, isFollowing, isFollowedBy, isFriend, isBlocked bool, mutualFriends int64) UserResponse {
	response := u.ToUserResponse()
	response.IsFollowing = isFollowing
	response.IsFollowedBy = isFollowedBy
	response.IsFriend = isFriend
	response.IsBlocked = isBlocked
	response.MutualFriends = mutualFriends

	// Apply privacy settings for email visibility
	if u.PrivacySettings.EmailVisibility != PrivacyPublic && u.ID != currentUserID {
		if u.PrivacySettings.EmailVisibility == PrivacyFriends && !isFriend {
			response.Email = ""
		} else if u.PrivacySettings.EmailVisibility == PrivacyPrivate {
			response.Email = ""
		}
	} else if u.ID == currentUserID {
		response.Email = u.Email // Always show own email
	}

	// Apply privacy settings for online status
	if !u.PrivacySettings.ShowOnlineStatus && u.ID != currentUserID {
		response.OnlineStatus = ""
		response.LastActiveAt = nil
	} else {
		response.OnlineStatus = u.OnlineStatus
		response.LastActiveAt = u.LastActiveAt
	}

	return response
}

// ToProfileResponse converts User model to detailed ProfileResponse
func (u *User) ToProfileResponse() ProfileResponse {
	return ProfileResponse{
		UserResponse:          u.ToUserResponse(),
		TotalLikesReceived:    u.TotalLikesReceived,
		TotalCommentsReceived: u.TotalCommentsReceived,
		ProfileViews:          u.ProfileViews,
	}
}

// ToUserSearchResponse converts User model to UserSearchResponse
func (u *User) ToUserSearchResponse() UserSearchResponse {
	return UserSearchResponse{
		ID:          u.ID.Hex(),
		Username:    u.Username,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		DisplayName: u.DisplayName,
		ProfilePic:  u.ProfilePic,
		IsVerified:  u.IsVerified,
	}
}

// IsBlocking checks if this user is blocking another user
func (u *User) IsBlocking(userID primitive.ObjectID) bool {
	for _, blockedID := range u.BlockedUsers {
		if blockedID == userID {
			return true
		}
	}
	return false
}

// BlockUser adds a user to the blocked list
func (u *User) BlockUser(userID primitive.ObjectID) {
	if !u.IsBlocking(userID) {
		u.BlockedUsers = append(u.BlockedUsers, userID)
		u.BeforeUpdate()
	}
}

// UnblockUser removes a user from the blocked list
func (u *User) UnblockUser(userID primitive.ObjectID) {
	for i, blockedID := range u.BlockedUsers {
		if blockedID == userID {
			u.BlockedUsers = append(u.BlockedUsers[:i], u.BlockedUsers[i+1:]...)
			u.BeforeUpdate()
			break
		}
	}
}

// UpdateOnlineStatus updates the user's online status and last active time
func (u *User) UpdateOnlineStatus(status string) {
	u.OnlineStatus = status
	now := time.Now()
	u.LastActiveAt = &now
	u.BeforeUpdate()
}

// IncrementFollowersCount increments the followers count
func (u *User) IncrementFollowersCount() {
	u.FollowersCount++
	u.BeforeUpdate()
}

// DecrementFollowersCount decrements the followers count
func (u *User) DecrementFollowersCount() {
	if u.FollowersCount > 0 {
		u.FollowersCount--
	}
	u.BeforeUpdate()
}

// IncrementFollowingCount increments the following count
func (u *User) IncrementFollowingCount() {
	u.FollowingCount++
	u.BeforeUpdate()
}

// DecrementFollowingCount decrements the following count
func (u *User) DecrementFollowingCount() {
	if u.FollowingCount > 0 {
		u.FollowingCount--
	}
	u.BeforeUpdate()
}

// IncrementPostsCount increments the posts count
func (u *User) IncrementPostsCount() {
	u.PostsCount++
	u.BeforeUpdate()
}

// DecrementPostsCount decrements the posts count
func (u *User) DecrementPostsCount() {
	if u.PostsCount > 0 {
		u.PostsCount--
	}
	u.BeforeUpdate()
}

// CanViewProfile checks if the current user can view this profile
func (u *User) CanViewProfile(currentUserID primitive.ObjectID, isFollowing bool) bool {
	// User can always view their own profile
	if u.ID == currentUserID {
		return true
	}

	// If account is suspended, only the user themselves can view
	if u.IsSuspended {
		return false
	}

	// Check privacy settings
	switch u.PrivacySettings.ProfileVisibility {
	case PrivacyPublic:
		return true
	case PrivacyFriends:
		return isFollowing // In this context, "friends" means followers/following
	case PrivacyPrivate:
		return false
	default:
		return false
	}
}

// CanViewPosts checks if the current user can view this user's posts
func (u *User) CanViewPosts(currentUserID primitive.ObjectID, isFollowing bool) bool {
	// User can always view their own posts
	if u.ID == currentUserID {
		return true
	}

	// If account is suspended, only the user themselves can view
	if u.IsSuspended {
		return false
	}

	// Check privacy settings
	switch u.PrivacySettings.PostsVisibility {
	case PrivacyPublic:
		return true
	case PrivacyFriends:
		return isFollowing
	case PrivacyPrivate:
		return false
	default:
		return false
	}
}
