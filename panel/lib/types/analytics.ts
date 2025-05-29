import { BaseEntity, DateRange, ChartDataPoint, TimeSeriesData } from './common';

export interface PlatformAnalytics {
  overview: PlatformOverview;
  user_metrics: UserMetrics;
  content_metrics: ContentMetrics;
  engagement_metrics: EngagementMetrics;
  revenue_metrics?: RevenueMetrics;
  geographic_data: GeographicData;
  device_analytics: DeviceAnalytics;
  performance_metrics: PerformanceMetrics;
  date_range: DateRange;
  generated_at: string;
}

export interface PlatformOverview {
  total_users: number;
  active_users: number;
  daily_active_users: number;
  weekly_active_users: number;
  monthly_active_users: number;
  total_posts: number;
  total_comments: number;
  total_likes: number;
  total_shares: number;
  total_groups: number;
  total_stories: number;
  total_messages: number;
  storage_used: number;
  bandwidth_used: number;
}

export interface UserMetrics {
  growth: UserGrowth;
  retention: RetentionMetrics;
  demographics: UserDemographics;
  activity_patterns: UserActivityPatterns;
  churn_analysis: ChurnAnalysis;
  user_journey: UserJourneyMetrics;
}

export interface UserGrowth {
  new_users_today: number;
  new_users_week: number;
  new_users_month: number;
  growth_rate_daily: number;
  growth_rate_weekly: number;
  growth_rate_monthly: number;
  
  historical_data: Array<{
    date: string;
    new_users: number;
    total_users: number;
    growth_rate: number;
  }>;
  
  projections: Array<{
    date: string;
    projected_users: number;
    confidence_interval: {
      lower: number;
      upper: number;
    };
  }>;
}

export interface RetentionMetrics {
  day_1_retention: number;
  day_7_retention: number;
  day_30_retention: number;
  day_90_retention: number;
  
  cohort_analysis: Array<{
    cohort_date: string;
    cohort_size: number;
    retention_rates: number[]; // retention by day/week
  }>;
  
  retention_by_acquisition_channel: Record<string, {
    day_1: number;
    day_7: number;
    day_30: number;
  }>;
}

export interface UserDemographics {
  age_groups: Array<{
    range: string;
    count: number;
    percentage: number;
  }>;
  
  gender_distribution: Array<{
    gender: string;
    count: number;
    percentage: number;
  }>;
  
  geographic_distribution: Array<{
    country: string;
    region?: string;
    count: number;
    percentage: number;
  }>;
  
  account_types: Array<{
    type: string;
    count: number;
    percentage: number;
  }>;
  
  subscription_tiers: Array<{
    tier: string;
    count: number;
    percentage: number;
    revenue_contribution: number;
  }>;
}

export interface UserActivityPatterns {
  peak_hours: Array<{
    hour: number;
    active_users: number;
    activity_score: number;
  }>;
  
  peak_days: Array<{
    day_of_week: number;
    active_users: number;
    activity_score: number;
  }>;
  
  session_duration: {
    average: number;
    median: number;
    distribution: Array<{
      range: string;
      count: number;
      percentage: number;
    }>;
  };
  
  feature_usage: Array<{
    feature: string;
    usage_count: number;
    unique_users: number;
    engagement_rate: number;
  }>;
}

export interface ChurnAnalysis {
  churn_rate_monthly: number;
  churn_rate_quarterly: number;
  at_risk_users: number;
  
  churn_reasons: Array<{
    reason: string;
    count: number;
    percentage: number;
  }>;
  
  predictive_indicators: Array<{
    indicator: string;
    correlation_score: number;
    threshold_value: any;
  }>;
  
  retention_initiatives_performance: Array<{
    initiative: string;
    target_users: number;
    retained_users: number;
    effectiveness_rate: number;
  }>;
}

export interface UserJourneyMetrics {
  onboarding_completion_rate: number;
  first_post_conversion: number;
  first_interaction_time: number; // minutes
  feature_adoption_rates: Record<string, number>;
  
  conversion_funnel: Array<{
    step: string;
    users_entered: number;
    users_completed: number;
    completion_rate: number;
    avg_time_to_complete: number;
  }>;
}

export interface ContentMetrics {
  creation_stats: ContentCreationStats;
  engagement_stats: ContentEngagementStats;
  quality_metrics: ContentQualityMetrics;
  moderation_stats: ContentModerationStats;
  trending_analysis: TrendingAnalysis;
}

export interface ContentCreationStats {
  posts_created_today: number;
  posts_created_week: number;
  posts_created_month: number;
  
  content_type_distribution: Record<string, {
    count: number;
    percentage: number;
    engagement_rate: number;
  }>;
  
  posting_frequency: {
    avg_posts_per_user: number;
    median_posts_per_user: number;
    power_users_percentage: number; // users creating 80% of content
  };
  
  content_length_analysis: {
    avg_post_length: number;
    avg_comment_length: number;
    optimal_length_for_engagement: number;
  };
}

export interface ContentEngagementStats {
  total_likes: number;
  total_comments: number;
  total_shares: number;
  total_saves: number;
  
  engagement_rates: {
    overall: number;
    by_content_type: Record<string, number>;
    by_time_of_day: Array<{
      hour: number;
      engagement_rate: number;
    }>;
    by_day_of_week: Array<{
      day: number;
      engagement_rate: number;
    }>;
  };
  
  viral_content: Array<{
    content_id: string;
    content_type: string;
    viral_score: number;
    reach: number;
    engagement_rate: number;
  }>;
}

export interface ContentQualityMetrics {
  avg_content_score: number;
  quality_distribution: Array<{
    score_range: string;
    count: number;
    percentage: number;
  }>;
  
  top_performing_content: Array<{
    content_id: string;
    score: number;
    engagement_metrics: Record<string, number>;
  }>;
  
  content_health_indicators: {
    spam_rate: number;
    duplicate_content_rate: number;
    low_quality_content_rate: number;
    user_generated_quality_score: number;
  };
}

export interface ContentModerationStats {
  flagged_content: number;
  auto_moderated: number;
  human_reviewed: number;
  false_positive_rate: number;
  
  moderation_actions: Record<string, {
    count: number;
    accuracy_rate: number;
  }>;
  
  response_times: {
    avg_first_response: number; // minutes
    avg_resolution_time: number; // minutes
    sla_compliance_rate: number;
  };
}

export interface TrendingAnalysis {
  trending_hashtags: Array<{
    hashtag: string;
    usage_count: number;
    growth_rate: number;
    engagement_rate: number;
  }>;
  
  trending_topics: Array<{
    topic: string;
    mention_count: number;
    sentiment_score: number;
    geographic_concentration: string[];
  }>;
  
  viral_patterns: {
    avg_time_to_viral: number; // hours
    viral_threshold_engagement: number;
    peak_viral_hours: number[];
  };
}

export interface EngagementMetrics {
  overall_engagement_rate: number;
  engagement_trends: EngagementTrends;
  interaction_patterns: InteractionPatterns;
  community_health: CommunityHealth;
  social_graph_metrics: SocialGraphMetrics;
}

export interface EngagementTrends {
  daily_engagement: ChartDataPoint[];
  weekly_engagement: ChartDataPoint[];
  monthly_engagement: ChartDataPoint[];
  
  engagement_by_feature: Record<string, {
    total_interactions: number;
    unique_users: number;
    avg_time_spent: number;
  }>;
}

export interface InteractionPatterns {
  like_to_comment_ratio: number;
  comment_to_share_ratio: number;
  response_rates: {
    posts_to_comments: number;
    comments_to_replies: number;
    messages_to_responses: number;
  };
  
  interaction_depth: {
    surface_level: number; // likes only
    moderate: number; // likes + comments
    deep: number; // likes + comments + shares + saves
  };
}

export interface CommunityHealth {
  toxicity_score: number; // 0-100, lower is better
  positivity_ratio: number;
  constructive_interaction_rate: number;
  
  community_metrics: {
    helpful_content_rate: number;
    knowledge_sharing_score: number;
    mentorship_connections: number;
    collaborative_projects: number;
  };
  
  conflict_resolution: {
    reported_conflicts: number;
    resolved_conflicts: number;
    avg_resolution_time: number;
    user_satisfaction_score: number;
  };
}

export interface SocialGraphMetrics {
  avg_connections_per_user: number;
  network_density: number;
  clustering_coefficient: number;
  
  connection_patterns: {
    mutual_connections_rate: number;
    cross_demographic_connections: number;
    geographic_connection_spread: number;
  };
  
  influence_metrics: {
    top_influencers: Array<{
      user_id: string;
      influence_score: number;
      reach: number;
      engagement_amplification: number;
    }>;
    
    influence_distribution: Array<{
      influence_range: string;
      user_count: number;
      percentage: number;
    }>;
  };
}

export interface RevenueMetrics {
  total_revenue: number;
  monthly_recurring_revenue: number;
  annual_recurring_revenue: number;
  
  revenue_streams: Array<{
    source: string;
    amount: number;
    percentage: number;
    growth_rate: number;
  }>;
  
  subscription_metrics: {
    new_subscriptions: number;
    cancelled_subscriptions: number;
    upgraded_subscriptions: number;
    downgraded_subscriptions: number;
    churn_rate: number;
    ltv: number; // lifetime value
    cac: number; // customer acquisition cost
    ltv_cac_ratio: number;
  };
  
  advertising_metrics?: {
    ad_impressions: number;
    ad_clicks: number;
    ad_revenue: number;
    cpm: number; // cost per mille
    ctr: number; // click-through rate
  };
}

export interface GeographicData {
  user_distribution: Array<{
    country: string;
    country_code: string;
    users: number;
    percentage: number;
    growth_rate: number;
  }>;
  
  regional_engagement: Array<{
    region: string;
    engagement_rate: number;
    avg_session_duration: number;
    peak_usage_hours: number[];
  }>;
  
  content_preferences: Array<{
    region: string;
    popular_content_types: string[];
    trending_topics: string[];
  }>;
  
  localization_metrics: {
    languages_supported: number;
    translation_accuracy: number;
    local_content_percentage: Record<string, number>;
  };
}

export interface DeviceAnalytics {
  device_distribution: Array<{
    device_type: string;
    count: number;
    percentage: number;
    avg_session_duration: number;
  }>;
  
  platform_distribution: Array<{
    platform: string;
    version: string;
    count: number;
    percentage: number;
  }>;
  
  browser_distribution: Array<{
    browser: string;
    version: string;
    count: number;
    percentage: number;
  }>;
  
  performance_by_device: Array<{
    device_category: string;
    avg_load_time: number;
    crash_rate: number;
    feature_usage_rate: Record<string, number>;
  }>;
}

export interface PerformanceMetrics {
  response_times: {
    avg_api_response: number;
    avg_page_load: number;
    avg_database_query: number;
  };
  
  availability: {
    uptime_percentage: number;
    downtime_incidents: number;
    avg_incident_duration: number;
  };
  
  error_rates: {
    client_errors: number;
    server_errors: number;
    network_errors: number;
  };
  
  resource_usage: {
    cpu_usage: number;
    memory_usage: number;
    storage_usage: number;
    bandwidth_usage: number;
  };
}

// Custom analytics types
export interface CustomAnalyticsQuery {
  id: string;
  name: string;
  description: string;
  query_type: 'user_segment' | 'content_performance' | 'engagement_analysis' | 'custom_metric';
  parameters: Record<string, any>;
  date_range: DateRange;
  filters: Record<string, any>;
  grouping: string[];
  metrics: string[];
  created_by: string;
  created_at: string;
  last_run_at?: string;
  is_scheduled: boolean;
  schedule_frequency?: 'daily' | 'weekly' | 'monthly';
}

export interface AnalyticsReport extends BaseEntity {
  title: string;
  description: string;
  report_type: 'dashboard' | 'executive_summary' | 'detailed_analysis' | 'custom';
  data: PlatformAnalytics;
  insights: AnalyticsInsight[];
  visualizations: AnalyticsVisualization[];
  date_range: DateRange;
  generated_by: string;
  is_public: boolean;
  tags: string[];
  export_formats: ('pdf' | 'excel' | 'csv' | 'json')[];
}

export interface AnalyticsInsight {
  type: 'trend' | 'anomaly' | 'opportunity' | 'risk' | 'achievement';
  title: string;
  description: string;
  confidence_score: number; // 0-100
  impact_score: number; // 0-100
  recommendation?: string;
  supporting_data: Record<string, any>;
  visualization_id?: string;
}

export interface AnalyticsVisualization {
  id: string;
  type: 'line_chart' | 'bar_chart' | 'pie_chart' | 'heatmap' | 'scatter_plot' | 'funnel' | 'gauge';
  title: string;
  data: TimeSeriesData | any;
  configuration: {
    x_axis?: string;
    y_axis?: string;
    color_scheme?: string[];
    show_legend?: boolean;
    show_grid?: boolean;
    animation_enabled?: boolean;
  };
}

// Real-time analytics
export interface RealTimeMetrics {
  current_active_users: number;
  posts_last_hour: number;
  comments_last_hour: number;
  likes_last_hour: number;
  new_registrations_last_hour: number;
  
  live_engagement: {
    trending_posts: string[];
    active_conversations: number;
    viral_content_emerging: string[];
  };
  
  system_health: {
    response_time: number;
    error_rate: number;
    cpu_usage: number;
    memory_usage: number;
  };
  
  alerts: Array<{
    type: 'performance' | 'security' | 'content' | 'user_behavior';
    severity: 'low' | 'medium' | 'high' | 'critical';
    message: string;
    timestamp: string;
    auto_resolved: boolean;
  }>;
}

// Export and sharing types
export interface AnalyticsExport {
  format: 'pdf' | 'excel' | 'csv' | 'json';
  data_range: DateRange;
  included_metrics: string[];
  filters: Record<string, any>;
  download_url?: string;
  expires_at?: string;
  file_size?: number;
  status: 'generating' | 'ready' | 'expired' | 'failed';
}