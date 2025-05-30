// app/admin/dashboard/page.tsx - Final Version with Request Deduplication
"use client";

import { useEffect, useState, useCallback } from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { ChartAreaInteractive } from "@/components/chart-area-interactive";
import { SiteHeader } from "@/components/site-header";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { withAuth, useAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import { requestManager } from "@/lib/request-manager";
import { DashboardStats } from "@/types/admin";
import {
  IconTrendingUp,
  IconTrendingDown,
  IconUsers,
  IconMessages,
  IconReport,
  IconHeart,
  IconRefresh,
} from "@tabler/icons-react";
import { Button } from "@/components/ui/button";

const DASHBOARD_STATS_KEY = "dashboard-stats";

function DashboardPage() {
  const { user } = useAuth();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchDashboardStats = useCallback(async (forceRefresh = false) => {
    try {
      setLoading(true);
      setError(null);

      console.log(`🔄 Fetching dashboard stats (force: ${forceRefresh})...`);

      // Clear cache if force refresh
      if (forceRefresh) {
        requestManager.clearCache(DASHBOARD_STATS_KEY);
      }

      // Use request manager to prevent duplicate calls
      const response = await requestManager.request(
        DASHBOARD_STATS_KEY,
        () => apiClient.getDashboardStats(),
        {
          cache: !forceRefresh,
          cacheDuration: 30000, // 30 seconds
        }
      );

      console.log("✅ Dashboard stats fetched successfully");
      setStats(response.data);
    } catch (error: any) {
      console.error("❌ Failed to fetch dashboard stats:", error);
      const errorMessage =
        error.response?.data?.message ||
        error.message ||
        "Failed to load dashboard data";
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  // Initial load
  useEffect(() => {
    fetchDashboardStats();
  }, [fetchDashboardStats]);

  const handleRefresh = useCallback(() => {
    fetchDashboardStats(true); // Force refresh
  }, [fetchDashboardStats]);

  if (error && !loading) {
    return (
      <SidebarProvider
        style={
          {
            "--sidebar-width": "calc(var(--spacing) * 72)",
            "--header-height": "calc(var(--spacing) * 12)",
          } as React.CSSProperties
        }
      >
        <AppSidebar variant="inset" />
        <SidebarInset>
          <SiteHeader />
          <div className="flex h-screen items-center justify-center">
            <Alert variant="destructive" className="max-w-md">
              <AlertDescription className="space-y-4">
                <div>{error}</div>
                <Button
                  onClick={handleRefresh}
                  className="w-full"
                  disabled={loading}
                >
                  <IconRefresh className="h-4 w-4 mr-2" />
                  Retry
                </Button>
              </AlertDescription>
            </Alert>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  return (
    <SidebarProvider
      style={
        {
          "--sidebar-width": "calc(var(--spacing) * 72)",
          "--header-height": "calc(var(--spacing) * 12)",
        } as React.CSSProperties
      }
    >
      <AppSidebar variant="inset" />
      <SidebarInset>
        <SiteHeader />
        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
              <div className="px-4 lg:px-6">
                <div className="flex items-center justify-between">
                  <div>
                    <h1 className="text-2xl font-bold">
                      Welcome back, {user?.first_name || user?.username}!
                    </h1>
                    <p className="text-muted-foreground">
                      Here's what's happening with your platform today.
                    </p>
                  </div>
                  <Button
                    onClick={handleRefresh}
                    variant="outline"
                    size="sm"
                    disabled={loading}
                  >
                    <IconRefresh
                      className={`h-4 w-4 mr-2 ${
                        loading ? "animate-spin" : ""
                      }`}
                    />
                    Refresh
                  </Button>
                </div>
              </div>

              {/* Stats Cards */}
              {loading ? <StatsCardsSkeleton /> : <StatsCards stats={stats} />}

              {/* Charts */}
              <div className="px-4 lg:px-6">
                {loading ? (
                  <Card>
                    <CardHeader>
                      <Skeleton className="h-6 w-48" />
                      <Skeleton className="h-4 w-32" />
                    </CardHeader>
                    <CardContent>
                      <Skeleton className="h-[250px] w-full" />
                    </CardContent>
                  </Card>
                ) : (
                  <ChartAreaInteractive />
                )}
              </div>

              {/* Recent Activities */}
              {loading ? (
                <RecentActivitiesSkeleton />
              ) : (
                <RecentActivities activities={stats?.recent_activities || []} />
              )}

              {/* System Health */}
              {loading ? (
                <SystemHealthSkeleton />
              ) : (
                <SystemHealthCard health={stats?.system_health} />
              )}
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

function StatsCards({ stats }: { stats: DashboardStats | null }) {
  if (!stats) return <StatsCardsSkeleton />;

  const cards = [
    {
      title: "Total Users",
      value: stats.total_users?.toLocaleString() || "0",
      icon: IconUsers,
      change: `+${stats.new_users_today || 0}`,
      changeType: "positive" as const,
      description: "New users today",
    },
    {
      title: "Total Posts",
      value: stats.total_posts?.toLocaleString() || "0",
      icon: IconMessages,
      change: `+${stats.new_posts_today || 0}`,
      changeType: "positive" as const,
      description: "Posts created today",
    },
    {
      title: "Pending Reports",
      value: stats.pending_reports?.toLocaleString() || "0",
      icon: IconReport,
      change: (stats.pending_reports || 0) > 10 ? "High" : "Normal",
      changeType: (stats.pending_reports || 0) > 10 ? "negative" : "neutral",
      description: "Require attention",
    },
    {
      title: "Active Users",
      value: stats.active_users?.toLocaleString() || "0",
      icon: IconHeart,
      change: `${(
        ((stats.active_users || 0) / (stats.total_users || 1)) *
        100
      ).toFixed(1)}%`,
      changeType: "positive" as const,
      description: "Of total users",
    },
  ];

  return (
    <div className="grid grid-cols-1 gap-4 px-4 lg:px-6 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
      {cards.map((card, index) => (
        <Card
          key={index}
          className="bg-gradient-to-t from-primary/5 to-card shadow-xs"
        >
          <CardHeader>
            <CardDescription>{card.title}</CardDescription>
            <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl flex items-center gap-2">
              <card.icon className="h-6 w-6 text-primary" />
              {card.value}
            </CardTitle>
            <div className="flex items-center gap-2">
              <Badge
                variant="outline"
                className={`
                ${
                  card.changeType === "positive"
                    ? "text-green-700 border-green-200 bg-green-50"
                    : ""
                }
                ${
                  card.changeType === "negative"
                    ? "text-red-700 border-red-200 bg-red-50"
                    : ""
                }
                ${
                  card.changeType === "neutral"
                    ? "text-gray-700 border-gray-200 bg-gray-50"
                    : ""
                }
              `}
              >
                {card.changeType === "positive" && (
                  <IconTrendingUp className="h-3 w-3" />
                )}
                {card.changeType === "negative" && (
                  <IconTrendingDown className="h-3 w-3" />
                )}
                {card.change}
              </Badge>
            </div>
          </CardHeader>
          <div className="px-6 pb-4">
            <p className="text-sm text-muted-foreground">{card.description}</p>
          </div>
        </Card>
      ))}
    </div>
  );
}

function RecentActivities({ activities }: { activities: any[] }) {
  return (
    <div className="px-4 lg:px-6">
      <Card>
        <CardHeader>
          <CardTitle>Recent Activities</CardTitle>
          <CardDescription>
            Latest admin activities and system events
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {activities.length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-4">
                No recent activities found
              </p>
            ) : (
              activities.slice(0, 5).map((activity, index) => (
                <div
                  key={activity.id || index}
                  className="flex items-start gap-3 pb-3 border-b last:border-b-0"
                >
                  <div className="w-2 h-2 bg-primary rounded-full mt-2" />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium">
                      {activity.description}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {new Date(activity.created_at).toLocaleString()}
                    </p>
                  </div>
                  <Badge variant="secondary" className="text-xs">
                    {activity.type}
                  </Badge>
                </div>
              ))
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function SystemHealthCard({ health }: { health: any }) {
  if (!health) return <SystemHealthSkeleton />;

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case "healthy":
      case "connected":
      case "active":
        return "text-green-600 bg-green-100";
      case "warning":
        return "text-yellow-600 bg-yellow-100";
      case "error":
      case "disconnected":
        return "text-red-600 bg-red-100";
      default:
        return "text-gray-600 bg-gray-100";
    }
  };

  return (
    <div className="px-4 lg:px-6">
      <Card>
        <CardHeader>
          <CardTitle>System Health</CardTitle>
          <CardDescription>
            Current system status and performance metrics
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="text-center">
              <div
                className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(
                  health.database_status
                )}`}
              >
                {health.database_status || "Unknown"}
              </div>
              <p className="text-sm text-muted-foreground mt-1">Database</p>
            </div>
            <div className="text-center">
              <div
                className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(
                  health.cache_status
                )}`}
              >
                {health.cache_status || "Unknown"}
              </div>
              <p className="text-sm text-muted-foreground mt-1">Cache</p>
            </div>
            <div className="text-center">
              <div className="text-lg font-semibold">
                {(health.memory_usage || 0).toFixed(1)}%
              </div>
              <p className="text-sm text-muted-foreground">Memory</p>
            </div>
            <div className="text-center">
              <div className="text-lg font-semibold">
                {(health.cpu_usage || 0).toFixed(1)}%
              </div>
              <p className="text-sm text-muted-foreground">CPU</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

// Skeleton components
function StatsCardsSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-4 px-4 lg:px-6 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
      {Array.from({ length: 4 }).map((_, i) => (
        <Card key={i}>
          <CardHeader>
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-8 w-20" />
            <Skeleton className="h-5 w-16" />
          </CardHeader>
        </Card>
      ))}
    </div>
  );
}

function RecentActivitiesSkeleton() {
  return (
    <div className="px-4 lg:px-6">
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-40" />
          <Skeleton className="h-4 w-60" />
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex items-start gap-3">
                <Skeleton className="w-2 h-2 rounded-full mt-2" />
                <div className="flex-1">
                  <Skeleton className="h-4 w-full mb-1" />
                  <Skeleton className="h-3 w-24" />
                </div>
                <Skeleton className="h-5 w-12" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function SystemHealthSkeleton() {
  return (
    <div className="px-4 lg:px-6">
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-32" />
          <Skeleton className="h-4 w-48" />
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="text-center">
                <Skeleton className="h-6 w-16 mx-auto mb-1" />
                <Skeleton className="h-4 w-12 mx-auto" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default withAuth(DashboardPage);
