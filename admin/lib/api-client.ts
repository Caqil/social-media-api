// lib/api-client.ts - Complete API Client Based on Go Routes
import { DashboardStats, User } from '@/types/admin'

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

  async refreshToken(refreshToken: string): Promise<ApiResponse<LoginResponse>> {
    return this.post('/api/v1/admin/public/auth/refresh', { refresh_token: refreshToken })
  }

  async forgotPassword(email: string): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/public/auth/forgot-password', { email })
  }

  async resetPassword(token: string, password: string): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/public/auth/reset-password', { token, password })
  }

  // ==================== DASHBOARD ====================
  async getDashboard(): Promise<ApiResponse<DashboardStats>> {
    return this.get<DashboardStats>('/api/v1/admin/dashboard')
  }

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
    search?: string
    role?: string
    is_verified?: boolean
    is_active?: boolean
    is_suspended?: boolean
    sort_by?: string
    sort_order?: 'asc' | 'desc'
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
  async getUsersByIds(userIds: string[]): Promise<ApiResponse<User[]>> {
    if (!userIds || userIds.length === 0) {
      return {
        success: true,
        message: 'No user IDs provided',
        data: []
      };
    }
    
    try {
      // Convert array of IDs to comma-separated string
      const ids = userIds.join(',');
      console.log(`🔄 Fetching users by IDs: ${ids}`);
      
      // Try first with the ids parameter
      const response = await this.get<User[]>('/api/v1/admin/users', { ids });
      
      // If no users were returned, try with individual requests
      if (!response.data || !Array.isArray(response.data) || response.data.length === 0) {
        console.log('⚠️ No users found with batch request, trying individual requests');
        
        const users: User[] = [];
        for (const userId of userIds) {
          try {
            const userResponse = await this.get<User>(`/api/v1/admin/users/${userId}`);
            if (userResponse.data) {
              users.push(userResponse.data);
            }
          } catch (error) {
            console.error(`❌ Failed to fetch user ${userId}:`, error);
            // Continue with the next user
          }
        }
        
        if (users.length > 0) {
          return {
            success: true,
            message: 'Users fetched individually',
            data: users
          };
        }
      }
      
      return response;
    } catch (error) {
      console.error('❌ Failed to fetch users by IDs:', error);
      throw error;
    }
  }
  async searchUsers(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/users/search', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Users search completed successfully',
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
      console.error('❌ Search users error:', error)
      throw error
    }
  }

  async getUser(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/users/${id}`)
  }

  async getUserStats(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/users/${id}/stats`)
  }

  async createUser(userData: {
    username: string
    email: string
    password: string
    first_name?: string
    last_name?: string
    bio?: string
    role?: string
    is_active?: boolean
    is_verified?: boolean
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/users', userData)
  }

  async updateUser(id: string, userData: {
    username?: string
    email?: string
    first_name?: string
    last_name?: string
    bio?: string
    role?: string
    is_active?: boolean
    is_verified?: boolean
  }): Promise<ApiResponse<any>> {
    return this.put(`/api/v1/admin/users/${id}`, userData)
  }

  async updateUserStatus(id: string, data: { 
    is_active?: boolean
    is_suspended?: boolean
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.put(`/api/v1/admin/users/${id}/status`, data)
  }

  async verifyUser(id: string): Promise<ApiResponse<any>> {
    return this.put(`/api/v1/admin/users/${id}/verify`)
  }

  async deleteUser(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.delete(`/api/v1/admin/users/${id}`, { reason })
  }

  async bulkUserAction(data: { 
    user_ids?: string[]
    action: string
    reason?: string
    duration?: string
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/users/bulk/actions', data)
  }

  async exportUsers(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/users/export')
  }

  // ==================== POST MANAGEMENT ====================
  async getPosts(params?: {
    page?: number
    limit?: number
    user_id?: string
    type?: string
    visibility?: string
    is_reported?: boolean
    is_hidden?: boolean
    search?: string
    date_from?: string
    date_to?: string
  }): Promise<PaginatedResponse> {
    try {
      // Always include these parameters to ensure we get user data
      const enhancedParams = {
        ...params,
        include_user: true,     // Request user data inclusion
        expand: 'user,media',   // Explicitly request user expansion if API supports it
      };
      
      console.log('🔄 Fetching posts with enhanced params:', enhancedParams);
      const response = await this.get<any>('/api/v1/admin/posts', enhancedParams);
      
      // If posts are returned without user data, fetch the user data separately
      if (response.data && Array.isArray(response.data)) {
        const needsUserData = response.data.filter(post => 
          post.user_id && (!post.user || !post.user.username)
        );
        
        if (needsUserData.length > 0) {
          console.log(`⚠️ Found ${needsUserData.length} posts without complete user data, fetching users...`);
          
          // Get unique user IDs that need to be fetched
          const userIds = [...new Set(needsUserData.map(post => post.user_id))];
          
          // Fetch users data if there are posts missing user information
          if (userIds.length > 0) {
            try {
              // Batch user fetch by user IDs if API supports it
              const usersResponse = await this.get<any>('/api/v1/admin/users', {
                ids: userIds.join(',')
              });
              
              if (usersResponse.data && Array.isArray(usersResponse.data)) {
                const usersMap = new Map();
                usersResponse.data.forEach(user => {
                  usersMap.set(user.id, user);
                });
                
                // Update posts with fetched user data
                response.data = response.data.map(post => {
                  if (post.user_id && (!post.user || !post.user.username)) {
                    const userData = usersMap.get(post.user_id);
                    if (userData) {
                      post.user = userData;
                    }
                  }
                  return post;
                });
                
                console.log('✅ Updated posts with user data');
              }
            } catch (userFetchError) {
              console.error('❌ Failed to fetch users data:', userFetchError);
            }
          }
        }
      }
      
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
      
      return response as PaginatedResponse;
    } catch (error) {
      console.error('❌ Get posts error:', error);
      throw error;
    }
  }

  async searchPosts(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/posts/search', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Posts search completed successfully',
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
      console.error('❌ Search posts error:', error)
      throw error
    }
  }

  async getPost(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/posts/${id}`)
  }

  async getPostStats(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/posts/${id}/stats`)
  }

  async hidePost(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.put(`/api/v1/admin/posts/${id}/hide`, { reason })
  }

  async deletePost(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.delete(`/api/v1/admin/posts/${id}`, { reason })
  }

  async bulkPostAction(data: { 
    post_ids: string[]
    action: string
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/posts/bulk/actions', data)
  }

  async exportPosts(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/posts/export')
  }

  // ==================== COMMENT MANAGEMENT ====================
  async getComments(params?: {
    page?: number
    limit?: number
    post_id?: string
    user_id?: string
    is_hidden?: boolean
    is_reported?: boolean
    search?: string
    date_from?: string
    date_to?: string
  }): Promise<PaginatedResponse> {
    try {
      // Always include parameters to ensure we get user data
      const enhancedParams = {
        ...params,
        include_user: true,     // Request user data inclusion
        expand: 'user,post',    // Explicitly request user and post expansion
      };
      
      console.log('📡 Fetching comments with enhanced params:', enhancedParams);
      const response = await this.get<any>('/api/v1/admin/comments', enhancedParams);
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Comments fetched successfully',
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
      
      return response as PaginatedResponse;
    } catch (error) {
      console.error('❌ Get comments error:', error);
      throw error;
    }
  }
  
  async getComment(id: string): Promise<ApiResponse<any>> {
    // Validate ID
    if (!id || typeof id !== 'string' || id === 'undefined') {
      return Promise.reject(new Error('Invalid comment ID'));
    }
    return this.get(`/api/v1/admin/comments/${id}`);
  }
  
  async updateComment(id: string, data: { content: string }): Promise<ApiResponse<any>> {
    // Validate ID
    if (!id || typeof id !== 'string' || id === 'undefined') {
      return Promise.reject(new Error('Invalid comment ID'));
    }
    return this.put(`/api/v1/admin/comments/${id}`, data);
  }
  
  async hideComment(id: string, reason?: string): Promise<ApiResponse<any>> {
    // Validate ID
    if (!id || typeof id !== 'string' || id === 'undefined') {
      return Promise.reject(new Error('Invalid comment ID'));
    }
    return this.put(`/api/v1/admin/comments/${id}/hide`, { reason });
  }
  
  async showComment(id: string): Promise<ApiResponse<any>> {
    // Validate ID
    if (!id || typeof id !== 'string' || id === 'undefined') {
      return Promise.reject(new Error('Invalid comment ID'));
    }
    return this.put(`/api/v1/admin/comments/${id}/show`);
  }
  
  async deleteComment(id: string, reason?: string): Promise<ApiResponse<any>> {
    // Validate ID
    if (!id || typeof id !== 'string' || id === 'undefined') {
      return Promise.reject(new Error('Invalid comment ID'));
    }
    return this.delete(`/api/v1/admin/comments/${id}`, { reason });
  }
  
  async bulkCommentAction(data: { 
    comment_ids: string[];
    action: string;
    reason?: string; 
  }): Promise<ApiResponse<any>> {
    // Validate comment IDs
    if (!data.comment_ids || !Array.isArray(data.comment_ids) || data.comment_ids.length === 0) {
      return Promise.reject(new Error('Invalid comment IDs for bulk action'));
    }
    
    // Filter out any invalid IDs
    data.comment_ids = data.comment_ids.filter(id => 
      id && typeof id === 'string' && id !== 'undefined'
    );
    
    if (data.comment_ids.length === 0) {
      return Promise.reject(new Error('No valid comment IDs for bulk action'));
    }
    
    return this.post('/api/v1/admin/comments/bulk/actions', data);
  }

  // ==================== GROUP MANAGEMENT ====================
  async getGroups(params?: {
    page?: number
    limit?: number
    type?: string
    category?: string
    is_verified?: boolean
    is_active?: boolean
    search?: string
    created_by?: string
    min_members?: number
    date_from?: string
    date_to?: string
    sort_by?: string
    sort_order?: 'asc' | 'desc'
    include_creator?: boolean
    expand?: string
  }): Promise<PaginatedResponse> {
    try {
      console.log('🔍 getGroups called with original params:', params);
      
      // Start with the original params and only add what we need
      const apiParams: any = {
        ...params,
        include_creator: true,
        expand: 'creator,stats'
      };
  
      // Clean up any undefined values to avoid sending them
      Object.keys(apiParams).forEach(key => {
        if (apiParams[key] === undefined || apiParams[key] === null || apiParams[key] === '') {
          delete apiParams[key];
        }
      });
  
      // Convert "all" values to undefined (don't send them)
      if (apiParams.type === 'all') delete apiParams.type;
      if (apiParams.category === 'all') delete apiParams.category;
      if (apiParams.is_verified === 'all') delete apiParams.is_verified;
      if (apiParams.is_active === 'all') delete apiParams.is_active;
  
      // Log what we're actually sending to the API
      console.log('📡 Sending to API:', apiParams);
      console.log('📡 API URL will be:', `${this.baseURL}/api/v1/admin/groups?${new URLSearchParams(apiParams).toString()}`);
  
      const response = await this.get<any>('/api/v1/admin/groups', apiParams);
      
      console.log('📥 API Response:', {
        success: response.success,
        dataLength: response.data?.length || 0,
        hasData: !!response.data,
        pagination: response.pagination
      });
  
      // If we got no data but expected some, log it
      if (!response.data || response.data.length === 0) {
        console.warn('⚠️ No groups returned from API. This could mean:');
        console.warn('1. No groups match the filters');
        console.warn('2. Backend is not processing filters correctly');
        console.warn('3. Database query issue');
      }
  
      // Process the response to normalize creator data if needed
      if (response.data && Array.isArray(response.data)) {
        console.log(`✅ Processing ${response.data.length} groups`);
  
        // Map through groups to normalize creator data
        response.data = response.data.map(group => {
          // If there's a created_by ID but no creator object
          if (group.created_by && !group.creator) {
            // The API might return creator under a different name
            const creatorData = group.owner || group.admin || group.created_by_user;
  
            if (creatorData) {
              group.creator = creatorData;
              console.log(`🔄 Normalized creator data for group ${group.id}`);
            }
          }
          return group;
        });
      }
  
      // Handle both paginated and simple array responses
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Groups fetched successfully',
          data: response.data,
          pagination: response.pagination || {
            current_page: params?.page || 1,
            per_page: response.data.length,
            total: response.data.length,
            total_pages: 1,
            has_next: false,
            has_previous: false
          }
        };
      }
  
      return response as PaginatedResponse;
    } catch (error) {
      console.error('❌ Get groups error:', error);
      throw error;
    }
  }

async getGroup(id: string): Promise<ApiResponse<any>> {
  return this.get(`/api/v1/admin/groups/${id}`);
}

async getGroupMembers(groupId: string, params?: {
  page?: number
  limit?: number
  role?: string
  search?: string
  include_user?: boolean  // Add this property to the type
}): Promise<PaginatedResponse> {
  try {
    const enhancedParams = {
      ...params,
      include_user: true,  // Request user data inclusion
    };
    
    const response = await this.get<any>(`/api/v1/admin/groups/${groupId}/members`, enhancedParams);
    
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
    
    return response as PaginatedResponse;
  } catch (error) {
    console.error('❌ Get group members error:', error);
    throw error;
  }
}

async updateGroupStatus(id: string, data: {
  is_active?: boolean
  status?: string
  reason?: string
}): Promise<ApiResponse<any>> {
  return this.put(`/api/v1/admin/groups/${id}/status`, data);
}

async deleteGroup(id: string, reason?: string): Promise<ApiResponse<any>> {
  return this.delete(`/api/v1/admin/groups/${id}`, { reason });
}

async bulkGroupAction(data: {
  group_ids: string[]
  action: string
  reason?: string
}): Promise<ApiResponse<any>> {
  return this.post('/api/v1/admin/groups/bulk/actions', data);
}


  // ==================== EVENT MANAGEMENT ====================
  async getEvents(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/events', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Events fetched successfully',
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
      console.error('❌ Get events error:', error)
      throw error
    }
  }

  async getEvent(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/events/${id}`)
  }

  async getEventAttendees(id: string, params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>(`/api/v1/admin/events/${id}/attendees`, params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Event attendees fetched successfully',
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
      console.error('❌ Get event attendees error:', error)
      throw error
    }
  }

  async updateEventStatus(id: string, data: { 
    status?: string
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.put(`/api/v1/admin/events/${id}/status`, data)
  }

  async deleteEvent(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.delete(`/api/v1/admin/events/${id}`, { reason })
  }

  async bulkEventAction(data: { 
    event_ids: string[]
    action: string
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/events/bulk/actions', data)
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

  async getStory(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/stories/${id}`)
  }

  async hideStory(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.put(`/api/v1/admin/stories/${id}/hide`, { reason })
  }

  async deleteStory(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.delete(`/api/v1/admin/stories/${id}`, { reason })
  }

  async bulkStoryAction(data: { 
    story_ids: string[]
    action: string
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/stories/bulk/actions', data)
  }

  // ==================== MESSAGE MANAGEMENT ====================
  async getMessages(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/messages', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Messages fetched successfully',
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
      console.error('❌ Get messages error:', error)
      throw error
    }
  }

  async getMessage(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/messages/${id}`)
  }

  async getConversations(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/messages/conversations', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Conversations fetched successfully',
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
      console.error('❌ Get conversations error:', error)
      throw error
    }
  }

  async getConversation(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/messages/conversations/${id}`)
  }

  async deleteMessage(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.delete(`/api/v1/admin/messages/${id}`, { reason })
  }

  async bulkMessageAction(data: { 
    message_ids: string[]
    action: string
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/messages/bulk/actions', data)
  }

  // ==================== REPORT MANAGEMENT ====================
  async getReports(params?: {
    status?: string
    target_type?: string
    reason?: string
    priority?: string
    assigned_to?: string
    resolved_by?: string
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

  async getReport(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/reports/${id}`)
  }

  async updateReportStatus(id: string, data: {
    status?: string
    resolution?: string
    note?: string
  }): Promise<ApiResponse<any>> {
    return this.put(`/api/v1/admin/reports/${id}/status`, data)
  }

  async assignReport(id: string, adminId: string): Promise<ApiResponse<any>> {
    return this.put(`/api/v1/admin/reports/${id}/assign`, { admin_id: adminId })
  }

  async resolveReport(id: string, data: { 
    resolution: string
    note?: string 
  }): Promise<ApiResponse<any>> {
    return this.post(`/api/v1/admin/reports/${id}/resolve`, data)
  }

  async rejectReport(id: string, data: { 
    note: string 
  }): Promise<ApiResponse<any>> {
    return this.post(`/api/v1/admin/reports/${id}/reject`, data)
  }

  async bulkReportAction(data: { 
    report_ids: string[]
    action: string
    resolution?: string
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/reports/bulk/actions', data)
  }

  async getReportStats(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/reports/stats')
  }

  async getReportSummary(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/reports/stats/summary')
  }

  // ==================== FOLLOW MANAGEMENT ====================
  async getFollows(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/follows', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Follows fetched successfully',
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
      console.error('❌ Get follows error:', error)
      throw error
    }
  }

  async getFollow(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/follows/${id}`)
  }

  async deleteFollow(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.delete(`/api/v1/admin/follows/${id}`, { reason })
  }

  async getRelationships(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/follows/relationships', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Relationships fetched successfully',
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
      console.error('❌ Get relationships error:', error)
      throw error
    }
  }

  async bulkFollowAction(data: { 
    follow_ids: string[]
    action: string
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/follows/bulk/actions', data)
  }

  // ==================== LIKE MANAGEMENT ====================
  async getLikes(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/likes', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Likes fetched successfully',
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
      console.error('❌ Get likes error:', error)
      throw error
    }
  }

  async getLikeStats(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/likes/stats')
  }

  async deleteLike(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.delete(`/api/v1/admin/likes/${id}`, { reason })
  }

  async bulkLikeAction(data: { 
    like_ids: string[]
    action: string
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/likes/bulk/actions', data)
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

  async getHashtag(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/hashtags/${id}`)
  }

  async getTrendingHashtags(limit: number = 20): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/hashtags/trending', { limit })
  }

  async blockHashtag(id: string, reason: string): Promise<ApiResponse<any>> {
    return this.put(`/api/v1/admin/hashtags/${id}/block`, { reason })
  }

  async unblockHashtag(id: string): Promise<ApiResponse<any>> {
    return this.put(`/api/v1/admin/hashtags/${id}/unblock`)
  }

  async deleteHashtag(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.delete(`/api/v1/admin/hashtags/${id}`, { reason })
  }

  async bulkHashtagAction(data: { 
    hashtag_ids: string[]
    action: string
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/hashtags/bulk/actions', data)
  }

  // ==================== MENTION MANAGEMENT ====================
  async getMentions(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/mentions', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Mentions fetched successfully',
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
      console.error('❌ Get mentions error:', error)
      throw error
    }
  }

  async getMention(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/mentions/${id}`)
  }

  async deleteMention(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.delete(`/api/v1/admin/mentions/${id}`, { reason })
  }

  async bulkMentionAction(data: { 
    mention_ids: string[]
    action: string
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/mentions/bulk/actions', data)
  }

  // ==================== MEDIA MANAGEMENT ====================
  async getMedia(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/media', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Media fetched successfully',
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
      console.error('❌ Get media error:', error)
      throw error
    }
  }

  async getMediaItem(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/media/${id}`)
  }

  async getMediaStats(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/media/stats')
  }

  async moderateMedia(id: string, data: {
    moderation_status: string
    reason?: string
  }): Promise<ApiResponse<any>> {
    return this.put(`/api/v1/admin/media/${id}/moderate`, data)
  }

  async deleteMedia(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.delete(`/api/v1/admin/media/${id}`, { reason })
  }

  async bulkMediaAction(data: { 
    media_ids: string[]
    action: string
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/media/bulk/actions', data)
  }

  async getStorageStats(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/media/storage/stats')
  }

  async cleanupStorage(): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/media/storage/cleanup')
  }

  // ==================== NOTIFICATION MANAGEMENT ====================
  async getNotifications(params?: any): Promise<PaginatedResponse> {
    try {
      const response = await this.get<any>('/api/v1/admin/notifications', params)
      
      if (Array.isArray(response.data)) {
        return {
          success: true,
          message: 'Notifications fetched successfully',
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
      console.error('❌ Get notifications error:', error)
      throw error
    }
  }

  async getNotification(id: string): Promise<ApiResponse<any>> {
    return this.get(`/api/v1/admin/notifications/${id}`)
  }

  async sendNotificationToUsers(data: {
    user_ids: string[]
    title: string
    message: string
    type?: string
    data?: any
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/notifications/send', data)
  }

  async broadcastNotification(data: {
    title: string
    message: string
    type?: string
    data?: any
    criteria?: any
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/notifications/broadcast', data)
  }

  async getNotificationStats(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/notifications/stats')
  }

  async deleteNotification(id: string, reason?: string): Promise<ApiResponse<any>> {
    return this.delete(`/api/v1/admin/notifications/${id}`, { reason })
  }

  async bulkNotificationAction(data: { 
    notification_ids: string[]
    action: string
    reason?: string 
  }): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/notifications/bulk/actions', data)
  }

  // ==================== ANALYTICS ====================
  async getUserAnalytics(params?: any): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/analytics/users', params)
  }

  async getContentAnalytics(params?: any): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/analytics/content', params)
  }

  async getEngagementAnalytics(params?: any): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/analytics/engagement', params)
  }

  async getGrowthAnalytics(params?: any): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/analytics/growth', params)
  }

  async getDemographicAnalytics(params?: any): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/analytics/demographics', params)
  }

  async getRevenueAnalytics(params?: any): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/analytics/revenue', params)
  }

  async getCustomReport(params?: any): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/analytics/reports/custom', params)
  }

  async getRealtimeAnalytics(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/analytics/realtime')
  }

  async getLiveStats(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/analytics/live-stats')
  }

  // ==================== SYSTEM MANAGEMENT (Super Admin only) ====================
  async getSystemHealth(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/system/health')
  }

  async getSystemInfo(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/system/info')
  }

  async getSystemLogs(params?: any): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/system/logs', params)
  }

  async getPerformanceMetrics(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/system/performance')
  }

  async getDatabaseStats(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/system/database/stats')
  }

  async getCacheStats(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/system/cache/stats')
  }

  async clearCache(): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/system/cache/clear')
  }

  async warmCache(): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/system/cache/warm')
  }

  async enableMaintenanceMode(): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/system/maintenance/enable')
  }

  async disableMaintenanceMode(): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/system/maintenance/disable')
  }

  async backupDatabase(): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/system/database/backup')
  }

  async getDatabaseBackups(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/system/database/backups')
  }

  async restoreDatabase(backupId: string): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/system/database/restore', { backup_id: backupId })
  }

  async optimizeDatabase(): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/system/database/optimize')
  }

  // ==================== CONFIGURATION MANAGEMENT (Super Admin only) ====================
  async getConfiguration(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/config')
  }

  async updateConfiguration(config: any): Promise<ApiResponse<any>> {
    return this.put('/api/v1/admin/config', config)
  }

  async getConfigurationHistory(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/config/history')
  }

  async rollbackConfiguration(version: string): Promise<ApiResponse<any>> {
    return this.post('/api/v1/admin/config/rollback', { version })
  }

  async validateConfiguration(config: any): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/config/validate', config)
  }

  async getFeatureFlags(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/config/features')
  }

  async updateFeatureFlags(features: any): Promise<ApiResponse<any>> {
    return this.put('/api/v1/admin/config/features', features)
  }

  async toggleFeature(feature: string): Promise<ApiResponse<any>> {
    return this.put(`/api/v1/admin/config/features/${feature}/toggle`)
  }

  async getRateLimits(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/config/rate-limits')
  }

  async updateRateLimits(limits: any): Promise<ApiResponse<any>> {
    return this.put('/api/v1/admin/config/rate-limits', limits)
  }

  // ==================== PUBLIC ADMIN ROUTES ====================
  async getPublicSystemStatus(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/public/status')
  }

  async getPublicHealthCheck(): Promise<ApiResponse<any>> {
    return this.get('/api/v1/admin/public/health')
  }

  // ==================== WEBSOCKET CONNECTIONS ====================
  connectWebSocket(endpoint: string, onMessage?: (data: any) => void): WebSocket | null {
    if (typeof window === 'undefined') return null

    try {
      const token = this.getStoredToken()
      const wsUrl = `${process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'}/api/v1/admin/ws/${endpoint}`
      
      const ws = new WebSocket(wsUrl, ['Authorization', `Bearer ${token}`])

      ws.onopen = () => {
        console.log(`✅ WebSocket connected: ${endpoint}`)
      }

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          onMessage?.(data)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      ws.onclose = () => {
        console.log(`🔌 WebSocket disconnected: ${endpoint}`)
      }

      ws.onerror = (error) => {
        console.error(`❌ WebSocket error on ${endpoint}:`, error)
      }

      return ws
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error)
      return null
    }
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