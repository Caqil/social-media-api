import { apiClient } from '../api';
import {
  PlatformAnalytics,
  UserMetrics,
  ContentMetrics,
  EngagementMetrics,
  RevenueMetrics,
  GeographicData,
  DeviceAnalytics,
  PerformanceMetrics,
  CustomAnalyticsQuery,
  AnalyticsReport,
  AnalyticsInsight,
  AnalyticsVisualization,
  RealTimeMetrics,
  AnalyticsExport,
  DateRange,
  TimeSeriesData,
} from '../types/analytics';

export class AnalyticsService {
  // Platform overview
  async getPlatformAnalytics(dateRange: DateRange): Promise<PlatformAnalytics> {
    return apiClient.get<PlatformAnalytics>('/admin/analytics/platform', {
      start_date: dateRange.start,
      end_date: dateRange.end
    });
  }

  async getDashboardStats(): Promise<{
    total_users: number;
    active_users: number;
    total_posts: number;
    total_comments: number;
    total_groups: number;
    total_reports: number;
    pending_reports: number;
    storage_used: number;
    storage_limit: number;
  }> {
    return apiClient.get('/admin/dashboard/stats');
  }

  async getSystemHealth(): Promise<{
    status: 'healthy' | 'warning' | 'critical';
    uptime: number;
    response_time: number;
    error_rate: number;
    active_users: number;
    server_load: number;
    memory_usage: number;
    disk_usage: number;
    database_connections: number;
  }> {
    return apiClient.get('/admin/system/health');
  }

  // User analytics
  async getUserMetrics(dateRange: DateRange): Promise<UserMetrics> {
    return apiClient.get<UserMetrics>('/admin/analytics/users', {
      start_date: dateRange.start,
      end_date: dateRange.end
    });
  }

  async getUserGrowthData(period = '30d'): Promise<Array<{
    date: string;
    new_users: number;
    active_users: number;
    total_users: number;
    growth_rate: number;
  }>> {
    return apiClient.get('/admin/analytics/users/growth', { period });
  }

  async getUserRetentionData(cohortDate: string): Promise<{
    cohort_size: number;
    retention_rates: number[];
    retention_by_day: Array<{
      day: number;
      retained_users: number;
      retention_rate: number;
    }>;
  }> {
    return apiClient.get('/admin/analytics/users/retention', { cohort_date: cohortDate });
  }

  async getUserDemographics(): Promise<{
    age_distribution: Array<{ range: string; count: number; percentage: number }>;
    gender_distribution: Array<{ gender: string; count: number; percentage: number }>;
    location_distribution: Array<{ country: string; count: number; percentage: number }>;
    device_distribution: Array<{ device: string; count: number; percentage: number }>;
  }> {
    return apiClient.get('/admin/analytics/users/demographics');
  }

  async getUserActivityPatterns(timeRange = 'week'): Promise<{
    hourly_distribution: Array<{ hour: number; active_users: number }>;
    daily_distribution: Array<{ day: number; active_users: number }>;
    session_duration: {
      average: number;
      median: number;
      distribution: Array<{ range: string; count: number }>;
    };
    feature_usage: Array<{
      feature: string;
      usage_count: number;
      unique_users: number;
    }>;
  }> {
    return apiClient.get('/admin/analytics/users/activity-patterns', { time_range: timeRange });
  }

  // Content analytics
  async getContentMetrics(dateRange: DateRange): Promise<ContentMetrics> {
    return apiClient.get<ContentMetrics>('/admin/analytics/content', {
      start_date: dateRange.start,
      end_date: dateRange.end
    });
  }

  async getContentPerformance(timeRange = 'week'): Promise<{
    top_posts: Array<{
      post_id: string;
      title: string;
      author: string;
      views: number;
      likes: number;
      comments: number;
      shares: number;
      engagement_rate: number;
    }>;
    content_type_performance: Record<string, {
      count: number;
      avg_engagement: number;
      total_views: number;
    }>;
    trending_hashtags: Array<{
      hashtag: string;
      usage_count: number;
      growth_rate: number;
    }>;
  }> {
    return apiClient.get('/admin/analytics/content/performance', { time_range: timeRange });
  }

  async getContentGrowthData(period = '30d'): Promise<Array<{
    date: string;
    posts_created: number;
    comments_created: number;
    stories_created: number;
    total_content: number;
  }>> {
    return apiClient.get('/admin/analytics/content/growth', { period });
  }

  async getHashtagAnalytics(timeRange = 'week'): Promise<Array<{
    hashtag: string;
    post_count: number;
    unique_users: number;
    engagement_rate: number;
    reach: number;
    trending_score: number;
  }>> {
    return apiClient.get('/admin/analytics/hashtags', { time_range: timeRange });
  }

  // Engagement analytics
  async getEngagementMetrics(dateRange: DateRange): Promise<EngagementMetrics> {
    return apiClient.get<EngagementMetrics>('/admin/analytics/engagement', {
      start_date: dateRange.start,
      end_date: dateRange.end
    });
  }

  async getEngagementTrends(timeRange = 'month'): Promise<{
    overall_engagement: Array<{ date: string; engagement_rate: number }>;
    by_content_type: Record<string, Array<{ date: string; engagement_rate: number }>>;
    by_time_of_day: Array<{ hour: number; engagement_rate: number }>;
    by_day_of_week: Array<{ day: number; engagement_rate: number }>;
  }> {
    return apiClient.get('/admin/analytics/engagement/trends', { time_range: timeRange });
  }

  async getTopEngagers(timeRange = 'week', limit = 50): Promise<Array<{
    user_id: string;
    username: string;
    total_interactions: number;
    posts_created: number;
    comments_made: number;
    likes_given: number;
    shares_made: number;
    engagement_score: number;
  }>> {
    return apiClient.get('/admin/analytics/engagement/top-engagers', { 
      time_range: timeRange, 
      limit 
    });
  }

  // Geographic analytics
  async getGeographicData(): Promise<GeographicData> {
    return apiClient.get<GeographicData>('/admin/analytics/geographic');
  }

  async getCountryStats(): Promise<Array<{
    country: string;
    country_code: string;
    users: number;
    posts: number;
    engagement_rate: number;
    growth_rate: number;
  }>> {
    return apiClient.get('/admin/analytics/geographic/countries');
  }

  async getCityStats(limit = 20): Promise<Array<{
    city: string;
    country: string;
    users: number;
    posts: number;
    engagement_rate: number;
  }>> {
    return apiClient.get('/admin/analytics/geographic/cities', { limit });
  }

  // Device and platform analytics
  async getDeviceAnalytics(): Promise<DeviceAnalytics> {
    return apiClient.get<DeviceAnalytics>('/admin/analytics/devices');
  }

  async getPlatformStats(): Promise<Array<{
    platform: string;
    users: number;
    sessions: number;
    avg_session_duration: number;
    bounce_rate: number;
    conversion_rate: number;
  }>> {
    return apiClient.get('/admin/analytics/platforms');
  }

  async getBrowserStats(): Promise<Array<{
    browser: string;
    version: string;
    users: number;
    percentage: number;
  }>> {
    return apiClient.get('/admin/analytics/browsers');
  }

  // Performance metrics
  async getPerformanceMetrics(): Promise<PerformanceMetrics> {
    return apiClient.get<PerformanceMetrics>('/admin/analytics/performance');
  }

  async getApiPerformance(timeRange = 'day'): Promise<{
    avg_response_time: number;
    median_response_time: number;
    p95_response_time: number;
    error_rate: number;
    requests_per_minute: number;
    slowest_endpoints: Array<{
      endpoint: string;
      avg_response_time: number;
      request_count: number;
    }>;
  }> {
    return apiClient.get('/admin/analytics/performance/api', { time_range: timeRange });
  }

  async getPageLoadMetrics(timeRange = 'day'): Promise<{
    avg_load_time: number;
    median_load_time: number;
    p95_load_time: number;
    bounce_rate: number;
    slowest_pages: Array<{
      page: string;
      avg_load_time: number;
      visits: number;
    }>;
  }> {
    return apiClient.get('/admin/analytics/performance/pages', { time_range: timeRange });
  }

  // Revenue analytics (if applicable)
  async getRevenueMetrics(dateRange: DateRange): Promise<RevenueMetrics> {
    return apiClient.get<RevenueMetrics>('/admin/analytics/revenue', {
      start_date: dateRange.start,
      end_date: dateRange.end
    });
  }

  async getSubscriptionMetrics(): Promise<{
    total_subscribers: number;
    new_subscribers: number;
    cancelled_subscribers: number;
    churn_rate: number;
    mrr: number;
    arr: number;
    ltv: number;
    cac: number;
    ltv_cac_ratio: number;
  }> {
    return apiClient.get('/admin/analytics/revenue/subscriptions');
  }

  // Real-time analytics
  async getRealTimeMetrics(): Promise<RealTimeMetrics> {
    return apiClient.get<RealTimeMetrics>('/admin/analytics/real-time');
  }

  async getCurrentActiveUsers(): Promise<{
    total_active: number;
    by_page: Array<{ page: string; active_users: number }>;
    by_device: Array<{ device: string; active_users: number }>;
    by_location: Array<{ country: string; active_users: number }>;
  }> {
    return apiClient.get('/admin/analytics/real-time/active-users');
  }

  async getLiveActivity(): Promise<{
    posts_last_hour: number;
    comments_last_hour: number;
    registrations_last_hour: number;
    active_conversations: number;
    trending_topics: string[];
  }> {
    return apiClient.get('/admin/analytics/real-time/activity');
  }

  // Custom analytics
  async createCustomQuery(query: Omit<CustomAnalyticsQuery, 'id' | 'created_by' | 'created_at' | 'last_run_at'>): Promise<CustomAnalyticsQuery> {
    return apiClient.post<CustomAnalyticsQuery>('/admin/analytics/custom-queries', query);
  }

  async getCustomQueries(): Promise<CustomAnalyticsQuery[]> {
    return apiClient.get<CustomAnalyticsQuery[]>('/admin/analytics/custom-queries');
  }

  async runCustomQuery(id: string): Promise<{
    query_id: string;
    results: any[];
    execution_time: number;
    row_count: number;
    columns: string[];
  }> {
    return apiClient.post(`/admin/analytics/custom-queries/${id}/run`);
  }

  async deleteCustomQuery(id: string): Promise<void> {
    return apiClient.delete(`/admin/analytics/custom-queries/${id}`);
  }

  // Analytics reports
  async createReport(report: Omit<AnalyticsReport, 'id' | 'generated_by' | 'created_at' | 'updated_at'>): Promise<AnalyticsReport> {
    return apiClient.post<AnalyticsReport>('/admin/analytics/reports', report);
  }

  async getReports(): Promise<AnalyticsReport[]> {
    return apiClient.get<AnalyticsReport[]>('/admin/analytics/reports');
  }

  async getReport(id: string): Promise<AnalyticsReport> {
    return apiClient.get<AnalyticsReport>(`/admin/analytics/reports/${id}`);
  }

  async updateReport(id: string, updates: Partial<AnalyticsReport>): Promise<AnalyticsReport> {
    return apiClient.put<AnalyticsReport>(`/admin/analytics/reports/${id}`, updates);
  }

  async deleteReport(id: string): Promise<void> {
    return apiClient.delete(`/admin/analytics/reports/${id}`);
  }

  async scheduleReport(id: string, schedule: {
    frequency: 'daily' | 'weekly' | 'monthly';
    recipients: string[];
    time?: string;
    day_of_week?: number;
    day_of_month?: number;
  }): Promise<void> {
    return apiClient.post(`/admin/analytics/reports/${id}/schedule`, schedule);
  }

  // Data export
  async exportAnalytics(exportConfig: Omit<AnalyticsExport, 'download_url' | 'expires_at' | 'file_size' | 'status'>): Promise<AnalyticsExport> {
    return apiClient.post<AnalyticsExport>('/admin/analytics/export', exportConfig);
  }

  async getExportStatus(exportId: string): Promise<{
    status: 'generating' | 'ready' | 'expired' | 'failed';
    progress: number;
    download_url?: string;
    expires_at?: string;
    file_size?: number;
    error_message?: string;
  }> {
    return apiClient.get(`/admin/analytics/export/${exportId}/status`);
  }

  async downloadExport(exportId: string): Promise<Blob> {
    const response = await fetch(`${apiClient['baseUrl']}/admin/analytics/export/${exportId}/download`);
    return response.blob();
  }

  // Insights and recommendations
  async getInsights(timeRange = 'week'): Promise<AnalyticsInsight[]> {
    return apiClient.get<AnalyticsInsight[]>('/admin/analytics/insights', { time_range: timeRange });
  }

  async getRecommendations(): Promise<Array<{
    type: 'growth' | 'engagement' | 'retention' | 'performance';
    title: string;
    description: string;
    impact: 'low' | 'medium' | 'high';
    effort: 'low' | 'medium' | 'high';
    priority_score: number;
    action_items: string[];
  }>> {
    return apiClient.get('/admin/analytics/recommendations');
  }

  async getAnomalyDetection(timeRange = 'week'): Promise<Array<{
    metric: string;
    anomaly_type: 'spike' | 'drop' | 'trend_change';
    severity: 'low' | 'medium' | 'high';
    detected_at: string;
    current_value: number;
    expected_value: number;
    deviation_percentage: number;
    description: string;
  }>> {
    return apiClient.get('/admin/analytics/anomalies', { time_range: timeRange });
  }

  // Comparative analytics
  async getComparativeAnalytics(periods: string[]): Promise<{
    user_growth: Array<{
      period: string;
      new_users: number;
      growth_rate: number;
    }>;
    engagement: Array<{
      period: string;
      engagement_rate: number;
      change: number;
    }>;
    content_creation: Array<{
      period: string;
      posts_created: number;
      change: number;
    }>;
  }> {
    return apiClient.get('/admin/analytics/comparative', { periods });
  }

  async getBenchmarkData(): Promise<{
    industry_averages: {
      engagement_rate: number;
      retention_rate: number;
      growth_rate: number;
      session_duration: number;
    };
    your_metrics: {
      engagement_rate: number;
      retention_rate: number;
      growth_rate: number;
      session_duration: number;
    };
    percentile_rankings: {
      engagement_rate: number;
      retention_rate: number;
      growth_rate: number;
      session_duration: number;
    };
  }> {
    return apiClient.get('/admin/analytics/benchmarks');
  }

  // Visualization data
  async getVisualizationData(type: string, params: Record<string, any>): Promise<TimeSeriesData | any> {
    return apiClient.get(`/admin/analytics/visualization/${type}`, params);
  }

  async createVisualization(config: Omit<AnalyticsVisualization, 'id'>): Promise<AnalyticsVisualization> {
    return apiClient.post<AnalyticsVisualization>('/admin/analytics/visualizations', config);
  }

  async getVisualizations(): Promise<AnalyticsVisualization[]> {
    return apiClient.get<AnalyticsVisualization[]>('/admin/analytics/visualizations');
  }

  async updateVisualization(id: string, config: Partial<AnalyticsVisualization>): Promise<AnalyticsVisualization> {
    return apiClient.put<AnalyticsVisualization>(`/admin/analytics/visualizations/${id}`, config);
  }

  async deleteVisualization(id: string): Promise<void> {
    return apiClient.delete(`/admin/analytics/visualizations/${id}`);
  }
}

export const analyticsService = new AnalyticsService();