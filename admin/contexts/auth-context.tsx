// contexts/auth-context.tsx - Fixed for SSR
"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
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
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AdminUser | null>(null);
  const [loading, setLoading] = useState(true);
  const [mounted, setMounted] = useState(false);
  const router = useRouter();

  const isAuthenticated = !!user;

  // Handle hydration
  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    // Only run on client side after mounting
    if (!mounted) return;

    console.log("🔄 Checking existing authentication...");

    // Check for existing token on mount
    const token =
      typeof window !== "undefined"
        ? localStorage.getItem("admin_token")
        : null;
    const userData =
      typeof window !== "undefined" ? localStorage.getItem("admin_user") : null;

    if (token && userData) {
      try {
        const parsedUser = JSON.parse(userData);
        setUser(parsedUser);
        console.log(
          "✅ Token and user restored from localStorage:",
          parsedUser.username
        );
      } catch (error) {
        console.error("❌ Error parsing user data:", error);
        // Clear corrupted data
        if (typeof window !== "undefined") {
          localStorage.removeItem("admin_token");
          localStorage.removeItem("admin_user");
          localStorage.removeItem("admin_refresh_token");
        }
      }
    } else {
      console.log("ℹ️ No existing authentication found");
    }

    setLoading(false);
  }, [mounted]);

  const login = async (email: string, password: string) => {
    try {
      setLoading(true);
      console.log("🔄 Starting login process...");

      const response = await apiClient.adminLogin(email, password);
      console.log("✅ Login API response received");

      // Verify response structure
      if (!response.success || !response.data) {
        throw new Error("Invalid login response structure");
      }

      const { access_token, refresh_token, user: userData } = response.data;

      if (!access_token || !userData) {
        throw new Error("Missing required login data");
      }

      // Only store if we're on the client side
      if (typeof window !== "undefined") {
        // Store tokens and user data
        localStorage.setItem("admin_token", access_token);
        localStorage.setItem("admin_refresh_token", refresh_token || "");
        localStorage.setItem("admin_user", JSON.stringify(userData));

        console.log("✅ Login data stored in localStorage");
      }

      // Update state
      setUser(userData);

      console.log("✅ Login successful, user state updated");

      // Small delay to ensure everything is set
      await new Promise((resolve) => setTimeout(resolve, 100));

      // Verify token was stored (only on client)
      if (typeof window !== "undefined") {
        const storedToken = localStorage.getItem("admin_token");
        if (!storedToken) {
          throw new Error("Failed to store authentication token");
        }
        console.log("✅ Token verified in localStorage");
      }

      console.log("✅ Redirecting to dashboard...");
      router.push("/admin/dashboard");
    } catch (error: any) {
      console.error("❌ Login error:", error);

      // Clear any partial data on error (only on client)
      if (typeof window !== "undefined") {
        localStorage.removeItem("admin_token");
        localStorage.removeItem("admin_refresh_token");
        localStorage.removeItem("admin_user");
      }
      setUser(null);

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
      await apiClient.adminLogout();
    } catch (error) {
      console.error("❌ Logout error:", error);
    } finally {
      // Clear local storage regardless of API call result (only on client)
      if (typeof window !== "undefined") {
        localStorage.removeItem("admin_token");
        localStorage.removeItem("admin_refresh_token");
        localStorage.removeItem("admin_user");
      }
      setUser(null);
      console.log("✅ Logout complete, redirecting to login");
      router.push("/admin/login");
    }
  };

  const refreshToken = async () => {
    try {
      console.log("🔄 Refreshing token...");
      const response = await apiClient.refreshToken();

      if (response.success && response.data?.access_token) {
        if (typeof window !== "undefined") {
          localStorage.setItem("admin_token", response.data.access_token);
        }
        console.log("✅ Token refreshed successfully");
      } else {
        throw new Error("Invalid refresh response");
      }
    } catch (error) {
      console.error("❌ Token refresh error:", error);
      logout();
    }
  };

  const value = {
    user,
    loading,
    login,
    logout,
    refreshToken,
    isAuthenticated,
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
    const { isAuthenticated, loading } = useAuth();
    const router = useRouter();
    const [mounted, setMounted] = useState(false);

    useEffect(() => {
      setMounted(true);
    }, []);

    useEffect(() => {
      if (mounted && !loading && !isAuthenticated) {
        console.log("🔄 Not authenticated, redirecting to login");
        router.push("/admin/login");
      }
    }, [mounted, isAuthenticated, loading, router]);

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

    if (!isAuthenticated) {
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
