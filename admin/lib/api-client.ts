// lib/api-client.ts - Fixed Version with Better Error Handling
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

class FetchApiClient {
  private baseURL: string
  private isRefreshing = false
  private refreshPromise: Promise<string> | null = null

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    console.log('🔧 API Client initialized:', this.baseURL)
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
      console.log('✅ Token stored successfully')
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
    const fullUrl = `${this.baseURL}${url}`

    const token = this.getStoredToken()
    const headers = new Headers({
      'Content-Type': 'application/json',
      ...options.headers,
    })

    // Add auth header if token exists and not auth endpoint
    if (token && !url.includes('/auth/')) {
      headers.set('Authorization', `Bearer ${token}`)
      console.log(`🔑 Request: ${options.method || 'GET'} ${url} (with token)`)
    } else {
      console.log(`🔓 Request: ${options.method || 'GET'} ${url} (no token)`)
    }

    try {
      const response = await fetch(fullUrl, {
        ...options,
        headers,
      })

      console.log(`📡 Response: ${options.method || 'GET'} ${url} - ${response.status}`)

      // Handle 401 with retry
      if (response.status === 401 && retryOnAuth && !url.includes('/auth/')) {
        console.log('🔄 Got 401, attempting token refresh...')

        try {
          await this.refreshTokenIfNeeded()
          return this.makeRequest<T>(url, options, false)
        } catch (refreshError) {
          console.error('❌ Token refresh failed:', refreshError)
          this.clearAuthData()

          if (typeof window !== 'undefined') {
            window.location.href = '/admin/login'
          }
          throw new Error('Authentication failed')
        }
      }

      // Parse response
      let data: any
      const contentType = response.headers.get('content-type')
      
      if (contentType && contentType.includes('application/json')) {
        data = await response.json()
      } else {
        const text = await response.text()
        console.warn('Non-JSON response:', text)
        data = { success: false, message: 'Invalid response format', data: null }
      }

      // Handle different response structures
      if (!response.ok) {
        console.error(`❌ Request failed: ${response.status}`)
        throw new Error(data.message || `Request failed: ${response.status}`)
      }

      // Ensure consistent response structure
      if (typeof data !== 'object' || data === null) {
        console.warn('Invalid response data structure')
        return {
          success: true,
          message: 'Success',
          data: data as T
        }
      }

      // If response doesn't have expected structure, wrap it
      if (!('success' in data)) {
        return {
          success: true,
          message: 'Success',
          data: data as T
        }
      }

      return data as ApiResponse<T>
    } catch (error) {
      console.error(`❌ Request error:`, error)
      if (error instanceof Error) {
        throw error
      }
      throw new Error('Network error occurred')
    }
  }

  // Generic HTTP methods with better error handling
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

  async delete<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.makeRequest<T>(url, {
      method: 'DELETE',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  // ==================== AUTHENTICATION ====================
  async adminLogin(email: string, password: string): Promise<ApiResponse<LoginResponse>> {
    console.log('🔄 Starting admin login...')
    
    this.clearAuthData()
    
    try {
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

      let data: any
      try {
        data = await response.json()
      } catch (parseError) {
        console.error('Failed to parse login response:', parseError)
        throw new Error('Invalid server response')
      }

      if (!response.ok) {
        console.error(`Login failed: ${response.status}`, data)
        throw new Error(data.message || `Login failed: ${response.status}`)
      }

      console.log('✅ Login API call successful')
      
      // Handle different response structures from Go API
      let loginData: LoginResponse
      
      if (data.data) {
        // Standard API response format
        loginData = data.data
      } else if (data.access_token) {
        // Direct token response
        loginData = data
      } else {
        throw new Error('Invalid login response format')
      }
      
      if (typeof window !== 'undefined') {
        if (loginData.access_token) {
          this.setStoredToken(loginData.access_token)
          console.log('✅ Access token stored')
        }
        if (loginData.refresh_token) {
          localStorage.setItem('admin_refresh_token', loginData.refresh_token)
          console.log('✅ Refresh token stored')
        }
        if (loginData.user) {
          localStorage.setItem('admin_user', JSON.stringify(loginData.user))
          console.log('✅ User data stored')
        }
      }
      
      return {
        success: true,
        message: 'Login successful',
        data: loginData
      }
    } catch (error) {
      console.error('❌ Login error:', error)
      this.clearAuthData()
      throw error
    }
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

  // ==================== DASHBOARD ====================
  async getDashboardStats(): Promise<ApiResponse<DashboardStats>> {
    console.log('🔄 Fetching dashboard stats...')
    try {
      const response = await this.get<DashboardStats>('/api/v1/admin/dashboard/stats')
      
      // Ensure we have valid data
      if (!response || !response.data) {
        console.warn('Dashboard stats response is null or invalid')
        return {
          success: false,
          message: 'No dashboard data available',
          data: {} as DashboardStats
        }
      }
      
      return response
    } catch (error) {
      console.error('❌ Dashboard stats error:', error)
      throw error
    }
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
  }): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/users', params)
      
      // Handle both direct array and paginated response
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Users fetched successfully',
          data: response.data,
          pagination: {
            current_page: 1,
            per_page: response.data.length,
            total: response.data.length,
            total_pages: 1,
            has_next: false,
            has_previous: false
          }
        }
      }
      
      return response as PaginatedResponse
    } catch (error) {
      console.error('❌ Get users error:', error)
      throw error
    }
  }

  async getUser(id: string) {
    return this.get(`/api/v1/admin/users/${id}`)
  }

  async updateUserStatus(id: string, data: { 
    is_active?: boolean
    is_suspended?: boolean
    reason?: string 
  }) {
    return this.put(`/api/v1/admin/users/${id}/status`, data)
  }

  async bulkUserAction(data: { 
    user_ids: string[]
    action: string
    reason?: string
    duration?: string
  }) {
    return this.post('/api/v1/admin/users/bulk/actions', data)
  }

  // ==================== POST MANAGEMENT ====================
  async getPosts(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/posts', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Posts fetched successfully',
          data: response.data,
          pagination: {
            current_page: 1,
            per_page: response.data.length,
            total: response.data.length,
            total_pages: 1,
            has_next: false,
            has_previous: false
          }
        }
      }
      
      return response as PaginatedResponse
    } catch (error) {
      console.error('❌ Get posts error:', error)
      throw error
    }
  }

  async getPost(id: string) {
    return this.get(`/api/v1/admin/posts/${id}`)
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

  // ==================== GROUP MANAGEMENT ====================
  async getGroups(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/groups', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Groups fetched successfully',
          data: response.data,
          pagination: {
            current_page: 1,
            per_page: response.data.length,
            total: response.data.length,
            total_pages: 1,
            has_next: false,
            has_previous: false
          }
        }
      }
      
      return response as PaginatedResponse
    } catch (error) {
      console.error('❌ Get groups error:', error)
      throw error
    }
  }

  async getGroup(id: string) {
    return this.get(`/api/v1/admin/groups/${id}`)
  }

  async getGroupMembers(id: string, params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>(`/api/v1/admin/groups/${id}/members`, params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Group members fetched successfully',
          data: response.data,
          pagination: {
            current_page: 1,
            per_page: response.data.length,
            total: response.data.length,
            total_pages: 1,
            has_next: false,
            has_previous: false
          }
        }
      }
      
      return response as PaginatedResponse
    } catch (error) {
      console.error('❌ Get group members error:', error)
      throw error
    }
  }

  async updateGroupStatus(id: string, data: { 
    is_active?: boolean
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

  // ==================== HASHTAG MANAGEMENT ====================
  async getHashtags(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/hashtags', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Hashtags fetched successfully',
          data: response.data,
          pagination: {
            current_page: 1,
            per_page: response.data.length,
            total: response.data.length,
            total_pages: 1,
            has_next: false,
            has_previous: false
          }
        }
      }
      
      return response as PaginatedResponse
    } catch (error) {
      console.error('❌ Get hashtags error:', error)
      throw error
    }
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

  // ==================== STORY MANAGEMENT ====================
  async getStories(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/stories', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Stories fetched successfully',
          data: response.data,
          pagination: {
            current_page: 1,
            per_page: response.data.length,
            total: response.data.length,
            total_pages: 1,
            has_next: false,
            has_previous: false
          }
        }
      }
      
      return response as PaginatedResponse
    } catch (error) {
      console.error('❌ Get stories error:', error)
      throw error
    }
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

  // ==================== REPORT MANAGEMENT ====================
  async getReports(params?: {
    status?: string
    target_type?: string
    limit?: number
    skip?: number
    page?: number
  }): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/reports', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Reports fetched successfully',
          data: response.data,
          pagination: {
            current_page: 1,
            per_page: response.data.length,
            total: response.data.length,
            total_pages: 1,
            has_next: false,
            has_previous: false
          }
        }
      }
      
      return response as PaginatedResponse
    } catch (error) {
      console.error('❌ Get reports error:', error)
      throw error
    }
  }

  async getReport(id: string) {
    return this.get(`/api/v1/admin/reports/${id}`)
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

  // ==================== UTILITY METHODS ====================
  
  getCurrentToken(): string | null {
    return this.getStoredToken()
  }

  getCurrentUser(): any | null {
    if (typeof window === 'undefined') return null
    try {
      const user = localStorage.getItem('admin_user')
      return user ? JSON.parse(user) : null
    } catch {
      return null
    }
  }

  isOnline(): boolean {
    return typeof window !== 'undefined' ? navigator.onLine : true
  }

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