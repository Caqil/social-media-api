import { apiClient } from '../api';
import { Notification, PaginatedResponse } from '../types/common';

export interface NotificationSettings {
  likes: {
    enabled: boolean;
    push: boolean;
    email: boolean;
    sms: boolean;
  };
  comments: {
    enabled: boolean;
    push: boolean;
    email: boolean;
    sms: boolean;
  };
  follows: {
    enabled: boolean;
    push: boolean;
    email: boolean;
    sms: boolean;
  };
  mentions: {
    enabled: boolean;
    push: boolean;
    email: boolean;
    sms: boolean;
  };
  messages: {
    enabled: boolean;
    push: boolean;
    email: boolean;
    sms: boolean;
  };
  group_invites: {
    enabled: boolean;
    push: boolean;
    email: boolean;
    sms: boolean;
  };
  event_invites: {
    enabled: boolean;
    push: boolean;
    email: boolean;
    sms: boolean;
  };
  friend_requests: {
    enabled: boolean;
    push: boolean;
    email: boolean;
    sms: boolean;
  };
  digest_frequency: 'daily' | 'weekly' | 'monthly' | 'never';
  quiet_hours_start?: string;
  quiet_hours_end?: string;
}

export interface CreateNotificationRequest {
  recipient_id: string;
  type: string;
  title: string;
  message: string;
  action_text?: string;
  target_id?: string;
  target_type?: string;
  target_url?: string;
  priority?: 'low' | 'medium' | 'high' | 'urgent';
  send_via_email?: boolean;
  send_via_push?: boolean;
  send_via_sms?: boolean;
  data?: Record<string, any>;
}

export interface BulkNotificationRequest {
  recipient_ids: string[];
  type: string;
  title: string;
  message: string;
  action_text?: string;
  target_id?: string;
  target_type?: string;
  target_url?: string;
  priority?: 'low' | 'medium' | 'high' | 'urgent';
  send_via_email?: boolean;
  send_via_push?: boolean;
  send_via_sms?: boolean;
  data?: Record<string, any>;
}

export interface NotificationTemplate {
  id: string;
  name: string;
  type: string;
  category: string;
  title_template: string;
  message_template: string;
  action_text?: string;
  default_channels: string[];
  variables: Array<{
    name: string;
    description: string;
    required: boolean;
    default_value?: any;
  }>;
  is_system: boolean;
  is_active: boolean;
  usage_count: number;
  created_at: string;
  updated_at: string;
}

export interface NotificationStats {
  total_notifications: number;
  unread_notifications: number;
  notifications_today: number;
  notifications_week: number;
  notifications_month: number;
  
  by_type: Record<string, {
    count: number;
    unread_count: number;
  }>;
  
  by_channel: Record<string, {
    sent: number;
    delivered: number;
    opened: number;
    clicked: number;
  }>;
  
  engagement_rates: {
    overall_open_rate: number;
    overall_click_rate: number;
    by_type: Record<string, {
      open_rate: number;
      click_rate: number;
    }>;
  };
}

export interface PushTokenRegistration {
  token: string;
  platform: 'android' | 'ios' | 'web';
  device_info?: string;
  app_version?: string;
}

export class NotificationService {
  // User notifications
  async getNotifications(params?: {
    unread_only?: boolean;
    type?: string;
    limit?: number;
    skip?: number;
  }): Promise<PaginatedResponse<Notification>> {
    return apiClient.get<PaginatedResponse<Notification>>('/notifications', params);
  }

  async getNotificationById(id: string): Promise<Notification> {
    return apiClient.get<Notification>(`/notifications/${id}`);
  }

  async markAsRead(notificationIds: string[]): Promise<void> {
    return apiClient.post('/notifications/mark-read', { notification_ids: notificationIds });
  }

  async markAllAsRead(): Promise<void> {
    return apiClient.post('/notifications/mark-all-read');
  }

  async deleteNotifications(notificationIds: string[]): Promise<void> {
    return apiClient.delete('/notifications', { notification_ids: notificationIds });
  }

  async deleteAllNotifications(): Promise<void> {
    return apiClient.delete('/notifications/all');
  }

  async getUnreadCount(): Promise<{ count: number }> {
    return apiClient.get<{ count: number }>('/notifications/unread-count');
  }

  // Notification settings
  async getNotificationSettings(): Promise<NotificationSettings> {
    return apiClient.get<NotificationSettings>('/notifications/preferences');
  }

  async updateNotificationSettings(settings: Partial<NotificationSettings>): Promise<NotificationSettings> {
    return apiClient.put<NotificationSettings>('/notifications/preferences', settings);
  }

  async getNotificationTypes(): Promise<Array<{
    type: string;
    name: string;
    description: string;
    category: string;
    default_enabled: boolean;
    channels: string[];
  }>> {
    return apiClient.get('/notifications/types');
  }

  // Push notifications
  async registerPushToken(data: PushTokenRegistration): Promise<void> {
    return apiClient.post('/push/register', data);
  }

  async unregisterPushToken(token: string): Promise<void> {
    return apiClient.delete('/push/token', { token });
  }

  async getPushTokens(): Promise<Array<{
    id: string;
    token: string;
    platform: string;
    device_info?: string;
    is_active: boolean;
    registered_at: string;
    last_used_at?: string;
  }>> {
    return apiClient.get('/push/tokens');
  }

  async testPushNotification(data: {
    title: string;
    body: string;
    data?: Record<string, any>;
  }): Promise<void> {
    return apiClient.post('/push/test', data);
  }

  // Admin notification management
  async createNotification(data: CreateNotificationRequest): Promise<Notification> {
    return apiClient.post<Notification>('/admin/notifications', data);
  }

  async createBulkNotification(data: BulkNotificationRequest): Promise<{
    job_id: string;
    estimated_recipients: number;
  }> {
    return apiClient.post('/admin/notifications/bulk', data);
  }

  async getBulkNotificationStatus(jobId: string): Promise<{
    status: 'pending' | 'processing' | 'completed' | 'failed';
    progress: number;
    sent_count: number;
    failed_count: number;
    errors?: string[];
  }> {
    return apiClient.get(`/admin/notifications/bulk/${jobId}/status`);
  }

  // Notification templates
  async getNotificationTemplates(): Promise<NotificationTemplate[]> {
    return apiClient.get<NotificationTemplate[]>('/admin/notifications/templates');
  }

  async getNotificationTemplate(id: string): Promise<NotificationTemplate> {
    return apiClient.get<NotificationTemplate>(`/admin/notifications/templates/${id}`);
  }

  async createNotificationTemplate(data: Omit<NotificationTemplate, 'id' | 'usage_count' | 'created_at' | 'updated_at'>): Promise<NotificationTemplate> {
    return apiClient.post<NotificationTemplate>('/admin/notifications/templates', data);
  }

  async updateNotificationTemplate(id: string, data: Partial<NotificationTemplate>): Promise<NotificationTemplate> {
    return apiClient.put<NotificationTemplate>(`/admin/notifications/templates/${id}`, data);
  }

  async deleteNotificationTemplate(id: string): Promise<void> {
    return apiClient.delete(`/admin/notifications/templates/${id}`);
  }

  async sendNotificationFromTemplate(templateId: string, data: {
    recipient_ids: string[];
    variables: Record<string, any>;
    override_channels?: string[];
  }): Promise<{
    job_id: string;
    estimated_recipients: number;
  }> {
    return apiClient.post(`/admin/notifications/templates/${templateId}/send`, data);
  }

  // Notification analytics
  async getNotificationStats(timeRange = 'week'): Promise<NotificationStats> {
    return apiClient.get<NotificationStats>('/admin/notifications/stats', { time_range: timeRange });
  }

  async getUserNotificationStats(userId: string, timeRange = 'week'): Promise<{
    total_received: number;
    total_read: number;
    read_rate: number;
    avg_response_time: number; // in minutes
    preferred_channels: string[];
    most_engaging_types: Array<{
      type: string;
      open_rate: number;
      click_rate: number;
    }>;
  }> {
    return apiClient.get(`/admin/users/${userId}/notification-stats`, { time_range: timeRange });
  }

  async getNotificationPerformance(timeRange = 'week'): Promise<{
    delivery_rates: Record<string, number>; // by channel
    engagement_rates: Record<string, number>; // by channel
    best_performing_templates: Array<{
      template_id: string;
      name: string;
      open_rate: number;
      click_rate: number;
      usage_count: number;
    }>;
    optimal_send_times: Array<{
      hour: number;
      day_of_week: number;
      engagement_score: number;
    }>;
  }> {
    return apiClient.get('/admin/notifications/performance', { time_range: timeRange });
  }

  // System notifications
  async createSystemNotification(data: {
    title: string;
    message: string;
    type: 'info' | 'warning' | 'success' | 'error';
    target_audience: 'all' | 'admins' | 'moderators' | 'premium_users';
    action_text?: string;
    action_url?: string;
    expires_at?: string;
    is_dismissible?: boolean;
  }): Promise<{
    id: string;
    estimated_recipients: number;
  }> {
    return apiClient.post('/admin/notifications/system', data);
  }

  async getSystemNotifications(): Promise<Array<{
    id: string;
    title: string;
    message: string;
    type: string;
    is_active: boolean;
    recipient_count: number;
    read_count: number;
    created_at: string;
    expires_at?: string;
  }>> {
    return apiClient.get('/admin/notifications/system');
  }

  async updateSystemNotification(id: string, data: {
    title?: string;
    message?: string;
    is_active?: boolean;
    expires_at?: string;
  }): Promise<void> {
    return apiClient.put(`/admin/notifications/system/${id}`, data);
  }

  async deleteSystemNotification(id: string): Promise<void> {
    return apiClient.delete(`/admin/notifications/system/${id}`);
  }

  // Email notifications
  async getEmailTemplates(): Promise<Array<{
    id: string;
    name: string;
    subject: string;
    type: string;
    language: string;
    is_active: boolean;
    usage_count: number;
  }>> {
    return apiClient.get('/admin/notifications/email/templates');
  }

  async updateEmailTemplate(id: string, data: {
    subject?: string;
    template_content?: string;
    is_active?: boolean;
  }): Promise<void> {
    return apiClient.put(`/admin/notifications/email/templates/${id}`, data);
  }

  async testEmailTemplate(id: string, data: {
    recipient_email: string;
    test_variables: Record<string, any>;
  }): Promise<void> {
    return apiClient.post(`/admin/notifications/email/templates/${id}/test`, data);
  }

  async getEmailStats(timeRange = 'week'): Promise<{
    total_sent: number;
    delivered: number;
    opened: number;
    clicked: number;
    bounced: number;
    delivery_rate: number;
    open_rate: number;
    click_rate: number;
    bounce_rate: number;
  }> {
    return apiClient.get('/admin/notifications/email/stats', { time_range: timeRange });
  }

  // Push notification management
  async sendBulkPush(data: {
    title: string;
    body: string;
    user_ids?: string[];
    user_segments?: string[];
    platforms?: ('android' | 'ios' | 'web')[];
    data?: Record<string, any>;
    scheduled_for?: string;
  }): Promise<{
    job_id: string;
    estimated_recipients: number;
  }> {
    return apiClient.post('/admin/push/bulk', data);
  }

  async getPushStats(timeRange = 'week'): Promise<{
    total_sent: number;
    delivered: number;
    opened: number;
    clicked: number;
    failed: number;
    delivery_rate: number;
    open_rate: number;
    click_rate: number;
    by_platform: Record<string, {
      sent: number;
      delivered: number;
      opened: number;
      clicked: number;
    }>;
  }> {
    return apiClient.get('/admin/push/stats', { time_range: timeRange });
  }

  async cleanupInactiveTokens(daysInactive = 30): Promise<{
    cleaned_tokens: number;
  }> {
    return apiClient.post('/admin/push/cleanup', { days_inactive: daysInactive });
  }

  // Advanced features
  async scheduleNotification(data: CreateNotificationRequest & {
    scheduled_for: string;
    timezone?: string;
  }): Promise<{
    scheduled_notification_id: string;
    scheduled_for: string;
  }> {
    return apiClient.post('/admin/notifications/schedule', data);
  }

  async getScheduledNotifications(): Promise<Array<{
    id: string;
    title: string;
    message: string;
    recipient_count: number;
    scheduled_for: string;
    status: 'pending' | 'sent' | 'cancelled';
    created_at: string;
  }>> {
    return apiClient.get('/admin/notifications/scheduled');
  }

  async cancelScheduledNotification(id: string): Promise<void> {
    return apiClient.delete(`/admin/notifications/scheduled/${id}`);
  }

  async createNotificationCampaign(data: {
    name: string;
    description: string;
    template_id: string;
    target_segments: string[];
    schedule: {
      start_date: string;
      end_date?: string;
      frequency: 'once' | 'daily' | 'weekly' | 'monthly';
      time_of_day?: string;
    };
    variables: Record<string, any>;
  }): Promise<{
    campaign_id: string;
    estimated_sends: number;
  }> {
    return apiClient.post('/admin/notifications/campaigns', data);
  }

  async getCampaigns(): Promise<Array<{
    id: string;
    name: string;
    status: 'draft' | 'active' | 'paused' | 'completed';
    total_sends: number;
    engagement_rate: number;
    created_at: string;
  }>> {
    return apiClient.get('/admin/notifications/campaigns');
  }

  async getCampaignAnalytics(id: string): Promise<{
    total_sends: number;
    delivered: number;
    opened: number;
    clicked: number;
    unsubscribed: number;
    conversion_rate: number;
    engagement_timeline: Array<{
      date: string;
      sends: number;
      opens: number;
      clicks: number;
    }>;
  }> {
    return apiClient.get(`/admin/notifications/campaigns/${id}/analytics`);
  }
}

export const notificationService = new NotificationService();