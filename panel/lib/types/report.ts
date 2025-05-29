import { BaseEntity, Media, Priority, PaginatedResponse } from './common';
import { User } from './auth';

export interface Report extends BaseEntity {
  reporter_id: string;
  target_id: string;
  target_type: 'user' | 'post' | 'comment' | 'story' | 'message' | 'group' | 'event';
  reason: ReportReason;
  category: ReportCategory;
  description: string;
  
  // Evidence
  screenshots: Media[];
  evidence: string[];
  additional_context?: string;
  
  // Status and priority
  status: ReportStatus;
  priority: Priority;
  severity: 'low' | 'medium' | 'high' | 'critical';
  
  // Assignment and handling
  assigned_to?: string;
  assigned_at?: string;
  escalated_to?: string;
  escalated_at?: string;
  escalation_reason?: string;
  
  // Resolution
  resolved_at?: string;
  resolved_by?: string;
  resolution: ReportResolution;
  resolution_note?: string;
  actions_taken: ModerationAction[];
  
  // Follow-up
  follow_up_required: boolean;
  follow_up_date?: string;
  follow_up_note?: string;
  
  // Meta information
  source: 'user_report' | 'auto_detection' | 'moderator_review' | 'bulk_report';
  confidence_score?: number; // for auto-detected reports
  duplicate_of?: string; // reference to original report if this is a duplicate
  related_reports: string[]; // IDs of related reports
  
  // Appeal information
  appeal_id?: string;
  appeal_status?: 'none' | 'pending' | 'approved' | 'rejected';
  
  // Relationships
  reporter?: User;
  assigned_moderator?: User;
  target_data?: any; // The actual reported content/user
}

export type ReportReason = 
  | 'spam'
  | 'harassment'
  | 'bullying'
  | 'hate_speech'
  | 'violence'
  | 'threats'
  | 'self_harm'
  | 'suicide'
  | 'nudity'
  | 'sexual_content'
  | 'child_safety'
  | 'misinformation'
  | 'fake_news'
  | 'scam'
  | 'fraud'
  | 'impersonation'
  | 'identity_theft'
  | 'copyright'
  | 'trademark'
  | 'privacy_violation'
  | 'doxxing'
  | 'drug_related'
  | 'weapon_sales'
  | 'human_trafficking'
  | 'terrorism'
  | 'extremism'
  | 'inappropriate_content'
  | 'off_topic'
  | 'technical_issue'
  | 'other';

export type ReportCategory = 
  | 'safety'
  | 'harassment'
  | 'content_policy'
  | 'intellectual_property'
  | 'privacy'
  | 'security'
  | 'technical'
  | 'community_standards'
  | 'legal_compliance';

export type ReportStatus = 
  | 'pending'
  | 'investigating'
  | 'escalated'
  | 'resolved'
  | 'rejected'
  | 'duplicate'
  | 'spam_report'
  | 'on_hold';

export type ReportResolution = 
  | 'no_action'
  | 'content_removed'
  | 'content_edited'
  | 'content_flagged'
  | 'user_warned'
  | 'user_suspended'
  | 'user_banned'
  | 'account_deleted'
  | 'visibility_reduced'
  | 'age_restricted'
  | 'false_positive'
  | 'duplicate_removed'
  | 'policy_updated'
  | 'escalated_legal';

export interface ModerationAction extends BaseEntity {
  report_id: string;
  moderator_id: string;
  action_type: 'warning' | 'content_removal' | 'account_suspension' | 'account_ban' | 'content_edit' | 'visibility_change' | 'age_restriction' | 'community_guideline_strike';
  target_id: string;
  target_type: string;
  reason: string;
  duration?: number; // in hours for temporary actions
  expires_at?: string;
  is_reversible: boolean;
  
  // Action details
  previous_state?: Record<string, any>;
  new_state?: Record<string, any>;
  automated: boolean;
  confidence_score?: number;
  
  // Appeal information
  appeal_deadline?: string;
  is_appealable: boolean;
  
  // Relationships
  moderator?: User;
  report?: Report;
}

export interface ReportAppeal extends BaseEntity {
  report_id: string;
  action_id: string;
  user_id: string;
  reason: string;
  evidence: string[];
  additional_info?: string;
  
  // Status
  status: 'pending' | 'reviewing' | 'approved' | 'rejected' | 'expired';
  reviewed_by?: string;
  reviewed_at?: string;
  review_note?: string;
  
  // Resolution
  outcome: 'action_reversed' | 'action_modified' | 'appeal_denied' | 'case_reopened';
  new_action?: ModerationAction;
  
  // Relationships
  user?: User;
  report?: Report;
  action?: ModerationAction;
  reviewer?: User;
}

export interface AutoModerationRule extends BaseEntity {
  name: string;
  description: string;
  category: ReportCategory;
  is_active: boolean;
  priority: number;
  
  // Rule conditions
  conditions: AutoModerationCondition[];
  logic_operator: 'AND' | 'OR'; // How to combine conditions
  
  // Actions to take when rule is triggered
  actions: AutoModerationActionConfig[];
  
  // Confidence and thresholds
  confidence_threshold: number; // 0-100
  false_positive_rate?: number;
  
  // Scope
  applies_to: ('user' | 'post' | 'comment' | 'story' | 'message' | 'group')[];
  excluded_user_roles: string[];
  excluded_users: string[];
  
  // Timing and frequency
  rate_limit?: {
    max_triggers_per_hour: number;
    max_triggers_per_day: number;
  };
  
  // Statistics
  trigger_count: number;
  accuracy_rate: number;
  last_triggered_at?: string;
  
  // Relationships
  created_by: string;
  updated_by?: string;
}

export interface AutoModerationCondition {
  type: 'keyword' | 'pattern' | 'sentiment' | 'toxicity' | 'spam_score' | 'user_history' | 'content_similarity' | 'frequency' | 'account_age' | 'reputation_score';
  operator: 'contains' | 'equals' | 'starts_with' | 'ends_with' | 'matches_regex' | 'greater_than' | 'less_than' | 'in_range';
  value: any;
  weight: number; // 1-10, importance of this condition
  case_sensitive?: boolean;
  normalized?: boolean; // for text processing
}

export interface AutoModerationActionConfig {
  type: 'flag' | 'hold_for_review' | 'auto_remove' | 'reduce_visibility' | 'add_warning' | 'create_report' | 'notify_moderators' | 'rate_limit_user';
  severity: 'low' | 'medium' | 'high';
  notify_user: boolean;
  notify_moderators: boolean;
  requires_human_review: boolean;
  escalation_threshold?: number; // auto-escalate after X triggers
  parameters?: Record<string, any>;
}

export interface AutoModerationResult extends BaseEntity {
  target_id: string;
  target_type: string;
  rule_id: string;
  rule_name: string;
  triggered_conditions: string[];
  confidence_score: number;
  actions_taken: string[];
  requires_review: boolean;
  false_positive: boolean; // marked by human moderator
  
  // Processing details
  processing_time: number; // milliseconds
  model_version?: string;
  
  // Relationships
  rule?: AutoModerationRule;
  generated_report?: Report;
}

// Queue and workflow types
export interface ModerationQueue {
  pending_reports: Report[];
  high_priority_reports: Report[];
  escalated_reports: Report[];
  auto_flagged_content: Array<{
    content_id: string;
    content_type: string;
    rule_triggered: string;
    confidence_score: number;
    created_at: string;
  }>;
  appeals_to_review: ReportAppeal[];
  follow_up_required: Report[];
}

export interface ModeratorWorkload {
  moderator_id: string;
  assigned_reports: number;
  pending_reviews: number;
  completed_today: number;
  avg_resolution_time: number; // in minutes
  accuracy_rate: number;
  current_shift_start?: string;
  current_shift_end?: string;
  availability_status: 'available' | 'busy' | 'offline';
}

// Search and filter types
export interface ReportSearchParams {
  query?: string;
  reporter_id?: string;
  target_type?: 'user' | 'post' | 'comment' | 'story' | 'message' | 'group';
  target_id?: string;
  reason?: ReportReason[];
  category?: ReportCategory[];
  status?: ReportStatus[];
  priority?: Priority[];
  severity?: ('low' | 'medium' | 'high' | 'critical')[];
  assigned_to?: string;
  resolved_by?: string;
  source?: ('user_report' | 'auto_detection' | 'moderator_review' | 'bulk_report')[];
  date_from?: string;
  date_to?: string;
  resolution?: ReportResolution[];
  follow_up_required?: boolean;
  has_appeal?: boolean;
  sort_by?: 'created_at' | 'priority' | 'severity' | 'resolved_at';
  sort_order?: 'asc' | 'desc';
  page?: number;
  limit?: number;
}

export interface ReportFilters {
  target_types: string[];
  reasons: string[];
  categories: string[];
  statuses: string[];
  priorities: string[];
  severities: string[];
  sources: string[];
  assigned_moderators: string[];
  date_range: {
    start?: string;
    end?: string;
  };
  resolution_types: string[];
  follow_up_status: boolean | null;
  appeal_status: string[];
}

// Form types
export interface CreateReportRequest {
  target_id: string;
  target_type: 'user' | 'post' | 'comment' | 'story' | 'message' | 'group';
  reason: ReportReason;
  category?: ReportCategory;
  description: string;
  screenshots?: string[]; // media IDs
  evidence?: string[];
  additional_context?: string;
}

export interface UpdateReportRequest {
  status?: ReportStatus;
  priority?: Priority;
  severity?: 'low' | 'medium' | 'high' | 'critical';
  assigned_to?: string;
  escalation_reason?: string;
  follow_up_required?: boolean;
  follow_up_date?: string;
  follow_up_note?: string;
  resolution_note?: string;
}

export interface ResolveReportRequest {
  resolution: ReportResolution;
  resolution_note: string;
  actions_taken: Array<{
    action_type: string;
    target_id: string;
    target_type: string;
    reason: string;
    duration?: number;
    parameters?: Record<string, any>;
  }>;
  notify_reporter?: boolean;
  follow_up_required?: boolean;
  follow_up_date?: string;
}

export interface BulkReportAction {
  action: 'assign' | 'resolve' | 'reject' | 'escalate' | 'close';
  report_ids: string[];
  data?: {
    assigned_to?: string;
    resolution?: ReportResolution;
    resolution_note?: string;
    escalation_reason?: string;
  };
}

export interface CreateAppealRequest {
  reason: string;
  evidence?: string[];
  additional_info?: string;
}

// API response types
export type ReportsResponse = PaginatedResponse<Report>;
export type ModerationActionsResponse = PaginatedResponse<ModerationAction>;
export type ReportAppealsResponse = PaginatedResponse<ReportAppeal>;
export type AutoModerationRulesResponse = PaginatedResponse<AutoModerationRule>;
export type AutoModerationResultsResponse = PaginatedResponse<AutoModerationResult>;

// Statistics and analytics types
export interface ReportStats {
  total_reports: number;
  pending_reports: number;
  resolved_reports: number;
  rejected_reports: number;
  reports_today: number;
  reports_week: number;
  reports_month: number;
  avg_resolution_time: number; // in hours
  accuracy_rate: number;
  
  reports_by_reason: Record<ReportReason, number>;
  reports_by_category: Record<ReportCategory, number>;
  reports_by_target_type: Record<string, number>;
  resolution_distribution: Record<ReportResolution, number>;
  
  top_reporters: Array<{
    user_id: string;
    username: string;
    report_count: number;
    accuracy_rate: number;
  }>;
  
  moderator_performance: Array<{
    moderator_id: string;
    username: string;
    reports_handled: number;
    avg_resolution_time: number;
    accuracy_rate: number;
  }>;
  
  auto_moderation_stats: {
    rules_active: number;
    auto_actions_taken: number;
    false_positive_rate: number;
    human_review_required_rate: number;
  };
}

export interface ModerationInsights {
  trending_violations: Array<{
    reason: ReportReason;
    count: number;
    growth_rate: number;
    severity_distribution: Record<string, number>;
  }>;
  
  content_health_score: number; // 0-100
  community_safety_index: number; // 0-100
  
  escalation_patterns: Array<{
    issue_type: string;
    escalation_rate: number;
    common_causes: string[];
  }>;
  
  appeal_success_rates: Record<ReportResolution, number>;
  
  time_based_patterns: {
    peak_report_hours: number[];
    seasonal_trends: Array<{
      period: string;
      report_volume: number;
      common_issues: string[];
    }>;
  };
}