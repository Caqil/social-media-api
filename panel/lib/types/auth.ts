import { BaseEntity, Permission, Role } from './common';

export interface User extends BaseEntity {
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  display_name: string;
  bio?: string;
  avatar?: string;
  phone?: string;
  date_of_birth?: string;
  gender?: 'male' | 'female' | 'other' | 'prefer_not_to_say';
  location?: string;
  website?: string;
  is_verified: boolean;
  is_admin: boolean;
  is_moderator: boolean;
  role: string;
  status: 'active' | 'inactive' | 'pending' | 'suspended';
  email_verified_at?: string;
  phone_verified_at?: string;
  last_login_at?: string;
  last_seen_at?: string;
  login_count: number;
  failed_login_attempts: number;
  locked_until?: string;
  two_factor_enabled: boolean;
  permissions: Permission[];
  roles: Role[];
}

export interface LoginRequest {
  email_or_username: string;
  password: string;
  device_info?: string;
  remember_me?: boolean;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  user: User;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  password_confirmation: string;
  first_name: string;
  last_name: string;
  display_name?: string;
  bio?: string;
  date_of_birth?: string;
  gender?: string;
  phone?: string;
  terms_accepted: boolean;
  privacy_accepted: boolean;
}

export interface RegisterResponse {
  user: User;
  verification_required: boolean;
  message: string;
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
  password_confirmation: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
  password_confirmation: string;
}

export interface VerifyEmailRequest {
  token: string;
}

export interface Session extends BaseEntity {
  user_id: string;
  device_info: string;
  ip_address: string;
  location?: string;
  user_agent: string;
  last_used_at: string;
  expires_at: string;
  is_current: boolean;
  is_active: boolean;
}

export interface LoginHistory extends BaseEntity {
  user_id: string;
  ip_address: string;
  user_agent: string;
  device_info: string;
  location?: string;
  success: boolean;
  failure_reason?: string;
  session_id?: string;
}

export interface SecuritySettings {
  two_factor_enabled: boolean;
  login_notifications: boolean;
  session_timeout: number;
  max_sessions: number;
  password_expiry_days: number;
  failed_login_threshold: number;
  account_lockout_duration: number;
}

export interface TwoFactorSettings {
  enabled: boolean;
  method: 'totp' | 'sms' | 'email';
  backup_codes: string[];
  verified_at?: string;
}

export interface PasswordPolicy {
  min_length: number;
  require_uppercase: boolean;
  require_lowercase: boolean;
  require_numbers: boolean;
  require_special_chars: boolean;
  prevent_common_passwords: boolean;
  prevent_personal_info: boolean;
  history_count: number;
  expiry_days: number;
}

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  permissions: Permission[];
  roles: Role[];
  session?: Session;
  lastActivity: string;
}

export interface AuthContextType {
  user: User | null;
  login: (credentials: LoginRequest) => Promise<LoginResponse>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<LoginResponse>;
  updateProfile: (data: Partial<User>) => Promise<User>;
  changePassword: (data: ChangePasswordRequest) => Promise<void>;
  isAuthenticated: boolean;
  isLoading: boolean;
  hasPermission: (permission: Permission) => boolean;
  hasRole: (role: string) => boolean;
  checkAuth: () => Promise<boolean>;
}

export interface AuthError {
  code: string;
  message: string;
  field?: string;
}

export type AuthErrorCode = 
  | 'INVALID_CREDENTIALS'
  | 'USER_NOT_FOUND'
  | 'USER_SUSPENDED'
  | 'EMAIL_NOT_VERIFIED'
  | 'ACCOUNT_LOCKED'
  | 'TOKEN_EXPIRED'
  | 'TOKEN_INVALID'
  | 'REFRESH_TOKEN_EXPIRED'
  | 'TWO_FACTOR_REQUIRED'
  | 'PERMISSION_DENIED'
  | 'SESSION_EXPIRED';

// OAuth types (if needed)
export interface OAuthProvider {
  id: string;
  name: string;
  enabled: boolean;
  client_id: string;
  authorize_url: string;
  token_url: string;
  user_info_url: string;
  scopes: string[];
}

export interface OAuthRequest {
  provider: string;
  code: string;
  state?: string;
  redirect_uri: string;
}

export interface OAuthResponse {
  user: User;
  is_new_user: boolean;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

// Admin auth types
export interface AdminLoginRequest extends LoginRequest {
  admin_code?: string;
}

export interface ImpersonateRequest {
  user_id: string;
  reason: string;
}

export interface ImpersonateResponse {
  original_user: User;
  impersonated_user: User;
  impersonation_token: string;
  expires_at: string;
}

export interface AdminAction {
  id: string;
  admin_id: string;
  action: string;
  target_type: string;
  target_id: string;
  reason?: string;
  metadata?: Record<string, any>;
  created_at: string;
}

// Form types
export interface LoginFormData {
  email_or_username: string;
  password: string;
  remember_me: boolean;
}

export interface RegisterFormData {
  username: string;
  email: string;
  password: string;
  password_confirmation: string;
  first_name: string;
  last_name: string;
  terms_accepted: boolean;
}

export interface ForgotPasswordFormData {
  email: string;
}

export interface ResetPasswordFormData {
  new_password: string;
  password_confirmation: string;
}

export interface ChangePasswordFormData {
  current_password: string;
  new_password: string;
  password_confirmation: string;
}