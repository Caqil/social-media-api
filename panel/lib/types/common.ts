// Common types used across the application
export interface PaginatedResponse<T> {
    data: T[];
    total: number;
    page: number;
    limit: number;
    has_more: boolean;
  }
  
  export interface ApiResponse<T> {
    success: boolean;
    data: T;
    message?: string;
    errors?: Record<string, string[]>;
  }
  
  export interface ApiError {
    message: string;
    status: number;
    code?: string;
    errors?: Record<string, string[]>;
  }
  
  export interface BaseEntity {
    id: string;
    created_at: string;
    updated_at: string;
  }
  
  export interface SoftDeleteEntity extends BaseEntity {
    deleted_at?: string;
    is_deleted: boolean;
  }
  
  export interface TimestampEntity {
    created_at: string;
    updated_at: string;
  }
  
  export interface Location {
    city?: string;
    state?: string;
    country?: string;
    latitude?: number;
    longitude?: number;
  }
  
  export interface Media {
    id: string;
    url: string;
    type: MediaType;
    filename: string;
    size: number;
    mime_type: string;
    alt_text?: string;
    width?: number;
    height?: number;
    duration?: number; // for videos
    thumbnail_url?: string;
  }
  
  export type MediaType = 'image' | 'video' | 'audio' | 'document';
  
  export interface SocialLinks {
    twitter?: string;
    facebook?: string;
    instagram?: string;
    linkedin?: string;
    youtube?: string;
    tiktok?: string;
    website?: string;
  }
  
  export type Status = 'active' | 'inactive' | 'pending' | 'suspended' | 'deleted';
  
  export type Visibility = 'public' | 'friends' | 'private' | 'custom';
  
  export type Priority = 'low' | 'medium' | 'high' | 'urgent';
  
  export type SortOrder = 'asc' | 'desc';
  
  export interface SortOption {
    field: string;
    order: SortOrder;
  }
  
  export interface FilterOption {
    field: string;
    operator: 'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte' | 'in' | 'like' | 'between';
    value: any;
  }
  
  export interface SearchParams {
    query?: string;
    page?: number;
    limit?: number;
    sort_by?: string;
    sort_order?: SortOrder;
    filters?: FilterOption[];
  }
  
  export interface BulkActionRequest {
    action: string;
    target_ids: string[];
    data?: Record<string, any>;
  }
  
  export interface BulkActionResponse {
    success_count: number;
    failed_count: number;
    errors: Array<{
      id: string;
      error: string;
    }>;
  }
  
  export interface ActivityLog {
    id: string;
    user_id: string;
    action: string;
    target_type: string;
    target_id: string;
    metadata?: Record<string, any>;
    ip_address: string;
    user_agent: string;
    created_at: string;
  }
  
  export interface AuditLog extends BaseEntity {
    user_id: string;
    action: string;
    resource_type: string;
    resource_id: string;
    old_values?: Record<string, any>;
    new_values?: Record<string, any>;
    ip_address: string;
    user_agent: string;
  }
  
  export interface Notification {
    id: string;
    recipient_id: string;
    type: string;
    title: string;
    message: string;
    data?: Record<string, any>;
    is_read: boolean;
    created_at: string;
  }
  
  export interface SystemHealth {
    status: 'healthy' | 'warning' | 'critical';
    services: Array<{
      name: string;
      status: 'up' | 'down' | 'degraded';
      response_time?: number;
      last_check: string;
    }>;
    metrics: {
      cpu_usage: number;
      memory_usage: number;
      disk_usage: number;
      active_connections: number;
    };
  }
  
  export interface DateRange {
    start: string;
    end: string;
  }
  
  export interface SelectOption {
    label: string;
    value: string;
    disabled?: boolean;
  }
  
  export interface TableColumn<T = any> {
    key: keyof T;
    label: string;
    sortable?: boolean;
    filterable?: boolean;
    width?: string;
    align?: 'left' | 'center' | 'right';
    render?: (value: any, item: T) => React.ReactNode;
  }
  
  export interface DashboardStats {
    total_users: number;
    active_users: number;
    total_posts: number;
    total_comments: number;
    total_groups: number;
    total_reports: number;
    pending_reports: number;
    storage_used: number;
    storage_limit: number;
  }
  
  export interface ChartDataPoint {
    date: string;
    value: number;
    label?: string;
  }
  
  export interface TimeSeriesData {
    labels: string[];
    datasets: Array<{
      label: string;
      data: number[];
      color?: string;
    }>;
  }
  
  // Form validation types
  export interface ValidationError {
    field: string;
    message: string;
  }
  
  export interface FormState<T> {
    data: T;
    errors: Record<keyof T, string>;
    isSubmitting: boolean;
    isDirty: boolean;
    isValid: boolean;
  }
  
  // Permission types
  export type Permission = 
    | 'users.view'
    | 'users.create'
    | 'users.edit'
    | 'users.delete'
    | 'users.suspend'
    | 'posts.view'
    | 'posts.edit'
    | 'posts.delete'
    | 'comments.view'
    | 'comments.edit'
    | 'comments.delete'
    | 'reports.view'
    | 'reports.resolve'
    | 'analytics.view'
    | 'settings.view'
    | 'settings.edit';
  
  export interface Role {
    id: string;
    name: string;
    description: string;
    permissions: Permission[];
    is_system: boolean;
  }
  
  // Export utility types
  export type Partial<T> = {
    [P in keyof T]?: T[P];
  };
  
  export type Required<T> = {
    [P in keyof T]-?: T[P];
  };
  
  export type Pick<T, K extends keyof T> = {
    [P in K]: T[P];
  };
  
  export type Omit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;