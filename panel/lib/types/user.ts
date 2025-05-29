import { BaseEntity, SoftDeleteEntity, Location, SocialLinks, Media, PaginatedResponse } from './common';
import { User } from './auth';

export interface UserProfile extends User {
  follower_count: number;
  following_count: number;
  post_count: number;
  story_count: number;
  comment_count: number;
  like_count: number;
  group_count: number;
  verification_badge?: VerificationBadge;
  privacy_settings: UserPrivacySettings;
  notification_settings: UserNotificationSettings;
  social_links?: SocialLinks;
  cover_image?: string;
  reputation_score: number;
  trust_level: number;
  account_type: 'personal' | 'business' | 'creator';
  subscription_tier?: 'free' | 'premium' | 'pro';
  subscription_expires_at?: string;
}

export interface VerificationBadge {
  type: 'verified' | 'premium' | 'business' | 'government';
  verified_at: string;
  verified_by: string;
  reason?: string;
}

export interface UserPrivacySettings {
  profile_visibility: 'public' | 'friends' | 'private';
  post_visibility: 'public' | 'friends' | 'private';
  story_visibility: 'public' | 'friends' | 'private';
  allow_message_requests: boolean;
  show_online_status: boolean;
  show_last_seen: boolean;
  allow_tagging: boolean;
  allow_search_by_email: boolean;
  allow_search_by_phone: boolean;
  show_activity_status: boolean;
  allow_friend_suggestions: boolean;
}

export interface UserNotificationSettings {
  likes: boolean;
  comments: boolean;
  follows: boolean;
  mentions: boolean;
  messages: boolean;
  group_invites: boolean;
  event_invites: boolean;
  friend_requests: boolean;
  email_notifications: boolean;
  push_notifications: boolean;
  sms_notifications: boolean;
  digest_frequency: 'daily' | 'weekly' | 'monthly' | 'never';
  quiet_hours_start?: string;
  quiet_hours_end?: string;
}

export interface UserStats {
  total_users: number;
  active_users: number;
  new_users_today: number;
  new_users_week: number;
  new_users_month: number;
  verified_users: number;
  suspended_users: number;
  premium_users: number;
  retention_rate_7d: number;
  retention_rate_30d: number;
  avg_session_duration: number;
  avg_posts_per_user: number;
}

export interface UserActivity extends BaseEntity {
  user_id: string;
  activity_type: 'login' | 'logout' | 'post_created' | 'comment_created' | 'like_given' | 'follow' | 'unfollow' | 'profile_updated';
  metadata?: Record<string, any>;
  ip_address: string;
  user_agent: string;
  location?: Location;
}

export interface UserBehavior {
  user_id: string;
  interests: string[];
  preferred_content_types: string[];
  engagement_rate: number;
  avg_session_duration: number;
  daily_active_hours: number[];
  content_creation_frequency: number;
  interaction_patterns: Record<string, number>;
  location_data?: Location;
  device_preferences: string[];
  language_preferences: string[];
}

export interface UserConnection extends BaseEntity {
  follower_id: string;
  following_id: string;
  status: 'active' | 'blocked' | 'muted';
  followed_at: string;
  unfollowed_at?: string;
  connection_type: 'follow' | 'friend' | 'block' | 'mute';
}

export interface FollowRequest extends BaseEntity {
  from_user_id: string;
  to_user_id: string;
  status: 'pending' | 'accepted' | 'rejected' | 'cancelled';
  message?: string;
  responded_at?: string;
}

export interface UserReport extends BaseEntity {
  reported_user_id: string;
  reporter_id: string;
  reason: ReportReason;
  description: string;
  status: 'pending' | 'investigating' | 'resolved' | 'rejected';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  assigned_to?: string;
  resolved_at?: string;
  resolution?: string;
  evidence?: Media[];
}

export type ReportReason = 
  | 'harassment'
  | 'spam'
  | 'fake_account'
  | 'inappropriate_content'
  | 'impersonation'
  | 'violence'
  | 'hate_speech'
  | 'self_harm'
  | 'misinformation'
  | 'copyright'
  | 'other';

export interface UserSuspension extends BaseEntity {
  user_id: string;
  suspended_by: string;
  reason: string;
  duration: number; // in hours, 0 for permanent
  expires_at?: string;
  is_active: boolean;
  lifted_at?: string;
  lifted_by?: string;
  lift_reason?: string;
}

export interface UserWarning extends BaseEntity {
  user_id: string;
  issued_by: string;
  reason: string;
  severity: 'low' | 'medium' | 'high';
  acknowledged: boolean;
  acknowledged_at?: string;
  expires_at?: string;
}

export interface UserDevice extends BaseEntity {
  user_id: string;
  device_type: 'mobile' | 'tablet' | 'desktop' | 'tv' | 'watch';
  device_name: string;
  platform: string;
  browser?: string;
  app_version?: string;
  push_token?: string;
  is_active: boolean;
  last_used_at: string;
  location?: Location;
}

export interface UserEngagement {
  user_id: string;
  date: string;
  posts_created: number;
  comments_made: number;
  likes_given: number;
  shares_made: number;
  stories_created: number;
  messages_sent: number;
  time_spent: number; // in minutes
  sessions_count: number;
  unique_interactions: number;
}

// Search and filter types
export interface UserSearchParams {
  query?: string;
  status?: 'active' | 'inactive' | 'suspended' | 'pending';
  role?: string;
  is_verified?: boolean;
  account_type?: 'personal' | 'business' | 'creator';
  subscription_tier?: 'free' | 'premium' | 'pro';
  registration_date_from?: string;
  registration_date_to?: string;
  last_login_from?: string;
  last_login_to?: string;
  location?: string;
  has_posts?: boolean;
  min_followers?: number;
  max_followers?: number;
  sort_by?: 'created_at' | 'last_login_at' | 'follower_count' | 'post_count';
  sort_order?: 'asc' | 'desc';
  page?: number;
  limit?: number;
}

export interface UserFilters {
  status: string[];
  roles: string[];
  account_types: string[];
  verification_status: boolean | null;
  subscription_tiers: string[];
  date_range: {
    start?: string;
    end?: string;
  };
  location: string[];
  engagement_level: 'low' | 'medium' | 'high' | null;
}

// Form types for user management
export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  display_name?: string;
  bio?: string;
  phone?: string;
  date_of_birth?: string;
  gender?: string;
  role: string;
  is_verified?: boolean;
  account_type?: 'personal' | 'business' | 'creator';
  send_welcome_email?: boolean;
}

export interface UpdateUserRequest {
  first_name?: string;
  last_name?: string;
  display_name?: string;
  bio?: string;
  phone?: string;
  date_of_birth?: string;
  gender?: string;
  location?: string;
  website?: string;
  social_links?: SocialLinks;
  role?: string;
  is_verified?: boolean;
  account_type?: 'personal' | 'business' | 'creator';
  status?: 'active' | 'inactive' | 'suspended';
}

export interface BulkUserAction {
  action: 'suspend' | 'unsuspend' | 'verify' | 'unverify' | 'delete' | 'activate' | 'deactivate';
  user_ids: string[];
  reason?: string;
  duration?: number; // for suspension
  send_notification?: boolean;
}

export interface UserImportData {
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  phone?: string;
  role?: string;
  send_welcome_email?: boolean;
}

export interface UserExportOptions {
  format: 'csv' | 'excel' | 'json';
  fields: string[];
  filters?: UserFilters;
  include_private_data?: boolean;
}

// API response types
export type UsersResponse = PaginatedResponse<UserProfile>;
export type UserActivitiesResponse = PaginatedResponse<UserActivity>;
export type UserConnectionsResponse = PaginatedResponse<UserConnection>;
export type FollowRequestsResponse = PaginatedResponse<FollowRequest>;
export type UserReportsResponse = PaginatedResponse<UserReport>;

// Analytics types
export interface UserAnalytics {
  growth: {
    daily: Array<{ date: string; new_users: number; active_users: number }>;
    weekly: Array<{ week: string; new_users: number; active_users: number }>;
    monthly: Array<{ month: string; new_users: number; active_users: number }>;
  };
  demographics: {
    age_groups: Array<{ range: string; count: number; percentage: number }>;
    gender: Array<{ gender: string; count: number; percentage: number }>;
    locations: Array<{ country: string; count: number; percentage: number }>;
  };
  engagement: {
    daily_active_users: number;
    weekly_active_users: number;
    monthly_active_users: number;
    avg_session_duration: number;
    retention_rates: {
      day_1: number;
      day_7: number;
      day_30: number;
    };
  };
}