"use client";

import { useEffect, useState, useCallback, useMemo } from "react";
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
  CardFooter,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { withAuth, useAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import { requestManager } from "@/lib/request-manager";
import { 
  DashboardStats, 
  AdminActivity, 
  SystemHealth,
  ChartData
} from "@/types/admin";
import {
  IconTrendingUp,
  IconTrendingDown,
  IconUsers,
  IconMessages,
  IconReport,
  IconHeart,
  IconRefresh,
  IconLoader2,
  IconAlertCircle,
  IconActivity,
  IconChartBar,
  IconUserPlus,
  IconFileUpload,
  IconEye,
  IconSettings
} from "@tabler/icons-react";
import { Button } from "@/components/ui/button";

// Define cache keys
const DASHBOARD_STATS_KEY = "dashboard-stats";

// Type for API errors
interface ApiError {
  response?: {
    data?: {
      message?: string;
    };
  };
  message?: string;
}

// Custom hook for fetching data with requestManager
function useDashboardData<T>(
  cacheKey: string,
  fetchFn: () => Promise<any>,
  dependencies: any[] = []
) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(
    async (forceRefresh = false) => {
      try {
        setLoading(true);
        setError(null);

        console.log(`🔄 Fetching ${cacheKey}${forceRefresh ? ' (force refresh)' : ''}...`);

        if (forceRefresh) {
          requestManager.clearCache(cacheKey);
        }

        const response = await requestManager.request(
          cacheKey,
          fetchFn,
          {
            cache: !forceRefresh,
            cacheDuration: 30000, // 30 seconds
          }
        );

        console.log(`✅ ${cacheKey} fetched successfully`);
        setData(response.data);
      } catch (err) {
        console.error(`❌ Failed to fetch ${cacheKey}:`, err);
        
        const error = err as ApiError;
        const errorMessage =
          error.response?.data?.message ||
          error.message ||
          `Failed to load ${cacheKey}`;
        
        setError(errorMessage);
      } finally {
        setLoading(false);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [cacheKey, ...dependencies]
  );

  useEffect(() => {
    fetchData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

function DashboardPage() {
  const { user } = useAuth();
  
  // Fetch main dashboard stats
  const {
    data: stats,
    loading: statsLoading,
    error: statsError,
    refetch: refetchStats
  } = useDashboardData<DashboardStats>(
    DASHBOARD_STATS_KEY,
    () => apiClient.getDashboardStats()
  );

  // Handle refresh for all data
  const handleRefresh = useCallback(() => {
    refetchStats(true);
  }, [refetchStats]);

  // Render error state if main stats has an error
  if (statsError && !statsLoading) {
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
          <div className="flex h-screen items-center justify-center p-6">
            <Alert variant="destructive" className="max-w-md">
              <IconAlertCircle className="h-5 w-5 mr-2" />
              <AlertDescription className="space-y-4">
                <div className="text-lg font-medium">Failed to load dashboard data</div>
                <div>{statsError}</div>
                <Button
                  onClick={handleRefresh}
                  className="w-full"
                  disabled={statsLoading}
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
        
        {/* Loading progress bar */}
        {statsLoading && (
          <div className="fixed top-0 left-0 right-0 z-50">
            <div 
              className="h-1 bg-primary transition-all duration-300 ease-in-out" 
              style={{ width: statsLoading ? '70%' : '100%' }}
            />
          </div>
        )}
        
        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
              {/* Welcome Message with time-sensitive greeting */}
              <div className="px-4 lg:px-6">
                <div className="flex items-center justify-between">
                  <div>
                    <h1 className="text-2xl font-bold">
                      {getTimeBasedGreeting()}, {user?.first_name || user?.username}!
                    </h1>
                    <p className="text-muted-foreground">
                      Here's what's happening with your platform today.
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      onClick={handleRefresh}
                      variant="outline"
                      size="sm"
                      disabled={statsLoading}
                    >
                      {statsLoading ? (
                        <>
                          <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                          Updating...
                        </>
                      ) : (
                        <>
                          <IconRefresh className="h-4 w-4 mr-2" />
                          Refresh
                        </>
                      )}
                    </Button>
                    <Button size="sm" variant="default">
                      <IconEye className="h-4 w-4 mr-2" />
                      Live View
                    </Button>
                  </div>
                </div>
              </div>

              {/* Stats Cards */}
              {statsLoading ? <StatsCardsSkeleton /> : <StatsCards stats={stats} />}

              {/* Activity Overview */}
              <div className="px-4 lg:px-6">
                <h2 className="text-xl font-semibold mb-4">Activity Overview</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {/* User Growth Chart */}
                  <Card className="overflow-hidden">
                    <CardHeader className="pb-0">
                      <div className="flex items-center justify-between">
                        <div>
                          <CardTitle>User Growth</CardTitle>
                          <CardDescription>New user signups over time</CardDescription>
                        </div>
                        <Badge variant="outline" className="text-green-700 border-green-200 bg-green-50">
                          <IconTrendingUp className="h-3 w-3 mr-1" />
                          +{stats?.new_users_today || 0} today
                        </Badge>
                      </div>
                    </CardHeader>
                    <CardContent>
                      {statsLoading ? (
                        <Skeleton className="h-[250px] w-full" />
                      ) : (
                        <ChartAreaInteractive 
                          data={stats?.user_growth_chart || []} 
                          height={250}
                          dataKey="count"
                          title="User Growth"
                          description="Daily new user registrations"
                        />
                      )}
                    </CardContent>
                  </Card>
                  
                  {/* Content Chart */}
                  <Card className="overflow-hidden">
                    <CardHeader className="pb-0">
                      <div className="flex items-center justify-between">
                        <div>
                          <CardTitle>Content Activity</CardTitle>
                          <CardDescription>Posts and engagement</CardDescription>
                        </div>
                        <Badge variant="outline" className="text-blue-700 border-blue-200 bg-blue-50">
                          <IconActivity className="h-3 w-3 mr-1" />
                          {stats?.new_posts_today || 0} new posts
                        </Badge>
                      </div>
                    </CardHeader>
                    <CardContent>
                      {statsLoading ? (
                        <Skeleton className="h-[250px] w-full" />
                      ) : (
                        <ChartAreaInteractive 
                          data={stats?.post_growth_chart || []} 
                          height={250}
                          dataKey="count"
                          title="Content Activity"
                          description="Post metrics over time"
                        />
                      )}
                    </CardContent>
                  </Card>
                </div>
              </div>

              {/* Activity + System Health Cards */}
              <div className="px-4 lg:px-6">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  {/* Recent Activities */}
                  <div className="md:col-span-2">
                    {statsLoading ? (
                      <RecentActivitiesSkeleton />
                    ) : (
                      <RecentActivities activities={stats?.recent_activities || []} />
                    )}
                  </div>
                  
                  {/* System Health */}
                  <div>
                    {statsLoading ? (
                      <SystemHealthSkeleton />
                    ) : (
                      <SystemHealthCard health={stats?.system_health} />
                    )}
                  </div>
                </div>
              </div>
              
              {/* At-Risk Items */}
              <div className="px-4 lg:px-6">
                <Card>
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <CardTitle>At-Risk Items</CardTitle>
                      {!statsLoading && stats?.pending_reports && stats.pending_reports > 0 && (
                        <Badge variant="destructive">
                          Requires Attention
                        </Badge>
                      )}
                    </div>
                    <CardDescription>
                      Items requiring moderation or attention
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    {statsLoading ? (
                      <div className="space-y-3">
                        {Array.from({ length: 3 }).map((_, i) => (
                          <Skeleton key={i} className="h-12 w-full" />
                        ))}
                      </div>
                    ) : stats?.pending_reports && stats.pending_reports > 0 ? (
                      <div className="space-y-3">
                        <div className="flex items-center justify-between p-3 bg-red-50 border border-red-100 rounded-md">
                          <div className="flex items-center gap-3">
                            <IconReport className="h-5 w-5 text-red-600" />
                            <div>
                              <p className="font-medium">Pending Reports</p>
                              <p className="text-sm text-muted-foreground">Requires moderation</p>
                            </div>
                          </div>
                          <Badge variant="destructive">{stats.pending_reports}</Badge>
                        </div>
                        
                        {stats.suspended_users > 0 && (
                          <div className="flex items-center justify-between p-3 bg-amber-50 border border-amber-100 rounded-md">
                            <div className="flex items-center gap-3">
                              <IconUsers className="h-5 w-5 text-amber-600" />
                              <div>
                                <p className="font-medium">Suspended Users</p>
                                <p className="text-sm text-muted-foreground">Review required</p>
                              </div>
                            </div>
                            <Badge variant="outline" className="bg-amber-100 text-amber-800">
                              {stats.suspended_users}
                            </Badge>
                          </div>
                        )}
                      </div>
                    ) : (
                      <div className="flex flex-col items-center justify-center py-6 text-center">
                        <IconChartBar className="h-12 w-12 text-muted-foreground/50 mb-3" />
                        <p className="text-muted-foreground">No pending items requiring attention</p>
                      </div>
                    )}
                  </CardContent>
                  <CardFooter className="bg-muted/30 border-t">
                    <Button variant="outline" className="w-full">
                      View All Moderation Tasks
                    </Button>
                  </CardFooter>
                </Card>
              </div>
              
              {/* Top Hashtags Section */}
              {stats?.top_hashtags && stats.top_hashtags.length > 0 && (
                <div className="px-4 lg:px-6">
                  <Card>
                    <CardHeader>
                      <CardTitle>Trending Hashtags</CardTitle>
                      <CardDescription>
                        Most popular hashtags being used across the platform
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="flex flex-wrap gap-2">
                        {stats.top_hashtags.map((hashtag, index) => (
                          <Badge 
                            key={index}
                            variant="secondary" 
                            className={`text-sm py-1.5 px-3 ${
                              index === 0 ? 'bg-blue-100 text-blue-800' :
                              index === 1 ? 'bg-purple-100 text-purple-800' :
                              index === 2 ? 'bg-green-100 text-green-800' :
                              ''
                            }`}
                          >
                            #{hashtag.tag}
                            <span className="ml-1 opacity-70">{hashtag.count}</span>
                          </Badge>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                </div>
              )}
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

function getTimeBasedGreeting(): string {
  const hour = new Date().getHours();
  if (hour < 12) return "Good morning";
  if (hour < 18) return "Good afternoon";
  return "Good evening";
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
      bgClass: "from-blue-50 to-blue-100/30 dark:from-blue-900/20 dark:to-blue-800/10",
      iconClass: "text-blue-600",
    },
    {
      title: "Total Posts",
      value: stats.total_posts?.toLocaleString() || "0",
      icon: IconMessages,
      change: `+${stats.new_posts_today || 0}`,
      changeType: "positive" as const,
      description: "Posts created today",
      bgClass: "from-purple-50 to-purple-100/30 dark:from-purple-900/20 dark:to-purple-800/10",
      iconClass: "text-purple-600",
    },
    {
      title: "Pending Reports",
      value: stats.pending_reports?.toLocaleString() || "0",
      icon: IconReport,
      change: (stats.pending_reports || 0) > 10 ? "High" : "Normal",
      changeType: (stats.pending_reports || 0) > 10 ? "negative" : "neutral",
      description: "Require attention",
      bgClass: "from-amber-50 to-amber-100/30 dark:from-amber-900/20 dark:to-amber-800/10",
      iconClass: "text-amber-600",
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
      bgClass: "from-green-50 to-green-100/30 dark:from-green-900/20 dark:to-green-800/10",
      iconClass: "text-green-600",
    },
  ];

  return (
    <div className="grid grid-cols-1 gap-4 px-4 lg:px-6 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
      {cards.map((card, index) => (
        <Card
          key={index}
          className={`bg-gradient-to-t ${card.bgClass} shadow-sm hover:shadow transition-all duration-200`}
        >
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardDescription>{card.title}</CardDescription>
              <div className={`p-2 rounded-full bg-white/80 ${card.iconClass}`}>
                <card.icon className="h-5 w-5" />
              </div>
            </div>
            <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
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
                  <IconTrendingUp className="h-3 w-3 mr-1" />
                )}
                {card.changeType === "negative" && (
                  <IconTrendingDown className="h-3 w-3 mr-1" />
                )}
                {card.change}
              </Badge>
            </div>
          </CardHeader>
          <CardFooter className="pt-0">
            <p className="text-sm text-muted-foreground">{card.description}</p>
          </CardFooter>
        </Card>
      ))}
    </div>
  );
}

function RecentActivities({ activities }: { activities: AdminActivity[] }) {
  return (
    <Card className="h-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Recent Activities</CardTitle>
            <CardDescription>
              Latest admin activities and system events
            </CardDescription>
          </div>
          <Button variant="ghost" size="sm">
            View All
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {activities.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-6 text-center">
              <IconActivity className="h-12 w-12 text-muted-foreground/50 mb-3" />
              <p className="text-muted-foreground">No recent activities found</p>
            </div>
          ) : (
            activities.slice(0, 5).map((activity, index) => (
              <div
                key={activity.id || index}
                className="flex items-start gap-3 pb-3 border-b last:border-b-0"
              >
                <div className="flex h-8 w-8 items-center justify-center rounded-full bg-muted">
                  <ActivityTypeIcon type={activity.type} />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium">
                    {activity.description}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {formatActivityTime(activity.created_at)}
                  </p>
                </div>
                <Badge 
                  variant="secondary" 
                  className="text-xs capitalize"
                  style={{ backgroundColor: getActivityTypeColor(activity.type) }}
                >
                  {activity.type.replace(/_/g, ' ')}
                </Badge>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function ActivityTypeIcon({ type }: { type: string }) {
  switch (type) {
    case 'user_login':
    case 'user_logout':
    case 'user_status_update':
      return <IconUsers className="h-4 w-4" />;
    case 'post_hidden':
    case 'post_deletion':
      return <IconMessages className="h-4 w-4" />;
    case 'config_update':
    case 'feature_toggle':
      return <IconSettings className="h-4 w-4" />;
    case 'user_verification':
      return <IconUserPlus className="h-4 w-4" />;
    case 'media_deletion':
    case 'media_moderation':
      return <IconFileUpload className="h-4 w-4" />;
    default:
      return <IconActivity className="h-4 w-4" />;
  }
}

function getActivityTypeColor(type: string): string {
  if (type.includes('user')) return 'rgba(59, 130, 246, 0.1)'; // blue
  if (type.includes('post')) return 'rgba(124, 58, 237, 0.1)'; // purple
  if (type.includes('config')) return 'rgba(245, 158, 11, 0.1)'; // amber
  if (type.includes('media')) return 'rgba(16, 185, 129, 0.1)'; // green
  if (type.includes('report')) return 'rgba(239, 68, 68, 0.1)'; // red
  return 'rgba(156, 163, 175, 0.1)'; // gray
}

function formatActivityTime(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / (1000 * 60));
  
  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins} minute${diffMins === 1 ? '' : 's'} ago`;
  
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
  
  const diffDays = Math.floor(diffHours / 24);
  if (diffDays < 7) return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;
  
  return date.toLocaleDateString();
}

function SystemHealthCard({ health }: { health: SystemHealth | undefined }) {
  if (!health) return <SystemHealthSkeleton />;

  const getStatusColor = (status: string | undefined) => {
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

  const getStatusBg = (value: number) => {
    if (value < 30) return "bg-green-500";
    if (value < 60) return "bg-amber-500";
    return "bg-red-500";
  };

  return (
    <Card className="h-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>System Health</CardTitle>
          <Badge
            className={`${
              health.status === "healthy"
                ? "bg-green-100 text-green-800"
                : health.status === "warning"
                ? "bg-amber-100 text-amber-800"
                : "bg-red-100 text-red-800"
            }`}
          >
            {health.status}
          </Badge>
        </div>
        <CardDescription>
          System status and performance metrics
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="text-center p-3 bg-muted/40 rounded-lg">
              <div
                className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(
                  health.database_status
                )}`}
              >
                {health.database_status || "Unknown"}
              </div>
              <p className="text-sm text-muted-foreground mt-1">Database</p>
            </div>
            <div className="text-center p-3 bg-muted/40 rounded-lg">
              <div
                className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(
                  health.cache_status
                )}`}
              >
                {health.cache_status || "Unknown"}
              </div>
              <p className="text-sm text-muted-foreground mt-1">Cache</p>
            </div>
          </div>
          
          <div className="space-y-3">
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm">CPU</span>
                <span className="text-sm font-medium">{health.cpu_usage.toFixed(1)}%</span>
              </div>
              <div className="h-2 w-full bg-muted rounded-full overflow-hidden">
                <div 
                  className={`h-full ${getStatusBg(health.cpu_usage)}`} 
                  style={{ width: `${health.cpu_usage}%` }}
                />
              </div>
            </div>
            
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm">Memory</span>
                <span className="text-sm font-medium">{health.memory_usage.toFixed(1)}%</span>
              </div>
              <div className="h-2 w-full bg-muted rounded-full overflow-hidden">
                <div 
                  className={`h-full ${getStatusBg(health.memory_usage)}`} 
                  style={{ width: `${health.memory_usage}%` }}
                />
              </div>
            </div>
            
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm">Disk</span>
                <span className="text-sm font-medium">{health.disk_usage.toFixed(1)}%</span>
              </div>
              <div className="h-2 w-full bg-muted rounded-full overflow-hidden">
                <div 
                  className={`h-full ${getStatusBg(health.disk_usage)}`} 
                  style={{ width: `${health.disk_usage}%` }}
                />
              </div>
            </div>
          </div>
          
          {health.alerts && health.alerts.length > 0 && (
            <div className="mt-4">
              <h4 className="text-sm font-medium mb-2">System Alerts</h4>
              {health.alerts.slice(0, 2).map((alert, index) => (
                <div 
                  key={index}
                  className={`p-2 rounded text-xs mb-2 ${
                    alert.level === 'critical' ? 'bg-red-100 text-red-800' :
                    alert.level === 'error' ? 'bg-red-50 text-red-700' :
                    alert.level === 'warning' ? 'bg-amber-50 text-amber-700' :
                    'bg-blue-50 text-blue-700'
                  }`}
                >
                  {alert.message}
                </div>
              ))}
            </div>
          )}
        </div>
      </CardContent>
      <CardFooter className="pt-0">
        <Button variant="outline" size="sm" className="w-full">
          View System Details
        </Button>
      </CardFooter>
    </Card>
  );
}

// Skeleton components
function StatsCardsSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-4 px-4 lg:px-6 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
      {Array.from({ length: 4 }).map((_, i) => (
        <Card key={i} className="overflow-hidden">
          <CardHeader>
            <div className="flex items-center justify-between">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-8 w-8 rounded-full" />
            </div>
            <Skeleton className="h-8 w-20 mt-2" />
            <Skeleton className="h-5 w-16 mt-1" />
          </CardHeader>
          <CardFooter className="pt-0">
            <Skeleton className="h-4 w-32" />
          </CardFooter>
        </Card>
      ))}
    </div>
  );
}

function RecentActivitiesSkeleton() {
  return (
    <Card className="h-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <Skeleton className="h-6 w-40" />
            <Skeleton className="h-4 w-60 mt-1" />
          </div>
          <Skeleton className="h-8 w-20" />
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="flex items-start gap-3 pb-3">
              <Skeleton className="h-8 w-8 rounded-full" />
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
  );
}

function SystemHealthSkeleton() {
  return (
    <Card className="h-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <Skeleton className="h-6 w-32" />
          <Skeleton className="h-5 w-16" />
        </div>
        <Skeleton className="h-4 w-48 mt-1" />
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            {Array.from({ length: 2 }).map((_, i) => (
              <div key={i} className="text-center p-3 bg-muted/40 rounded-lg">
                <Skeleton className="h-6 w-16 mx-auto" />
                <Skeleton className="h-4 w-12 mx-auto mt-1" />
              </div>
            ))}
          </div>
          
          <div className="space-y-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i}>
                <div className="flex items-center justify-between mb-1">
                  <Skeleton className="h-4 w-12" />
                  <Skeleton className="h-4 w-10" />
                </div>
                <Skeleton className="h-2 w-full rounded-full" />
              </div>
            ))}
          </div>
        </div>
      </CardContent>
      <CardFooter className="pt-0">
        <Skeleton className="h-8 w-full" />
      </CardFooter>
    </Card>
  );
}

export default withAuth(DashboardPage);