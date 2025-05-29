import { BaseEntity, Media, Location, Visibility, PaginatedResponse } from './common';
import { User } from './auth';

export interface Story extends BaseEntity {
  author_id: string;
  content?: string;
  content_type: 'text' | 'image' | 'video' | 'audio';
  media: Media[];
  
  // Duration and timing
  duration: number; // in seconds
  expires_at: string;
  is_expired: boolean;
  
  // Visibility and access
  visibility: Visibility;
  allowed_viewers: string[]; // user IDs for custom visibility
  blocked_viewers: string[]; // user IDs who can't view
  
  // Interaction settings
  allow_replies: boolean;
  allow_reactions: boolean;
  allow_sharing: boolean;
  allow_screenshot: boolean;
  
  // Visual styling
  background_color?: string;
  text_color?: string;
  font_family?: string;
  font_size?: number;
  text_alignment?: 'left' | 'center' | 'right';
  
  // Interactive elements
  stickers: StorySticker[];
  polls: StoryPoll[];
  questions: StoryQuestion[];
  mentions: string[];
  hashtags: string[];
  location?: Location;
  
  // Music and audio
  music?: {
    title: string;
    artist: string;
    url: string;
    start_time?: number; // seconds
    duration?: number; // seconds
  };
  
  // Analytics
  view_count: number;
  unique_view_count: number;
  reaction_count: number;
  reply_count: number;
  share_count: number;
  screenshot_count: number;
  completion_rate: number; // percentage of viewers who watched till end
  
  // Status
  status: 'active' | 'expired' | 'archived' | 'deleted' | 'flagged';
  is_featured: boolean;
  is_highlighted: boolean;
  
  // Moderation
  moderation_status: 'approved' | 'pending' | 'rejected' | 'flagged';
  moderated_by?: string;
  moderated_at?: string;
  
  // Relationships
  author?: User;
  views?: StoryView[];
  reactions?: StoryReaction[];
  replies?: StoryReply[];
}

export interface StorySticker {
  id: string;
  type: 'emoji' | 'gif' | 'location' | 'mention' | 'hashtag' | 'time' | 'weather' | 'custom';
  content: string;
  x: number; // position percentage (0-100)
  y: number; // position percentage (0-100)
  width?: number;
  height?: number;
  rotation?: number; // degrees
  scale?: number; // 0.1 to 2.0
  z_index?: number;
  animation?: 'none' | 'bounce' | 'pulse' | 'shake' | 'rotate';
}

export interface StoryPoll {
  id: string;
  question: string;
  options: Array<{
    id: string;
    text: string;
    vote_count: number;
    percentage: number;
  }>;
  total_votes: number;
  expires_at?: string;
  x: number;
  y: number;
}

export interface StoryQuestion {
  id: string;
  question: string;
  responses: Array<{
    id: string;
    user_id: string;
    answer: string;
    created_at: string;
  }>;
  x: number;
  y: number;
}

export interface StoryView extends BaseEntity {
  story_id: string;
  viewer_id?: string; // null for anonymous views
  session_id: string;
  view_duration: number; // seconds
  completion_percentage: number;
  source: 'timeline' | 'profile' | 'direct' | 'story_ring';
  device_type: 'mobile' | 'tablet' | 'desktop';
  ip_address: string;
  location?: Location;
}

export interface StoryReaction extends BaseEntity {
  story_id: string;
  user_id: string;
  reaction_type: 'like' | 'love' | 'laugh' | 'wow' | 'sad' | 'angry' | 'fire' | 'clap';
  is_active: boolean;
}

export interface StoryReply extends BaseEntity {
  story_id: string;
  user_id: string;
  content: string;
  content_type: 'text' | 'image' | 'voice';
  media?: Media;
  is_read: boolean;
  read_at?: string;
}

export interface StoryHighlight extends BaseEntity {
  user_id: string;
  title: string;
  cover_image: string;
  story_ids: string[];
  is_public: boolean;
  view_count: number;
  stories?: Story[];
}

export interface StoryTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  background_type: 'color' | 'gradient' | 'image';
  background_value: string;
  text_settings: {
    color: string;
    font_family: string;
    font_size: number;
    alignment: 'left' | 'center' | 'right';
  };
  stickers: StorySticker[];
  is_premium: boolean;
  usage_count: number;
}

export interface StoryAnalytics {
  story_id: string;
  date: string;
  views: number;
  unique_views: number;
  reactions: number;
  replies: number;
  shares: number;
  screenshots: number;
  completion_rate: number;
  avg_view_duration: number;
  reach: number;
  impressions: number;
  engagement_rate: number;
  
  // Demographic breakdown
  viewer_demographics: {
    age_groups: Record<string, number>;
    genders: Record<string, number>;
    locations: Record<string, number>;
  };
  
  // Device and source analytics
  device_breakdown: Record<string, number>;
  source_breakdown: Record<string, number>;
  
  // Time-based analytics
  view_times: Array<{
    hour: number;
    view_count: number;
  }>;
}

// Search and filter types
export interface StorySearchParams {
  query?: string;
  author_id?: string;
  content_type?: 'text' | 'image' | 'video' | 'audio';
  visibility?: Visibility;
  status?: 'active' | 'expired' | 'archived' | 'deleted' | 'flagged';
  moderation_status?: 'approved' | 'pending' | 'rejected' | 'flagged';
  has_media?: boolean;
  has_music?: boolean;
  has_location?: boolean;
  is_featured?: boolean;
  is_highlighted?: boolean;
  date_from?: string;
  date_to?: string;
  min_views?: number;
  max_views?: number;
  hashtags?: string[];
  mentions?: string[];
  location?: string;
  sort_by?: 'created_at' | 'view_count' | 'reaction_count' | 'completion_rate';
  sort_order?: 'asc' | 'desc';
  page?: number;
  limit?: number;
}

export interface StoryFilters {
  content_types: string[];
  visibility_levels: string[];
  statuses: string[];
  moderation_statuses: string[];
  authors: string[];
  date_range: {
    start?: string;
    end?: string;
  };
  view_count_range: {
    min?: number;
    max?: number;
  };
  duration_range: {
    min?: number;
    max?: number;
  };
  has_media: boolean | null;
  has_music: boolean | null;
  has_location: boolean | null;
  is_featured: boolean | null;
  hashtags: string[];
  mentions: string[];
  locations: string[];
}

// Form types
export interface CreateStoryRequest {
  content?: string;
  content_type: 'text' | 'image' | 'video' | 'audio';
  media?: string[]; // media IDs
  duration?: number;
  visibility: Visibility;
  allowed_viewers?: string[];
  blocked_viewers?: string[];
  allow_replies?: boolean;
  allow_reactions?: boolean;
  allow_sharing?: boolean;
  allow_screenshot?: boolean;
  background_color?: string;
  text_color?: string;
  font_family?: string;
  stickers?: Omit<StorySticker, 'id'>[];
  polls?: Omit<StoryPoll, 'id' | 'total_votes'>[];
  questions?: Omit<StoryQuestion, 'id' | 'responses'>[];
  mentions?: string[];
  hashtags?: string[];
  location?: Location;
  music?: {
    title: string;
    artist: string;
    url: string;
    start_time?: number;
    duration?: number;
  };
}

export interface UpdateStoryRequest {
  visibility?: Visibility;
  allowed_viewers?: string[];
  blocked_viewers?: string[];
  allow_replies?: boolean;
  allow_reactions?: boolean;
  allow_sharing?: boolean;
  allow_screenshot?: boolean;
  is_featured?: boolean;
  is_highlighted?: boolean;
}

export interface CreateStoryHighlightRequest {
  title: string;
  story_ids: string[];
  cover_image?: string;
  is_public?: boolean;
}

export interface BulkStoryAction {
  action: 'delete' | 'archive' | 'feature' | 'unfeature' | 'approve' | 'reject';
  story_ids: string[];
  reason?: string;
}

// API response types
export type StoriesResponse = PaginatedResponse<Story>;
export type StoryViewsResponse = PaginatedResponse<StoryView>;
export type StoryReactionsResponse = PaginatedResponse<StoryReaction>;
export type StoryRepliesResponse = PaginatedResponse<StoryReply>;
export type StoryHighlightsResponse = PaginatedResponse<StoryHighlight>;
export type StoryAnalyticsResponse = PaginatedResponse<StoryAnalytics>;

// Statistics types
export interface StoryStats {
  total_stories: number;
  active_stories: number;
  expired_stories: number;
  archived_stories: number;
  flagged_stories: number;
  stories_today: number;
  stories_week: number;
  stories_month: number;
  avg_views_per_story: number;
  avg_completion_rate: number;
  total_story_views: number;
  unique_story_viewers: number;
  
  content_type_distribution: Record<string, number>;
  avg_duration_by_type: Record<string, number>;
  engagement_by_type: Record<string, {
    avg_views: number;
    avg_reactions: number;
    avg_replies: number;
    completion_rate: number;
  }>;
  
  top_story_creators: Array<{
    user_id: string;
    username: string;
    story_count: number;
    total_views: number;
    avg_engagement: number;
  }>;
}

export interface StoryTrends {
  trending_hashtags: Array<{
    hashtag: string;
    story_count: number;
    view_count: number;
    growth_rate: number;
  }>;
  
  popular_music: Array<{
    title: string;
    artist: string;
    usage_count: number;
    total_views: number;
  }>;
  
  trending_locations: Array<{
    location: string;
    story_count: number;
    unique_users: number;
  }>;
  
  viral_stories: Array<Story & {
    viral_score: number;
    growth_velocity: number;
  }>;
}

export interface StoryModeration {
  flagged_stories: Story[];
  pending_review: Story[];
  auto_moderation_results: Array<{
    story_id: string;
    confidence_score: number;
    flagged_reasons: string[];
    action_taken: string;
  }>;
  user_reports: Array<{
    story_id: string;
    report_count: number;
    most_common_reason: string;
  }>;
}