// contexts/auth-context.tsx - Final Fix for Duplicate Calls
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
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<boolean>;
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
  const router = useRouter();

  // Use refs to prevent duplicate calls
  const initializingRef = useRef(false);
  const loginInProgressRef = useRef(false);

  // Simple token validation - just check existence
  const validateToken = useCallback(() => {
    const token = authStorage.getToken();
    const storedUser = authStorage.getUser<User>();

    if (!token || !storedUser) {
      console.log("❌ No token or user found");
      return false;
    }

    console.log("✅ Token and user found");
    return true;
  }, []);

  // Initialize auth state - prevent duplicate calls
  useEffect(() => {
    const initializeAuth = async () => {
      // Prevent duplicate initialization
      if (initializingRef.current) {
        console.log("⚠️ Auth initialization already in progress, skipping...");
        return;
      }

      initializingRef.current = true;
      console.log("🔄 Initializing auth state...");

      try {
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
          console.log("❌ No valid session found");
        }
      } catch (error) {
        console.error("❌ Auth initialization error:", error);
        authStorage.clearAll();
        setUser(null);
        setIsAuthenticated(false);
        setHasValidToken(false);
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
    // Prevent duplicate login calls
    if (loginInProgressRef.current) {
      console.log("⚠️ Login already in progress, skipping...");
      return;
    }

    loginInProgressRef.current = true;
    console.log("🔄 Starting login process...");
    setIsLoading(true);

    try {
      const response = await apiClient.adminLogin(email, password);

      if (response.success && response.data) {
        const { user: userData } = response.data;

        setUser(userData);
        setIsAuthenticated(true);
        setHasValidToken(true);

        console.log("✅ Login successful, user state updated");
      } else {
        throw new Error(response.message || "Login failed");
      }
    } catch (error: any) {
      console.error("❌ Login failed:", error);

      // Clear any partial auth state
      authStorage.clearAll();
      setUser(null);
      setIsAuthenticated(false);
      setHasValidToken(false);

      throw new Error(
        error.response?.data?.message ||
          error.message ||
          "Login failed. Please try again."
      );
    } finally {
      setIsLoading(false);
      loginInProgressRef.current = false;
    }
  }, []);

  const logout = useCallback(async () => {
    console.log("🔄 Starting logout process...");
    setIsLoading(true);

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

      console.log("✅ Logout completed");

      // Redirect to login
      router.push("/admin/login");
    }
  }, [router]);

  const checkAuth = useCallback(async (): Promise<boolean> => {
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
  }, [isAuthenticated, validateToken]);

  const value: AuthContextType = {
    user,
    isAuthenticated,
    hasValidToken,
    isLoading,
    login,
    logout,
    checkAuth,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// Higher-order component for protecting routes
export function withAuth<P extends object>(Component: React.ComponentType<P>) {
  return function AuthenticatedComponent(props: P) {
    const { isAuthenticated, hasValidToken, isLoading, user } = useAuth();
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
