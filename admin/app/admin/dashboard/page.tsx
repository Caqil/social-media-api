// app/admin/dashboard/page.tsx
"use client";

import { useEffect, useState } from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { ChartAreaInteractive } from "@/components/chart-area-interactive";
import { DataTable } from "@/components/data-table";
import { SectionCards } from "@/components/section-cards";
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
import { withAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import { DashboardStats } from "@/types/admin";
import {
  IconTrendingUp,
  IconTrendingDown,
  IconUsers,
  IconMessages,
  IconReport,
  IconHeart,
} from "@tabler/icons-react";

function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchDashboardStats = async () => {
      try {
        setLoading(true);
        const response = await apiClient.getDashboardStats();
        setStats(response.data);
      } catch (error: any) {
        console.error("Failed to fetch dashboard stats:", error);
        setError(
          error.response?.data?.message || "Failed to load dashboard data"
        );
      } finally {
        setLoading(false);
      }
    };

    fetchDashboardStats();
  }, []);

  if (error) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Alert variant="destructive" className="max-w-md">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
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
                  <ChartAreaInteractive
                    data={stats?.user_growth_chart || []}
                    title="User Growth"
                    description="New user registrations over time"
                  />
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
  if (!stats) return null;

  const cards = [
    {
      title: "Total Users",
      value: stats.total_users.toLocaleString(),
      icon: IconUsers,
      change: `+${stats.new_users_today}`,
      changeType: "positive" as const,
      description: "New users today",
    },
    {
      title: "Total Posts",
      value: stats.total_posts.toLocaleString(),
      icon: IconMessages,
      change: `+${stats.new_posts_today}`,
      changeType: "positive" as const,
      description: "Posts created today",
    },
    {
      title: "Pending Reports",
      value: stats.pending_reports.toLocaleString(),
      icon: IconReport,
      change: stats.pending_reports > 10 ? "High" : "Normal",
      changeType: stats.pending_reports > 10 ? "negative" : "neutral",
      description: "Require attention",
    },
    {
      title: "Active Users",
      value: stats.active_users.toLocaleString(),
      icon: IconHeart,
      change: `${((stats.active_users / stats.total_users) * 100).toFixed(1)}%`,
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
  if (!health) return null;

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
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
                {health.database_status}
              </div>
              <p className="text-sm text-muted-foreground mt-1">Database</p>
            </div>
            <div className="text-center">
              <div
                className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(
                  health.cache_status
                )}`}
              >
                {health.cache_status}
              </div>
              <p className="text-sm text-muted-foreground mt-1">Cache</p>
            </div>
            <div className="text-center">
              <div className="text-lg font-semibold">
                {health.memory_usage.toFixed(1)}%
              </div>
              <p className="text-sm text-muted-foreground">Memory</p>
            </div>
            <div className="text-center">
              <div className="text-lg font-semibold">
                {health.cpu_usage.toFixed(1)}%
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
