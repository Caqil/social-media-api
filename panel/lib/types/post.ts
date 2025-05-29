import { BaseEntity, SoftDeleteEntity, Media, Location, Visibility, PaginatedResponse } from './common';
import { User } from './auth';

export interface Post extends SoftDeleteEntity {
  author_id: string;
  content: string;
  content_type: 'text' | 'image' | 'video' | 'audio' | 'poll' | 'link' | 'event';
  type: 'regular' | 'story' | 'reel' | 'live' | 'announcement';
  visibility: Visibility;
  status: 'draft' | 'published' | 'archived' | 'deleted' | 'flagged' | 'under_review';
  language: string;
  hashtags: string[];
  mentions: string[];
  media: Media[];
  location?: Location;
  
  // Engagement metrics
  like_count: number;
  comment_count: number;
  share_count: number;
  view_count: number;
  save_count: number;
  report_count: number;
  
  // Settings
  comments_enabled: boolean;
  likes_enabled: boolean;
  shares_enabled: boolean;
  save_enabled: boolean;
  
  // Scheduling
  published_at?: string;
  scheduled_for?: string;
  expires_at?: string;
  
  // Moderation
  is_pinned: boolean;
  is_featured: boolean;
  is_promoted: boolean;
  moderation_status: 'approved' | 'pending' | 'rejected' | 'flagged';
  moderated_by?: string;
  moderated_at?: string;
  moderation_reason?: string;
  
  // Analytics
  reach: number;
  impressions: number;
  engagement_rate: number;
  click_through_rate?: number;
  
  // Group/Community posts
  group_id?: string;
  event_id?: string;
  
  // Threading
  parent_post_id?: string; // for reposts/quotes
  original_post_id?: string; // for shares
  
  // Poll data
  poll_options?: PollOption[];
  poll_expires_at?: string;
  poll_multiple_choice?: boolean;
  poll_show_results?: boolean;
  
  // Relationships
  author?: User;
  group?: any; // Will be defined in group types
  event?: any; // Will be defined in event types
  parent_post?: Post;
  original_post?: Post;
}

export interface PollOption {
  id: string;
  text: string;
  vote_count: number;
  percentage: number;
  media?: Media;
}

export interface PostDraft extends Omit<Post, 'id' | 'created_at' | 'updated_at' | 'author' | 'like_count' | 'comment_count' | 'share_count' | 'view_count'> {
  id?: string;
  title?: string;
  auto_save_at?: string;
}

export interface PostEngagement extends BaseEntity {
  post_id: string;
  user_id: string;
  type: 'like' | 'comment' | 'share' | 'save' | 'view' | 'click' | 'poll_vote';
  reaction_type?: 'like' | 'love' | 'laugh' | 'wow' | 'sad' | 'angry';
  metadata?: Record<string, any>;
  duration?: number; // for view duration
}

export interface PostView extends BaseEntity {
  post_id: string;
  user_id?: string; // null for anonymous views
  session_id: string;
  duration: number; // in seconds
  source: 'feed' | 'profile' | 'search' | 'direct' | 'share';
  device_type: 'mobile' | 'tablet' | 'desktop';
  location?: Location;
}

export interface PostShare extends BaseEntity {
  post_id: string;
  user_id: string;
  platform: 'internal' | 'facebook' | 'twitter' | 'instagram' | 'linkedin' | 'whatsapp' | 'telegram' | 'email' | 'copy_link';
  message?: string;
  recipients?: string[]; // for internal shares
}

export interface PostReport extends BaseEntity {
  post_id: string;
  reporter_id: string;
  reason: PostReportReason;
  description: string;
  status: 'pending' | 'investigating' | 'resolved' | 'rejected';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  assigned_to?: string;
  resolved_at?: string;
  resolution?: string;
  action_taken?: PostModerationAction;
}

export type PostReportReason = 
  | 'spam'
  | 'harassment'
  | 'hate_speech'
  | 'violence'
  | 'nudity'
  | 'misinformation'
  | 'copyright'
  | 'self_harm'
  | 'inappropriate_content'
  | 'fake_news'
  | 'scam'
  | 'other';

export type PostModerationAction = 
  | 'no_action'
  | 'warning_issued'
  | 'content_removed'
  | 'user_suspended'
  | 'account_deleted'
  | 'content_flagged'
  | 'visibility_reduced';

export interface PostModeration extends BaseEntity {
  post_id: string;
  moderator_id: string;
  action: PostModerationAction;
  reason: string;
  previous_status: string;
  new_status: string;
  notes?: string;
  automated: boolean;
  confidence_score?: number; // for automated moderation
}

export interface PostAnalytics {
  post_id: string;
  date: string;
  views: number;
  unique_views: number;
  likes: number;
  comments: number;
  shares: number;
  saves: number;
  reach: number;
  impressions: number;
  engagement_rate: number;
  click_rate: number;
  completion_rate?: number; // for videos
  demographics: {
    age_groups: Record<string, number>;
    genders: Record<string, number>;
    locations: Record<string, number>;
  };
  traffic_sources: Record<string, number>;
  devices: Record<string, number>;
}

// Search and filter types
export interface PostSearchParams {
  query?: string;
  author_id?: string;
  group_id?: string;
  content_type?: 'text' | 'image' | 'video' | 'audio' | 'poll' | 'link';
  post_type?: 'regular' | 'story' | 'reel' | 'live' | 'announcement';
  status?: 'draft' | 'published' | 'archived' | 'deleted' | 'flagged';
  visibility?: Visibility;
  hashtags?: string[];
  has_media?: boolean;
  is_pinned?: boolean;
  is_featured?: boolean;
  moderation_status?: 'approved' | 'pending' | 'rejected' | 'flagged';
  date_from?: string;
  date_to?: string;
  min_likes?: number;
  max_likes?: number;
  min_comments?: number;
  max_comments?: number;
  location?: string;
  language?: string;
  sort_by?: 'created_at' | 'like_count' | 'comment_count' | 'view_count' | 'engagement_rate';
  sort_order?: 'asc' | 'desc';
  page?: number;
  limit?: number;
}

export interface PostFilters {
  content_types: string[];
  post_types: string[];
  statuses: string[];
  visibility_levels: string[];
  moderation_statuses: string[];
  authors: string[];
  groups: string[];
  hashtags: string[];
  date_range: {
    start?: string;
    end?: string;
  };
  engagement_range: {
    min_likes?: number;
    max_likes?: number;
    min_comments?: number;
    max_comments?: number;
  };
  has_media: boolean | null;
  is_pinned: boolean | null;
  is_featured: boolean | null;
  languages: string[];
  locations: string[];
}

// Form types
export interface CreatePostRequest {
  content: string;
  content_type: 'text' | 'image' | 'video' | 'audio' | 'poll' | 'link';
  type?: 'regular' | 'story' | 'reel' | 'announcement';
  visibility: Visibility;
  hashtags?: string[];
  mentions?: string[];
  media?: string[]; // media IDs
  location?: Location;
  group_id?: string;
  event_id?: string;
  scheduled_for?: string;
  expires_at?: string;
  comments_enabled?: boolean;
  likes_enabled?: boolean;
  shares_enabled?: boolean;
  poll_options?: Array<{
    text: string;
    media_id?: string;
  }>;
  poll_expires_at?: string;
  poll_multiple_choice?: boolean;
}

export interface UpdatePostRequest {
  content?: string;
  visibility?: Visibility;
  hashtags?: string[];
  mentions?: string[];
  location?: Location;
  scheduled_for?: string;
  expires_at?: string;
  comments_enabled?: boolean;
  likes_enabled?: boolean;
  shares_enabled?: boolean;
  is_pinned?: boolean;
  is_featured?: boolean;
  status?: 'draft' | 'published' | 'archived';
}

export interface BulkPostAction {
  action: 'delete' | 'archive' | 'restore' | 'pin' | 'unpin' | 'feature' | 'unfeature' | 'approve' | 'reject';
  post_ids: string[];
  reason?: string;
  send_notification?: boolean;
}

export interface PostModerationRequest {
  action: PostModerationAction;
  reason: string;
  notes?: string;
  notify_author?: boolean;
}

// API response types
export type PostsResponse = PaginatedResponse<Post>;
export type PostEngagementsResponse = PaginatedResponse<PostEngagement>;
export type PostViewsResponse = PaginatedResponse<PostView>;
export type PostReportsResponse = PaginatedResponse<PostReport>;
export type PostAnalyticsResponse = PaginatedResponse<PostAnalytics>;

// Statistics types
export interface PostStats {
  total_posts: number;
  published_posts: number;
  draft_posts: number;
  archived_posts: number;
  flagged_posts: number;
  posts_today: number;
  posts_week: number;
  posts_month: number;
  avg_engagement_rate: number;
  top_hashtags: Array<{
    tag: string;
    count: number;
    growth: number;
  }>;
  content_type_distribution: Record<string, number>;
  visibility_distribution: Record<string, number>;
}

export interface TrendingContent {
  hashtags: Array<{
    tag: string;
    post_count: number;
    engagement_rate: number;
    growth_rate: number;
  }>;
  posts: Array<Post & {
    trending_score: number;
    viral_coefficient: number;
  }>;
  topics: Array<{
    topic: string;
    post_count: number;
    engagement: number;
  }>;
}

// Content creation helpers
export interface PostTemplate {
  id: string;
  name: string;
  description: string;
  content_type: string;
  template_data: {
    content?: string;
    hashtags?: string[];
    media_slots?: number;
    poll_options?: string[];
  };
  is_public: boolean;
  usage_count: number;
  created_by: string;
}

export interface ContentSchedule {
  id: string;
  user_id: string;
  title: string;
  posts: Array<{
    post_data: CreatePostRequest;
    scheduled_for: string;
    status: 'pending' | 'published' | 'failed';
  }>;
  frequency: 'once' | 'daily' | 'weekly' | 'monthly';
  is_active: boolean;
}