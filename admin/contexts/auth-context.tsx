// contexts/auth-context.tsx - Fixed for proper token synchronization
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
import { AdminUser, LoginResponse } from "@/types/admin";

interface AuthContextType {
  user: AdminUser | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
  isAuthenticated: boolean;
  hasValidToken: boolean; // New: tracks if we have a valid token
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AdminUser | null>(null);
  const [loading, setLoading] = useState(true);
  const [mounted, setMounted] = useState(false);
  const [hasValidToken, setHasValidToken] = useState(false);
  const router = useRouter();

  const isAuthenticated = !!user && hasValidToken;

  // Handle hydration
  useEffect(() => {
    setMounted(true);
  }, []);

  // Check if token exists and is valid
  const checkTokenValidity = useCallback(() => {
    if (typeof window === "undefined") return false;

    const token = localStorage.getItem("admin_token");
    const userData = localStorage.getItem("admin_user");

    return !!(token && userData);
  }, []);

  useEffect(() => {
    // Only run on client side after mounting
    if (!mounted) return;

    console.log("🔄 Checking existing authentication...");

    const token = localStorage.getItem("admin_token");
    const userData = localStorage.getItem("admin_user");

    if (token && userData) {
      try {
        const parsedUser = JSON.parse(userData);
        setUser(parsedUser);
        setHasValidToken(true);
        console.log(
          "✅ Token and user restored from localStorage:",
          parsedUser.username
        );
      } catch (error) {
        console.error("❌ Error parsing user data:", error);
        // Clear corrupted data
        localStorage.removeItem("admin_token");
        localStorage.removeItem("admin_user");
        localStorage.removeItem("admin_refresh_token");
        setHasValidToken(false);
      }
    } else {
      console.log("ℹ️ No existing authentication found");
      setHasValidToken(false);
    }

    setLoading(false);
  }, [mounted]);

  const login = async (email: string, password: string) => {
    try {
      setLoading(true);
      console.log("🔄 Starting login process...");

      const response = await apiClient.adminLogin(email, password);
      console.log("✅ Login API response received");

      if (!response.success || !response.data) {
        throw new Error("Invalid login response structure");
      }

      const { access_token, refresh_token, user: userData } = response.data;

      if (!access_token || !userData) {
        throw new Error("Missing required login data");
      }

      // Store tokens and user data atomically
      if (typeof window !== "undefined") {
        localStorage.setItem("admin_token", access_token);
        localStorage.setItem("admin_refresh_token", refresh_token || "");
        localStorage.setItem("admin_user", JSON.stringify(userData));
        console.log("✅ Login data stored in localStorage");
      }

      // Update state atomically
      setUser(userData);
      setHasValidToken(true);

      console.log("✅ Login successful, user state updated");

      // Wait a bit longer to ensure everything is properly set
      await new Promise((resolve) => setTimeout(resolve, 200));

      // Verify token was stored
      if (typeof window !== "undefined") {
        const storedToken = localStorage.getItem("admin_token");
        if (!storedToken) {
          throw new Error("Failed to store authentication token");
        }
        console.log("✅ Token verified in localStorage");
      }
    } catch (error: any) {
      console.error("❌ Login error:", error);

      // Clear any partial data on error
      if (typeof window !== "undefined") {
        localStorage.removeItem("admin_token");
        localStorage.removeItem("admin_refresh_token");
        localStorage.removeItem("admin_user");
      }
      setUser(null);
      setHasValidToken(false);

      const errorMessage =
        error.response?.data?.message || error.message || "Login failed";
      throw new Error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    try {
      console.log("🔄 Logging out...");
      setHasValidToken(false); // Immediately mark as invalid
      await apiClient.adminLogout();
    } catch (error) {
      console.error("❌ Logout error:", error);
    } finally {
      // Clear local storage regardless of API call result
      if (typeof window !== "undefined") {
        localStorage.removeItem("admin_token");
        localStorage.removeItem("admin_refresh_token");
        localStorage.removeItem("admin_user");
      }
      setUser(null);
      setHasValidToken(false);
      console.log("✅ Logout complete, redirecting to login");
      router.push("/admin/login");
    }
  };

  const refreshToken = async () => {
    try {
      console.log("🔄 Refreshing token...");
      setHasValidToken(false); // Temporarily mark as invalid

      const response = await apiClient.refreshToken();

      if (response.success && response.data?.access_token) {
        if (typeof window !== "undefined") {
          localStorage.setItem("admin_token", response.data.access_token);
        }
        setHasValidToken(true); // Mark as valid again
        console.log("✅ Token refreshed successfully");
      } else {
        throw new Error("Invalid refresh response");
      }
    } catch (error) {
      console.error("❌ Token refresh error:", error);
      setHasValidToken(false);
      logout();
    }
  };

  // Listen for storage changes (token updates from API client)
  useEffect(() => {
    if (!mounted) return;

    const handleStorageChange = (event: StorageEvent) => {
      if (event.key === "admin_token") {
        const hasToken = !!event.newValue;
        setHasValidToken(hasToken);
        console.log("🔄 Token storage changed, hasValidToken:", hasToken);
      }
    };

    window.addEventListener("storage", handleStorageChange);
    return () => window.removeEventListener("storage", handleStorageChange);
  }, [mounted]);

  const value = {
    user,
    loading,
    login,
    logout,
    refreshToken,
    isAuthenticated,
    hasValidToken,
  };

  // Don't render children until mounted to prevent hydration mismatch
  if (!mounted) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-32 w-32 animate-spin rounded-full border-b-2 border-primary"></div>
      </div>
    );
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

// Higher-order component for protecting admin routes
export function withAuth<T extends object>(Component: React.ComponentType<T>) {
  return function AuthenticatedComponent(props: T) {
    const { isAuthenticated, loading, hasValidToken } = useAuth();
    const router = useRouter();
    const [mounted, setMounted] = useState(false);

    useEffect(() => {
      setMounted(true);
    }, []);

    useEffect(() => {
      if (mounted && !loading && (!isAuthenticated || !hasValidToken)) {
        console.log(
          "🔄 Not authenticated or invalid token, redirecting to login"
        );
        router.push("/admin/login");
      }
    }, [mounted, isAuthenticated, loading, hasValidToken, router]);

    if (!mounted || loading) {
      return (
        <div className="flex h-screen items-center justify-center">
          <div className="text-center">
            <div className="h-16 w-16 animate-spin rounded-full border-b-2 border-primary mx-auto mb-4"></div>
            <p className="text-muted-foreground">Loading...</p>
          </div>
        </div>
      );
    }

    if (!isAuthenticated || !hasValidToken) {
      return (
        <div className="flex h-screen items-center justify-center">
          <div className="text-center">
            <div className="h-16 w-16 animate-spin rounded-full border-b-2 border-primary mx-auto mb-4"></div>
            <p className="text-muted-foreground">Redirecting to login...</p>
          </div>
        </div>
      );
    }

    return <Component {...props} />;
  };
}

// Admin Providers Component
export function AdminProviders({ children }: { children: React.ReactNode }) {
  return <AuthProvider>{children}</AuthProvider>;
}
