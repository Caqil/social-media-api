// types/admin.ts - Complete Type Definitions

// ==================== BASE TYPES ====================
export interface BaseEntity {
    id: string
    created_at: string
    updated_at: string
    deleted_at?: string
  }
  
  export interface PaginationMeta {
    current_page: number
    per_page: number
    total: number
    total_pages: number
    has_next: boolean
    has_previous: boolean
  }
  
  export interface PaginationLinks {
    self: string
    next?: string
    previous?: string
    first?: string
    last?: string
  }
  
  // ==================== ENUMS ====================
  export enum UserRole {
    USER = 'user',
    MODERATOR = 'moderator',
    ADMIN = 'admin',
    SUPER_ADMIN = 'super_admin'
  }
  
  export enum PrivacyLevel {
    PUBLIC = 'public',
    FRIENDS = 'friends',
    PRIVATE = 'private'
  }
  
  export enum ReportReason {
    SPAM = 'spam',
    HARASSMENT = 'harassment',
    HATE_SPEECH = 'hate_speech',
    INAPPROPRIATE_CONTENT = 'inappropriate_content',
    COPYRIGHT = 'copyright',
    FAKE_NEWS = 'fake_news',
    NUDITY = 'nudity',
    VIOLENCE = 'violence',
    SELF_HARM = 'self_harm',
    OTHER = 'other'
  }
  
  export enum ReportStatus {
    PENDING = 'pending',
    REVIEWING = 'reviewing',
    RESOLVED = 'resolved',
    REJECTED = 'rejected'
  }
  
  export enum MediaType {
    IMAGE = 'image',
    VIDEO = 'video',
    AUDIO = 'audio',
    DOCUMENT = 'document'
  }
  
  export enum NotificationType {
    LIKE = 'like',
    COMMENT = 'comment',
    FOLLOW = 'follow',
    MENTION = 'mention',
    SYSTEM = 'system',
    ADMIN = 'admin'
  }
  
  export enum EventStatus {
    DRAFT = 'draft',
    PUBLISHED = 'published',
    CANCELLED = 'cancelled',
    COMPLETED = 'completed'
  }
  
  export enum GroupType {
    PUBLIC = 'public',
    PRIVATE = 'private',
    SECRET = 'secret'
  }
  
  // ==================== USER TYPES ====================
  export interface User extends BaseEntity {
    username: string
    email: string
    first_name?: string
    last_name?: string
    bio?: string
    profile_picture?: string
    cover_image?: string
    date_of_birth?: string
    gender?: string
    country?: string
    city?: string
    website?: string
    phone_number?: string
    is_verified: boolean
    is_active: boolean
    is_suspended: boolean
    is_private: boolean
    role: UserRole
    followers_count: number
    following_count: number
    posts_count: number
    email_verified: boolean
    phone_verified: boolean
    two_factor_enabled: boolean
    privacy_level: PrivacyLevel
    last_active_at?: string
    last_login_at?: string
    login_count: number
    suspension_reason?: string
    suspended_until?: string
    suspended_by?: string
    verification_level: number
    reputation_score: number
    settings: UserSettings
  }
  
  export interface UserSettings {
    notifications_enabled: boolean
    email_notifications: boolean
    push_notifications: boolean
    privacy_level: PrivacyLevel
    show_online_status: boolean
    allow_messages_from: 'everyone' | 'friends' | 'none'
    language: string
    timezone: string
    theme: 'light' | 'dark' | 'auto'
  }
  
  export interface UserResponse {
    id: string
    username: string
    email: string
    first_name?: string
    last_name?: string
    profile_picture?: string
    is_verified: boolean
    is_active: boolean
    role: UserRole
    created_at: string
    last_active_at?: string
  }
  
  // ==================== POST TYPES ====================
  export interface Post extends BaseEntity {
    user_id: string
    content: string
    type: 'text' | 'image' | 'video' | 'poll' | 'shared'
    visibility: PrivacyLevel
    media_urls?: string[]
    media_objects?: Media[]
    hashtags?: string[]
    mentions?: string[]
    location?: string
    likes_count: number
    comments_count: number
    shares_count: number
    views_count: number
    is_reported: boolean
    is_hidden: boolean
    is_pinned: boolean
    is_promoted: boolean
    original_post_id?: string
    group_id?: string
    event_id?: string
    poll_data?: PollData
    scheduled_at?: string
    expires_at?: string
    edit_history?: PostEdit[]
    user?: UserResponse
  }
  
  export interface PostEdit {
    content: string
    edited_at: string
    reason?: string
  }
  
  export interface PollData {
    question: string
    options: PollOption[]
    multiple_choice: boolean
    expires_at?: string
    total_votes: number
  }
  
  export interface PollOption {
    id: string
    text: string
    votes_count: number
  }
  
  export interface PostResponse extends Post {
    user: UserResponse
    media_objects: Media[]
  }
  
  // ==================== COMMENT TYPES ====================
  export interface Comment extends BaseEntity {
    post_id: string
    user_id: string
    parent_id?: string
    content: string
    media_urls?: string[]
    likes_count: number
    replies_count: number
    is_hidden: boolean
    is_reported: boolean
    depth: number
    user?: UserResponse
    post?: PostResponse
    replies?: Comment[]
  }
  
  export interface CommentResponse extends Comment {
    user: UserResponse
    replies: CommentResponse[]
  }
  
  // ==================== GROUP TYPES ====================
  export interface Group extends BaseEntity {
    name: string
    description?: string
    avatar?: string
    cover_image?: string
    type: GroupType
    category: string
    location?: string
    website?: string
    rules?: string[]
    tags: string[]
    members_count: number
    posts_count: number
    is_verified: boolean
    is_active: boolean
    created_by: string
    admin_ids: string[]
    moderator_ids: string[]
    settings: GroupSettings
    creator?: UserResponse
  }
  
  export interface GroupSettings {
    require_approval: boolean
    allow_member_posts: boolean
    allow_member_invites: boolean
    posting_permissions: 'admin_only' | 'admin_moderator' | 'all_members'
    content_moderation: boolean
  }
  
  export interface GroupResponse extends Group {
    creator: UserResponse
    member_role?: 'member' | 'moderator' | 'admin'
    is_member: boolean
  }
  
  export interface GroupMember extends BaseEntity {
    group_id: string
    user_id: string
    role: 'member' | 'moderator' | 'admin'
    joined_at: string
    invited_by?: string
    user?: UserResponse
  }
  
  // ==================== EVENT TYPES ====================
  export interface Event extends BaseEntity {
    title: string
    description?: string
    image?: string
    start_date: string
    end_date?: string
    location?: string
    address?: string
    latitude?: number
    longitude?: number
    price?: number
    currency?: string
    capacity?: number
    category: string
    tags: string[]
    status: EventStatus
    visibility: PrivacyLevel
    created_by: string
    group_id?: string
    attendees_count: number
    interested_count: number
    settings: EventSettings
    creator?: UserResponse
    group?: GroupResponse
  }
  
  export interface EventSettings {
    require_approval: boolean
    allow_guests: boolean
    show_attendee_list: boolean
    send_reminders: boolean
    allow_posts: boolean
  }
  
  export interface EventResponse extends Event {
    creator: UserResponse
    attendance_status?: 'attending' | 'interested' | 'not_attending'
  }
  
  export interface EventAttendee extends BaseEntity {
    event_id: string
    user_id: string
    status: 'attending' | 'interested' | 'not_attending'
    response_date: string
    user?: UserResponse
  }
  
  // ==================== STORY TYPES ====================
  export interface Story extends BaseEntity {
    user_id: string
    content?: string
    media_url: string
    media_type: MediaType
    duration: number
    background_color?: string
    text_color?: string
    font_family?: string
    views_count: number
    likes_count: number
    is_hidden: boolean
    expires_at: string
    user?: UserResponse
  }
  
  export interface StoryResponse extends Story {
    user: UserResponse
    is_viewed: boolean
    is_liked: boolean
  }
  
  // ==================== MESSAGE TYPES ====================
  export interface Message extends BaseEntity {
    conversation_id: string
    sender_id: string
    recipient_id?: string
    content?: string
    message_type: 'text' | 'image' | 'video' | 'audio' | 'file' | 'location'
    media_url?: string
    file_name?: string
    file_size?: number
    is_read: boolean
    read_at?: string
    is_edited: boolean
    edited_at?: string
    reply_to_id?: string
    sender?: UserResponse
    recipient?: UserResponse
  }
  
  export interface Conversation extends BaseEntity {
    type: 'direct' | 'group'
    title?: string
    avatar?: string
    participant_ids: string[]
    last_message_id?: string
    last_message_at?: string
    is_archived: boolean
    is_muted: boolean
    unread_count: number
    participants?: UserResponse[]
    last_message?: Message
  }
  
  export interface MessageResponse extends Message {
    sender: UserResponse
    conversation: Conversation
  }
  
  // ==================== REPORT TYPES ====================
  export interface Report extends BaseEntity {
    reporter_id: string
    target_id: string
    target_type: 'user' | 'post' | 'comment' | 'group' | 'event' | 'message'
    target_user_id?: string
    reason: ReportReason
    description?: string
    status: ReportStatus
    priority: 'low' | 'medium' | 'high' | 'critical'
    resolution?: string
    resolution_note?: string
    resolved_by?: string
    resolved_at?: string
    assigned_to?: string
    evidence_urls?: string[]
    reporter?: UserResponse
    target_user?: UserResponse
    assigned_admin?: UserResponse
    resolved_admin?: UserResponse
  }
  
  export interface ReportResponse extends Report {
    reporter: UserResponse
    target_user?: UserResponse
    assigned_admin?: UserResponse
    resolved_admin?: UserResponse
  }
  
  // ==================== FOLLOW TYPES ====================
  export interface Follow extends BaseEntity {
    follower_id: string
    following_id: string
    status: 'active' | 'blocked' | 'pending'
    follower?: UserResponse
    following?: UserResponse
  }
  
  export interface FollowResponse extends Follow {
    follower: UserResponse
    following: UserResponse
  }
  
  // ==================== LIKE TYPES ====================
  export interface Like extends BaseEntity {
    user_id: string
    target_id: string
    target_type: 'post' | 'comment' | 'story'
    target_user_id?: string
    reaction_type: 'like' | 'love' | 'laugh' | 'angry' | 'sad' | 'wow'
    user?: UserResponse
  }
  
  export interface LikeResponse extends Like {
    user: UserResponse
  }
  
  // ==================== HASHTAG TYPES ====================
  export interface Hashtag extends BaseEntity {
    tag: string
    usage_count: number
    total_usage: number
    trending_score: number
    is_blocked: boolean
    block_reason?: string
    category?: string
    description?: string
    last_used_at?: string
  }
  
  export interface HashtagResponse extends Hashtag {
    recent_posts?: PostResponse[]
  }
  
  export interface HashtagStats {
    tag: string
    count: number
    growth_rate?: number
  }
  
  // ==================== MENTION TYPES ====================
  export interface Mention extends BaseEntity {
    mentioner_id: string
    mentioned_user_id: string
    post_id?: string
    comment_id?: string
    context_type: 'post' | 'comment' | 'story'
    context_id: string
    mentioner?: UserResponse
    mentioned_user?: UserResponse
  }
  
  export interface MentionResponse extends Mention {
    mentioner: UserResponse
    mentioned_user: UserResponse
  }
  
  // ==================== MEDIA TYPES ====================
  export interface Media extends BaseEntity {
    user_id: string
    filename: string
    original_name: string
    file_path: string
    file_size: number
    mime_type: string
    media_type: MediaType
    width?: number
    height?: number
    duration?: number
    thumbnail_url?: string
    is_processed: boolean
    moderation_status: 'pending' | 'approved' | 'rejected' | 'flagged'
    moderation_reason?: string
    storage_provider: string
    public_url: string
    user?: UserResponse
  }
  
  export interface MediaResponse extends Media {
    user: UserResponse
  }
  
  // ==================== NOTIFICATION TYPES ====================
  export interface Notification extends BaseEntity {
    user_id: string
    title: string
    message: string
    type: NotificationType
    data?: any
    is_read: boolean
    read_at?: string
    action_url?: string
    sender_id?: string
    user?: UserResponse
    sender?: UserResponse
  }
  
  export interface NotificationResponse extends Notification {
    user: UserResponse
    sender?: UserResponse
  }
  
  // ==================== ANALYTICS TYPES ====================
  export interface ChartData {
    date: string
    count: number
    label?: string
  }
  
  export interface UserAnalytics {
    total_users: number
    new_users: number
    active_users: number
    user_retention: number
    user_growth_chart: ChartData[]
    age_groups: AgeStats[]
    gender_groups: GenderStats[]
    location_groups: CountryStats[]
    user_activity_by_hour: ChartData[]
    top_countries: CountryStats[]
    signup_sources: SourceStats[]
  }
  
  export interface ContentAnalytics {
    total_posts: number
    total_comments: number
    total_likes: number
    total_shares: number
    engagement_rate: number
    posts_by_category: CategoryStats[]
    top_hashtags: HashtagStats[]
    content_by_hour: ChartData[]
    viral_content: PostResponse[]
    engagement_trends: ChartData[]
  }
  
  export interface EngagementAnalytics {
    total_engagements: number
    engagement_rate: number
    likes_per_post: number
    comments_per_post: number
    shares_per_post: number
    engagement_by_hour: ChartData[]
    engagement_by_content_type: CategoryStats[]
    top_engaging_users: UserResponse[]
  }
  
  export interface RevenueAnalytics {
    total_revenue: number
    monthly_revenue: number
    revenue_growth: number
    revenue_by_source: SourceStats[]
    revenue_trends: ChartData[]
    subscription_metrics: SubscriptionMetrics
    conversion_rates: ConversionStats
  }
  
  export interface AgeStats {
    age_group: string
    count: number
    percentage: number
  }
  
  export interface GenderStats {
    gender: string
    count: number
    percentage: number
  }
  
  export interface CountryStats {
    country: string
    count: number
    percentage: number
  }
  
  export interface CategoryStats {
    category: string
    count: number
    percentage: number
  }
  
  export interface SourceStats {
    source: string
    count: number
    percentage: number
  }
  
  export interface SubscriptionMetrics {
    total_subscribers: number
    monthly_recurring_revenue: number
    churn_rate: number
    lifetime_value: number
  }
  
  export interface ConversionStats {
    visitor_to_signup: number
    signup_to_active: number
    active_to_subscriber: number
  }
  
  // ==================== DASHBOARD TYPES ====================
  export interface DashboardStats {
    total_users: number
    total_posts: number
    total_comments: number
    total_groups: number
    total_events: number
    total_stories: number
    total_messages: number
    total_reports: number
    total_likes: number
    total_follows: number
    active_users: number
    new_users_today: number
    new_posts_today: number
    pending_reports: number
    suspended_users: number
    user_growth_chart: ChartData[]
    post_growth_chart: ChartData[]
    top_hashtags: HashtagStats[]
    top_users: UserResponse[]
    recent_activities: AdminActivity[]
    system_health: SystemHealth
    content_distribution: ContentDistributionStats
  }
  
  export interface AdminActivity {
    id: string
    type: string
    description: string
    admin_id: string
    admin_name: string
    ip_address: string
    user_agent: string
    created_at: string
  }
  
  export interface SystemHealth {
    status: 'healthy' | 'warning' | 'error'
    database_status: string
    cache_status: string
    storage_status: string
    response_time: number
    memory_usage: number
    cpu_usage: number
    disk_usage: number
    uptime: number
    last_updated: string
    alerts: SystemAlert[]
  }
  
  export interface SystemAlert {
    level: 'info' | 'warning' | 'error' | 'critical'
    message: string
    timestamp: string
    resolved: boolean
  }
  
  export interface ContentDistributionStats {
    posts_by_type: Record<string, number>
    users_by_country: Record<string, number>
    content_by_hour: ChartData[]
    engagement_rates: Record<string, number>
    popular_hashtags: HashtagStats[]
  }
  
  // ==================== FILTER TYPES ====================
  export interface BaseFilter {
    page?: number
    limit?: number
    search?: string
    sort_by?: string
    sort_order?: 'asc' | 'desc'
    date_from?: string
    date_to?: string
  }
  
  export interface UserFilter extends BaseFilter {
    is_verified?: boolean
    is_active?: boolean
    is_suspended?: boolean
    role?: UserRole
    country?: string
    age_min?: number
    age_max?: number
    has_profile_picture?: boolean
  }
  
  export interface PostFilter extends BaseFilter {
    user_id?: string
    type?: string
    visibility?: PrivacyLevel
    is_reported?: boolean
    is_hidden?: boolean
    has_media?: boolean
    min_likes?: number
    min_comments?: number
    hashtags?: string[]
  }
  
  export interface ReportFilter extends BaseFilter {
    status?: ReportStatus
    target_type?: string
    reason?: ReportReason
    priority?: string
    assigned_to?: string
    resolved_by?: string
  }
  
  export interface GroupFilter extends BaseFilter {
    type?: GroupType
    category?: string
    is_verified?: boolean
    is_active?: boolean
    min_members?: number
    created_by?: string
  }
  
  export interface EventFilter extends BaseFilter {
    status?: EventStatus
    category?: string
    location?: string
    price_min?: number
    price_max?: number
    created_by?: string
  }
  
  // ==================== ACTION TYPES ====================
  export interface BulkAction {
    ids: string[]
    action: string
    reason?: string
    data?: any
  }
  
  export interface BulkActionResult {
    total: number
    success_count: number
    failure_count: number
    errors: string[]
  }
  
  // ==================== AUTH TYPES ====================
  export interface AdminUser {
    id: string
    email: string
    username: string
    first_name?: string
    last_name?: string
    role: UserRole
    permissions: string[]
    last_login_at?: string
    created_at: string
  }
  
  export interface LoginResponse {
    access_token: string
    refresh_token: string
    expires_in: number
    user: AdminUser
  }
  
  export interface AuthToken {
    access_token: string
    refresh_token: string
    expires_at: string
  }
  
  // ==================== CONFIGURATION TYPES ====================
  export interface SystemConfiguration {
    app_name: string
    app_version: string
    environment: string
    features: FeatureFlags
    rate_limits: RateLimits
    email_settings: EmailSettings
    storage_settings: StorageSettings
    security_settings: SecuritySettings
    content_moderation: ModerationSettings
  }
  
  export interface FeatureFlags {
    [key: string]: boolean
  }
  
  export interface RateLimits {
    posts_per_hour: number
    comments_per_hour: number
    messages_per_hour: number
    likes_per_minute: number
    follows_per_hour: number
    api_requests_per_minute: number
  }
  
  export interface EmailSettings {
    provider: string
    smtp_host?: string
    smtp_port?: number
    smtp_username?: string
    from_email: string
    from_name: string
    templates: EmailTemplates
  }
  
  export interface EmailTemplates {
    welcome: string
    verification: string
    password_reset: string
    notification: string
  }
  
  export interface StorageSettings {
    provider: 'local' | 'aws' | 'gcp' | 'azure'
    bucket_name?: string
    region?: string
    max_file_size: number
    allowed_file_types: string[]
    cdn_url?: string
  }
  
  export interface SecuritySettings {
    password_min_length: number
    require_email_verification: boolean
    enable_two_factor: boolean
    session_timeout: number
    max_login_attempts: number
    lockout_duration: number
  }
  
  export interface ModerationSettings {
    auto_moderation: boolean
    profanity_filter: boolean
    spam_detection: boolean
    adult_content_filter: boolean
    manual_review_threshold: number
  }
  
  // ==================== TABLE TYPES ====================
  export interface TableColumn {
    key: string
    label: string
    sortable?: boolean
    filterable?: boolean
    width?: string
    align?: 'left' | 'center' | 'right'
    render?: (value: any, row: any) => React.ReactNode
  }
  
  export interface TableData {
    [key: string]: any
  }
  
  export interface TableProps {
    data: TableData[]
    columns: TableColumn[]
    loading?: boolean
    pagination?: PaginationMeta
    onPageChange?: (page: number) => void
    onSort?: (column: string, direction: 'asc' | 'desc') => void
    onFilter?: (filters: Record<string, any>) => void
    onRowSelect?: (selectedRows: string[]) => void
    bulkActions?: Array<{
      label: string
      action: string
      variant?: 'default' | 'destructive'
      icon?: React.ComponentType
    }>
    onBulkAction?: (action: string, selectedIds: string[]) => void
  }
  
  // ==================== FORM TYPES ====================
  export interface FormField {
    name: string
    label: string
    type: 'text' | 'email' | 'password' | 'number' | 'select' | 'textarea' | 'checkbox' | 'radio' | 'date' | 'file'
    required?: boolean
    placeholder?: string
    description?: string
    options?: Array<{ label: string; value: string }>
    validation?: {
      min?: number
      max?: number
      pattern?: string
      message?: string
    }
    disabled?: boolean
    hidden?: boolean
  }
  
  export interface FormConfig {
    title: string
    description?: string
    fields: FormField[]
    submitLabel?: string
    cancelLabel?: string
  }
  
  // ==================== WEBSOCKET TYPES ====================
  export interface WebSocketMessage {
    type: string
    data: any
    timestamp: string
  }
  
  export interface RealtimeStats {
    online_users: number
    active_sessions: number
    new_posts: number
    new_comments: number
    new_likes: number
    new_users: number
    new_reports: number
    system_load: number
    timestamp: string
  }
  
  // ==================== EXPORT TYPES ====================
  export interface ExportOptions {
    format: 'csv' | 'excel' | 'json' | 'pdf'
    fields?: string[]
    filters?: any
    date_range?: {
      start: string
      end: string
    }
  }
  
  export interface ExportJob {
    id: string
    type: string
    status: 'pending' | 'processing' | 'completed' | 'failed'
    progress: number
    file_url?: string
    error_message?: string
    created_at: string
    completed_at?: string
  }
  
  // ==================== API RESPONSE TYPES ====================
  export interface ApiError {
    code: string
    message: string
    details?: any
  }
  
  export interface ValidationError {
    field: string
    message: string
    code: string
  }
  
  export interface ApiValidationResponse {
    success: false
    message: string
    errors: ValidationError[]
  }
  
  // ==================== UTILITY TYPES ====================
  export type EntityType = 'user' | 'post' | 'comment' | 'group' | 'event' | 'story' | 'message' | 'media'
  
  export type ActionType = 'create' | 'read' | 'update' | 'delete' | 'moderate' | 'approve' | 'reject'
  
  export type Permission = `${EntityType}.${ActionType}` | 'admin.*' | 'system.*'
  
  export interface AuditLog {
    id: string
    admin_id: string
    action: string
    entity_type: EntityType
    entity_id?: string
    old_values?: any
    new_values?: any
    ip_address: string
    user_agent: string
    timestamp: string
  }
  
  // ==================== COMPONENT PROPS TYPES ====================
  export interface BaseComponentProps {
    className?: string
    children?: React.ReactNode
  }
  
  export interface DataTableProps extends TableProps {
    title?: string
    description?: string
    searchPlaceholder?: string
    emptyMessage?: string
    showSearch?: boolean
    showRefresh?: boolean
    showExport?: boolean
    onRefresh?: () => void
    onExport?: () => void
  }
  
  export interface ModalProps extends BaseComponentProps {
    open: boolean
    onClose: () => void
    title?: string
    description?: string
    size?: 'sm' | 'md' | 'lg' | 'xl'
  }
  
  export interface MetricCardProps {
    title: string
    value: string | number
    change?: string
    changeType?: 'positive' | 'negative' | 'neutral'
    icon?: React.ComponentType
    loading?: boolean
  }
  
  export interface ChartProps {
    data: ChartData[]
    title?: string
    description?: string
    height?: number
    color?: string
    type?: 'line' | 'area' | 'bar' | 'pie'
  }
  
  // ==================== THEME TYPES ====================
  export interface ThemeConfig {
    colors: {
      primary: string
      secondary: string
      success: string
      warning: string
      error: string
      info: string
    }
    fonts: {
      primary: string
      secondary: string
    }
    spacing: {
      xs: string
      sm: string
      md: string
      lg: string
      xl: string
    }
  }
  
  // ==================== SEARCH TYPES ====================
  export interface SearchResult {
    type: EntityType
    id: string
    title: string
    description?: string
    image?: string
    url: string
    relevance_score: number
  }
  
  export interface SearchFilters {
    types?: EntityType[]
    date_range?: {
      start: string
      end: string
    }
    user_id?: string
    status?: string[]
    tags?: string[]
  }
  
  // ==================== INTEGRATION TYPES ====================
  export interface WebhookConfig {
    id: string
    name: string
    url: string
    events: string[]
    active: boolean
    secret?: string
    headers?: Record<string, string>
    created_at: string
  }
  
  export interface APIKey {
    id: string
    name: string
    key: string
    permissions: string[]
    expires_at?: string
    last_used_at?: string
    is_active: boolean
    created_at: string
  }
  
  export interface Integration {
    id: string
    name: string
    type: string
    config: any
    status: 'active' | 'inactive' | 'error'
    last_sync_at?: string
    created_at: string
  }
  
  export default {}