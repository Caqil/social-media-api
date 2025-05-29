import { apiClient } from '../api';
import {
  MediaFile,
  MediaFilesResponse,
  MediaFoldersResponse,
  MediaUploadSessionsResponse,
  MediaProcessingJobsResponse,
  MediaFolder,
  MediaUploadSession,
  MediaProcessingJob,
  MediaLibrary,
  MediaSearchParams,
  MediaUploadRequest,
  MediaUpdateRequest,
  BulkMediaAction,
  MediaModerationRequest,
  CreateFolderRequest,
  UpdateFolderRequest,
  MediaStats,
  MediaAnalytics,
  MediaConfig,
  MediaUploadProgress,
  MediaDownloadRequest,
  MediaEmbedCode,
} from '../types/media';
import { BulkActionResponse } from '../types/common';

export class MediaService {
  // Media file management
  async getMediaFiles(params?: MediaSearchParams): Promise<MediaFilesResponse> {
    return apiClient.get<MediaFilesResponse>('/media', params);
  }

  async getMediaFileById(id: string): Promise<MediaFile> {
    return apiClient.get<MediaFile>(`/media/${id}`);
  }

  async updateMediaFile(id: string, data: MediaUpdateRequest): Promise<MediaFile> {
    return apiClient.put<MediaFile>(`/media/${id}`, data);
  }

  async deleteMediaFile(id: string): Promise<void> {
    return apiClient.delete(`/media/${id}`);
  }

  async restoreMediaFile(id: string): Promise<MediaFile> {
    return apiClient.post<MediaFile>(`/media/${id}/restore`);
  }

  // Media upload
  async uploadMedia(data: MediaUploadRequest): Promise<MediaFile> {
    const formData = new FormData();
    formData.append('file', data.file);
    
    if (data.filename) formData.append('filename', data.filename);
    if (data.alt_text) formData.append('alt_text', data.alt_text);
    if (data.description) formData.append('description', data.description);
    if (data.title) formData.append('title', data.title);
    if (data.tags) formData.append('tags', JSON.stringify(data.tags));
    if (data.category) formData.append('category', data.category);
    if (data.folder_id) formData.append('folder_id', data.folder_id);
    if (data.is_public !== undefined) formData.append('is_public', String(data.is_public));
    if (data.related_to) formData.append('related_to', data.related_to);
    if (data.related_type) formData.append('related_type', data.related_type);
    if (data.auto_delete_at) formData.append('auto_delete_at', data.auto_delete_at);

    return apiClient.upload<MediaFile>('/media/upload', formData);
  }

  async createUploadSession(data: {
    filename: string;
    file_size: number;
    chunk_size?: number;
  }): Promise<MediaUploadSession> {
    return apiClient.post<MediaUploadSession>('/media/upload/session', data);
  }

  async uploadChunk(sessionId: string, chunkIndex: number, chunk: Blob): Promise<{
    uploaded_chunks: number;
    total_chunks: number;
    progress: number;
  }> {
    const formData = new FormData();
    formData.append('chunk', chunk);
    formData.append('chunk_index', String(chunkIndex));

    return apiClient.upload(`/media/upload/session/${sessionId}/chunk`, formData);
  }

  async completeUpload(sessionId: string): Promise<MediaFile> {
    return apiClient.post<MediaFile>(`/media/upload/session/${sessionId}/complete`);
  }

  async cancelUpload(sessionId: string): Promise<void> {
    return apiClient.delete(`/media/upload/session/${sessionId}`);
  }

  async getUploadProgress(sessionId: string): Promise<MediaUploadProgress> {
    return apiClient.get<MediaUploadProgress>(`/media/upload/session/${sessionId}/progress`);
  }

  // Media processing
  async getProcessingJobs(params?: { 
    media_id?: string; 
    status?: string; 
    limit?: number; 
    skip?: number; 
  }): Promise<MediaProcessingJobsResponse> {
    return apiClient.get<MediaProcessingJobsResponse>('/media/processing/jobs', params);
  }

  async getProcessingJob(id: string): Promise<MediaProcessingJob> {
    return apiClient.get<MediaProcessingJob>(`/media/processing/jobs/${id}`);
  }

  async retryProcessingJob(id: string): Promise<MediaProcessingJob> {
    return apiClient.post<MediaProcessingJob>(`/media/processing/jobs/${id}/retry`);
  }

  async cancelProcessingJob(id: string): Promise<void> {
    return apiClient.delete(`/media/processing/jobs/${id}`);
  }

  async reprocessMedia(id: string, jobTypes?: string[]): Promise<MediaProcessingJob[]> {
    return apiClient.post<MediaProcessingJob[]>(`/media/${id}/reprocess`, { job_types: jobTypes });
  }

  // Media folders
  async getFolders(params?: { user_id?: string; parent_folder_id?: string }): Promise<MediaFoldersResponse> {
    return apiClient.get<MediaFoldersResponse>('/media/folders', params);
  }

  async getFolderById(id: string): Promise<MediaFolder> {
    return apiClient.get<MediaFolder>(`/media/folders/${id}`);
  }

  async createFolder(data: CreateFolderRequest): Promise<MediaFolder> {
    return apiClient.post<MediaFolder>('/media/folders', data);
  }

  async updateFolder(id: string, data: UpdateFolderRequest): Promise<MediaFolder> {
    return apiClient.put<MediaFolder>(`/media/folders/${id}`, data);
  }

  async deleteFolder(id: string, deleteContents = false): Promise<void> {
    return apiClient.delete(`/media/folders/${id}`, { delete_contents: deleteContents });
  }

  async moveMediaToFolder(mediaIds: string[], folderId: string | null): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/media/move-to-folder', { 
      media_ids: mediaIds, 
      folder_id: folderId 
    });
  }

  // Media search
  async searchMedia(query: string, filters?: Partial<MediaSearchParams>): Promise<MediaFilesResponse> {
    return apiClient.get<MediaFilesResponse>('/media/search', { query, ...filters });
  }

  async getUserMedia(userId: string, params?: MediaSearchParams): Promise<MediaFilesResponse> {
    return apiClient.get<MediaFilesResponse>(`/media/user/${userId}`, params);
  }

  async getMyMedia(params?: MediaSearchParams): Promise<MediaFilesResponse> {
    return apiClient.get<MediaFilesResponse>('/media/my-media', params);
  }

  async getMediaLibrary(userId?: string): Promise<MediaLibrary> {
    const endpoint = userId ? `/media/library/${userId}` : '/media/library';
    return apiClient.get<MediaLibrary>(endpoint);
  }

  // Media moderation
  async moderateMedia(id: string, data: MediaModerationRequest): Promise<MediaFile> {
    return apiClient.post<MediaFile>(`/media/${id}/moderate`, data);
  }

  async flagMedia(id: string, reason: string): Promise<void> {
    return apiClient.post(`/media/${id}/flag`, { reason });
  }

  async unflagMedia(id: string): Promise<void> {
    return apiClient.post(`/media/${id}/unflag`);
  }

  async approveMedia(id: string, notes?: string): Promise<MediaFile> {
    return apiClient.post<MediaFile>(`/media/${id}/approve`, { notes });
  }

  async rejectMedia(id: string, reason: string): Promise<MediaFile> {
    return apiClient.post<MediaFile>(`/media/${id}/reject`, { reason });
  }

  async getPendingMedia(params?: { limit?: number; skip?: number }): Promise<MediaFilesResponse> {
    return apiClient.get<MediaFilesResponse>('/admin/media/pending', params);
  }

  async getFlaggedMedia(params?: { limit?: number; skip?: number }): Promise<MediaFilesResponse> {
    return apiClient.get<MediaFilesResponse>('/admin/media/flagged', params);
  }

  // Bulk operations
  async bulkMediaAction(action: BulkMediaAction): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/media/bulk-action', action);
  }

  async bulkDeleteMedia(mediaIds: string[], reason?: string): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/media/bulk-delete', { 
      media_ids: mediaIds, 
      reason 
    });
  }

  async bulkModerateMedia(mediaIds: string[], action: string, reason: string): Promise<BulkActionResponse> {
    return apiClient.post<BulkActionResponse>('/media/bulk-moderate', {
      media_ids: mediaIds,
      action,
      reason
    });
  }

  // Media variants and URLs
  async getMediaUrl(id: string, variant?: string): Promise<{ url: string; expires_at?: string }> {
    return apiClient.get(`/media/${id}/url`, { variant });
  }

  async getMediaVariants(id: string): Promise<Array<{
    name: string;
    url: string;
    width?: number;
    height?: number;
    file_size: number;
  }>> {
    return apiClient.get(`/media/${id}/variants`);
  }

  async generateMediaVariant(id: string, variantName: string, options: {
    width?: number;
    height?: number;
    quality?: number;
    format?: string;
  }): Promise<{ url: string; job_id: string }> {
    return apiClient.post(`/media/${id}/variants/${variantName}`, options);
  }

  async deleteMediaVariant(id: string, variantName: string): Promise<void> {
    return apiClient.delete(`/media/${id}/variants/${variantName}`);
  }

  // Media download and sharing
  async downloadMedia(data: MediaDownloadRequest): Promise<Blob> {
    const response = await fetch(
      `${apiClient['baseUrl']}/media/${data.media_id}/download?${new URLSearchParams({
        ...(data.variant && { variant: data.variant }),
        ...(data.track_download && { track_download: 'true' })
      })}`
    );
    return response.blob();
  }

  async getMediaEmbedCode(id: string, options: {
    embed_type?: 'direct' | 'iframe' | 'responsive';
    width?: number;
    height?: number;
    autoplay?: boolean;
    controls?: boolean;
  }): Promise<MediaEmbedCode> {
    return apiClient.get<MediaEmbedCode>(`/media/${id}/embed`, options);
  }

  async shareMedia(id: string, data: {
    platform: string;
    message?: string;
    recipients?: string[];
  }): Promise<void> {
    return apiClient.post(`/media/${id}/share`, data);
  }

  // Media analytics
  async getMediaStats(): Promise<MediaStats> {
    return apiClient.get<MediaStats>('/admin/media/stats');
  }

  async getMediaAnalytics(timeRange = 'week'): Promise<MediaAnalytics> {
    return apiClient.get<MediaAnalytics>('/admin/analytics/media', { time_range: timeRange });
  }

  async getMediaFileAnalytics(id: string, timeRange = 'week'): Promise<{
    views: number;
    downloads: number;
    shares: number;
    unique_viewers: number;
    avg_view_duration: number;
    engagement_rate: number;
    geographic_distribution: Record<string, number>;
    device_breakdown: Record<string, number>;
    referrer_breakdown: Record<string, number>;
  }> {
    return apiClient.get(`/media/${id}/analytics`, { time_range: timeRange });
  }

  async getPopularMedia(params?: { 
    time_range?: string; 
    type?: string; 
    limit?: number; 
  }): Promise<MediaFile[]> {
    return apiClient.get<MediaFile[]>('/media/popular', params);
  }

  async getTrendingMedia(params?: { 
    time_range?: string; 
    type?: string; 
    limit?: number; 
  }): Promise<MediaFile[]> {
    return apiClient.get<MediaFile[]>('/media/trending', params);
  }

  // Media configuration
  async getMediaConfig(): Promise<MediaConfig> {
    return apiClient.get<MediaConfig>('/admin/media/config');
  }

  async updateMediaConfig(config: Partial<MediaConfig>): Promise<MediaConfig> {
    return apiClient.put<MediaConfig>('/admin/media/config', config);
  }

  async getStorageInfo(): Promise<{
    total_storage: number;
    used_storage: number;
    available_storage: number;
    storage_by_type: Record<string, number>;
    storage_by_user: Array<{
      user_id: string;
      username: string;
      storage_used: number;
      file_count: number;
    }>;
  }> {
    return apiClient.get('/admin/media/storage-info');
  }

  // Media cleanup and maintenance
  async cleanupExpiredMedia(): Promise<{
    cleaned_files: number;
    reclaimed_storage: number;
  }> {
    return apiClient.post('/admin/media/cleanup-expired');
  }

  async cleanupUnusedMedia(olderThanDays = 30): Promise<{
    cleaned_files: number;
    reclaimed_storage: number;
  }> {
    return apiClient.post('/admin/media/cleanup-unused', { older_than_days: olderThanDays });
  }

  async optimizeStorage(): Promise<{
    optimized_files: number;
    space_saved: number;
  }> {
    return apiClient.post('/admin/media/optimize-storage');
  }

  async rebuildThumbnails(mediaIds?: string[]): Promise<{
    job_id: string;
    total_files: number;
  }> {
    return apiClient.post('/admin/media/rebuild-thumbnails', { media_ids: mediaIds });
  }

  // Media export and backup
  async exportMedia(filters?: MediaSearchParams, format = 'zip'): Promise<{
    download_url: string;
    expires_at: string;
    file_size: number;
  }> {
    return apiClient.post('/admin/media/export', { filters, format });
  }

  async createBackup(mediaIds?: string[]): Promise<{
    backup_id: string;
    estimated_size: number;
    estimated_time: number;
  }> {
    return apiClient.post('/admin/media/backup', { media_ids: mediaIds });
  }

  async getBackupStatus(backupId: string): Promise<{
    status: 'pending' | 'processing' | 'completed' | 'failed';
    progress: number;
    download_url?: string;
    expires_at?: string;
  }> {
    return apiClient.get(`/admin/media/backup/${backupId}/status`);
  }

  // Content analysis
  async analyzeMediaContent(id: string, analysisTypes?: string[]): Promise<{
    job_id: string;
    estimated_time: number;
  }> {
    return apiClient.post(`/media/${id}/analyze`, { analysis_types: analysisTypes });
  }

  async getContentAnalysis(id: string): Promise<{
    nsfw_score?: number;
    violence_score?: number;
    toxicity_score?: number;
    dominant_colors?: string[];
    objects_detected?: string[];
    faces_detected?: number;
    extracted_text?: string;
    sentiment_score?: number;
    content_tags?: string[];
  }> {
    return apiClient.get(`/media/${id}/content-analysis`);
  }

  async getBulkContentAnalysis(mediaIds: string[]): Promise<Array<{
    media_id: string;
    analysis_results: Record<string, any>;
    confidence_scores: Record<string, number>;
  }>> {
    return apiClient.post('/media/bulk-content-analysis', { media_ids: mediaIds });
  }

  // EXIF and metadata
  async getMediaMetadata(id: string): Promise<{
    exif_data?: Record<string, any>;
    file_info: {
      format: string;
      color_space?: string;
      bit_depth?: number;
      compression?: string;
    };
    camera_info?: {
      make?: string;
      model?: string;
      lens?: string;
      settings?: Record<string, any>;
    };
    location_data?: {
      latitude: number;
      longitude: number;
      altitude?: number;
      location_name?: string;
    };
    creation_date?: string;
    modification_date?: string;
  }> {
    return apiClient.get(`/media/${id}/metadata`);
  }

  async updateMediaMetadata(id: string, metadata: Record<string, any>): Promise<MediaFile> {
    return apiClient.put<MediaFile>(`/media/${id}/metadata`, { metadata });
  }

  async stripMetadata(id: string, metadataTypes: string[]): Promise<MediaFile> {
    return apiClient.post<MediaFile>(`/media/${id}/strip-metadata`, { metadata_types: metadataTypes });
  }
}

export const mediaService = new MediaService();