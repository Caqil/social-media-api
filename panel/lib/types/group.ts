import { BaseEntity, SoftDeleteEntity, Location, Media, PaginatedResponse } from './common';
import { User } from './auth';

export interface Group extends SoftDeleteEntity {
  name: string;
  slug: string;
  description: string;
  privacy: 'public' | 'private' | 'secret';
  category: string;
  subcategory?: string;
  tags: string[];
  
  // Media
  avatar?: string;
  cover_image?: string;
  banner_image?: string;
  
  // Location and contact
  location?: Location;
  website?: string;
  email?: string;
  phone?: string;
  
  // Membership
  member_count: number;
  max_members?: number;
  pending_requests_count: number;
  
  // Settings
  post_approval_required: boolean;
  member_approval_required: boolean;
  allow_member_invites: boolean;
  allow_external_sharing: boolean;
  allow_polls: boolean;
  allow_events: boolean;
  allow_discussions: boolean;
  allow_media_posts: boolean;
  
  // Moderation
  content_moderation_level: 'strict' | 'moderate' | 'relaxed';
  auto_moderation_enabled: boolean;
  profanity_filter_enabled: boolean;
  spam_detection_enabled: boolean;
  
  // Guidelines
  rules: string[];
  guidelines: string;
  welcome_message?: string;
  
  // Status
  status: 'active' | 'inactive' | 'suspended' | 'archived';
  is_verified: boolean;
  is_featured: boolean;
  
  // Analytics
  engagement_rate: number;
  activity_score: number;
  growth_rate: number;
  
  // Metadata
  founded_at: string;
  founder_id: string;
  language: string;
  time_zone: string;
  
  // Relationships
  founder?: User;
  admins?: GroupMember[];
  moderators?: GroupMember[];
}

export interface GroupMember extends BaseEntity {
  group_id: string;
  user_id: string;
  role: 'admin' | 'moderator' | 'member';
  status: 'active' | 'banned' | 'left' | 'pending';
  joined_at: string;
  invited_by?: string;
  invitation_message?: string;
  
  // Member settings
  notifications_enabled: boolean;
  nickname?: string;
  title?: string;
  bio?: string;
  
  // Permissions
  can_post: boolean;
  can_comment: boolean;
  can_invite_members: boolean;
  can_approve_posts: boolean;
  can_moderate_comments: boolean;
  
  // Analytics
  posts_count: number;
  comments_count: number;
  likes_given: number;
  activity_score: number;
  last_active_at: string;
  
  // Relationships
  user?: User;
  group?: Group;
}

export interface GroupInvitation extends BaseEntity {
  group_id: string;
  inviter_id: string;
  invitee_id: string;
  message?: string;
  status: 'pending' | 'accepted' | 'rejected' | 'expired';
  expires_at: string;
  responded_at?: string;
  
  // Relationships
  group?: Group;
  inviter?: User;
  invitee?: User;
}

export interface GroupJoinRequest extends BaseEntity {
  group_id: string;
  user_id: string;
  message?: string;
  status: 'pending' | 'approved' | 'rejected';
  reviewed_by?: string;
  reviewed_at?: string;
  review_note?: string;
  
  // Relationships
  group?: Group;
  user?: User;
  reviewer?: User;
}

export interface GroupPost extends BaseEntity {
  group_id: string;
  author_id: string;
  content: string;
  content_type: 'text' | 'image' | 'video' | 'poll' | 'event' | 'link';
  media: Media[];
  hashtags: string[];
  mentions: string[];
  
  // Status
  status: 'published' | 'pending' | 'approved' | 'rejected' | 'deleted';
  is_pinned: boolean;
  is_announcement: boolean;
  
  // Engagement
  like_count: number;
  comment_count: number;
  share_count: number;
  
  // Moderation
  requires_approval: boolean;
  approved_by?: string;
  approved_at?: string;
  rejection_reason?: string;
  
  // Relationships
  group?: Group;
  author?: User;
}

export interface GroupEvent extends BaseEntity {
  group_id: string;
  organizer_id: string;
  title: string;
  description: string;
  event_type: 'online' | 'in_person' | 'hybrid';
  
  // Timing
  start_date: string;
  end_date: string;
  timezone: string;
  
  // Location (for in-person events)
  location?: Location;
  venue_name?: string;
  venue_address?: string;
  
  // Online event details
  meeting_url?: string;
  meeting_id?: string;
  meeting_password?: string;
  
  // Settings
  max_attendees?: number;
  registration_required: boolean;
  is_public: boolean;
  allow_guests: boolean;
  
  // Status
  status: 'draft' | 'published' | 'cancelled' | 'completed';
  attendee_count: number;
  waitlist_count: number;
  
  // Media
  cover_image?: string;
  images: string[];
  
  // Relationships
  group?: Group;
  organizer?: User;
}

export interface GroupAnalytics {
  group_id: string;
  date: string;
  member_count: number;
  new_members: number;
  active_members: number;
  posts_count: number;
  comments_count: number;
  likes_count: number;
  shares_count: number;
  engagement_rate: number;
  avg_session_duration: number;
  top_contributors: Array<{
    user_id: string;
    posts: number;
    comments: number;
    likes: number;
  }>;
}

// Search and filter types
export interface GroupSearchParams {
  query?: string;
  category?: string;
  subcategory?: string;
  privacy?: 'public' | 'private';
  tags?: string[];
  location?: string;
  min_members?: number;
  max_members?: number;
  is_verified?: boolean;
  is_featured?: boolean;
  has_recent_activity?: boolean;
  founded_after?: string;
  founded_before?: string;
  sort_by?: 'created_at' | 'member_count' | 'activity_score' | 'name';
  sort_order?: 'asc' | 'desc';
  page?: number;
  limit?: number;
}

export interface GroupFilters {
  categories: string[];
  subcategories: string[];
  privacy_levels: string[];
  member_count_ranges: Array<{
    min: number;
    max: number;
    label: string;
  }>;
  tags: string[];
  locations: string[];
  verification_status: boolean | null;
  featured_status: boolean | null;
  activity_levels: string[];
  founding_date_range: {
    start?: string;
    end?: string;
  };
}

// Form types
export interface CreateGroupRequest {
  name: string;
  description: string;
  privacy: 'public' | 'private' | 'secret';
  category: string;
  subcategory?: string;
  tags: string[];
  location?: Location;
  website?: string;
  rules: string[];
  guidelines?: string;
  welcome_message?: string;
  post_approval_required?: boolean;
  member_approval_required?: boolean;
  allow_member_invites?: boolean;
  max_members?: number;
  avatar?: string;
  cover_image?: string;
}

export interface UpdateGroupRequest {
  name?: string;
  description?: string;
  privacy?: 'public' | 'private' | 'secret';
  category?: string;
  subcategory?: string;
  tags?: string[];
  location?: Location;
  website?: string;
  rules?: string[];
  guidelines?: string;
  welcome_message?: string;
  post_approval_required?: boolean;
  member_approval_required?: boolean;
  allow_member_invites?: boolean;
  max_members?: number;
  avatar?: string;
  cover_image?: string;
  status?: 'active' | 'inactive' | 'archived';
}

export interface InviteToGroupRequest {
  user_ids: string[];
  message?: string;
}

export interface JoinGroupRequest {
  message?: string;
}

export interface UpdateMemberRoleRequest {
  user_id: string;
  role: 'admin' | 'moderator' | 'member';
  permissions?: {
    can_post?: boolean;
    can_comment?: boolean;
    can_invite_members?: boolean;
    can_approve_posts?: boolean;
    can_moderate_comments?: boolean;
  };
}

export interface BulkGroupAction {
  action: 'archive' | 'unarchive' | 'feature' | 'unfeature' | 'verify' | 'unverify' | 'delete';
  group_ids: string[];
  reason?: string;
}

// API response types
export type GroupsResponse = PaginatedResponse<Group>;
export type GroupMembersResponse = PaginatedResponse<GroupMember>;
export type GroupInvitationsResponse = PaginatedResponse<GroupInvitation>;
export type GroupJoinRequestsResponse = PaginatedResponse<GroupJoinRequest>;
export type GroupPostsResponse = PaginatedResponse<GroupPost>;
export type GroupEventsResponse = PaginatedResponse<GroupEvent>;
export type GroupAnalyticsResponse = PaginatedResponse<GroupAnalytics>;

// Statistics types
export interface GroupStats {
  total_groups: number;
  public_groups: number;
  private_groups: number;
  secret_groups: number;
  verified_groups: number;
  featured_groups: number;
  active_groups: number;
  groups_created_today: number;
  groups_created_week: number;
  groups_created_month: number;
  avg_members_per_group: number;
  avg_posts_per_group: number;
  most_popular_categories: Array<{
    category: string;
    count: number;
    growth_rate: number;
  }>;
  engagement_metrics: {
    avg_posts_per_day: number;
    avg_comments_per_post: number;
    avg_member_activity_rate: number;
  };
}

export interface GroupCategory {
  id: string;
  name: string;
  description: string;
  icon?: string;
  color?: string;
  subcategories: Array<{
    id: string;
    name: string;
    description: string;
  }>;
  group_count: number;
  is_featured: boolean;
  sort_order: number;
}

export interface GroupRecommendation {
  group: Group;
  reason: 'similar_interests' | 'mutual_friends' | 'location_based' | 'trending' | 'category_match';
  score: number;
  explanation: string;
  mutual_members?: User[];
}

export interface GroupInsights {
  growth_trend: Array<{
    date: string;
    new_members: number;
    total_members: number;
  }>;
  engagement_trend: Array<{
    date: string;
    posts: number;
    comments: number;
    likes: number;
    active_members: number;
  }>;
  member_demographics: {
    age_groups: Record<string, number>;
    locations: Record<string, number>;
    join_sources: Record<string, number>;
  };
  content_performance: {
    top_posts: GroupPost[];
    popular_topics: Array<{
      topic: string;
      mentions: number;
      engagement: number;
    }>;
    posting_patterns: Array<{
      hour: number;
      post_count: number;
      engagement_rate: number;
    }>;
  };
}