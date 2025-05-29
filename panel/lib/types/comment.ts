import { BaseEntity, SoftDeleteEntity, Media, PaginatedResponse } from './common';
import { User } from './auth';
import { Post } from './post';

export interface Comment extends SoftDeleteEntity {
  post_id: string;
  author_id: string;
  parent_comment_id?: string; // for nested comments/replies
  content: string;
  content_type: 'text' | 'image' | 'gif' | 'sticker' | 'emoji_reaction';
  
  // Engagement metrics
  like_count: number;
  reply_count: number;
  report_count: number;
  
  // Content details
  mentions: string[];
  hashtags: string[];
  media?: Media[];
  
  // Status and moderation
  status: 'published' | 'pending' | 'approved' | 'rejected' | 'flagged' | 'deleted';
  is_pinned: boolean;
  is_featured: boolean;
  moderation_status: 'approved' | 'pending' | 'rejected' | 'flagged' | 'auto_approved';
  moderated_by?: string;
  moderated_at?: string;
  moderation_reason?: string;
  
  // Analytics
  engagement_score: number;
  sentiment_score?: number; // -1 to 1, negative to positive
  toxicity_score?: number; // 0 to 1, safe to toxic
  
  // Threading and hierarchy
  thread_level: number; // 0 for top-level, 1+ for nested
  thread_root_id?: string; // points to the root comment of the thread
  
  // Editing history
  is_edited: boolean;
  edited_at?: string;
  edit_count: number;
  
  // Relationships
  author?: User;
  post?: Post;
  parent_comment?: Comment;
  replies?: Comment[];
  reactions?: CommentReaction[];
}

export interface CommentReaction extends BaseEntity {
  comment_id: string;
  user_id: string;
  reaction_type: 'like' | 'love' | 'laugh' | 'wow' | 'sad' | 'angry' | 'dislike';
  is_active: boolean;
}

export interface CommentReport extends BaseEntity {
  comment_id: string;
  reporter_id: string;
  reason: CommentReportReason;
  description: string;
  status: 'pending' | 'investigating' | 'resolved' | 'rejected';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  assigned_to?: string;
  resolved_at?: string;
  resolution?: string;
  action_taken?: CommentModerationAction;
  automated: boolean;
  confidence_score?: number;
}

export type CommentReportReason = 
  | 'spam'
  | 'harassment'
  | 'hate_speech'
  | 'bullying'
  | 'inappropriate_content'
  | 'violence'
  | 'self_harm'
  | 'misinformation'
  | 'scam'
  | 'impersonation'
  | 'off_topic'
  | 'copyright'
  | 'other';

export type CommentModerationAction = 
  | 'no_action'
  | 'comment_removed'
  | 'comment_hidden'
  | 'user_warned'
  | 'user_suspended'
  | 'comment_flagged'
  | 'escalated';

export interface CommentModeration extends BaseEntity {
  comment_id: string;
  moderator_id: string;
  action: CommentModerationAction;
  reason: string;
  previous_status: string;
  new_status: string;
  notes?: string;
  automated: boolean;
  confidence_score?: number;
  appeal_eligible: boolean;
}

export interface CommentEdit extends BaseEntity {
  comment_id: string;
  old_content: string;
  new_content: string;
  edit_reason?: string;
  edited_by: string; // usually the author, but can be a moderator
  is_moderator_edit: boolean;
}

export interface CommentThread {
  root_comment: Comment;
  replies: Comment[];
  total_replies: number;
  has_more: boolean;
  depth: number;
  engagement_summary: {
    total_likes: number;
    total_replies: number;
    unique_participants: number;
    avg_sentiment: number;
  };
}

export interface CommentAnalytics {
  comment_id: string;
  date: string;
  views: number;
  likes: number;
  replies: number;
  reports: number;
  engagement_rate: number;
  sentiment_score: number;
  toxicity_score: number;
  response_time: number; // time between post and comment in minutes
  author_influence_score: number;
}

// Search and filter types
export interface CommentSearchParams {
  query?: string;
  post_id?: string;
  author_id?: string;
  parent_comment_id?: string;
  status?: 'published' | 'pending' | 'approved' | 'rejected' | 'flagged';
  moderation_status?: 'approved' | 'pending' | 'rejected' | 'flagged';
  content_type?: 'text' | 'image' | 'gif' | 'sticker';
  is_pinned?: boolean;
  has_media?: boolean;
  has_replies?: boolean;
  thread_level?: number;
  date_from?: string;
  date_to?: string;
  min_likes?: number;
  max_likes?: number;
  min_replies?: number;
  max_replies?: number;
  sentiment?: 'positive' | 'neutral' | 'negative';
  toxicity_threshold?: number;
  hashtags?: string[];
  mentions?: string[];
  sort_by?: 'created_at' | 'like_count' | 'reply_count' | 'engagement_score';
  sort_order?: 'asc' | 'desc';
  page?: number;
  limit?: number;
}

export interface CommentFilters {
  statuses: string[];
  moderation_statuses: string[];
  content_types: string[];
  authors: string[];
  posts: string[];
  date_range: {
    start?: string;
    end?: string;
  };
  engagement_range: {
    min_likes?: number;
    max_likes?: number;
    min_replies?: number;
    max_replies?: number;
  };
  sentiment_filter: 'positive' | 'neutral' | 'negative' | null;
  toxicity_range: {
    min?: number;
    max?: number;
  };
  thread_level_range: {
    min?: number;
    max?: number;
  };
  has_media: boolean | null;
  is_pinned: boolean | null;
  is_edited: boolean | null;
  hashtags: string[];
  mentions: string[];
}

// Form types
export interface CreateCommentRequest {
  post_id: string;
  parent_comment_id?: string;
  content: string;
  content_type?: 'text' | 'image' | 'gif' | 'sticker';
  mentions?: string[];
  media?: string[]; // media IDs
}

export interface UpdateCommentRequest {
  content?: string;
  edit_reason?: string;
  mentions?: string[];
}

export interface BulkCommentAction {
  action: 'delete' | 'approve' | 'reject' | 'flag' | 'unflag' | 'pin' | 'unpin';
  comment_ids: string[];
  reason?: string;
  send_notification?: boolean;
}

export interface CommentModerationRequest {
  action: CommentModerationAction;
  reason: string;
  notes?: string;
  notify_author?: boolean;
  escalate?: boolean;
}

// API response types
export type CommentsResponse = PaginatedResponse<Comment>;
export type CommentReactionsResponse = PaginatedResponse<CommentReaction>;
export type CommentReportsResponse = PaginatedResponse<CommentReport>;
export type CommentThreadsResponse = PaginatedResponse<CommentThread>;
export type CommentAnalyticsResponse = PaginatedResponse<CommentAnalytics>;

// Statistics types
export interface CommentStats {
  total_comments: number;
  published_comments: number;
  pending_comments: number;
  flagged_comments: number;
  deleted_comments: number;
  comments_today: number;
  comments_week: number;
  comments_month: number;
  avg_comments_per_post: number;
  avg_response_time: number; // in minutes
  top_commenters: Array<{
    user_id: string;
    username: string;
    comment_count: number;
    avg_engagement: number;
  }>;
  sentiment_distribution: {
    positive: number;
    neutral: number;
    negative: number;
  };
  toxicity_levels: {
    safe: number;
    moderate: number;
    high: number;
  };
}

// Moderation queue types
export interface CommentModerationQueue {
  pending_approval: Comment[];
  flagged_content: Comment[];
  reported_comments: Array<Comment & {
    report_count: number;
    latest_report: CommentReport;
  }>;
  high_toxicity: Comment[];
  escalated_issues: Array<Comment & {
    escalation_reason: string;
    escalated_by: string;
    escalated_at: string;
  }>;
}

// Auto-moderation types
export interface CommentAutoModerationRule {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  priority: number;
  conditions: Array<{
    type: 'keyword' | 'pattern' | 'sentiment' | 'toxicity' | 'length' | 'caps_ratio';
    operator: 'contains' | 'equals' | 'starts_with' | 'ends_with' | 'greater_than' | 'less_than';
    value: any;
    case_sensitive?: boolean;
  }>;
  actions: Array<{
    type: 'flag' | 'hold_for_review' | 'auto_delete' | 'warn_user' | 'reduce_visibility';
    parameters?: Record<string, any>;
  }>;
  exceptions: string[]; // user IDs or roles exempt from this rule
  created_by: string;
}

export interface CommentAutoModerationResult {
  comment_id: string;
  rule_id: string;
  rule_name: string;
  action_taken: string;
  confidence_score: number;
  triggered_conditions: string[];
  requires_human_review: boolean;
  processed_at: string;
}

// Engagement insights
export interface CommentEngagementInsights {
  most_engaging_comments: Array<Comment & {
    engagement_score: number;
    viral_coefficient: number;
  }>;
  conversation_starters: Array<Comment & {
    reply_tree_size: number;
    unique_participants: number;
  }>;
  sentiment_trends: Array<{
    date: string;
    positive_ratio: number;
    negative_ratio: number;
    neutral_ratio: number;
  }>;
  user_behavior_patterns: Array<{
    user_id: string;
    avg_comment_length: number;
    preferred_content_types: string[];
    peak_activity_hours: number[];
    engagement_style: 'conversational' | 'reactive' | 'informative';
  }>;
}