import { apiClient } from '../api';
import {
  Group,
  GroupsResponse,
  GroupMembersResponse,
  GroupInvitationsResponse,
  GroupJoinRequestsResponse,
  GroupPostsResponse,
  GroupEventsResponse,
  GroupAnalyticsResponse,
  GroupMember,
  GroupInvitation,
  GroupJoinRequest,
  GroupPost,
  GroupEvent,
  GroupAnalytics,
  GroupSearchParams,
  CreateGroupRequest,
  UpdateGroupRequest,
  InviteToGroupRequest,
  JoinGroupRequest,
  UpdateMemberRoleRequest,
  BulkGroupAction,
  GroupStats,
  GroupCategory,
  GroupRecommendation,
  GroupInsights,
} from '../types/group';
import { BulkActionResponse } from '../types/common';

export class GroupService {
  // Group management
  async getGroups(params?: GroupSearchParams): Promise<GroupsResponse> {
    return apiClient.get<GroupsResponse>('/groups', params);
  }

  async getGroupById(id: string): Promise<Group> {
    return apiClient.get<Group>(`/groups/${id}`);
  }

  async getGroupBySlug(slug: string): Promise<Group> {
    return apiClient.get<Group>(`/groups/slug/${slug}`);
  }

  async createGroup(data: CreateGroupRequest): Promise<Group> {
    return apiClient.post<Group>('/groups', data);
  }

  async updateGroup(id: string, data: UpdateGroupRequest): Promise<Group> {
    return apiClient.put<Group>(`/groups/${id}`, data);
  }

  async deleteGroup(id: string): Promise<void> {
    return apiClient.delete(`/groups/${id}`);
  }

  async archiveGroup(id: string): Promise<Group> {
    return apiClient.post<Group>(`/groups/${id}/archive`);
  }

  async unarchiveGroup(id: string): Promise<Group> {
    return apiClient.post<Group>(`/groups/${id}/unarchive`);
  }

  // Group membership
  async joinGroup(id: string, data?: JoinGroupRequest): Promise<GroupMember | GroupJoinRequest> {
    return apiClient.post(`/groups/${id}/join`, data);
  }

  async leaveGroup(id: string): Promise<void> {
    return apiClient.post(`/groups/${id}/leave`);
  }

  async getGroupMembers(id: string, params?: { 
    role?: string; 
    status?: string; 
    limit?: number; 
    offset?: number; 
  }): Promise<GroupMembersResponse> {
    return apiClient.get<GroupMembersResponse>(`/groups/${id}/members`, params);
  }

  async removeMember(groupId: string, memberId: string): Promise<void> {
    return apiClient.delete(`/groups/${groupId}/members/${memberId}`);
  }

  async updateMemberRole(groupId: string, data: UpdateMemberRoleRequest): Promise<GroupMember> {
    return apiClient.put<GroupMember>(`/groups/${groupId}/members/${data.user_id}/role`, data);
  }

  async banMember(groupId: string, memberId: string, reason: string): Promise<void> {
    return apiClient.post(`/groups/${groupId}/members/${memberId}/ban`, { reason });
  }

  async unbanMember(groupId: string, memberId: string): Promise<void> {
    return apiClient.post(`/groups/${groupId}/members/${memberId}/unban`);
  }

  // Group invitations
  async inviteToGroup(id: string, data: InviteToGroupRequest): Promise<GroupInvitation[]> {
    return apiClient.post<GroupInvitation[]>(`/groups/${id}/invite`, data);
  }

  async getGroupInvitations(params?: { 
    group_id?: string; 
    status?: string; 
    limit?: number; 
    skip?: number; 
  }): Promise<GroupInvitationsResponse> {
    return apiClient.get<GroupInvitationsResponse>('/group-invites', params);
  }

  async acceptGroupInvite(inviteId: string): Promise<GroupMember> {
    return apiClient.post<GroupMember>(`/group-invites/${inviteId}/accept`);
  }

  async rejectGroupInvite(inviteId: string): Promise<void> {
    return apiClient.post(`/group-invites/${inviteId}/reject`);
  }

  async cancelGroupInvite(inviteId: string): Promise<void> {
    return apiClient.delete(`/group-invites/${inviteId}`);
  }

  // Join requests
  async getJoinRequests(groupId: string, params?: { 
    status?: string; 
    limit?: number; 
    skip?: number; 
  }): Promise<GroupJoinRequestsResponse> {
    return apiClient.get<GroupJoinRequestsResponse>(`/groups/${groupId}/join-requests`, params);
  }

  async approveJoinRequest(requestId: string, note?: string): Promise<GroupMember> {
    return apiClient.post<GroupMember>(`/group-join-requests/${requestId}/approve`, { note });
  }

  async rejectJoinRequest(requestId: string, reason: string): Promise<void> {
    return apiClient.post(`/group-join-requests/${requestId}/reject`, { reason });
  }

  async getPendingJoinRequests(params?: { limit?: number; skip?: number }): Promise<GroupJoinRequestsResponse> {
    return apiClient.get<GroupJoinRequestsResponse>('/group-join-requests/pending', params);
  }

  // Group posts
  async getGroupPosts(groupId: string, params?: { 
    status?: string; 
    author_id?: string; 
    limit?: number; 
    skip?: number; 
  }): Promise<GroupPostsResponse> {
    return apiClient.get<GroupPostsResponse>(`/groups/${groupId}/posts`, params);
  }

  async createGroupPost(groupId: string, data: {
    content: string;
    content_type?: string;
    media?: string[];
    hashtags?: string[];
    mentions?: string[];
  }): Promise<GroupPost> {
    return apiClient.post<GroupPost>(`/groups/${groupId}/posts`, data);
  }

  async approveGroupPost(postId: string, notes?: string): Promise<GroupPost> {
    return apiClient.post<GroupPost>(`/group-posts/${postId}/approve`, { notes });
  }

  async rejectGroupPost(postId: string, reason: string): Promise<GroupPost> {
    return apiClient.post<GroupPost>(`/group-posts/${postId}/reject`, { reason });
  }

  async pinGroupPost(postId: string): Promise<GroupPost> {
    return apiClient.post<GroupPost>(`/group-posts/${postId}/pin`);
  }

  async unpinGroupPost(postId: string): Promise<GroupPost> {
    return apiClient.post<GroupPost>(`/group-posts/${postId}/unpin`);
  }

  // Group events
  async getGroupEvents(groupId: string, params?: { 
    status?: string; 
    upcoming?: boolean; 
    limit?: number; 
    skip?: number; 
  }): Promise<GroupEventsResponse> {
    return apiClient.get<GroupEventsResponse>(`/groups/${groupId}/events`, params);
  }

  async createGroupEvent(groupId: string, data: Omit<GroupEvent, 'id' | 'created_at' | 'updated_at' | 'group' | 'organizer'>): Promise<GroupEvent> {
    return apiClient.post<GroupEvent>(`/groups/${groupId}/events`, data);
  }

  async updateGroupEvent(eventId: string, data: Partial<GroupEvent>): Promise<GroupEvent> {
    return apiClient.put<GroupEvent>(`/group-events/${eventId}`, data);
  }

  async deleteGroupEvent(eventId: string): Promise<void> {
    return apiClient.delete(`/group-events/${eventId}`);
  }

  async attendEvent(eventId: string): Promise<void> {
    return apiClient.post(`/group-events/${eventId}/attend`);
  }

  async unattendEvent(eventId: string): Promise<void> {
    return apiClient.delete(`/group-events/${eventId}/attend`);
  }

  // Group search and discovery
  async searchGroups(query: string, filters?: Partial<GroupSearchParams>): Promise<GroupsResponse> {
    return apiClient.get<GroupsResponse>('/groups/search', { query, ...filters });
  }

  async getPublicGroups(params?: { limit?: number; offset?: number }): Promise<GroupsResponse> {
    return apiClient.get<GroupsResponse>('/groups/public', params);
  }

  async getTrendingGroups(params?: { 
    time_range?: string; 
    limit?: number; 
    offset?: number; 
  }): Promise<GroupsResponse> {
    return apiClient.get<GroupsResponse>('/groups/trending', params);
  }

  async getFeaturedGroups(params?: { limit?: number; offset?: number }): Promise<GroupsResponse> {
    return apiClient.get<GroupsResponse>('/groups/featured', params);
  }

  async getRecommendedGroups(params?: { limit?: number }): Promise<GroupRecommendation[]> {
    return apiClient.get<GroupRecommendation[]>('/groups/recommendations', params);
  }

  async getSuggestedGroups(userId?: string, params?: { limit?: number }): Promise<Group[]> {
    const endpoint = userId ? `/groups/suggestions/${userId}` : '/groups/suggestions';
    return apiClient.get<Group[]>(endpoint, params);
  }

  // User's groups
  async getUserGroups(userId: string, params?: { 
    role?: string; 
    limit?: number; 
    offset?: number; 
  }): Promise<GroupsResponse> {
    return apiClient.get<GroupsResponse>(`/users/${userId}/groups`, params);
  }

  async getMyGroups(params?: { 
    role?: string; 
    status?: string; 
    limit?: number; 
    offset?: number; 
  }): Promise<GroupsResponse> {
    return apiClient.get<GroupsResponse>('/groups/my-groups', params);
  }

  async getGroupsByCategory(category: string, params?: { limit?: number; offset?: number }): Promise<GroupsResponse> {
    return apiClient.get<GroupsResponse>(`/groups/category/${category}`, params);
  }

  // Group categories
  async getGroupCategories(): Promise<GroupCategory[]> {
    return apiClient.get<GroupCategory[]>('/groups/categories');
  }

  async getCategory(id: string): Promise<GroupCategory> {
    return apiClient.get<GroupCategory>(`/groups/categories/${id}`);
  }

  // Admin group management
  async getGroupStats(): Promise<GroupStats> {
    return apiClient.get<GroupStats>('/admin/groups/stats');
  }

  async getGroupAnalytics(groupId: string, timeRange = 'week'): Promise<GroupAnalytics> {
    return apiClient.get<GroupAnalytics>(`/groups/${groupId}/analytics`, { time_range: timeRange });
  }

  async getGroupInsights(groupId: string, timeRange = 'month'): Promise<GroupInsights> {
    return apiClient.get<GroupInsights>(`/groups/${groupId}/insights`, { time_range: timeRange });
  }

  async verifyGroup(id: string): Promise<Group> {
    return apiClient.post<Group>(`/admin/groups/${id}/verify`);
  }

  async unverifyGroup(id: string): Promise<Group> {
    return apiClient.post<Group>(`/admin/groups/${id}/unverify`);
  }

  async featureGroup(id: string): Promise<Group> {
    return apiClient.post<Group>(`/admin/groups/${id}/feature`);
  }

  async unfeatureGroup(id: string): Promise<Group> {
    return apiClient.post<Group>(`/admin/groups/${id}/unfeature`);
  }

  async suspendGroup(id: string, reason: string): Promise<Group> {
    return apiClient.post<Group>(`/admin/groups/${id}/suspend`, { reason });
  }

  async unsuspendGroup(id: string): Promise<Group> {
    return apiClient.post<Group>(`/admin/groups/${id}/unsuspend`);
  }

  // Bulk operations
  async bulkGroupAction(action: BulkGroupAction): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/admin/groups/bulk-action', action);
  }

  async bulkDeleteGroups(groupIds: string[], reason?: string): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/admin/groups/bulk-delete', { 
      group_ids: groupIds, 
      reason 
    });
  }

  async bulkFeatureGroups(groupIds: string[]): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/admin/groups/bulk-feature', { group_ids: groupIds });
  }

  async bulkVerifyGroups(groupIds: string[]): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/admin/groups/bulk-verify', { group_ids: groupIds });
  }

  // Group settings and configuration
  async updateGroupSettings(groupId: string, settings: {
    post_approval_required?: boolean;
    member_approval_required?: boolean;
    allow_member_invites?: boolean;
    allow_external_sharing?: boolean;
    content_moderation_level?: string;
    auto_moderation_enabled?: boolean;
    profanity_filter_enabled?: boolean;
  }): Promise<Group> {
    return apiClient.put<Group>(`/groups/${groupId}/settings`, settings);
  }

  async updateGroupRules(groupId: string, rules: string[]): Promise<Group> {
    return apiClient.put<Group>(`/groups/${groupId}/rules`, { rules });
  }

  async updateGroupWelcomeMessage(groupId: string, welcomeMessage: string): Promise<Group> {
    return apiClient.put<Group>(`/groups/${groupId}/welcome-message`, { welcome_message: welcomeMessage });
  }

  // Group moderation
  async getGroupModerationQueue(groupId: string, params?: { 
    type?: string; 
    limit?: number; 
    skip?: number; 
  }): Promise<{
    pending_posts: GroupPost[];
    pending_members: GroupJoinRequest[];
    reported_content: any[];
    flagged_members: GroupMember[];
  }> {
    return apiClient.get(`/groups/${groupId}/moderation-queue`, params);
  }

  async reportGroup(id: string, data: {
    reason: string;
    description: string;
    evidence?: string[];
  }): Promise<void> {
    return apiClient.post(`/groups/${id}/report`, data);
  }

  // Group export and backup
  async exportGroupData(groupId: string, options: {
    include_posts?: boolean;
    include_members?: boolean;
    include_events?: boolean;
    format?: 'json' | 'csv';
  }): Promise<{
    download_url: string;
    expires_at: string;
  }> {
    return apiClient.post(`/groups/${groupId}/export`, options);
  }

  async exportGroups(filters?: GroupSearchParams, format = 'csv'): Promise<{
    download_url: string;
    expires_at: string;
  }> {
    return apiClient.post('/admin/groups/export', { filters, format });
  }

  // Group discovery algorithms
  async getNearbyGroups(params: {
    latitude: number;
    longitude: number;
    radius?: number; // in kilometers
    limit?: number;
  }): Promise<Group[]> {
    return apiClient.get<Group[]>('/groups/nearby', params);
  }

  async getGroupsWithMutualMembers(params?: { limit?: number }): Promise<Array<Group & {
    mutual_members_count: number;
    mutual_members: Array<{ id: string; username: string; avatar?: string }>;
  }>> {
    return apiClient.get('/groups/mutual-members', params);
  }

  async getGroupsByInterests(interests: string[], params?: { limit?: number }): Promise<Group[]> {
    return apiClient.post<Group[]>('/groups/by-interests', { interests, ...params });
  }

  // Group engagement metrics
  async getGroupEngagementMetrics(groupId: string, timeRange = 'week'): Promise<{
    active_members: number;
    posts_per_day: number;
    comments_per_day: number;
    engagement_rate: number;
    top_contributors: Array<{
      user_id: string;
      username: string;
      posts: number;
      comments: number;
      likes: number;
    }>;
    activity_timeline: Array<{
      date: string;
      posts: number;
      comments: number;
      active_members: number;
    }>;
  }> {
    return apiClient.get(`/groups/${groupId}/engagement`, { time_range: timeRange });
  }

  async getGroupGrowthMetrics(groupId: string, timeRange = 'month'): Promise<{
    member_growth: Array<{
      date: string;
      new_members: number;
      total_members: number;
      growth_rate: number;
    }>;
    retention_rate: number;
    churn_rate: number;
    acquisition_sources: Record<string, number>;
  }> {
    return apiClient.get(`/groups/${groupId}/growth`, { time_range: timeRange });
  }
}

export const groupService = new GroupService();