import { apiClient } from '../api';
import {
  Comment,
  CommentsResponse,
  CommentStats,
  CommentSearchParams,
  CreateCommentRequest,
  UpdateCommentRequest,
  BulkCommentAction,
  CommentReactionsResponse,
  CommentReportsResponse,
  CommentThreadsResponse,
  CommentReaction,
  CommentReport,
  CommentThread,
  CommentAnalytics,
  CommentModeration,
  CommentModerationRequest,
  CommentModerationQueue,
  CommentAutoModerationRule,
  CommentAutoModerationResult,
  CommentEngagementInsights,
} from '../types/comment';
import { BulkActionResponse } from '../types/common';

export class CommentService {
  // Comment management
  async getComments(params?: CommentSearchParams): Promise<CommentsResponse> {
    return apiClient.get<CommentsResponse>('/comments', params);
  }

  async getCommentById(id: string): Promise<Comment> {
    return apiClient.get<Comment>(`/comments/${id}`);
  }

  async createComment(data: CreateCommentRequest): Promise<Comment> {
    return apiClient.post<Comment>('/comments', data);
  }

  async updateComment(id: string, data: UpdateCommentRequest): Promise<Comment> {
    return apiClient.put<Comment>(`/comments/${id}`, data);
  }

  async deleteComment(id: string): Promise<void> {
    return apiClient.delete(`/comments/${id}`);
  }

  async restoreComment(id: string): Promise<Comment> {
    return apiClient.post<Comment>(`/comments/${id}/restore`);
  }

  // Post comments
  async getPostComments(postId: string, params?: {
    sort_by?: 'newest' | 'oldest' | 'most_liked' | 'most_replied';
    limit?: number;
    skip?: number;
  }): Promise<CommentsResponse> {
    return apiClient.get<CommentsResponse>(`/posts/${postId}/comments`, params);
  }

  async getCommentReplies(commentId: string, params?: {
    limit?: number;
    skip?: number;
  }): Promise<CommentsResponse> {
    return apiClient.get<CommentsResponse>(`/comments/${commentId}/replies`, params);
  }

  async getCommentThread(commentId: string): Promise<CommentThread> {
    return apiClient.get<CommentThread>(`/comments/${commentId}/thread`);
  }

  // Comment reactions
  async reactToComment(id: string, reactionType: string): Promise<CommentReaction> {
    return apiClient.post<CommentReaction>(`/comments/${id}/react`, { reaction_type: reactionType });
  }

  async removeCommentReaction(id: string): Promise<void> {
    return apiClient.delete(`/comments/${id}/react`);
  }

  async getCommentReactions(id: string, params?: { 
    reaction_type?: string; 
    limit?: number; 
    skip?: number; 
  }): Promise<CommentReactionsResponse> {
    return apiClient.get<CommentReactionsResponse>(`/comments/${id}/reactions`, params);
  }

  // Comment voting (for community features)
  async voteComment(id: string, voteType: 'upvote' | 'downvote'): Promise<void> {
    return apiClient.post(`/comments/${id}/vote`, { vote_type: voteType });
  }

  async removeCommentVote(id: string): Promise<void> {
    return apiClient.delete(`/comments/${id}/vote`);
  }

  // Comment moderation
  async moderateComment(id: string, data: CommentModerationRequest): Promise<CommentModeration> {
    return apiClient.post<CommentModeration>(`/comments/${id}/moderate`, data);
  }

  async flagComment(id: string, reason: string): Promise<void> {
    return apiClient.post(`/comments/${id}/flag`, { reason });
  }

  async unflagComment(id: string): Promise<void> {
    return apiClient.post(`/comments/${id}/unflag`);
  }

  async approveComment(id: string, notes?: string): Promise<Comment> {
    return apiClient.post<Comment>(`/comments/${id}/approve`, { notes });
  }

  async rejectComment(id: string, reason: string): Promise<Comment> {
    return apiClient.post<Comment>(`/comments/${id}/reject`, { reason });
  }

  async hideComment(id: string, reason?: string): Promise<Comment> {
    return apiClient.post<Comment>(`/comments/${id}/hide`, { reason });
  }

  async unhideComment(id: string): Promise<Comment> {
    return apiClient.post<Comment>(`/comments/${id}/unhide`);
  }

  // Pin/unpin comments (for post authors and moderators)
  async pinComment(id: string): Promise<Comment> {
    return apiClient.post<Comment>(`/comments/${id}/pin`);
  }

  async unpinComment(id: string): Promise<Comment> {
    return apiClient.post<Comment>(`/comments/${id}/unpin`);
  }

  // Comment reports
  async reportComment(id: string, data: {
    reason: string;
    description: string;
    evidence?: string[];
  }): Promise<CommentReport> {
    return apiClient.post<CommentReport>(`/comments/${id}/report`, data);
  }

  async getCommentReports(id: string, params?: { 
    limit?: number; 
    skip?: number; 
  }): Promise<CommentReportsResponse> {
    return apiClient.get<CommentReportsResponse>(`/admin/comments/${id}/reports`, params);
  }

  async resolveCommentReport(reportId: string, resolution: string, notes?: string): Promise<void> {
    return apiClient.post(`/admin/reports/${reportId}/resolve`, { resolution, notes });
  }

  // User comments
  async getUserComments(userId: string, params?: CommentSearchParams): Promise<CommentsResponse> {
    return apiClient.get<CommentsResponse>(`/users/${userId}/comments`, params);
  }

  async getCurrentUserComments(params?: CommentSearchParams): Promise<CommentsResponse> {
    return apiClient.get<CommentsResponse>('/comments/my-comments', params);
  }

  // Comment search
  async searchComments(query: string, filters?: Partial<CommentSearchParams>): Promise<CommentsResponse> {
    return apiClient.get<CommentsResponse>('/comments/search', { query, ...filters });
  }

  // Bulk operations
  async bulkCommentAction(action: BulkCommentAction): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/comments/bulk-action', action);
  }

  async bulkDeleteComments(commentIds: string[], reason?: string): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/comments/bulk-delete', { 
      comment_ids: commentIds, 
      reason 
    });
  }

  async bulkModerateComments(commentIds: string[], action: string, reason: string): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/comments/bulk-moderate', {
      comment_ids: commentIds,
      action,
      reason
    });
  }

  // Comment statistics
  async getCommentStats(): Promise<CommentStats> {
    return apiClient.get<CommentStats>('/admin/comments/stats');
  }

  async getCommentAnalytics(id: string, timeRange = 'week'): Promise<CommentAnalytics> {
    return apiClient.get<CommentAnalytics>(`/comments/${id}/analytics`, { time_range: timeRange });
  }

  async getCommentEngagementInsights(timeRange = 'week'): Promise<CommentEngagementInsights> {
    return apiClient.get<CommentEngagementInsights>('/admin/analytics/comments/engagement', { 
      time_range: timeRange 
    });
  }

  // Moderation queue
  async getModerationQueue(params?: { 
    status?: string; 
    priority?: string; 
    limit?: number; 
    skip?: number; 
  }): Promise<CommentModerationQueue> {
    return apiClient.get<CommentModerationQueue>('/admin/comments/moderation-queue', params);
  }

  async getPendingComments(params?: { limit?: number; skip?: number }): Promise<CommentsResponse> {
    return apiClient.get<CommentsResponse>('/admin/comments/pending', params);
  }

  async getFlaggedComments(params?: { limit?: number; skip?: number }): Promise<CommentsResponse> {
    return apiClient.get<CommentsResponse>('/admin/comments/flagged', params);
  }

  async getHighToxicityComments(params?: { 
    threshold?: number; 
    limit?: number; 
    skip?: number; 
  }): Promise<CommentsResponse> {
    return apiClient.get<CommentsResponse>('/admin/comments/high-toxicity', params);
  }

  // Auto-moderation rules
  async getAutoModerationRules(): Promise<CommentAutoModerationRule[]> {
    return apiClient.get<CommentAutoModerationRule[]>('/admin/comments/auto-moderation/rules');
  }

  async createAutoModerationRule(data: Omit<CommentAutoModerationRule, 'id' | 'trigger_count' | 'accuracy_rate' | 'last_triggered_at' | 'created_by' | 'created_at' | 'updated_at'>): Promise<CommentAutoModerationRule> {
    return apiClient.post<CommentAutoModerationRule>('/admin/comments/auto-moderation/rules', data);
  }

  async updateAutoModerationRule(id: string, data: Partial<CommentAutoModerationRule>): Promise<CommentAutoModerationRule> {
    return apiClient.put<CommentAutoModerationRule>(`/admin/comments/auto-moderation/rules/${id}`, data);
  }

  async deleteAutoModerationRule(id: string): Promise<void> {
    return apiClient.delete(`/admin/comments/auto-moderation/rules/${id}`);
  }

  async toggleAutoModerationRule(id: string, enabled: boolean): Promise<CommentAutoModerationRule> {
    return apiClient.patch<CommentAutoModerationRule>(`/admin/comments/auto-moderation/rules/${id}`, { is_active: enabled });
  }

  async testAutoModerationRule(id: string, testContent: string): Promise<{
    triggered: boolean;
    confidence_score: number;
    triggered_conditions: string[];
    proposed_actions: string[];
  }> {
    return apiClient.post(`/admin/comments/auto-moderation/rules/${id}/test`, { content: testContent });
  }

  // Auto-moderation results
  async getAutoModerationResults(params?: {
    rule_id?: string;
    date_from?: string;
    date_to?: string;
    limit?: number;
    skip?: number;
  }): Promise<{
    data: CommentAutoModerationResult[];
    total: number;
  }> {
    return apiClient.get('/admin/comments/auto-moderation/results', params);
  }

  async reviewAutoModerationResult(id: string, action: 'approve' | 'reject', notes?: string): Promise<void> {
    return apiClient.post(`/admin/comments/auto-moderation/results/${id}/review`, { 
      action, 
      notes 
    });
  }

  async markAutoModerationResultAsFalsePositive(id: string): Promise<void> {
    return apiClient.post(`/admin/comments/auto-moderation/results/${id}/false-positive`);
  }

  // Comment threads management
  async getCommentThreads(postId: string, params?: { 
    limit?: number; 
    skip?: number; 
    sort_by?: string; 
  }): Promise<CommentThreadsResponse> {
    return apiClient.get<CommentThreadsResponse>(`/posts/${postId}/comment-threads`, params);
  }

  async collapseCommentThread(commentId: string): Promise<void> {
    return apiClient.post(`/comments/${commentId}/collapse-thread`);
  }

  async expandCommentThread(commentId: string): Promise<void> {
    return apiClient.post(`/comments/${commentId}/expand-thread`);
  }

  // Comment sentiment analysis
  async getCommentSentiment(id: string): Promise<{
    sentiment_score: number;
    sentiment_label: 'positive' | 'neutral' | 'negative';
    confidence: number;
    emotions?: Record<string, number>;
  }> {
    return apiClient.get(`/comments/${id}/sentiment`);
  }

  async getBulkCommentSentiment(commentIds: string[]): Promise<Array<{
    comment_id: string;
    sentiment_score: number;
    sentiment_label: 'positive' | 'neutral' | 'negative';
    confidence: number;
  }>> {
    return apiClient.post('/comments/bulk-sentiment', { comment_ids: commentIds });
  }

  // Comment export
  async exportComments(filters?: CommentSearchParams, format = 'csv'): Promise<{
    download_url: string;
    expires_at: string;
  }> {
    return apiClient.post('/admin/comments/export', { filters, format });
  }

  // Comment editing history
  async getCommentEditHistory(id: string): Promise<Array<{
    id: string;
    old_content: string;
    new_content: string;
    edit_reason?: string;
    edited_by: string;
    edited_at: string;
    is_moderator_edit: boolean;
  }>> {
    return apiClient.get(`/comments/${id}/edit-history`);
  }

  // Advanced analytics
  async getCommentMetrics(timeRange = 'week'): Promise<{
    total_comments: number;
    avg_comments_per_post: number;
    comment_engagement_rate: number;
    top_commenters: Array<{
      user_id: string;
      username: string;
      comment_count: number;
      avg_likes: number;
    }>;
    sentiment_distribution: {
      positive: number;
      neutral: number;
      negative: number;
    };
    response_time_metrics: {
      avg_response_time: number;
      median_response_time: number;
    };
  }> {
    return apiClient.get('/admin/analytics/comments/metrics', { time_range: timeRange });
  }

  async getConversationQuality(postId: string): Promise<{
    quality_score: number;
    constructive_comments: number;
    toxic_comments: number;
    engagement_depth: number;
    unique_participants: number;
    avg_thread_length: number;
  }> {
    return apiClient.get(`/posts/${postId}/conversation-quality`);
  }
}

export const commentService = new CommentService();