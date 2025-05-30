// contexts/auth-context.tsx - Fixed Version with Better Error Handling
"use client";

import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  useRef,
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
  isAuthenticated: boolean;
  hasValidToken: boolean;
  isLoading: boolean;
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<boolean>;
  clearError: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [hasValidToken, setHasValidToken] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  // Use refs to prevent duplicate calls
  const initializingRef = useRef(false);
  const loginInProgressRef = useRef(false);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  // Simple token validation
  const validateToken = useCallback(() => {
    try {
      const token = authStorage.getToken();
      const storedUser = authStorage.getUser<User>();

      if (!token || !storedUser) {
        console.log("❌ No token or user found");
        return false;
      }

      // Basic JWT validation
      const parts = token.split(".");
      if (parts.length !== 3) {
        console.log("❌ Invalid token format");
        return false;
      }

      try {
        const payload = JSON.parse(atob(parts[1]));
        const currentTime = Math.floor(Date.now() / 1000);

        if (payload.exp && payload.exp < currentTime) {
          console.log("❌ Token expired");
          return false;
        }
      } catch (parseError) {
        console.log("❌ Token payload invalid");
        return false;
      }

      console.log("✅ Token and user valid");
      return true;
    } catch (error) {
      console.error("❌ Token validation error:", error);
      return false;
    }
  }, []);

  // Initialize auth state
  useEffect(() => {
    const initializeAuth = async () => {
      if (initializingRef.current) {
        console.log("⚠️ Auth initialization already in progress, skipping...");
        return;
      }

      initializingRef.current = true;
      console.log("🔄 Initializing auth state...");

      try {
        setError(null);

        const isValid = validateToken();
        const storedUser = authStorage.getUser<User>();

        if (isValid && storedUser) {
          setUser(storedUser);
          setIsAuthenticated(true);
          setHasValidToken(true);
          console.log("✅ Auth state initialized with existing session");
        } else {
          // Clear invalid data
          authStorage.clearAll();
          setUser(null);
          setIsAuthenticated(false);
          setHasValidToken(false);
          console.log("❌ No valid session found, cleared invalid data");
        }
      } catch (error) {
        console.error("❌ Auth initialization error:", error);
        authStorage.clearAll();
        setUser(null);
        setIsAuthenticated(false);
        setHasValidToken(false);
        setError("Failed to initialize authentication");
      } finally {
        setIsLoading(false);
        initializingRef.current = false;
      }
    };

    // Only run on client side and only once
    if (typeof window !== "undefined" && !initializingRef.current) {
      initializeAuth();
    } else if (typeof window === "undefined") {
      setIsLoading(false);
    }
  }, [validateToken]);

  const login = useCallback(async (email: string, password: string) => {
    if (loginInProgressRef.current) {
      console.log("⚠️ Login already in progress, skipping...");
      return;
    }

    loginInProgressRef.current = true;
    console.log("🔄 Starting login process...");
    setIsLoading(true);
    setError(null);

    try {
      // Validate inputs
      if (!email || !password) {
        throw new Error("Email and password are required");
      }

      const response = await apiClient.adminLogin(email, password);

      if (!response) {
        throw new Error("No response from server");
      }

      if (!response.success) {
        throw new Error(response.message || "Login failed");
      }

      if (!response.data) {
        throw new Error("Invalid login response data");
      }

      const { user: userData } = response.data;

      if (!userData) {
        throw new Error("User data not found in response");
      }

      setUser(userData);
      setIsAuthenticated(true);
      setHasValidToken(true);
      setError(null);

      console.log("✅ Login successful, user state updated");
    } catch (error: any) {
      console.error("❌ Login failed:", error);

      // Clear any partial auth state
      authStorage.clearAll();
      setUser(null);
      setIsAuthenticated(false);
      setHasValidToken(false);

      const errorMessage = error.message || "Login failed. Please try again.";
      setError(errorMessage);

      throw new Error(errorMessage);
    } finally {
      setIsLoading(false);
      loginInProgressRef.current = false;
    }
  }, []);

  const logout = useCallback(async () => {
    console.log("🔄 Starting logout process...");
    setIsLoading(true);
    setError(null);

    try {
      await apiClient.adminLogout();
    } catch (error) {
      console.warn(
        "⚠️ Logout API call failed, but continuing with cleanup:",
        error
      );
    } finally {
      // Always clear local state
      setUser(null);
      setIsAuthenticated(false);
      setHasValidToken(false);
      setIsLoading(false);
      setError(null);

      console.log("✅ Logout completed");

      // Redirect to login
      router.push("/admin/login");
    }
  }, [router]);

  const checkAuth = useCallback(async (): Promise<boolean> => {
    try {
      setError(null);

      const isValid = validateToken();
      const storedUser = authStorage.getUser<User>();

      if (isValid && storedUser) {
        if (!isAuthenticated) {
          setUser(storedUser);
          setIsAuthenticated(true);
          setHasValidToken(true);
        }
        return true;
      } else {
        authStorage.clearAll();
        setUser(null);
        setIsAuthenticated(false);
        setHasValidToken(false);
        return false;
      }
    } catch (error) {
      console.error("❌ Check auth error:", error);
      setError("Authentication check failed");
      authStorage.clearAll();
      setUser(null);
      setIsAuthenticated(false);
      setHasValidToken(false);
      return false;
    }
  }, [isAuthenticated, validateToken]);

  const value: AuthContextType = {
    user,
    isAuthenticated,
    hasValidToken,
    isLoading,
    error,
    login,
    logout,
    checkAuth,
    clearError,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// Higher-order component for protecting routes
export function withAuth<P extends object>(Component: React.ComponentType<P>) {
  return function AuthenticatedComponent(props: P) {
    const { isAuthenticated, hasValidToken, isLoading, user, error } =
      useAuth();
    const router = useRouter();
    const redirectingRef = useRef(false);

    useEffect(() => {
      if (!isLoading && (!isAuthenticated || !hasValidToken || !user)) {
        if (!redirectingRef.current) {
          redirectingRef.current = true;
          console.log("🔄 Redirecting to login - not authenticated");
          router.push("/admin/login");
        }
      }
    }, [isAuthenticated, hasValidToken, user, isLoading, router]);

    // Show loading while checking auth
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

    // Show error if authentication failed
    if (error) {
      return (
        <div className="flex h-screen items-center justify-center">
          <div className="text-center max-w-md">
            <div className="text-red-600 mb-4">Authentication Error</div>
            <p className="text-muted-foreground mb-4">{error}</p>
            <button
              onClick={() => router.push("/admin/login")}
              className="px-4 py-2 bg-primary text-white rounded hover:bg-primary/90"
            >
              Go to Login
            </button>
          </div>
        </div>
      );
    }

    // Show nothing while redirecting
    if (!isAuthenticated || !hasValidToken || !user) {
      return (
        <div className="flex h-screen items-center justify-center">
          <div className="text-center">
            <div className="h-16 w-16 animate-spin rounded-full border-b-2 border-primary mx-auto mb-4"></div>
            <p className="text-muted-foreground">Redirecting...</p>
          </div>
        </div>
      );
    }

    return <Component {...props} />;
  };
}

// Hook for checking specific permissions
export function usePermissions() {
  const { user } = useAuth();

  const hasPermission = useCallback(
    (permission: string): boolean => {
      if (!user) return false;

      // Super admin has all permissions
      if (user.role === "super_admin") return true;

      // Check specific permission
      return (
        user.permissions?.includes(permission) ||
        user.permissions?.includes("admin.*") ||
        user.permissions?.includes("*") ||
        false
      );
    },
    [user]
  );

  const hasRole = useCallback(
    (role: string): boolean => {
      return user?.role === role;
    },
    [user]
  );

  const hasAnyRole = useCallback(
    (roles: string[]): boolean => {
      return roles.some((role) => user?.role === role);
    },
    [user]
  );

  return {
    hasPermission,
    hasRole,
    hasAnyRole,
    userRole: user?.role,
    userPermissions: user?.permissions || [],
  };
}
