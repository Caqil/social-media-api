// lib/api-client.ts - Complete Implementation
import axios, { AxiosInstance, AxiosResponse } from 'axios'

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
    links: {
      self: string
      next?: string
      previous?: string
      first?: string
      last?: string
    }
  }

class ApiClient {
  private client: AxiosInstance
  private wsConnections: Map<string, WebSocket> = new Map()

  constructor() {
    this.client = axios.create({
      baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    this.setupInterceptors()
  }

  private setupInterceptors() {
    // Request interceptor for auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('admin_token')
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => Promise.reject(error)
    )

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response: AxiosResponse) => response,
      async (error) => {
        if (error.response?.status === 401) {
          // Try to refresh token first
          try {
            await this.refreshToken()
            // Retry the original request
            const originalRequest = error.config
            const token = localStorage.getItem('admin_token')
            if (token) {
              originalRequest.headers.Authorization = `Bearer ${token}`
              return this.client(originalRequest)
            }
          } catch (refreshError) {
            // Refresh failed, redirect to login
            localStorage.removeItem('admin_token')
            localStorage.removeItem('admin_refresh_token')
            localStorage.removeItem('admin_user')
            window.location.href = '/admin/login'
          }
        }
        return Promise.reject(error)
      }
    )
  }

  // Generic methods
  async get<T>(url: string, params?: any): Promise<ApiResponse<T>> {
    const response = await this.client.get(url, { params })
    return response.data
  }

  async post<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    const response = await this.client.post(url, data)
    return response.data
  }

  async put<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    const response = await this.client.put(url, data)
    return response.data
  }

  async patch<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    const response = await this.client.patch(url, data)
    return response.data
  }

  async delete<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    const response = await this.client.delete(url, { data })
    return response.data
  }

  // ==================== AUTHENTICATION ====================
  async adminLogin(email: string, password: string) {
    const response = await this.client.post('/admin/public/auth/login', {
      email,
      password,
    })
    return response.data
  }

  async adminLogout() {
    const response = await this.client.post('/admin/public/auth/logout')
    return response.data
  }

  async refreshToken() {
    const refreshToken = localStorage.getItem('admin_refresh_token')
    const response = await this.client.post('/admin/public/auth/refresh', {
      refresh_token: refreshToken,
    })
    if (response.data.success) {
      localStorage.setItem('admin_token', response.data.data.access_token)
    }
    return response.data
  }

  async adminForgotPassword(email: string) {
    return this.post('/admin/public/auth/forgot-password', { email })
  }

  async adminResetPassword(token: string, newPassword: string) {
    return this.post('/admin/public/auth/reset-password', { 
      token, 
      new_password: newPassword 
    })
  }

  // ==================== DASHBOARD ====================
  async getDashboardStats() {
    return this.get('/admin/dashboard/stats')
  }

  // ==================== USER MANAGEMENT ====================
  async getUsers(params?: any) {
    return this.get<PaginatedResponse>('/admin/users', params)
  }

  async searchUsers(query: string, params?: any) {
    return this.get<PaginatedResponse>('/admin/users/search', { q: query, ...params })
  }

  async getUser(id: string) {
    return this.get(`/admin/users/${id}`)
  }

  async getUserStats(id: string) {
    return this.get(`/admin/users/${id}/stats`)
  }

  async updateUserStatus(id: string, data: { is_active: boolean; is_suspended: boolean; reason?: string }) {
    return this.put(`/admin/users/${id}/status`, data)
  }

  async verifyUser(id: string) {
    return this.put(`/admin/users/${id}/verify`)
  }

  async deleteUser(id: string, reason: string) {
    return this.delete(`/admin/users/${id}`, { reason })
  }

  async bulkUserAction(data: { user_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/users/bulk/actions', data)
  }

  async exportUsers(params?: any) {
    return this.get('/admin/users/export', params)
  }

  // ==================== POST MANAGEMENT ====================
  async getPosts(params?: any) {
    return this.get<PaginatedResponse>('/admin/posts', params)
  }

  async searchPosts(query: string, params?: any) {
    return this.get<PaginatedResponse>('/admin/posts/search', { q: query, ...params })
  }

  async getPostStats(id: string) {
    return this.get(`/admin/posts/${id}/stats`)
  }

  async hidePost(id: string, reason: string) {
    return this.put(`/admin/posts/${id}/hide`, { reason })
  }

  async deletePost(id: string, reason: string) {
    return this.delete(`/admin/posts/${id}`, { reason })
  }

  async bulkPostAction(data: { post_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/posts/bulk/actions', data)
  }

  async exportPosts(params?: any) {
    return this.get('/admin/posts/export', params)
  }

  // ==================== COMMENT MANAGEMENT ====================
  async getComments(params?: any) {
    return this.get<PaginatedResponse>('/admin/comments', params)
  }

  async getComment(id: string) {
    return this.get(`/admin/comments/${id}`)
  }

  async hideComment(id: string, reason: string) {
    return this.put(`/admin/comments/${id}/hide`, { reason })
  }

  async deleteComment(id: string, reason: string) {
    return this.delete(`/admin/comments/${id}`, { reason })
  }

  async bulkCommentAction(data: { comment_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/comments/bulk/actions', data)
  }

  // ==================== GROUP MANAGEMENT ====================
  async getGroups(params?: any) {
    return this.get<PaginatedResponse>('/admin/groups', params)
  }

  async getGroup(id: string) {
    return this.get(`/admin/groups/${id}`)
  }

  async getGroupMembers(id: string, params?: any) {
    return this.get<PaginatedResponse>(`/admin/groups/${id}/members`, params)
  }

  async updateGroupStatus(id: string, data: { is_active: boolean; reason?: string }) {
    return this.put(`/admin/groups/${id}/status`, data)
  }

  async deleteGroup(id: string, reason: string) {
    return this.delete(`/admin/groups/${id}`, { reason })
  }

  async bulkGroupAction(data: { group_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/groups/bulk/actions', data)
  }

  // ==================== EVENT MANAGEMENT ====================
  async getEvents(params?: any) {
    return this.get<PaginatedResponse>('/admin/events', params)
  }

  async getEvent(id: string) {
    return this.get(`/admin/events/${id}`)
  }

  async getEventAttendees(id: string, params?: any) {
    return this.get<PaginatedResponse>(`/admin/events/${id}/attendees`, params)
  }

  async updateEventStatus(id: string, data: { status: string; reason?: string }) {
    return this.put(`/admin/events/${id}/status`, data)
  }

  async deleteEvent(id: string, reason: string) {
    return this.delete(`/admin/events/${id}`, { reason })
  }

  async bulkEventAction(data: { event_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/events/bulk/actions', data)
  }

  // ==================== STORY MANAGEMENT ====================
  async getStories(params?: any) {
    return this.get<PaginatedResponse>('/admin/stories', params)
  }

  async getStory(id: string) {
    return this.get(`/admin/stories/${id}`)
  }

  async hideStory(id: string, reason: string) {
    return this.put(`/admin/stories/${id}/hide`, { reason })
  }

  async deleteStory(id: string, reason: string) {
    return this.delete(`/admin/stories/${id}`, { reason })
  }

  async bulkStoryAction(data: { story_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/stories/bulk/actions', data)
  }

  // ==================== MESSAGE MANAGEMENT ====================
  async getMessages(params?: any) {
    return this.get<PaginatedResponse>('/admin/messages', params)
  }

  async getMessage(id: string) {
    return this.get(`/admin/messages/${id}`)
  }

  async getConversations(params?: any) {
    return this.get<PaginatedResponse>('/admin/messages/conversations', params)
  }

  async getConversation(id: string) {
    return this.get(`/admin/messages/conversations/${id}`)
  }

  async deleteMessage(id: string, reason: string) {
    return this.delete(`/admin/messages/${id}`, { reason })
  }

  async bulkMessageAction(data: { message_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/messages/bulk/actions', data)
  }

  // ==================== REPORT MANAGEMENT ====================
  async getReports(params?: any) {
    return this.get<PaginatedResponse>('/admin/reports', params)
  }

  async getReport(id: string) {
    return this.get(`/admin/reports/${id}`)
  }

  async updateReportStatus(id: string, data: {
    status: string
    resolution?: string
    note?: string
  }) {
    return this.put(`/admin/reports/${id}/status`, data)
  }

  async assignReport(id: string, assigneeId: string) {
    return this.put(`/admin/reports/${id}/assign`, { assignee_id: assigneeId })
  }

  async resolveReport(id: string, data: { resolution: string; note?: string }) {
    return this.post(`/admin/reports/${id}/resolve`, data)
  }

  async rejectReport(id: string, data: { reason: string }) {
    return this.post(`/admin/reports/${id}/reject`, data)
  }

  async bulkReportAction(data: { report_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/reports/bulk/actions', data)
  }

  async getReportStats() {
    return this.get('/admin/reports/stats')
  }

  async getReportSummary() {
    return this.get('/admin/reports/stats/summary')
  }

  // ==================== FOLLOW MANAGEMENT ====================
  async getFollows(params?: any) {
    return this.get<PaginatedResponse>('/admin/follows', params)
  }

  async getFollow(id: string) {
    return this.get(`/admin/follows/${id}`)
  }

  async deleteFollow(id: string, reason: string) {
    return this.delete(`/admin/follows/${id}`, { reason })
  }

  async getRelationships(userId: string) {
    return this.get('/admin/follows/relationships', { user_id: userId })
  }

  async bulkFollowAction(data: { follow_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/follows/bulk/actions', data)
  }

  // ==================== LIKE MANAGEMENT ====================
  async getLikes(params?: any) {
    return this.get<PaginatedResponse>('/admin/likes', params)
  }

  async getLikeStats() {
    return this.get('/admin/likes/stats')
  }

  async deleteLike(id: string, reason: string) {
    return this.delete(`/admin/likes/${id}`, { reason })
  }

  async bulkLikeAction(data: { like_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/likes/bulk/actions', data)
  }

  // ==================== HASHTAG MANAGEMENT ====================
  async getHashtags(params?: any) {
    return this.get<PaginatedResponse>('/admin/hashtags', params)
  }

  async getHashtag(id: string) {
    return this.get(`/admin/hashtags/${id}`)
  }

  async getTrendingHashtags(limit: number = 10) {
    return this.get('/admin/hashtags/trending', { limit })
  }

  async blockHashtag(id: string, reason: string) {
    return this.put(`/admin/hashtags/${id}/block`, { reason })
  }

  async unblockHashtag(id: string) {
    return this.put(`/admin/hashtags/${id}/unblock`)
  }

  async deleteHashtag(id: string, reason: string) {
    return this.delete(`/admin/hashtags/${id}`, { reason })
  }

  async bulkHashtagAction(data: { hashtag_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/hashtags/bulk/actions', data)
  }

  // ==================== MENTION MANAGEMENT ====================
  async getMentions(params?: any) {
    return this.get<PaginatedResponse>('/admin/mentions', params)
  }

  async getMention(id: string) {
    return this.get(`/admin/mentions/${id}`)
  }

  async deleteMention(id: string, reason: string) {
    return this.delete(`/admin/mentions/${id}`, { reason })
  }

  async bulkMentionAction(data: { mention_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/mentions/bulk/actions', data)
  }

  // ==================== MEDIA MANAGEMENT ====================
  async getMedia(params?: any) {
    return this.get<PaginatedResponse>('/admin/media', params)
  }

  async getMediaItem(id: string) {
    return this.get(`/admin/media/${id}`)
  }

  async getMediaStats() {
    return this.get('/admin/media/stats')
  }

  async moderateMedia(id: string, data: { action: string; reason?: string }) {
    return this.put(`/admin/media/${id}/moderate`, data)
  }

  async deleteMedia(id: string, reason: string) {
    return this.delete(`/admin/media/${id}`, { reason })
  }

  async bulkMediaAction(data: { media_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/media/bulk/actions', data)
  }

  async getStorageStats() {
    return this.get('/admin/media/storage/stats')
  }

  async cleanupStorage(data: { older_than_days: number; media_type?: string }) {
    return this.post('/admin/media/storage/cleanup', data)
  }

  // ==================== NOTIFICATION MANAGEMENT ====================
  async getNotifications(params?: any) {
    return this.get<PaginatedResponse>('/admin/notifications', params)
  }

  async getNotification(id: string) {
    return this.get(`/admin/notifications/${id}`)
  }

  async sendNotificationToUsers(data: {
    user_ids: string[]
    title: string
    message: string
    type: string
    data?: any
  }) {
    return this.post('/admin/notifications/send', data)
  }

  async broadcastNotification(data: {
    title: string
    message: string
    type: string
    data?: any
    filter?: any
  }) {
    return this.post('/admin/notifications/broadcast', data)
  }

  async getNotificationStats() {
    return this.get('/admin/notifications/stats')
  }

  async deleteNotification(id: string, reason: string) {
    return this.delete(`/admin/notifications/${id}`, { reason })
  }

  async bulkNotificationAction(data: { notification_ids: string[]; action: string; reason?: string }) {
    return this.post('/admin/notifications/bulk/actions', data)
  }

  // ==================== ANALYTICS ====================
  async getUserAnalytics(period: string = '30d') {
    return this.get('/admin/analytics/users', { period })
  }

  async getContentAnalytics(period: string = '30d') {
    return this.get('/admin/analytics/content', { period })
  }

  async getEngagementAnalytics(period: string = '30d') {
    return this.get('/admin/analytics/engagement', { period })
  }

  async getGrowthAnalytics(period: string = '30d') {
    return this.get('/admin/analytics/growth', { period })
  }

  async getDemographicAnalytics() {
    return this.get('/admin/analytics/demographics')
  }

  async getRevenueAnalytics(period: string = '30d') {
    return this.get('/admin/analytics/revenue', { period })
  }

  async getCustomReport(data: {
    report_type: string
    start_date: string
    end_date: string
    filters?: any
  }) {
    return this.post('/admin/analytics/reports/custom', data)
  }

  async getRealtimeAnalytics() {
    return this.get('/admin/analytics/realtime')
  }

  async getLiveStats() {
    return this.get('/admin/analytics/live-stats')
  }

  // ==================== SYSTEM MANAGEMENT ====================
  async getSystemHealth() {
    return this.get('/admin/system/health')
  }

  async getSystemInfo() {
    return this.get('/admin/system/info')
  }

  async getSystemLogs(params?: any) {
    return this.get('/admin/system/logs', params)
  }

  async getPerformanceMetrics() {
    return this.get('/admin/system/performance')
  }

  async getDatabaseStats() {
    return this.get('/admin/system/database/stats')
  }

  async getCacheStats() {
    return this.get('/admin/system/cache/stats')
  }

  // Super Admin only operations
  async clearCache(cacheType?: string) {
    return this.post('/admin/system/cache/clear', { cache_type: cacheType })
  }

  async warmCache(cacheType?: string) {
    return this.post('/admin/system/cache/warm', { cache_type: cacheType })
  }

  async enableMaintenanceMode(data: { message?: string; duration?: string }) {
    return this.post('/admin/system/maintenance/enable', data)
  }

  async disableMaintenanceMode() {
    return this.post('/admin/system/maintenance/disable')
  }

  async backupDatabase(data: { backup_type?: string; collections?: string[] }) {
    return this.post('/admin/system/database/backup', data)
  }

  async getDatabaseBackups() {
    return this.get('/admin/system/database/backups')
  }

  async restoreDatabase(data: { backup_id: string; restore_type?: string; collections?: string[] }) {
    return this.post('/admin/system/database/restore', data)
  }

  async optimizeDatabase() {
    return this.post('/admin/system/database/optimize')
  }

  // ==================== CONFIGURATION MANAGEMENT ====================
  async getConfiguration() {
    return this.get('/admin/config')
  }

  async updateConfiguration(data: any) {
    return this.put('/admin/config', data)
  }

  async getConfigurationHistory(params?: any) {
    return this.get('/admin/config/history', params)
  }

  async rollbackConfiguration(configId: string) {
    return this.post('/admin/config/rollback', { config_id: configId })
  }

  async validateConfiguration(data: any) {
    return this.post('/admin/config/validate', data)
  }

  async getFeatureFlags() {
    return this.get('/admin/config/features')
  }

  async updateFeatureFlags(data: any) {
    return this.put('/admin/config/features', data)
  }

  async toggleFeature(feature: string, enabled: boolean) {
    return this.put(`/admin/config/features/${feature}/toggle`, { enabled })
  }

  async getRateLimits() {
    return this.get('/admin/config/rate-limits')
  }

  async updateRateLimits(data: any) {
    return this.put('/admin/config/rate-limits', data)
  }

  // ==================== PUBLIC ADMIN ROUTES ====================
  async getPublicSystemStatus() {
    return this.get('/admin/public/status')
  }

  async getPublicHealthCheck() {
    return this.get('/admin/public/health')
  }

  // ==================== WEBSOCKET CONNECTIONS ====================
  connectWebSocket(endpoint: string, onMessage?: (data: any) => void): WebSocket | null {
    try {
      const token = localStorage.getItem('admin_token')
      const wsUrl = `${process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'}/api/admin/ws/${endpoint}`
      
      const ws = new WebSocket(wsUrl, ['Authorization', `Bearer ${token}`])
      
      ws.onopen = () => {
        console.log(`WebSocket connected: ${endpoint}`)
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
    const formData = new FormData()
    formData.append('file', file)
    formData.append('type', type)

    const response = await this.client.post('/admin/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  }

  // ==================== EXPORT UTILITIES ====================
  async exportData(type: string, params?: any): Promise<Blob> {
    const response = await this.client.get(`/admin/export/${type}`, {
      params,
      responseType: 'blob',
    })
    return response.data
  }

  downloadFile(blob: Blob, filename: string) {
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  }
}

export const apiClient = new ApiClient()
export default apiClient