import { apiClient } from '../api';
import {
  Report,
  ReportsResponse,
  ReportStats,
  ReportSearchParams,
  CreateReportRequest,
  UpdateReportRequest,
  ResolveReportRequest,
  BulkReportAction,
  ModerationActionsResponse,
  ReportAppealsResponse,
  AutoModerationRulesResponse,
  AutoModerationResultsResponse,
  ModerationAction,
  ReportAppeal,
  CreateAppealRequest,
  AutoModerationRule,
  AutoModerationResult,
  ModerationQueue,
  ModeratorWorkload,
  ModerationInsights,
} from '../types/report';
import { BulkActionResponse } from '../types/common';

export class ReportService {
  // Report management
  async getReports(params?: ReportSearchParams): Promise<ReportsResponse> {
    return apiClient.get<ReportsResponse>('/admin/reports', params);
  }

  async getReportById(id: string): Promise<Report> {
    return apiClient.get<Report>(`/admin/reports/${id}`);
  }

  async createReport(data: CreateReportRequest): Promise<Report> {
    return apiClient.post<Report>('/reports', data);
  }

  async updateReport(id: string, data: UpdateReportRequest): Promise<Report> {
    return apiClient.put<Report>(`/admin/reports/${id}`, data);
  }

  async deleteReport(id: string): Promise<void> {
    return apiClient.delete(`/admin/reports/${id}`);
  }

  async resolveReport(id: string, data: ResolveReportRequest): Promise<Report> {
    return apiClient.post<Report>(`/admin/reports/${id}/resolve`, data);
  }

  async rejectReport(id: string, reason: string): Promise<Report> {
    return apiClient.post<Report>(`/admin/reports/${id}/reject`, { reason });
  }

  async reopenReport(id: string, reason: string): Promise<Report> {
    return apiClient.post<Report>(`/admin/reports/${id}/reopen`, { reason });
  }

  // Report assignment
  async assignReport(id: string, moderatorId: string): Promise<Report> {
    return apiClient.post<Report>(`/admin/reports/${id}/assign`, { moderator_id: moderatorId });
  }

  async unassignReport(id: string): Promise<Report> {
    return apiClient.post<Report>(`/admin/reports/${id}/unassign`);
  }

  async escalateReport(id: string, reason: string, escalateTo?: string): Promise<Report> {
    return apiClient.post<Report>(`/admin/reports/${id}/escalate`, { 
      reason, 
      escalate_to: escalateTo 
    });
  }

  // Report priority and status
  async setPriority(id: string, priority: 'low' | 'medium' | 'high' | 'urgent'): Promise<Report> {
    return apiClient.patch<Report>(`/admin/reports/${id}`, { priority });
  }

  async setSeverity(id: string, severity: 'low' | 'medium' | 'high' | 'critical'): Promise<Report> {
    return apiClient.patch<Report>(`/admin/reports/${id}`, { severity });
  }

  async setFollowUp(id: string, required: boolean, date?: string, note?: string): Promise<Report> {
    return apiClient.patch<Report>(`/admin/reports/${id}`, { 
      follow_up_required: required,
      follow_up_date: date,
      follow_up_note: note
    });
  }

  // Bulk operations
  async bulkReportAction(action: BulkReportAction): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/admin/reports/bulk-action', action);
  }

  async bulkAssignReports(reportIds: string[], moderatorId: string): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/admin/reports/bulk-assign', { 
      report_ids: reportIds, 
      moderator_id: moderatorId 
    });
  }

  async bulkResolveReports(reportIds: string[], resolution: string, note: string): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/admin/reports/bulk-resolve', {
      report_ids: reportIds,
      resolution,
      note
    });
  }

  // Report queues
  async getModerationQueue(): Promise<ModerationQueue> {
    return apiClient.get<ModerationQueue>('/admin/reports/moderation-queue');
  }

  async getPendingReports(params?: { limit?: number; skip?: number }): Promise<ReportsResponse> {
    return apiClient.get<ReportsResponse>('/admin/reports/pending', params);
  }

  async getHighPriorityReports(params?: { limit?: number; skip?: number }): Promise<ReportsResponse> {
    return apiClient.get<ReportsResponse>('/admin/reports/high-priority', params);
  }

  async getEscalatedReports(params?: { limit?: number; skip?: number }): Promise<ReportsResponse> {
    return apiClient.get<ReportsResponse>('/admin/reports/escalated', params);
  }

  async getMyAssignedReports(params?: { limit?: number; skip?: number }): Promise<ReportsResponse> {
    return apiClient.get<ReportsResponse>('/admin/reports/my-assigned', params);
  }

  async getFollowUpReports(params?: { limit?: number; skip?: number }): Promise<ReportsResponse> {
    return apiClient.get<ReportsResponse>('/admin/reports/follow-up', params);
  }

  // Report search and filtering
  async searchReports(query: string, filters?: Partial<ReportSearchParams>): Promise<ReportsResponse> {
    return apiClient.get<ReportsResponse>('/admin/reports/search', { query, ...filters });
  }

  async getReportsByTarget(targetType: string, targetId: string, params?: { limit?: number; skip?: number }): Promise<ReportsResponse> {
    return apiClient.get<ReportsResponse>(`/admin/reports/target/${targetType}/${targetId}`, params);
  }

  async getReportsByUser(userId: string, params?: { limit?: number; skip?: number }): Promise<ReportsResponse> {
    return apiClient.get<ReportsResponse>(`/admin/reports/user/${userId}`, params);
  }

  async getReportsByReporter(reporterId: string, params?: { limit?: number; skip?: number }): Promise<ReportsResponse> {
    return apiClient.get<ReportsResponse>(`/admin/reports/reporter/${reporterId}`, params);
  }

  // Moderation actions
  async getModerationActions(params?: { 
    report_id?: string; 
    moderator_id?: string; 
    limit?: number; 
    skip?: number; 
  }): Promise<ModerationActionsResponse> {
    return apiClient.get<ModerationActionsResponse>('/admin/moderation-actions', params);
  }

  async getModerationAction(id: string): Promise<ModerationAction> {
    return apiClient.get<ModerationAction>(`/admin/moderation-actions/${id}`);
  }

  async reverseModerationAction(id: string, reason: string): Promise<ModerationAction> {
    return apiClient.post<ModerationAction>(`/admin/moderation-actions/${id}/reverse`, { reason });
  }

  async getModerationHistory(targetType: string, targetId: string): Promise<ModerationAction[]> {
    return apiClient.get<ModerationAction[]>(`/admin/moderation-history/${targetType}/${targetId}`);
  }

  // Report appeals
  async getReportAppeals(params?: { 
    status?: string; 
    limit?: number; 
    skip?: number; 
  }): Promise<ReportAppealsResponse> {
    return apiClient.get<ReportAppealsResponse>('/admin/appeals', params);
  }

  async getReportAppeal(id: string): Promise<ReportAppeal> {
    return apiClient.get<ReportAppeal>(`/admin/appeals/${id}`);
  }

  async createAppeal(reportId: string, data: CreateAppealRequest): Promise<ReportAppeal> {
    return apiClient.post<ReportAppeal>(`/reports/${reportId}/appeal`, data);
  }

  async reviewAppeal(id: string, decision: 'approve' | 'reject', notes: string): Promise<ReportAppeal> {
    return apiClient.post<ReportAppeal>(`/admin/appeals/${id}/review`, { 
      decision, 
      notes 
    });
  }

  async getMyAppeals(params?: { limit?: number; skip?: number }): Promise<ReportAppealsResponse> {
    return apiClient.get<ReportAppealsResponse>('/appeals/my-appeals', params);
  }

  // Auto-moderation
  async getAutoModerationRules(): Promise<AutoModerationRulesResponse> {
    return apiClient.get<AutoModerationRulesResponse>('/admin/auto-moderation/rules');
  }

  async createAutoModerationRule(data: Omit<AutoModerationRule, 'id' | 'trigger_count' | 'accuracy_rate' | 'last_triggered_at' | 'created_by' | 'created_at' | 'updated_at'>): Promise<AutoModerationRule> {
    return apiClient.post<AutoModerationRule>('/admin/auto-moderation/rules', data);
  }

  async updateAutoModerationRule(id: string, data: Partial<AutoModerationRule>): Promise<AutoModerationRule> {
    return apiClient.put<AutoModerationRule>(`/admin/auto-moderation/rules/${id}`, data);
  }

  async deleteAutoModerationRule(id: string): Promise<void> {
    return apiClient.delete(`/admin/auto-moderation/rules/${id}`);
  }

  async toggleAutoModerationRule(id: string, enabled: boolean): Promise<AutoModerationRule> {
    return apiClient.patch<AutoModerationRule>(`/admin/auto-moderation/rules/${id}`, { is_active: enabled });
  }

  async testAutoModerationRule(id: string, testData: any): Promise<{
    triggered: boolean;
    confidence_score: number;
    triggered_conditions: string[];
    proposed_actions: string[];
  }> {
    return apiClient.post(`/admin/auto-moderation/rules/${id}/test`, testData);
  }

  // Auto-moderation results
  async getAutoModerationResults(params?: {
    rule_id?: string;
    target_type?: string;
    confidence_threshold?: number;
    date_from?: string;
    date_to?: string;
    limit?: number;
    skip?: number;
  }): Promise<AutoModerationResultsResponse> {
    return apiClient.get<AutoModerationResultsResponse>('/admin/auto-moderation/results', params);
  }

  async reviewAutoModerationResult(id: string, action: 'approve' | 'reject', notes?: string): Promise<AutoModerationResult> {
    return apiClient.post<AutoModerationResult>(`/admin/auto-moderation/results/${id}/review`, { 
      action, 
      notes 
    });
  }

  async markAutoModerationResultAsFalsePositive(id: string): Promise<AutoModerationResult> {
    return apiClient.post<AutoModerationResult>(`/admin/auto-moderation/results/${id}/false-positive`);
  }

  // Report statistics
  async getReportStats(): Promise<ReportStats> {
    return apiClient.get<ReportStats>('/admin/reports/stats');
  }

  async getModerationInsights(timeRange = 'week'): Promise<ModerationInsights> {
    return apiClient.get<ModerationInsights>('/admin/moderation/insights', { time_range: timeRange });
  }

  async getModeratorWorkload(moderatorId?: string): Promise<ModeratorWorkload | ModeratorWorkload[]> {
    if (moderatorId) {
      return apiClient.get<ModeratorWorkload>(`/admin/moderators/${moderatorId}/workload`);
    }
    return apiClient.get<ModeratorWorkload[]>('/admin/moderators/workload');
  }

  async getReportTrends(timeRange = 'month'): Promise<Array<{
    date: string;
    total_reports: number;
    resolved_reports: number;
    pending_reports: number;
    avg_resolution_time: number;
  }>> {
    return apiClient.get('/admin/analytics/reports/trends', { time_range: timeRange });
  }

  // Report categories and reasons
  async getReportReasons(): Promise<Array<{
    reason: string;
    category: string;
    description: string;
    severity_level: string;
  }>> {
    return apiClient.get('/reports/reasons');
  }

  async getReportCategories(): Promise<Array<{
    category: string;
    description: string;
    reasons: string[];
  }>> {
    return apiClient.get('/reports/categories');
  }

  // Duplicate detection
  async findDuplicateReports(reportId: string): Promise<Report[]> {
    return apiClient.get<Report[]>(`/admin/reports/${reportId}/duplicates`);
  }

  async markReportAsDuplicate(reportId: string, originalReportId: string): Promise<Report> {
    return apiClient.post<Report>(`/admin/reports/${reportId}/mark-duplicate`, { 
      original_report_id: originalReportId 
    });
  }

  async mergeDuplicateReports(primaryReportId: string, duplicateReportIds: string[]): Promise<Report> {
    return apiClient.post<Report>(`/admin/reports/${primaryReportId}/merge-duplicates`, { 
      duplicate_report_ids: duplicateReportIds 
    });
  }

  // Report export
  async exportReports(filters?: ReportSearchParams, format = 'csv'): Promise<{
    download_url: string;
    expires_at: string;
  }> {
    return apiClient.post('/admin/reports/export', { filters, format });
  }

  // Report notifications
  async subscribeToReportNotifications(filters: {
    target_types?: string[];
    priorities?: string[];
    categories?: string[];
  }): Promise<void> {
    return apiClient.post('/admin/reports/notifications/subscribe', filters);
  }

  async unsubscribeFromReportNotifications(): Promise<void> {
    return apiClient.delete('/admin/reports/notifications/unsubscribe');
  }

  async getReportNotificationSettings(): Promise<{
    subscribed: boolean;
    filters: Record<string, any>;
    delivery_methods: string[];
  }> {
    return apiClient.get('/admin/reports/notifications/settings');
  }

  // Advanced analytics
  async getReportPatterns(timeRange = 'month'): Promise<{
    repeat_offenders: Array<{
      user_id: string;
      username: string;
      report_count: number;
      severity_score: number;
    }>;
    common_violations: Array<{
      reason: string;
      count: number;
      trend: 'increasing' | 'decreasing' | 'stable';
    }>;
    peak_reporting_times: Array<{
      hour: number;
      day_of_week: number;
      report_count: number;
    }>;
  }> {
    return apiClient.get('/admin/analytics/reports/patterns', { time_range: timeRange });
  }

  async getContentHealthMetrics(): Promise<{
    overall_health_score: number;
    violation_rate: number;
    community_engagement: number;
    moderation_effectiveness: number;
    trend_indicators: Record<string, 'improving' | 'declining' | 'stable'>;
  }> {
    return apiClient.get('/admin/analytics/content-health');
  }

  // Report templates (for common report types)
  async getReportTemplates(): Promise<Array<{
    id: string;
    name: string;
    category: string;
    reason: string;
    description_template: string;
    evidence_requirements: string[];
  }>> {
    return apiClient.get('/reports/templates');
  }

  async useReportTemplate(templateId: string, targetId: string, targetType: string, customData?: Record<string, any>): Promise<Report> {
    return apiClient.post<Report>(`/reports/templates/${templateId}/use`, {
      target_id: targetId,
      target_type: targetType,
      custom_data: customData
    });
  }
}

export const reportService = new ReportService();