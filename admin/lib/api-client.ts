// lib/api-client.ts - Complete Fetch Implementation (No Axios)
import { DashboardStats } from '@/types/admin'

export interface ApiResponse<T = any> {
  success: boolean
  message: string
  data: T
}

export interface PaginatedResponse<T = any> {
  success: boolean
  message: string
  data: T[]
  pagination: {
    current_page: number
    per_page: number
    total: number
    total_pages: number
    has_next: boolean
    has_previous: boolean
  }
  links?: {
    self: string
    next?: string
    previous?: string
    first?: string
    last?: string
  }
}

export interface LoginResponse {
  access_token: string
  refresh_token?: string
  expires_in?: number
  token_type?: string
  user: {
    id: string
    email: string
    username: string
    first_name?: string
    last_name?: string
    role: string
    permissions?: string[]
    is_verified: boolean
    created_at: string
  }
}

export interface RefreshResponse {
  access_token: string
  refresh_token?: string
  expires_in?: number
  token_type?: string
}

class FetchApiClient {
  private baseURL: string
  private isRefreshing = false
  private refreshPromise: Promise<string> | null = null
  private wsConnections: Map<string, WebSocket> = new Map()

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    console.log('🔧 Fetch API Client initialized:', this.baseURL)
  }

  private getStoredToken(): string | null {
    if (typeof window === 'undefined') return null
    try {
      return localStorage.getItem('admin_token')
    } catch {
      return null
    }
  }

  private setStoredToken(token: string): void {
    if (typeof window === 'undefined') return
    try {
      localStorage.setItem('admin_token', token)
      console.log('✅ Token stored')
    } catch (error) {
      console.error('❌ Failed to store token:', error)
    }
  }

  private clearAuthData(): void {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_refresh_token')
      localStorage.removeItem('admin_user')
      console.log('🧹 Auth data cleared')
    }
  }

  private async refreshTokenIfNeeded(): Promise<string> {
    // If already refreshing, wait for the existing refresh
    if (this.isRefreshing && this.refreshPromise) {
      console.log('🔄 Waiting for existing refresh...')
      return this.refreshPromise
    }

    const refreshToken = typeof window !== 'undefined' ? localStorage.getItem('admin_refresh_token') : null
    
    if (!refreshToken) {
      throw new Error('No refresh token available')
    }

    this.isRefreshing = true
    console.log('🔄 Starting token refresh...')
    
    this.refreshPromise = this.performRefresh(refreshToken)
    
    try {
      const newToken = await this.refreshPromise
      console.log('✅ Token refresh completed')
      return newToken
    } finally {
      this.isRefreshing = false
      this.refreshPromise = null
    }
  }

  private async performRefresh(refreshToken: string): Promise<string> {
    const response = await fetch(`${this.baseURL}/api/v1/admin/public/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        refresh_token: refreshToken
      })
    })

    if (!response.ok) {
      throw new Error(`Refresh failed: ${response.status}`)
    }

    const data = await response.json()

    if (data.success && data.data?.access_token) {
      const newToken = data.data.access_token
      
      // Store new tokens
      this.setStoredToken(newToken)
      
      if (data.data.refresh_token && typeof window !== 'undefined') {
        localStorage.setItem('admin_refresh_token', data.data.refresh_token)
      }
      
      return newToken
    } else {
      throw new Error('Invalid refresh response')
    }
  }
  private async makeRequest<T>(
    url: string,
    options: RequestInit = {},
    retryOnAuth = true
  ): Promise<ApiResponse<T>> {
    const fullUrl = `${this.baseURL}${url}`;

    // Get current token
    const token = this.getStoredToken();

    // Prepare headers as a Headers object
    const headers = new Headers({
      'Content-Type': 'application/json',
      ...options.headers,
    });

    // Add auth header if token exists and not auth endpoint
    if (token && !url.includes('/auth/')) {
      headers.set('Authorization', `Bearer ${token}`);
      console.log(`🔑 Request: ${options.method || 'GET'} ${url} (with token)`);
    } else {
      console.log(`🔓 Request: ${options.method || 'GET'} ${url} (no token)`);
    }

    // Make the request
    const response = await fetch(fullUrl, {
      ...options,
      headers,
    });

    console.log(`📡 Response: ${options.method || 'GET'} ${url} - ${response.status}`);

    // Handle 401 with retry
    if (response.status === 401 && retryOnAuth && !url.includes('/auth/')) {
      console.log('🔄 Got 401, attempting token refresh...');

      try {
        // Refresh token and retry
        await this.refreshTokenIfNeeded();

        // Retry the request with new token (no retry on second attempt)
        return this.makeRequest<T>(url, options, false);
      } catch (refreshError) {
        console.error('❌ Token refresh failed:', refreshError);
        this.clearAuthData();

        if (typeof window !== 'undefined') {
          window.location.href = '/admin/login';
        }
        throw refreshError;
      }
    }

    // Handle other errors
    if (!response.ok) {
      const errorText = await response.text();
      console.error(`❌ Request failed: ${response.status} - ${errorText}`);
      throw new Error(`Request failed: ${response.status}`);
    }

    // Parse response
    const data = await response.json();
    return data as ApiResponse<T>;
  }

  // Generic HTTP methods
  async get<T>(url: string, params?: any): Promise<ApiResponse<T>> {
    let finalUrl = url
    if (params) {
      const searchParams = new URLSearchParams()
      Object.keys(params).forEach(key => {
        if (params[key] !== undefined && params[key] !== null) {
          searchParams.append(key, params[key].toString())
        }
      })
      if (searchParams.toString()) {
        finalUrl += '?' + searchParams.toString()
      }
    }
    
    return this.makeRequest<T>(finalUrl, { method: 'GET' })
  }

  async post<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.makeRequest<T>(url, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  async put<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.makeRequest<T>(url, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  async patch<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.makeRequest<T>(url, {
      method: 'PATCH',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  async delete<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.makeRequest<T>(url, {
      method: 'DELETE',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  // ==================== AUTHENTICATION ====================
  async adminLogin(email: string, password: string): Promise<ApiResponse<LoginResponse>> {
    console.log('🔄 Starting admin login...')
    
    // Clear any existing auth data first
    this.clearAuthData()
    
    const response = await fetch(`${this.baseURL}/api/v1/admin/public/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email,
        password,
      })
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.message || `Login failed: ${response.status}`)
    }

    const data = await response.json()
    console.log('✅ Login API call successful')
    
    // Store tokens if login successful
    if (data.success && data.data) {
      const loginData = data.data as LoginResponse
      const { access_token, refresh_token, user } = loginData
      
      if (typeof window !== 'undefined') {
        if (access_token) {
          this.setStoredToken(access_token)
          console.log('✅ Access token stored')
        }
        if (refresh_token) {
          localStorage.setItem('admin_refresh_token', refresh_token)
          console.log('✅ Refresh token stored')
        }
        if (user) {
          localStorage.setItem('admin_user', JSON.stringify(user))
          console.log('✅ User data stored')
        }
      }
    }
    
    return data as ApiResponse<LoginResponse>
  }

  async adminLogout(): Promise<void> {
    console.log('🔄 Starting admin logout...')
    
    try {
      const token = this.getStoredToken()
      if (token) {
        await fetch(`${this.baseURL}/api/v1/admin/public/auth/logout`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
        })
      }
    } catch (error) {
      console.warn('⚠️ Logout API call failed:', error)
    } finally {
      this.clearAuthData()
      console.log('✅ Logout completed')
    }
  }

  async refreshToken(): Promise<ApiResponse<RefreshResponse>> {
    console.log('🔄 Manual token refresh requested...')
    try {
      const newToken = await this.refreshTokenIfNeeded()
      return { 
        success: true, 
        message: 'Token refreshed successfully',
        data: { access_token: newToken } as RefreshResponse 
      }
    } catch (error) {
      console.error('❌ Manual token refresh failed:', error)
      throw error
    }
  }

  async adminForgotPassword(email: string) {
    return this.post('/api/v1/admin/public/auth/forgot-password', { email })
  }

  async adminResetPassword(token: string, newPassword: string, confirmPassword: string) {
    return this.post('/api/v1/admin/public/auth/reset-password', { 
      token, 
      new_password: newPassword,
      confirm_password: confirmPassword
    })
  }

  // ==================== PUBLIC ADMIN ROUTES ====================
  async getPublicSystemStatus() {
    return this.get('/api/v1/admin/public/status')
  }

  async getPublicHealthCheck() {
    return this.get('/api/v1/admin/public/health')
  }

  // ==================== DASHBOARD ====================
  async getDashboard() {
    return this.get('/api/v1/admin/dashboard')
  }

  async getDashboardStats(): Promise<ApiResponse<DashboardStats>> {
    console.log('🔄 Fetching dashboard stats...')
    return this.get<DashboardStats>('/api/v1/admin/dashboard/stats')
  }

  // ==================== USER MANAGEMENT ====================
  async getUsers(params?: {
    page?: number
    limit?: number
    status?: string
    role?: string
    sort_by?: string
    sort_order?: 'asc' | 'desc'
    search?: string
    date_from?: string
    date_to?: string
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/users', params)
  }

  async searchUsers(params?: {
    query?: string
    limit?: number
    skip?: number
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/users/search', params)
  }

  async getUser(id: string) {
    return this.get(`/api/v1/admin/users/${id}`)
  }

  async getUserStats(id: string) {
    return this.get(`/api/v1/admin/users/${id}/stats`)
  }

  async updateUserStatus(id: string, data: { 
    is_active?: boolean
    is_suspended?: boolean
    reason?: string 
  }) {
    return this.put(`/api/v1/admin/users/${id}/status`, data)
  }

  async verifyUser(id: string) {
    return this.put(`/api/v1/admin/users/${id}/verify`)
  }

  async deleteUser(id: string, reason: string) {
    return this.delete(`/api/v1/admin/users/${id}`, { reason })
  }

  async bulkUserAction(data: { 
    user_ids: string[]
    action: string
    reason?: string
    duration?: string
  }) {
    return this.post('/api/v1/admin/users/bulk/actions', data)
  }

  async exportUsers(params?: {
    format?: string
    date_from?: string
    date_to?: string
  }) {
    return this.get('/api/v1/admin/users/export', params)
  }

  // ==================== POST MANAGEMENT ====================
  async getPosts(params?: {
    page?: number
    limit?: number
    status?: string
    sort_by?: string
    user_id?: string
    search?: string
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/posts', params)
  }

  async searchPosts(params?: {
    query?: string
    limit?: number
    skip?: number
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/posts/search', params)
  }

  async getPost(id: string) {
    return this.get(`/api/v1/admin/posts/${id}`)
  }

  async getPostStats(id: string) {
    return this.get(`/api/v1/admin/posts/${id}/stats`)
  }

  async hidePost(id: string, reason: string) {
    return this.put(`/api/v1/admin/posts/${id}/hide`, { reason })
  }

  async deletePost(id: string, reason: string) {
    return this.delete(`/api/v1/admin/posts/${id}`, { reason })
  }

  async bulkPostAction(data: { 
    post_ids: string[]
    action: string
    reason?: string 
  }) {
    return this.post('/api/v1/admin/posts/bulk/actions', data)
  }

  async exportPosts(params?: {
    format?: string
    date_from?: string
  }) {
    return this.get('/api/v1/admin/posts/export', params)
  }

  // ==================== COMMENT MANAGEMENT ====================
  async getComments(params?: {
    page?: number
    limit?: number
    status?: string
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/comments', params)
  }

  async getComment(id: string) {
    return this.get(`/api/v1/admin/comments/${id}`)
  }

  async hideComment(id: string, reason: string) {
    return this.put(`/api/v1/admin/comments/${id}/hide`, { reason })
  }

  async deleteComment(id: string, reason: string) {
    return this.delete(`/api/v1/admin/comments/${id}`, { reason })
  }

  async bulkCommentAction(data: { 
    comment_ids: string[]
    action: string
    reason?: string 
  }) {
    return this.post('/api/v1/admin/comments/bulk/actions', data)
  }

  // ==================== GROUP MANAGEMENT ====================
  async getGroups(params?: {
    page?: number
    limit?: number
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/groups', params)
  }

  async getGroup(id: string) {
    return this.get(`/api/v1/admin/groups/${id}`)
  }

  async getGroupMembers(id: string, params?: any) {
    return this.get<PaginatedResponse>(`/api/v1/admin/groups/${id}/members`, params)
  }

  async updateGroupStatus(id: string, data: { 
    status?: string
    reason?: string 
  }) {
    return this.put(`/api/v1/admin/groups/${id}/status`, data)
  }

  async deleteGroup(id: string, reason: string) {
    return this.delete(`/api/v1/admin/groups/${id}`, { reason })
  }

  async bulkGroupAction(data: { 
    group_ids: string[]
    action: string
    reason?: string 
  }) {
    return this.post('/api/v1/admin/groups/bulk/actions', data)
  }

  // ==================== EVENT MANAGEMENT ====================
  async getEvents(params?: {
    page?: number
    limit?: number
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/events', params)
  }

  async getEvent(id: string) {
    return this.get(`/api/v1/admin/events/${id}`)
  }

  async getEventAttendees(id: string, params?: any) {
    return this.get<PaginatedResponse>(`/api/v1/admin/events/${id}/attendees`, params)
  }

  async updateEventStatus(id: string, data: { 
    status: string
    reason?: string 
  }) {
    return this.put(`/api/v1/admin/events/${id}/status`, data)
  }

  async deleteEvent(id: string, reason: string) {
    return this.delete(`/api/v1/admin/events/${id}`, { reason })
  }

  async bulkEventAction(data: { 
    event_ids: string[]
    action: string
    reason?: string 
  }) {
    return this.post('/api/v1/admin/events/bulk/actions', data)
  }

  // ==================== STORY MANAGEMENT ====================
  async getStories(params?: {
    page?: number
    limit?: number
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/stories', params)
  }

  async getStory(id: string) {
    return this.get(`/api/v1/admin/stories/${id}`)
  }

  async hideStory(id: string, reason: string) {
    return this.put(`/api/v1/admin/stories/${id}/hide`, { reason })
  }

  async deleteStory(id: string, reason: string) {
    return this.delete(`/api/v1/admin/stories/${id}`, { reason })
  }

  async bulkStoryAction(data: { 
    story_ids: string[]
    action: string
    reason?: string 
  }) {
    return this.post('/api/v1/admin/stories/bulk/actions', data)
  }

  // ==================== MESSAGE MANAGEMENT ====================
  async getMessages(params?: {
    page?: number
    limit?: number
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/messages', params)
  }

  async getMessage(id: string) {
    return this.get(`/api/v1/admin/messages/${id}`)
  }

  async getConversations(params?: {
    page?: number
    limit?: number
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/messages/conversations', params)
  }

  async getConversation(id: string) {
    return this.get(`/api/v1/admin/messages/conversations/${id}`)
  }

  async deleteMessage(id: string, reason: string) {
    return this.delete(`/api/v1/admin/messages/${id}`, { reason })
  }

  async bulkMessageAction(data: { 
    message_ids: string[]
    action: string
    reason?: string 
  }) {
    return this.post('/api/v1/admin/messages/bulk/actions', data)
  }

  // ==================== REPORT MANAGEMENT ====================
  async getReports(params?: {
    status?: string
    target_type?: string
    limit?: number
    skip?: number
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/reports', params)
  }

  async getReport(id: string) {
    return this.get(`/api/v1/admin/reports/${id}`)
  }

  async updateReportStatus(id: string, data: {
    status?: string
    notes?: string
  }) {
    return this.put(`/api/v1/admin/reports/${id}/status`, data)
  }

  async assignReport(id: string, assigneeId: string) {
    return this.put(`/api/v1/admin/reports/${id}/assign`, { assigned_to: assigneeId })
  }

  async resolveReport(id: string, data: { 
    resolution: string
    note?: string 
  }) {
    return this.post(`/api/v1/admin/reports/${id}/resolve`, data)
  }

  async rejectReport(id: string, data: { 
    note: string 
  }) {
    return this.post(`/api/v1/admin/reports/${id}/reject`, data)
  }

  async bulkReportAction(data: { 
    report_ids: string[]
    action: string
    resolution?: string
    reason?: string 
  }) {
    return this.post('/api/v1/admin/reports/bulk/actions', data)
  }

  async getReportStats() {
    return this.get('/api/v1/admin/reports/stats')
  }

  async getReportSummary() {
    return this.get('/api/v1/admin/reports/stats/summary')
  }

  // ==================== FOLLOW MANAGEMENT ====================
  async getFollows(params?: {
    page?: number
    limit?: number
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/follows', params)
  }

  async getFollow(id: string) {
    return this.get(`/api/v1/admin/follows/${id}`)
  }

  async deleteFollow(id: string, reason: string) {
    return this.delete(`/api/v1/admin/follows/${id}`, { reason })
  }

  async getRelationships(userId: string) {
    return this.get('/api/v1/admin/follows/relationships', { user_id: userId })
  }

  async bulkFollowAction(data: { 
    follow_ids: string[]
    action: string
    reason?: string 
  }) {
    return this.post('/api/v1/admin/follows/bulk/actions', data)
  }

  // ==================== LIKE MANAGEMENT ====================
  async getLikes(params?: {
    page?: number
    limit?: number
    target_type?: string
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/likes', params)
  }

  async getLikeStats() {
    return this.get('/api/v1/admin/likes/stats')
  }

  async deleteLike(id: string, reason: string) {
    return this.delete(`/api/v1/admin/likes/${id}`, { reason })
  }

  async bulkLikeAction(data: { 
    like_ids: string[]
    action: string
    reason?: string 
  }) {
    return this.post('/api/v1/admin/likes/bulk/actions', data)
  }

  // ==================== HASHTAG MANAGEMENT ====================
  async getHashtags(params?: {
    page?: number
    limit?: number
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/hashtags', params)
  }

  async getHashtag(id: string) {
    return this.get(`/api/v1/admin/hashtags/${id}`)
  }

  async getTrendingHashtags(limit: number = 20) {
    return this.get('/api/v1/admin/hashtags/trending', { limit })
  }

  async blockHashtag(id: string, reason: string) {
    return this.put(`/api/v1/admin/hashtags/${id}/block`, { reason })
  }

  async unblockHashtag(id: string) {
    return this.put(`/api/v1/admin/hashtags/${id}/unblock`)
  }

  async deleteHashtag(id: string, reason: string) {
    return this.delete(`/api/v1/admin/hashtags/${id}`, { reason })
  }

  async bulkHashtagAction(data: { 
    hashtag_ids: string[]
    action: string
    reason?: string 
  }) {
    return this.post('/api/v1/admin/hashtags/bulk/actions', data)
  }

  // ==================== MENTION MANAGEMENT ====================
  async getMentions(params?: {
    page?: number
    limit?: number
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/mentions', params)
  }

  async getMention(id: string) {
    return this.get(`/api/v1/admin/mentions/${id}`)
  }

  async deleteMention(id: string, reason: string) {
    return this.delete(`/api/v1/admin/mentions/${id}`, { reason })
  }

  async bulkMentionAction(data: { 
    mention_ids: string[]
    action: string
    reason?: string 
  }) {
    return this.post('/api/v1/admin/mentions/bulk/actions', data)
  }

  // ==================== MEDIA MANAGEMENT ====================
  async getMedia(params?: {
    page?: number
    limit?: number
    type?: string
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/media', params)
  }

  async getMediaItem(id: string) {
    return this.get(`/api/v1/admin/media/${id}`)
  }

  async getMediaStats() {
    return this.get('/api/v1/admin/media/stats')
  }

  async moderateMedia(id: string, data: { 
    action: string
    reason?: string 
  }) {
    return this.put(`/api/v1/admin/media/${id}/moderate`, data)
  }

  async deleteMedia(id: string, reason: string) {
    return this.delete(`/api/v1/admin/media/${id}`, { reason })
  }

  async bulkMediaAction(data: { 
    media_ids: string[]
    action: string
    reason?: string 
  }) {
    return this.post('/api/v1/admin/media/bulk/actions', data)
  }

  async getStorageStats() {
    return this.get('/api/v1/admin/media/storage/stats')
  }

  async cleanupStorage(data: { 
    cleanup_type?: string
    older_than_days: number
    media_type?: string 
  }) {
    return this.post('/api/v1/admin/media/storage/cleanup', data)
  }

  // ==================== NOTIFICATION MANAGEMENT ====================
  async getNotifications(params?: {
    page?: number
    limit?: number
  }) {
    return this.get<PaginatedResponse>('/api/v1/admin/notifications', params)
  }

  async getNotification(id: string) {
    return this.get(`/api/v1/admin/notifications/${id}`)
  }

  async sendNotificationToUsers(data: {
    user_ids: string[]
    title: string
    message: string
    type: string
    data?: any
  }) {
    return this.post('/api/v1/admin/notifications/send', data)
  }

  async broadcastNotification(data: {
    title: string
    message: string
    type: string
    target_audience?: string
    data?: any
  }) {
    return this.post('/api/v1/admin/notifications/broadcast', data)
  }

  async getNotificationStats() {
    return this.get('/api/v1/admin/notifications/stats')
  }

  async deleteNotification(id: string, reason?: string) {
    return this.delete(`/api/v1/admin/notifications/${id}`, { reason })
  }

  async bulkNotificationAction(data: { 
    notification_ids: string[]
    action: string
  }) {
    return this.post('/api/v1/admin/notifications/bulk/actions', data)
  }

  // ==================== ANALYTICS ====================
  async getUserAnalytics(params?: { 
    period?: string
    date_from?: string
    date_to?: string 
  }) {
    return this.get('/api/v1/admin/analytics/users', params)
  }

  async getContentAnalytics(params?: { period?: string }) {
    return this.get('/api/v1/admin/analytics/content', params)
  }

  async getEngagementAnalytics(params?: { period?: string }) {
    return this.get('/api/v1/admin/analytics/engagement', params)
  }

  async getGrowthAnalytics(params?: { period?: string }) {
    return this.get('/api/v1/admin/analytics/growth', params)
  }

  async getDemographicAnalytics() {
    return this.get('/api/v1/admin/analytics/demographics')
  }

  async getRevenueAnalytics(params?: { period?: string }) {
    return this.get('/api/v1/admin/analytics/revenue', params)
  }

  async getCustomReport(params?: { 
    report_type?: string
    date_range?: string 
  }) {
    return this.get('/api/v1/admin/analytics/reports/custom', params)
  }

  async getRealtimeAnalytics() {
    return this.get('/api/v1/admin/analytics/realtime')
  }

  async getLiveStats() {
    return this.get('/api/v1/admin/analytics/live-stats')
  }

  // ==================== SYSTEM MANAGEMENT ====================
  async getSystemHealth() {
    return this.get('/api/v1/admin/system/health')
  }

  async getSystemInfo() {
    return this.get('/api/v1/admin/system/info')
  }

  async getSystemLogs(params?: {
    level?: string
    limit?: number
    date_from?: string
  }) {
    return this.get('/api/v1/admin/system/logs', params)
  }

  async getPerformanceMetrics() {
    return this.get('/api/v1/admin/system/performance')
  }

  async getDatabaseStats() {
    return this.get('/api/v1/admin/system/database/stats')
  }

  async getCacheStats() {
    return this.get('/api/v1/admin/system/cache/stats')
  }

  // Super Admin only operations
  async clearCache(cacheType?: string) {
    return this.post('/api/v1/admin/system/cache/clear', { cache_type: cacheType })
  }

  async warmCache(cacheTypes?: string[]) {
    return this.post('/api/v1/admin/system/cache/warm', { cache_types: cacheTypes })
  }

  async enableMaintenanceMode(data: { 
    message?: string
    estimated_duration?: string 
  }) {
    return this.post('/api/v1/admin/system/maintenance/enable', data)
  }

  async disableMaintenanceMode() {
    return this.post('/api/v1/admin/system/maintenance/disable')
  }

  async backupDatabase(data: { 
    backup_type?: string
    description?: string
  }) {
    return this.post('/api/v1/admin/system/database/backup', data)
  }

  async getDatabaseBackups() {
    return this.get('/api/v1/admin/system/database/backups')
  }

  async restoreDatabase(data: { 
    backup_id: string
    confirm: boolean
  }) {
    return this.post('/api/v1/admin/system/database/restore', data)
  }

  async optimizeDatabase() {
    return this.post('/api/v1/admin/system/database/optimize')
  }

  // ==================== CONFIGURATION MANAGEMENT (Super Admin) ====================
  async getConfiguration() {
    return this.get('/api/v1/admin/config')
  }

  async updateConfiguration(data: {
    max_post_length?: number
    max_file_size?: number
    registration_enabled?: boolean
    maintenance_mode?: boolean
  }) {
    return this.put('/api/v1/admin/config', data)
  }

  async getConfigurationHistory(params?: { limit?: number }) {
    return this.get('/api/v1/admin/config/history', params)
  }

  async rollbackConfiguration(configVersion: string) {
    return this.post('/api/v1/admin/config/rollback', { config_version: configVersion })
  }

  async validateConfiguration() {
    return this.get('/api/v1/admin/config/validate')
  }

  async getFeatureFlags() {
    return this.get('/api/v1/admin/config/features')
  }

  async updateFeatureFlags(data: {
    stories_enabled?: boolean
    groups_enabled?: boolean
    live_streaming?: boolean
  }) {
    return this.put('/api/v1/admin/config/features', data)
  }

  async toggleFeature(feature: string, enabled?: boolean) {
    if (enabled !== undefined) {
      return this.put(`/api/v1/admin/config/features/${feature}/toggle`, { enabled })
    }
    return this.put(`/api/v1/admin/config/features/${feature}/toggle`)
  }

  async getRateLimits() {
    return this.get('/api/v1/admin/config/rate-limits')
  }

  async updateRateLimits(data: {
    posts?: { requests: number; window: string }
    comments?: { requests: number; window: string }
    likes?: { requests: number; window: string }
  }) {
    return this.put('/api/v1/admin/config/rate-limits', data)
  }

  // ==================== WEBSOCKET CONNECTIONS ====================
  connectWebSocket(endpoint: string, onMessage?: (data: any) => void): WebSocket | null {
    try {
      const token = this.getStoredToken()
      if (!token) {
        console.error('No token available for WebSocket connection')
        return null
      }

      const wsUrl = `${process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'}/api/v1/admin/ws/${endpoint}`
      
      const ws = new WebSocket(wsUrl)
      
      ws.onopen = () => {
        console.log(`WebSocket connected: ${endpoint}`)
        // Send auth token after connection
        ws.send(JSON.stringify({ type: 'auth', token }))
      }
      
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          onMessage?.(data)
        } catch (error) {
          console.error('WebSocket message parse error:', error)
        }
      }
      
      ws.onerror = (error) => {
        console.error(`WebSocket error on ${endpoint}:`, error)
      }
      
      ws.onclose = () => {
        console.log(`WebSocket disconnected: ${endpoint}`)
        this.wsConnections.delete(endpoint)
      }
      
      this.wsConnections.set(endpoint, ws)
      return ws
    } catch (error) {
      console.error('WebSocket connection error:', error)
      return null
    }
  }

  // Connect to dashboard WebSocket
  connectDashboardWebSocket(onMessage?: (data: any) => void) {
    return this.connectWebSocket('dashboard', onMessage)
  }

  // Connect to monitoring WebSocket
  connectMonitoringWebSocket(onMessage?: (data: any) => void) {
    return this.connectWebSocket('monitoring', onMessage)
  }

  // Connect to moderation WebSocket
  connectModerationWebSocket(onMessage?: (data: any) => void) {
    return this.connectWebSocket('moderation', onMessage)
  }

  // Connect to activities WebSocket
  connectActivitiesWebSocket(onMessage?: (data: any) => void) {
    return this.connectWebSocket('activities', onMessage)
  }

  disconnectWebSocket(endpoint: string) {
    const ws = this.wsConnections.get(endpoint)
    if (ws) {
      ws.close()
      this.wsConnections.delete(endpoint)
    }
  }

  disconnectAllWebSockets() {
    this.wsConnections.forEach((ws, endpoint) => {
      ws.close()
    })
    this.wsConnections.clear()
  }

  // ==================== FILE UPLOAD ====================
  async uploadFile(file: File, type: 'avatar' | 'media' | 'document' = 'media'): Promise<ApiResponse> {
    const token = this.getStoredToken()
    const formData = new FormData()
    formData.append('file', file)
    formData.append('type', type)

    const response = await fetch(`${this.baseURL}/api/v1/admin/upload`, {
      method: 'POST',
      headers: {
        ...(token && { Authorization: `Bearer ${token}` }),
      },
      body: formData,
    })

    if (!response.ok) {
      throw new Error(`Upload failed: ${response.status}`)
    }

    return response.json()
  }

  // ==================== EXPORT UTILITIES ====================
  async exportData(type: string, params?: any): Promise<Blob> {
    const token = this.getStoredToken()
    let url = `${this.baseURL}/api/v1/admin/export/${type}`
    
    if (params) {
      const searchParams = new URLSearchParams()
      Object.keys(params).forEach(key => {
        if (params[key] !== undefined && params[key] !== null) {
          searchParams.append(key, params[key].toString())
        }
      })
      if (searchParams.toString()) {
        url += '?' + searchParams.toString()
      }
    }

    const response = await fetch(url, {
      method: 'GET',
      headers: {
        ...(token && { Authorization: `Bearer ${token}` }),
      },
    })

    if (!response.ok) {
      throw new Error(`Export failed: ${response.status}`)
    }

    return response.blob()
  }

  downloadFile(blob: Blob, filename: string) {
    if (typeof window === 'undefined') return

    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  }

  // ==================== HEALTH CHECK METHODS ====================
  async healthCheck() {
    return this.get('/health')
  }

  async apiInfo() {
    return this.get('/api/v1')
  }

  // ==================== UTILITY METHODS ====================
  
  // Check if user is online
  isOnline(): boolean {
    return typeof window !== 'undefined' ? navigator.onLine : true
  }

  // Get current token
  getCurrentToken(): string | null {
    return this.getStoredToken()
  }

  // Get current user
  getCurrentUser(): any | null {
    if (typeof window === 'undefined') return null
    try {
      const user = localStorage.getItem('admin_user')
      return user ? JSON.parse(user) : null
    } catch {
      return null
    }
  }

  // Check if user has permission
  hasPermission(permission: string): boolean {
    const user = this.getCurrentUser()
    if (!user) return false
    
    // Super admin has all permissions
    if (user.role === 'super_admin') return true
    
    // Check specific permission
    return user.permissions?.includes(permission) || 
           user.permissions?.includes('admin.*') ||
           user.permissions?.includes('*') || 
           false
  }

  // Get API base URL
  getBaseUrl(): string {
    return this.baseURL
  }

  // Debug method
  debugStatus() {
    console.log('🔍 API Client Debug:')
    console.log('- Base URL:', this.baseURL)
    console.log('- Token:', this.getStoredToken() ? 'Present' : 'Missing')
    console.log('- User:', this.getCurrentUser() ? 'Present' : 'Missing')
    console.log('- Is refreshing:', this.isRefreshing)
  }
}

export const apiClient = new FetchApiClient()
export default apiClient