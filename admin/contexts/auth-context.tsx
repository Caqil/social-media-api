// contexts/auth-context.tsx - Updated with simplified token handling
"use client";

import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
} from "react";
import { useRouter } from "next/navigation";
import { apiClient } from "@/lib/api-client";
import { authStorage } from "@/lib/storage-utils";

interface User {
  id: string;
  email: string;
  username: string;
  first_name?: string;
  last_name?: string;
  role: string;
  permissions?: string[];
  is_verified: boolean;
  created_at: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  hasValidToken: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<void>;
  checkAuth: () => Promise<boolean>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [hasValidToken, setHasValidToken] = useState(false);
  const router = useRouter();

  // Check if token exists and is valid
  const validateToken = useCallback((tokenToCheck?: string) => {
    const currentToken = tokenToCheck || authStorage.getToken();
    if (!currentToken) {
      return false;
    }

    try {
      // Basic JWT validation - check if it's not expired
      const payload = JSON.parse(atob(currentToken.split(".")[1]));
      const currentTime = Math.floor(Date.now() / 1000);

      if (payload.exp && payload.exp < currentTime) {
        console.log("Token is expired");
        return false;
      }

      return true;
    } catch (error) {
      console.error("Token validation error:", error);
      return false;
    }
  }, []);

  // Initialize auth state from storage
  const initializeAuth = useCallback(async () => {
    console.log("🔄 Initializing auth...");

    try {
      const storedToken = authStorage.getToken();
      const storedUser = authStorage.getUser<User>();

      console.log("📦 Found in storage:", {
        hasToken: !!storedToken,
        hasUser: !!storedUser,
        tokenLength: storedToken?.length,
      });

      if (storedToken && storedUser) {
        const isValidToken = validateToken(storedToken);

        if (isValidToken) {
          console.log("✅ Valid token found, setting auth state");
          setToken(storedToken);
          setUser(storedUser);
          setIsAuthenticated(true);
          setHasValidToken(true);
        } else {
          console.log("❌ Invalid token, trying to refresh...");
          try {
            await refreshTokens();
          } catch (error) {
            console.log("❌ Refresh failed, clearing auth");
            clearAuthState();
          }
        }
      } else {
        console.log("⚠️ No auth data found in storage");
        clearAuthState();
      }
    } catch (error) {
      console.error("❌ Auth initialization error:", error);
      clearAuthState();
    } finally {
      setIsLoading(false);
    }
  }, [validateToken]);

  // Clear auth state
  const clearAuthState = useCallback(() => {
    console.log("🗑️ Clearing auth state");
    setUser(null);
    setToken(null);
    setIsAuthenticated(false);
    setHasValidToken(false);
    authStorage.clearAll();
  }, []);

  // Refresh tokens
  const refreshTokens = useCallback(async () => {
    console.log("🔄 Refreshing tokens...");

    try {
      const refreshToken = authStorage.getRefreshToken();
      if (!refreshToken) {
        throw new Error("No refresh token available");
      }

      const response = await apiClient.refreshToken();

      if (response.success && response.data?.access_token) {
        const newToken = response.data.access_token;

        console.log("✅ Token refresh successful");
        setToken(newToken);
        setHasValidToken(true);
        authStorage.setToken(newToken);

        // Update refresh token if provided
        if (response.data.refresh_token) {
          authStorage.setRefreshToken(response.data.refresh_token);
        }
      } else {
        throw new Error("Invalid refresh response");
      }
    } catch (error) {
      console.error("❌ Token refresh failed:", error);
      throw error;
    }
  }, []);

  // Login function
  const login = useCallback(
    async (email: string, password: string) => {
      console.log("🔄 Starting login process...");
      setIsLoading(true);

      try {
        const response = await apiClient.adminLogin(email, password);

        console.log("📡 Login response:", {
          success: response.success,
          hasData: !!response.data,
          hasAccessToken: !!response.data?.access_token,
          hasRefreshToken: !!response.data?.refresh_token,
          hasUser: !!response.data?.user,
        });

        if (response.success && response.data) {
          const { access_token, refresh_token, user: userData } = response.data;

          if (!access_token || !userData) {
            throw new Error("Missing required login data");
          }

          console.log("✅ Login successful, setting auth state");

          // Set state
          setToken(access_token);
          setUser(userData);
          setIsAuthenticated(true);
          setHasValidToken(true);

          // Store in localStorage (already done by apiClient, but ensuring consistency)
          authStorage.setToken(access_token);
          // Only store refresh token if it exists
          if (refresh_token) {
            authStorage.setRefreshToken(refresh_token);
          } else {
            console.log("⚠️ No refresh token provided in login response");
          }
          authStorage.setUser(userData);

          console.log("✅ Auth state updated successfully");
        } else {
          const errorMessage = response.message || "Login failed";
          console.error("❌ Login failed:", errorMessage);
          throw new Error(errorMessage);
        }
      } catch (error: any) {
        console.error("❌ Login error:", error);
        clearAuthState();
        throw new Error(
          error.response?.data?.message || error.message || "Login failed"
        );
      } finally {
        setIsLoading(false);
      }
    },
    [clearAuthState]
  );

  // Logout function
  const logout = useCallback(async () => {
    console.log("🔄 Starting logout process...");
    setIsLoading(true);

    try {
      // Call logout API (this will clear tokens on server)
      await apiClient.adminLogout();
      console.log("✅ Server logout successful");
    } catch (error) {
      console.error(
        "⚠️ Server logout failed, but continuing with local logout:",
        error
      );
    } finally {
      // Always clear local state regardless of server response
      clearAuthState();
      setIsLoading(false);

      // Redirect to login page
      router.push("/admin/login");
      console.log("✅ Logout completed");
    }
  }, [clearAuthState, router]);

  // Check auth status
  const checkAuth = useCallback(async (): Promise<boolean> => {
    const currentToken = authStorage.getToken();
    const currentUser = authStorage.getUser<User>();

    if (!currentToken || !currentUser) {
      return false;
    }

    const isValidToken = validateToken(currentToken);

    if (isValidToken) {
      setToken(currentToken);
      setUser(currentUser);
      setIsAuthenticated(true);
      setHasValidToken(true);
      return true;
    } else {
      try {
        await refreshTokens();
        return true;
      } catch (error) {
        clearAuthState();
        return false;
      }
    }
  }, [validateToken, refreshTokens, clearAuthState]);

  // Initialize auth on mount
  useEffect(() => {
    initializeAuth();
  }, [initializeAuth]);

  // Set up token refresh interval
  useEffect(() => {
    if (!isAuthenticated || !token) {
      return;
    }

    // Check token validity every 5 minutes
    const interval = setInterval(() => {
      const isValid = validateToken();
      if (!isValid) {
        console.log("🔄 Token expired, attempting refresh...");
        refreshTokens().catch((error) => {
          console.error("❌ Auto-refresh failed:", error);
          clearAuthState();
          router.push("/admin/login");
        });
      }
    }, 5 * 60 * 1000); // 5 minutes

    return () => clearInterval(interval);
  }, [
    isAuthenticated,
    token,
    validateToken,
    refreshTokens,
    clearAuthState,
    router,
  ]);

  const contextValue: AuthContextType = {
    user,
    token,
    isAuthenticated,
    isLoading,
    hasValidToken,
    login,
    logout,
    refreshToken: refreshTokens,
    checkAuth,
  };

  return (
    <AuthContext.Provider value={contextValue}>{children}</AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

// HOC for protecting routes
export function withAuth<P extends object>(
  WrappedComponent: React.ComponentType<P>
) {
  return function AuthenticatedComponent(props: P) {
    const { isAuthenticated, isLoading, hasValidToken } = useAuth();
    const router = useRouter();

    useEffect(() => {
      if (!isLoading && (!isAuthenticated || !hasValidToken)) {
        console.log("🔒 Not authenticated, redirecting to login");
        router.push("/admin/login");
      }
    }, [isAuthenticated, isLoading, hasValidToken, router]);

    // Show loading spinner while checking auth
    if (isLoading) {
      return (
        <div className="flex h-screen items-center justify-center">
          <div className="text-center">
            <div className="h-16 w-16 animate-spin rounded-full border-b-2 border-primary mx-auto mb-4"></div>
            <p className="text-muted-foreground">Loading...</p>
          </div>
        </div>
      );
    }

    // Don't render the component if not authenticated
    if (!isAuthenticated || !hasValidToken) {
      return null;
    }

    return <WrappedComponent {...props} />;
  };
}

export default AuthProvider;
