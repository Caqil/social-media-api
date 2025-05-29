import { apiClient } from '../api';
import {
  UserProfile,
  UsersResponse,
  UserStats,
  UserSearchParams,
  CreateUserRequest,
  UpdateUserRequest,
  BulkUserAction,
  UserActivitiesResponse,
  UserConnectionsResponse,
  FollowRequestsResponse,
  UserAnalytics,
  UserActivity,
  UserConnection,
  FollowRequest,
  UserReport,
  UserSuspension,
  UserWarning,
  UserDevice,
  UserEngagement,
  BulkActionResponse,
} from '../types/user';
import { PaginatedResponse } from '../types/common';

export class UserService {
  // User management
  async getUsers(params?: UserSearchParams): Promise<UsersResponse> {
    return apiClient.get<UsersResponse>('/users', params);
  }

  async getUserById(id: string): Promise<UserProfile> {
    return apiClient.get<UserProfile>(`/users/${id}`);
  }

  async getUserByUsername(username: string): Promise<UserProfile> {
    return apiClient.get<UserProfile>(`/users/username/${username}`);
  }

  async getUserByEmail(email: string): Promise<UserProfile> {
    return apiClient.get<UserProfile>(`/users/email/${email}`);
  }

  async createUser(data: CreateUserRequest): Promise<UserProfile> {
    return apiClient.post<UserProfile>('/users', data);
  }

  async updateUser(id: string, data: UpdateUserRequest): Promise<UserProfile> {
    return apiClient.put<UserProfile>(`/users/${id}`, data);
  }

  async deleteUser(id: string, reason?: string): Promise<void> {
    return apiClient.delete(`/users/${id}`, { reason });
  }

  async suspendUser(id: string, reason: string, duration?: number): Promise<UserSuspension> {
    return apiClient.post<UserSuspension>(`/users/${id}/suspend`, { reason, duration });
  }

  async unsuspendUser(id: string): Promise<void> {
    return apiClient.post(`/users/${id}/unsuspend`);
  }

  async verifyUser(id: string): Promise<UserProfile> {
    return apiClient.post<UserProfile>(`/users/${id}/verify`);
  }

  async unverifyUser(id: string): Promise<UserProfile> {
    return apiClient.post<UserProfile>(`/users/${id}/unverify`);
  }

  async warnUser(id: string, reason: string, severity: 'low' | 'medium' | 'high'): Promise<UserWarning> {
    return apiClient.post<UserWarning>(`/users/${id}/warn`, { reason, severity });
  }

  // Bulk operations
  async bulkUserAction(action: BulkUserAction): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/users/bulk-action', action);
  }

  // User search and filtering
  async searchUsers(query: string, filters?: Partial<UserSearchParams>): Promise<UsersResponse> {
    return apiClient.get<UsersResponse>('/users/search', { query, ...filters });
  }

  async getSuggestedUsers(limit = 20): Promise<UserProfile[]> {
    return apiClient.get<UserProfile[]>('/users/suggestions', { limit });
  }

  // User statistics
  async getUserStats(): Promise<UserStats> {
    return apiClient.get<UserStats>('/admin/users/stats');
  }

  async getUserAnalytics(timeRange = 'week'): Promise<UserAnalytics> {
    return apiClient.get<UserAnalytics>('/admin/analytics/users', { time_range: timeRange });
  }

  // User activities and behavior
  async getUserActivities(userId: string, params?: { limit?: number; skip?: number }): Promise<UserActivitiesResponse> {
    return apiClient.get<UserActivitiesResponse>(`/users/${userId}/activities`, params);
  }

  async getUserEngagement(userId: string, timeRange = 'month'): Promise<UserEngagement[]> {
    return apiClient.get<UserEngagement[]>(`/users/${userId}/engagement`, { time_range: timeRange });
  }

  async getUserDevices(userId: string): Promise<UserDevice[]> {
    return apiClient.get<UserDevice[]>(`/users/${userId}/devices`);
  }

  // User connections and relationships
  async getUserFollowers(userId: string, params?: { limit?: number; skip?: number }): Promise<UserConnectionsResponse> {
    return apiClient.get<UserConnectionsResponse>(`/users/${userId}/followers`, params);
  }

  async getUserFollowing(userId: string, params?: { limit?: number; skip?: number }): Promise<UserConnectionsResponse> {
    return apiClient.get<UserConnectionsResponse>(`/users/${userId}/following`, params);
  }

  async getMutualFollows(userId: string, params?: { limit?: number; skip?: number }): Promise<UserConnectionsResponse> {
    return apiClient.get<UserConnectionsResponse>(`/users/${userId}/mutual-follows`, params);
  }

  async followUser(userId: string): Promise<UserConnection> {
    return apiClient.post<UserConnection>(`/users/${userId}/follow`);
  }

  async unfollowUser(userId: string): Promise<void> {
    return apiClient.delete(`/users/${userId}/follow`);
  }

  async getFollowRequests(params?: { limit?: number; skip?: number; type?: 'sent' | 'received' }): Promise<FollowRequestsResponse> {
    return apiClient.get<FollowRequestsResponse>('/follow-requests', params);
  }

  async acceptFollowRequest(requestId: string): Promise<void> {
    return apiClient.post(`/follow-requests/${requestId}/accept`);
  }

  async rejectFollowRequest(requestId: string): Promise<void> {
    return apiClient.post(`/follow-requests/${requestId}/reject`);
  }

  async cancelFollowRequest(requestId: string): Promise<void> {
    return apiClient.delete(`/follow-requests/${requestId}`);
  }

  async removeFollower(userId: string): Promise<void> {
    return apiClient.delete(`/followers/${userId}`);
  }

  async getFollowStatus(userId: string): Promise<{
    is_following: boolean;
    is_followed_by: boolean;
    follow_request_sent: boolean;
    follow_request_received: boolean;
  }> {
    return apiClient.get(`/users/${userId}/follow-status`);
  }

  // User blocking
  async blockUser(userId: string): Promise<void> {
    return apiClient.post(`/users/${userId}/block`);
  }

  async unblockUser(userId: string): Promise<void> {
    return apiClient.delete(`/users/${userId}/block`);
  }

  async getBlockedUsers(params?: { limit?: number; skip?: number }): Promise<PaginatedResponse<UserProfile>> {
    return apiClient.get<PaginatedResponse<UserProfile>>('/users/blocked', params);
  }

  // User reports
  async reportUser(userId: string, data: {
    reason: string;
    description: string;
    evidence?: string[];
  }): Promise<UserReport> {
    return apiClient.post<UserReport>(`/users/${userId}/report`, data);
  }

  async getUserReports(userId: string, params?: { limit?: number; skip?: number }): Promise<PaginatedResponse<UserReport>> {
    return apiClient.get<PaginatedResponse<UserReport>>(`/admin/users/${userId}/reports`, params);
  }

  // User profile management
  async updateUserProfile(userId: string, data: Partial<UserProfile>): Promise<UserProfile> {
    return apiClient.put<UserProfile>(`/users/${userId}/profile`, data);
  }

  async updateUserPrivacySettings(userId: string, settings: any): Promise<void> {
    return apiClient.put(`/users/${userId}/privacy-settings`, settings);
  }

  async updateUserNotificationSettings(userId: string, settings: any): Promise<void> {
    return apiClient.put(`/users/${userId}/notification-settings`, settings);
  }

  async deactivateUser(userId: string, reason?: string): Promise<void> {
    return apiClient.post(`/users/${userId}/deactivate`, { reason });
  }

  async reactivateUser(userId: string): Promise<UserProfile> {
    return apiClient.post<UserProfile>(`/users/${userId}/reactivate`);
  }

  // User impersonation (admin only)
  async impersonateUser(userId: string, reason: string): Promise<{
    impersonation_token: string;
    expires_at: string;
  }> {
    return apiClient.post(`/admin/users/${userId}/impersonate`, { reason });
  }

  async stopImpersonation(userId: string): Promise<void> {
    return apiClient.post(`/admin/users/${userId}/stop-impersonation`);
  }

  // User export and import
  async exportUsers(filters?: UserSearchParams, format = 'csv'): Promise<{
    download_url: string;
    expires_at: string;
  }> {
    return apiClient.post('/admin/users/export', { filters, format });
  }

  async importUsers(file: File): Promise<{
    job_id: string;
    total_records: number;
  }> {
    const formData = new FormData();
    formData.append('file', file);
    return apiClient.upload('/admin/users/import', formData);
  }

  async getImportStatus(jobId: string): Promise<{
    status: 'pending' | 'processing' | 'completed' | 'failed';
    progress: number;
    processed_records: number;
    failed_records: number;
    errors?: string[];
  }> {
    return apiClient.get(`/admin/users/import/${jobId}/status`);
  }

  // User sessions management
  async getUserSessions(userId: string): Promise<Array<{
    id: string;
    device_info: string;
    ip_address: string;
    location?: string;
    last_used_at: string;
    is_current: boolean;
  }>> {
    return apiClient.get(`/admin/users/${userId}/sessions`);
  }

  async revokeUserSession(userId: string, sessionId: string): Promise<void> {
    return apiClient.delete(`/admin/users/${userId}/sessions/${sessionId}`);
  }

  async revokeAllUserSessions(userId: string): Promise<void> {
    return apiClient.delete(`/admin/users/${userId}/sessions`);
  }

  // Advanced analytics
  async getUserGrowthData(period = '30d'): Promise<Array<{
    date: string;
    new_users: number;
    active_users: number;
    total_users: number;
  }>> {
    return apiClient.get('/admin/analytics/users/growth', { period });
  }

  async getUserRetentionData(cohortDate: string): Promise<{
    cohort_size: number;
    retention_rates: number[];
  }> {
    return apiClient.get('/admin/analytics/users/retention', { cohort_date: cohortDate });
  }

  async getUserSegmentation(): Promise<Array<{
    segment: string;
    count: number;
    percentage: number;
    characteristics: Record<string, any>;
  }>> {
    return apiClient.get('/admin/analytics/users/segmentation');
  }

  async getUserChurnPrediction(userId: string): Promise<{
    churn_probability: number;
    risk_factors: string[];
    recommended_actions: string[];
  }> {
    return apiClient.get(`/admin/analytics/users/${userId}/churn-prediction`);
  }

  // User content summary
  async getUserContentSummary(userId: string): Promise<{
    posts_count: number;
    comments_count: number;
    stories_count: number;
    likes_given: number;
    likes_received: number;
    followers_count: number;
    following_count: number;
    groups_count: number;
    last_activity_at: string;
  }> {
    return apiClient.get(`/users/${userId}/content-summary`);
  }

  // User moderation history
  async getUserModerationHistory(userId: string, params?: { limit?: number; skip?: number }): Promise<PaginatedResponse<{
    id: string;
    action: string;
    reason: string;
    moderator_id: string;
    created_at: string;
    expires_at?: string;
    is_active: boolean;
  }>> {
    return apiClient.get(`/admin/users/${userId}/moderation-history`, params);
  }

  async clearUserModerationHistory(userId: string, reason: string): Promise<void> {
    return apiClient.post(`/admin/users/${userId}/clear-moderation-history`, { reason });
  }
}

export const userService = new UserService();