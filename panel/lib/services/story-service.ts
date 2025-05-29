import { apiClient } from '../api';
import {
  Story,
  StoriesResponse,
  StoryViewsResponse,
  StoryReactionsResponse,
  StoryRepliesResponse,
  StoryHighlightsResponse,
  StoryAnalyticsResponse,
  StoryView,
  StoryReaction,
  StoryReply,
  StoryHighlight,
  StoryAnalytics,
  StoryTemplate,
  StorySearchParams,
  CreateStoryRequest,
  UpdateStoryRequest,
  CreateStoryHighlightRequest,
  BulkStoryAction,
  StoryStats,
  StoryTrends,
  StoryModeration,
} from '../types/story';
import { BulkActionResponse } from '../types/common';

export class StoryService {
  // Story management
  async getStories(params?: StorySearchParams): Promise<StoriesResponse> {
    return apiClient.get<StoriesResponse>('/stories', params);
  }

  async getStoryById(id: string): Promise<Story> {
    return apiClient.get<Story>(`/stories/${id}`);
  }

  async createStory(data: CreateStoryRequest): Promise<Story> {
    return apiClient.post<Story>('/stories', data);
  }

  async updateStory(id: string, data: UpdateStoryRequest): Promise<Story> {
    return apiClient.put<Story>(`/stories/${id}`, data);
  }

  async deleteStory(id: string): Promise<void> {
    return apiClient.delete(`/stories/${id}`);
  }

  async archiveStory(id: string): Promise<Story> {
    return apiClient.post<Story>(`/stories/${id}/archive`);
  }

  async unarchiveStory(id: string): Promise<Story> {
    return apiClient.post<Story>(`/stories/${id}/unarchive`);
  }

  // Story viewing and interactions
  async viewStory(id: string): Promise<StoryView> {
    return apiClient.post<StoryView>(`/stories/${id}/view`);
  }

  async getStoryViews(id: string, params?: { limit?: number; skip?: number }): Promise<StoryViewsResponse> {
    return apiClient.get<StoryViewsResponse>(`/stories/${id}/views`, params);
  }

  async reactToStory(id: string, reactionType: string): Promise<StoryReaction> {
    return apiClient.post<StoryReaction>(`/stories/${id}/react`, { reaction_type: reactionType });
  }

  async removeStoryReaction(id: string): Promise<void> {
    return apiClient.delete(`/stories/${id}/react`);
  }

  async getStoryReactions(id: string, params?: { 
    reaction_type?: string; 
    limit?: number; 
    skip?: number; 
  }): Promise<StoryReactionsResponse> {
    return apiClient.get<StoryReactionsResponse>(`/stories/${id}/reactions`, params);
  }

  async replyToStory(id: string, data: {
    content: string;
    content_type?: 'text' | 'image' | 'voice';
    media?: string;
  }): Promise<StoryReply> {
    return apiClient.post<StoryReply>(`/stories/${id}/reply`, data);
  }

  async getStoryReplies(id: string, params?: { limit?: number; skip?: number }): Promise<StoryRepliesResponse> {
    return apiClient.get<StoryRepliesResponse>(`/stories/${id}/replies`, params);
  }

  async markStoryRepliesAsRead(storyId: string): Promise<void> {
    return apiClient.post(`/stories/${storyId}/replies/mark-read`);
  }

  // User stories
  async getUserStories(userId: string): Promise<Story[]> {
    return apiClient.get<Story[]>(`/stories/user/${userId}`);
  }

  async getMyStories(params?: { status?: string; limit?: number; skip?: number }): Promise<StoriesResponse> {
    return apiClient.get<StoriesResponse>('/stories/my-stories', params);
  }

  async getFollowingStories(): Promise<Array<{
    user_id: string;
    username: string;
    avatar?: string;
    stories: Story[];
    has_unseen: boolean;
    latest_story_at: string;
  }>> {
    return apiClient.get('/stories/following');
  }

  async getActiveStories(params?: { limit?: number; skip?: number }): Promise<StoriesResponse> {
    return apiClient.get<StoriesResponse>('/stories/active', params);
  }

  async getArchivedStories(params?: { limit?: number; skip?: number }): Promise<StoriesResponse> {
    return apiClient.get<StoriesResponse>('/stories/archived', params);
  }

  // Story highlights
  async getStoryHighlights(userId: string): Promise<StoryHighlight[]> {
    return apiClient.get<StoryHighlight[]>(`/story-highlights/user/${userId}`);
  }

  async getMyStoryHighlights(): Promise<StoryHighlight[]> {
    return apiClient.get<StoryHighlight[]>('/story-highlights/my-highlights');
  }

  async createStoryHighlight(data: CreateStoryHighlightRequest): Promise<StoryHighlight> {
    return apiClient.post<StoryHighlight>('/story-highlights', data);
  }

  async updateStoryHighlight(id: string, data: {
    title?: string;
    story_ids?: string[];
    cover_image?: string;
    is_public?: boolean;
  }): Promise<StoryHighlight> {
    return apiClient.put<StoryHighlight>(`/story-highlights/${id}`, data);
  }

  async deleteStoryHighlight(id: string): Promise<void> {
    return apiClient.delete(`/story-highlights/${id}`);
  }

  async addStoryToHighlight(highlightId: string, storyId: string): Promise<StoryHighlight> {
    return apiClient.post<StoryHighlight>(`/story-highlights/${highlightId}/stories`, { story_id: storyId });
  }

  async removeStoryFromHighlight(highlightId: string, storyId: string): Promise<StoryHighlight> {
    return apiClient.delete<StoryHighlight>(`/story-highlights/${highlightId}/stories/${storyId}`);
  }

  // Story templates
  async getStoryTemplates(params?: { 
    category?: string; 
    is_premium?: boolean; 
    limit?: number; 
    skip?: number; 
  }): Promise<StoryTemplate[]> {
    return apiClient.get<StoryTemplate[]>('/stories/templates', params);
  }

  async getStoryTemplate(id: string): Promise<StoryTemplate> {
    return apiClient.get<StoryTemplate>(`/stories/templates/${id}`);
  }

  async useStoryTemplate(id: string, customizations?: {
    background_color?: string;
    text_color?: string;
    content?: string;
  }): Promise<{
    template_data: any;
    preview_url: string;
  }> {
    return apiClient.post(`/stories/templates/${id}/use`, customizations);
  }

  // Story search and discovery
  async searchStories(query: string, filters?: Partial<StorySearchParams>): Promise<StoriesResponse> {
    return apiClient.get<StoriesResponse>('/stories/search', { query, ...filters });
  }

  async getTrendingStories(params?: { 
    time_range?: string; 
    location?: string; 
    limit?: number; 
  }): Promise<Story[]> {
    return apiClient.get<Story[]>('/stories/trending', params);
  }

  async getStoriesByHashtag(hashtag: string, params?: { limit?: number; skip?: number }): Promise<StoriesResponse> {
    return apiClient.get<StoriesResponse>(`/stories/hashtag/${hashtag}`, params);
  }

  async getStoriesByLocation(location: string, params?: { limit?: number; skip?: number }): Promise<StoriesResponse> {
    return apiClient.get<StoriesResponse>('/stories/location', { location, ...params });
  }

  // Story moderation
  async flagStory(id: string, reason: string): Promise<void> {
    return apiClient.post(`/stories/${id}/flag`, { reason });
  }

  async unflagStory(id: string): Promise<void> {
    return apiClient.post(`/stories/${id}/unflag`);
  }

  async reportStory(id: string, data: {
    reason: string;
    description: string;
    evidence?: string[];
  }): Promise<void> {
    return apiClient.post(`/stories/${id}/report`, data);
  }

  async approveStory(id: string, notes?: string): Promise<Story> {
    return apiClient.post<Story>(`/admin/stories/${id}/approve`, { notes });
  }

  async rejectStory(id: string, reason: string): Promise<Story> {
    return apiClient.post<Story>(`/admin/stories/${id}/reject`, { reason });
  }

  async featureStory(id: string): Promise<Story> {
    return apiClient.post<Story>(`/admin/stories/${id}/feature`);
  }

  async unfeatureStory(id: string): Promise<Story> {
    return apiClient.post<Story>(`/admin/stories/${id}/unfeature`);
  }

  // Bulk operations
  async bulkStoryAction(action: BulkStoryAction): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/stories/bulk-action', action);
  }

  async bulkDeleteStories(storyIds: string[], reason?: string): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/stories/bulk-delete', { 
      story_ids: storyIds, 
      reason 
    });
  }

  async bulkModerateStories(storyIds: string[], action: string, reason: string): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/admin/stories/bulk-moderate', {
      story_ids: storyIds,
      action,
      reason
    });
  }

  // Story analytics
  async getStoryStats(): Promise<StoryStats> {
    return apiClient.get<StoryStats>('/admin/stories/stats');
  }

  async getStoryAnalytics(id: string, timeRange = 'week'): Promise<StoryAnalytics> {
    return apiClient.get<StoryAnalytics>(`/stories/${id}/analytics`, { time_range: timeRange });
  }

  async getMyStoryAnalytics(timeRange = 'week'): Promise<{
    total_views: number;
    unique_viewers: number;
    total_reactions: number;
    total_replies: number;
    avg_completion_rate: number;
    top_performing_stories: Array<Story & {
      views: number;
      completion_rate: number;
    }>;
    viewer_demographics: {
      age_groups: Record<string, number>;
      locations: Record<string, number>;
    };
  }> {
    return apiClient.get('/stories/my-analytics', { time_range: timeRange });
  }

  async getStoryTrends(timeRange = 'week'): Promise<StoryTrends> {
    return apiClient.get<StoryTrends>('/admin/analytics/stories/trends', { time_range: timeRange });
  }

  async getStoryEngagementMetrics(timeRange = 'week'): Promise<{
    avg_views_per_story: number;
    avg_completion_rate: number;
    avg_reactions_per_story: number;
    avg_replies_per_story: number;
    engagement_by_content_type: Record<string, {
      views: number;
      completion_rate: number;
      reactions: number;
    }>;
    peak_viewing_hours: Array<{
      hour: number;
      view_count: number;
    }>;
  }> {
    return apiClient.get('/admin/analytics/stories/engagement', { time_range: timeRange });
  }

  // Story moderation queue
  async getModerationQueue(params?: { 
    status?: string; 
    priority?: string; 
    limit?: number; 
    skip?: number; 
  }): Promise<StoryModeration> {
    return apiClient.get<StoryModeration>('/admin/stories/moderation-queue', params);
  }

  async getPendingStories(params?: { limit?: number; skip?: number }): Promise<StoriesResponse> {
    return apiClient.get<StoriesResponse>('/admin/stories/pending', params);
  }

  async getFlaggedStories(params?: { limit?: number; skip?: number }): Promise<StoriesResponse> {
    return apiClient.get<StoriesResponse>('/admin/stories/flagged', params);
  }

  async getReportedStories(params?: { limit?: number; skip?: number }): Promise<StoriesResponse> {
    return apiClient.get<StoriesResponse>('/admin/stories/reported', params);
  }

  // Story privacy and visibility
  async updateStoryPrivacy(id: string, settings: {
    visibility?: 'public' | 'friends' | 'private' | 'custom';
    allowed_viewers?: string[];
    blocked_viewers?: string[];
  }): Promise<Story> {
    return apiClient.put<Story>(`/stories/${id}/privacy`, settings);
  }

  async hideStoryFromUser(storyId: string, userId: string): Promise<void> {
    return apiClient.post(`/stories/${storyId}/hide-from/${userId}`);
  }

  async unhideStoryFromUser(storyId: string, userId: string): Promise<void> {
    return apiClient.delete(`/stories/${storyId}/hide-from/${userId}`);
  }

  // Story sharing and embedding
  async shareStory(id: string, data: {
    platform: string;
    message?: string;
    recipients?: string[];
  }): Promise<void> {
    return apiClient.post(`/stories/${id}/share`, data);
  }

  async getStoryEmbedCode(id: string, options?: {
    width?: number;
    height?: number;
    autoplay?: boolean;
  }): Promise<{
    embed_code: string;
    preview_url: string;
  }> {
    return apiClient.get(`/stories/${id}/embed`, options);
  }

  // Story export
  async exportStories(filters?: StorySearchParams, format = 'json'): Promise<{
    download_url: string;
    expires_at: string;
  }> {
    return apiClient.post('/admin/stories/export', { filters, format });
  }

  async exportUserStories(userId: string, format = 'json'): Promise<{
    download_url: string;
    expires_at: string;
  }> {
    return apiClient.post(`/stories/user/${userId}/export`, { format });
  }

  // Story cleanup and maintenance
  async cleanupExpiredStories(): Promise<{
    cleaned_stories: number;
    reclaimed_storage: number;
  }> {
    return apiClient.post('/admin/stories/cleanup-expired');
  }

  async getStorageUsage(): Promise<{
    total_stories: number;
    total_storage: number;
    storage_by_type: Record<string, number>;
    oldest_stories: Story[];
    largest_stories: Story[];
  }> {
    return apiClient.get('/admin/stories/storage-usage');
  }

  // Story viewer management
  async getStoryViewers(id: string, params?: { limit?: number; skip?: number }): Promise<Array<{
    user_id: string;
    username: string;
    avatar?: string;
    viewed_at: string;
    view_duration: number;
    completion_percentage: number;
  }>> {
    return apiClient.get(`/stories/${id}/viewers`, params);
  }

  async blockViewerFromStories(userId: string): Promise<void> {
    return apiClient.post(`/stories/block-viewer/${userId}`);
  }

  async unblockViewerFromStories(userId: string): Promise<void> {
    return apiClient.delete(`/stories/block-viewer/${userId}`);
  }

  async getBlockedViewers(): Promise<Array<{
    user_id: string;
    username: string;
    avatar?: string;
    blocked_at: string;
  }>> {
    return apiClient.get('/stories/blocked-viewers');
  }

  // Advanced story features
  async createStoryPoll(storyId: string, poll: {
    question: string;
    options: string[];
    expires_at?: string;
  }): Promise<void> {
    return apiClient.post(`/stories/${storyId}/poll`, poll);
  }

  async voteOnStoryPoll(storyId: string, pollId: string, optionId: string): Promise<void> {
    return apiClient.post(`/stories/${storyId}/poll/${pollId}/vote`, { option_id: optionId });
  }

  async createStoryQuestion(storyId: string, question: string): Promise<void> {
    return apiClient.post(`/stories/${storyId}/question`, { question });
  }

  async answerStoryQuestion(storyId: string, questionId: string, answer: string): Promise<void> {
    return apiClient.post(`/stories/${storyId}/question/${questionId}/answer`, { answer });
  }

  async addStorySticker(storyId: string, sticker: {
    type: string;
    content: string;
    x: number;
    y: number;
    scale?: number;
    rotation?: number;
  }): Promise<Story> {
    return apiClient.post<Story>(`/stories/${storyId}/stickers`, sticker);
  }

  async removeStorySticker(storyId: string, stickerId: string): Promise<Story> {
    return apiClient.delete<Story>(`/stories/${storyId}/stickers/${stickerId}`);
  }
}

export const storyService = new StoryService();