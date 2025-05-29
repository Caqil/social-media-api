import { apiClient } from '../api';

export interface LoginRequest {
  email_or_username: string;
  password: string;
  device_info?: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: {
    id: string;
    username: string;
    email: string;
    first_name: string;
    last_name: string;
    display_name: string;
    role: string;
    is_verified: boolean;
    is_admin: boolean;
    avatar?: string;
  };
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  new_password: string;
  confirm_password: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

export interface UserProfile {
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  display_name: string;
  bio?: string;
  avatar?: string;
  role: string;
  is_verified: boolean;
  is_admin: boolean;
  created_at: string;
  updated_at: string;
}

export interface Session {
  id: string;
  device_info: string;
  ip_address: string;
  created_at: string;
  last_used_at: string;
  is_current: boolean;
}

export class AuthService {
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await apiClient.post<LoginResponse>('/auth/login', credentials);
    
    // Set the token in the API client
    apiClient.setToken(response.access_token);
    
    // Store refresh token
    if (typeof window !== 'undefined') {
      localStorage.setItem('refresh_token', response.refresh_token);
    }
    
    return response;
  }

  async logout(): Promise<void> {
    try {
      await apiClient.post('/auth/logout');
    } catch (error) {
      // Continue with logout even if API call fails
      console.error('Logout API call failed:', error);
    } finally {
      // Clear tokens
      apiClient.clearToken();
      if (typeof window !== 'undefined') {
        localStorage.removeItem('refresh_token');
      }
    }
  }

  async logoutAll(): Promise<void> {
    try {
      await apiClient.post('/auth/logout-all');
    } catch (error) {
      console.error('Logout all API call failed:', error);
    } finally {
      // Clear tokens
      apiClient.clearToken();
      if (typeof window !== 'undefined') {
        localStorage.removeItem('refresh_token');
      }
    }
  }

  async refreshToken(): Promise<LoginResponse> {
    const refreshToken = typeof window !== 'undefined' 
      ? localStorage.getItem('refresh_token') 
      : null;
    
    if (!refreshToken) {
      throw new Error('No refresh token available');
    }

    const response = await apiClient.post<LoginResponse>('/auth/refresh', {
      refresh_token: refreshToken,
    });
    
    // Update tokens
    apiClient.setToken(response.access_token);
    if (typeof window !== 'undefined') {
      localStorage.setItem('refresh_token', response.refresh_token);
    }
    
    return response;
  }

  async getProfile(): Promise<UserProfile> {
    return apiClient.get<UserProfile>('/auth/profile');
  }

  async updateProfile(data: Partial<UserProfile>): Promise<UserProfile> {
    return apiClient.put<UserProfile>('/auth/profile', data);
  }

  async changePassword(data: ChangePasswordRequest): Promise<void> {
    await apiClient.post('/auth/change-password', data);
  }

  async forgotPassword(data: ForgotPasswordRequest): Promise<void> {
    await apiClient.post('/auth/forgot-password', data);
  }

  async resetPassword(data: ResetPasswordRequest): Promise<void> {
    await apiClient.post('/auth/reset-password', data);
  }

  async getSessions(): Promise<Session[]> {
    return apiClient.get<Session[]>('/auth/sessions');
  }

  async revokeSession(sessionId: string): Promise<void> {
    await apiClient.delete(`/auth/sessions/${sessionId}`);
  }

  async checkAuthStatus(): Promise<boolean> {
    try {
      await this.getProfile();
      return true;
    } catch (error) {
      // Try to refresh token
      try {
        await this.refreshToken();
        return true;
      } catch (refreshError) {
        // Both failed, user is not authenticated
        apiClient.clearToken();
        if (typeof window !== 'undefined') {
          localStorage.removeItem('refresh_token');
        }
        return false;
      }
    }
  }

  isAuthenticated(): boolean {
    if (typeof window === 'undefined') return false;
    
    const token = localStorage.getItem('access_token');
    return !!token;
  }

  getToken(): string | null {
    if (typeof window === 'undefined') return null;
    
    return localStorage.getItem('access_token');
  }
}

export const authService = new AuthService();