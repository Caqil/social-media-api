import { BaseEntity, SoftDeleteEntity, MediaType, PaginatedResponse } from './common';
import { User } from './auth';

export interface MediaFile extends SoftDeleteEntity {
  filename: string;
  original_filename: string;
  file_path: string;
  url: string;
  type: MediaType;
  mime_type: string;
  file_size: number;
  
  // Image/Video specific
  width?: number;
  height?: number;
  duration?: number; // for videos/audio in seconds
  aspect_ratio?: string;
  
  // Metadata
  alt_text?: string;
  description?: string;
  title?: string;
  tags: string[];
  
  // Storage and CDN
  storage_provider: 'local' | 'aws_s3' | 'cloudinary' | 'azure' | 'gcp';
  cdn_url?: string;
  variants: MediaVariant[];
  
  // Ownership and permissions
  uploaded_by: string;
  is_public: boolean;
  is_processed: boolean;
  processing_status: 'pending' | 'processing' | 'completed' | 'failed';
  processing_error?: string;
  
  // Usage tracking
  download_count: number;
  view_count: number;
  used_in_posts: number;
  used_in_stories: number;
  used_in_messages: number;
  
  // Content analysis
  content_analysis?: MediaContentAnalysis;
  moderation_status: 'approved' | 'pending' | 'rejected' | 'flagged';
  moderated_by?: string;
  moderated_at?: string;
  moderation_reason?: string;
  
  // Expiration and cleanup
  expires_at?: string;
  auto_delete_at?: string;
  
  // Relations
  category: string;
  related_to?: string; // ID of parent entity (post, story, etc.)
  related_type?: 'post' | 'story' | 'message' | 'user_avatar' | 'group_avatar' | 'cover_image';
  
  // Relationships
  uploader?: User;
}

export interface MediaVariant {
  id: string;
  name: string; // thumbnail, small, medium, large, original
  url: string;
  width?: number;
  height?: number;
  file_size: number;
  format: string;
  quality?: number; // 1-100 for images
  bitrate?: number; // for videos/audio
  generated_at: string;
}

export interface MediaContentAnalysis {
  // Image analysis
  dominant_colors?: string[];
  has_faces?: boolean;
  face_count?: number;
  objects_detected?: string[];
  scene_classification?: string[];
  
  // Content safety
  nsfw_score?: number; // 0-1
  violence_score?: number; // 0-1
  toxicity_score?: number; // 0-1
  
  // Text extraction (OCR)
  extracted_text?: string;
  text_language?: string;
  
  // Video analysis
  keyframes?: string[]; // URLs to keyframe images
  scene_changes?: number[]; // timestamps in seconds
  audio_analysis?: {
    has_speech: boolean;
    language_detected?: string;
    volume_levels?: number[];
  };
  
  // Technical metadata
  exif_data?: Record<string, any>;
  geolocation?: {
    latitude: number;
    longitude: number;
  };
  creation_date?: string;
  camera_info?: {
    make?: string;
    model?: string;
    lens?: string;
  };
}

export interface MediaFolder extends BaseEntity {
  name: string;
  description?: string;
  parent_folder_id?: string;
  user_id: string;
  is_public: boolean;
  media_count: number;
  total_size: number;
  
  // Permissions
  can_upload: boolean;
  can_delete: boolean;
  can_share: boolean;
  
  // Relationships
  parent_folder?: MediaFolder;
  subfolders?: MediaFolder[];
  media_files?: MediaFile[];
}

export interface MediaUploadSession extends BaseEntity {
  user_id: string;
  filename: string;
  file_size: number;
  chunk_size: number;
  total_chunks: number;
  uploaded_chunks: number;
  upload_id: string;
  status: 'pending' | 'uploading' | 'completed' | 'failed' | 'cancelled';
  error_message?: string;
  expires_at: string;
}

export interface MediaProcessingJob extends BaseEntity {
  media_id: string;
  job_type: 'thumbnail_generation' | 'video_transcoding' | 'content_analysis' | 'variant_generation';
  status: 'queued' | 'processing' | 'completed' | 'failed';
  progress: number; // 0-100
  started_at?: string;
  completed_at?: string;
  error_message?: string;
  result_data?: Record<string, any>;
}

export interface MediaLibrary {
  user_id: string;
  total_files: number;
  total_size: number;
  storage_limit: number;
  storage_used_percentage: number;
  
  breakdown_by_type: Record<MediaType, {
    count: number;
    size: number;
  }>;
  
  recent_uploads: MediaFile[];
  most_used_files: MediaFile[];
  
  folders: MediaFolder[];
}

// Search and filter types
export interface MediaSearchParams {
  query?: string;
  type?: MediaType[];
  category?: string[];
  uploaded_by?: string;
  is_public?: boolean;
  moderation_status?: ('approved' | 'pending' | 'rejected' | 'flagged')[];
  processing_status?: ('pending' | 'processing' | 'completed' | 'failed')[];
  has_analysis?: boolean;
  min_size?: number;
  max_size?: number;
  min_width?: number;
  max_width?: number;
  min_height?: number;
  max_height?: number;
  min_duration?: number;
  max_duration?: number;
  date_from?: string;
  date_to?: string;
  tags?: string[];
  folder_id?: string;
  related_type?: ('post' | 'story' | 'message' | 'user_avatar' | 'group_avatar')[];
  nsfw_threshold?: number;
  sort_by?: 'created_at' | 'file_size' | 'download_count' | 'view_count' | 'filename';
  sort_order?: 'asc' | 'desc';
  page?: number;
  limit?: number;
}

export interface MediaFilters {
  types: MediaType[];
  categories: string[];
  uploaders: string[];
  moderation_statuses: string[];
  processing_statuses: string[];
  size_ranges: Array<{
    min: number;
    max: number;
    label: string;
  }>;
  dimension_ranges: Array<{
    min_width: number;
    max_width: number;
    min_height: number;
    max_height: number;
    label: string;
  }>;
  duration_ranges: Array<{
    min: number;
    max: number;
    label: string;
  }>;
  date_range: {
    start?: string;
    end?: string;
  };
  tags: string[];
  folders: string[];
  usage_types: string[];
  has_analysis: boolean | null;
  is_public: boolean | null;
}

// Form types
export interface MediaUploadRequest {
  file: File;
  filename?: string;
  alt_text?: string;
  description?: string;
  title?: string;
  tags?: string[];
  category?: string;
  folder_id?: string;
  is_public?: boolean;
  related_to?: string;
  related_type?: 'post' | 'story' | 'message' | 'user_avatar' | 'group_avatar' | 'cover_image';
  auto_delete_at?: string;
}

export interface MediaUpdateRequest {
  filename?: string;
  alt_text?: string;
  description?: string;
  title?: string;
  tags?: string[];
  category?: string;
  folder_id?: string;
  is_public?: boolean;
  auto_delete_at?: string;
}

export interface BulkMediaAction {
  action: 'delete' | 'move_to_folder' | 'change_visibility' | 'add_tags' | 'remove_tags' | 'approve' | 'reject';
  media_ids: string[];
  data?: {
    folder_id?: string;
    is_public?: boolean;
    tags?: string[];
    reason?: string;
  };
}

export interface MediaModerationRequest {
  action: 'approve' | 'reject' | 'flag';
  reason?: string;
  notes?: string;
}

export interface CreateFolderRequest {
  name: string;
  description?: string;
  parent_folder_id?: string;
  is_public?: boolean;
}

export interface UpdateFolderRequest {
  name?: string;
  description?: string;
  parent_folder_id?: string;
  is_public?: boolean;
}

// API response types
export type MediaFilesResponse = PaginatedResponse<MediaFile>;
export type MediaFoldersResponse = PaginatedResponse<MediaFolder>;
export type MediaUploadSessionsResponse = PaginatedResponse<MediaUploadSession>;
export type MediaProcessingJobsResponse = PaginatedResponse<MediaProcessingJob>;

// Statistics types
export interface MediaStats {
  total_files: number;
  total_size: number; // in bytes
  files_today: number;
  files_week: number;
  files_month: number;
  storage_used: number; // in bytes
  storage_limit: number; // in bytes
  storage_used_percentage: number;
  
  breakdown_by_type: Record<MediaType, {
    count: number;
    size: number;
    percentage: number;
  }>;
  
  breakdown_by_category: Record<string, {
    count: number;
    size: number;
  }>;
  
  top_uploaders: Array<{
    user_id: string;
    username: string;
    file_count: number;
    total_size: number;
  }>;
  
  most_downloaded: MediaFile[];
  most_viewed: MediaFile[];
  
  processing_stats: {
    pending_jobs: number;
    processing_jobs: number;
    failed_jobs: number;
    avg_processing_time: number;
  };
  
  moderation_stats: {
    pending_review: number;
    approved: number;
    rejected: number;
    flagged: number;
  };
  
  cleanup_stats: {
    files_to_expire: number;
    files_to_delete: number;
    storage_to_reclaim: number;
  };
}

export interface MediaAnalytics {
  date: string;
  uploads: number;
  total_size_uploaded: number;
  downloads: number;
  views: number;
  processing_jobs: number;
  storage_used: number;
  
  popular_formats: Record<string, number>;
  upload_sources: Record<string, number>;
  
  content_analysis_results: {
    nsfw_flagged: number;
    violence_detected: number;
    faces_detected: number;
    text_extracted: number;
  };
}

// Configuration types
export interface MediaConfig {
  max_file_size: number; // in bytes
  allowed_types: MediaType[];
  allowed_extensions: string[];
  max_dimensions: {
    width: number;
    height: number;
  };
  max_duration: number; // in seconds for videos/audio
  
  thumbnail_sizes: Array<{
    name: string;
    width: number;
    height: number;
    quality: number;
  }>;
  
  video_transcoding: {
    enabled: boolean;
    formats: string[];
    qualities: number[];
    max_bitrate: number;
  };
  
  content_analysis: {
    enabled: boolean;
    nsfw_detection: boolean;
    violence_detection: boolean;
    face_detection: boolean;
    text_extraction: boolean;
    object_detection: boolean;
  };
  
  storage: {
    provider: 'local' | 'aws_s3' | 'cloudinary' | 'azure' | 'gcp';
    bucket_name?: string;
    region?: string;
    cdn_enabled: boolean;
    cdn_url?: string;
  };
  
  cleanup: {
    auto_delete_unused: boolean;
    unused_threshold_days: number;
    delete_expired: boolean;
    compress_old_files: boolean;
    compression_threshold_days: number;
  };
}

// Utility types
export interface MediaUploadProgress {
  session_id: string;
  filename: string;
  uploaded_bytes: number;
  total_bytes: number;
  percentage: number;
  estimated_time_remaining: number; // in seconds
  upload_speed: number; // bytes per second
  status: 'uploading' | 'processing' | 'completed' | 'failed';
  error_message?: string;
}

export interface MediaDownloadRequest {
  media_id: string;
  variant?: string; // specific variant name
  track_download?: boolean;
}

export interface MediaEmbedCode {
  media_id: string;
  embed_type: 'direct' | 'iframe' | 'responsive';
  width?: number;
  height?: number;
  autoplay?: boolean;
  controls?: boolean;
  html: string;
}