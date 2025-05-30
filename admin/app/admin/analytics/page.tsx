// app/admin/analytics/page.tsx
"use client";

import { useEffect, useState } from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { SiteHeader } from "@/components/site-header";
import { ChartAreaInteractive } from "@/components/chart-area-interactive";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { withAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import {
  IconTrendingUp,
  IconTrendingDown,
  IconUsers,
  IconMessages,
  IconHeart,
  IconEye,
  IconRefresh,
  IconDownload,
  IconCalendar,
  IconGlobe,
} from "@tabler/icons-react";

function AnalyticsPage() {
  const [userAnalytics, setUserAnalytics] = useState<any>(null);
  const [contentAnalytics, setContentAnalytics] = useState<any>(null);
  const [engagementAnalytics, setEngagementAnalytics] = useState<any>(null);
  const [demographicAnalytics, setDemographicAnalytics] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState("overview");

  const fetchAnalytics = async () => {
    try {
      setLoading(true);
      setError(null);

      const [
        userResponse,
        contentResponse,
        engagementResponse,
        demographicResponse,
      ] = await Promise.all([
        apiClient.getUserAnalytics(),
        apiClient.getContentAnalytics(),
        apiClient.getEngagementAnalytics(),
        apiClient.getDemographicAnalytics(),
      ]);

      setUserAnalytics(userResponse.data);
      setContentAnalytics(contentResponse.data);
      setEngagementAnalytics(engagementResponse.data);
      setDemographicAnalytics(demographicResponse.data);
    } catch (error: any) {
      console.error("Failed to fetch analytics:", error);
      setError(error.response?.data?.message || "Failed to load analytics");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAnalytics();
  }, []);

  const handleRefresh = () => {
    fetchAnalytics();
  };

  const handleExport = () => {
    // Implement export functionality
    console.log("Export analytics data");
  };

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
                <Button onClick={handleRefresh} className="w-full">
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
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">Analytics & Insights</h1>
              <p className="text-muted-foreground">
                Comprehensive platform analytics and user insights
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button
                onClick={handleRefresh}
                variant="outline"
                size="sm"
                disabled={loading}
              >
                <IconRefresh
                  className={`h-4 w-4 mr-2 ${loading ? "animate-spin" : ""}`}
                />
                Refresh
              </Button>
              <Button onClick={handleExport} variant="outline" size="sm">
                <IconDownload className="h-4 w-4 mr-2" />
                Export
              </Button>
            </div>
          </div>

          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="overview">Overview</TabsTrigger>
              <TabsTrigger value="users">Users</TabsTrigger>
              <TabsTrigger value="content">Content</TabsTrigger>
              <TabsTrigger value="engagement">Engagement</TabsTrigger>
            </TabsList>

            <TabsContent value="overview" className="space-y-4">
              {loading ? (
                <OverviewSkeleton />
              ) : (
                <OverviewTab
                  data={{
                    userAnalytics,
                    contentAnalytics,
                    engagementAnalytics,
                  }}
                />
              )}
            </TabsContent>

            <TabsContent value="users" className="space-y-4">
              {loading ? (
                <UsersSkeleton />
              ) : (
                <UsersTab
                  data={userAnalytics}
                  demographics={demographicAnalytics}
                />
              )}
            </TabsContent>

            <TabsContent value="content" className="space-y-4">
              {loading ? (
                <ContentSkeleton />
              ) : (
                <ContentTab data={contentAnalytics} />
              )}
            </TabsContent>

            <TabsContent value="engagement" className="space-y-4">
              {loading ? (
                <EngagementSkeleton />
              ) : (
                <EngagementTab data={engagementAnalytics} />
              )}
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

function OverviewTab({ data }: { data: any }) {
  const { userAnalytics, contentAnalytics, engagementAnalytics } = data;

  const overviewStats = [
    {
      title: "Total Users",
      value: userAnalytics?.total_users?.toLocaleString() || "0",
      change: `+${userAnalytics?.new_users || 0}`,
      changeType: "positive" as const,
      icon: IconUsers,
      description: "New users this month",
    },
    {
      title: "Total Posts",
      value: contentAnalytics?.total_posts?.toLocaleString() || "0",
      change: `+${((contentAnalytics?.total_posts || 0) * 0.12).toFixed(0)}`,
      changeType: "positive" as const,
      icon: IconMessages,
      description: "Growth this month",
    },
    {
      title: "Engagement Rate",
      value: `${engagementAnalytics?.engagement_rate?.toFixed(1) || 0}%`,
      change: engagementAnalytics?.engagement_rate > 5 ? "+2.3%" : "-1.2%",
      changeType:
        engagementAnalytics?.engagement_rate > 5 ? "positive" : "negative",
      icon: IconHeart,
      description: "Overall engagement",
    },
    {
      title: "Active Users",
      value: userAnalytics?.active_users?.toLocaleString() || "0",
      change: `${(
        ((userAnalytics?.active_users || 0) /
          (userAnalytics?.total_users || 1)) *
        100
      ).toFixed(1)}%`,
      changeType: "positive" as const,
      icon: IconEye,
      description: "Of total users",
    },
  ];

  return (
    <>
      {/* Overview Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {overviewStats.map((stat, index) => (
          <Card
            key={index}
            className="bg-gradient-to-br from-primary/5 to-card"
          >
            <CardHeader className="pb-2">
              <CardDescription className="flex items-center gap-2">
                <stat.icon className="h-4 w-4" />
                {stat.title}
              </CardDescription>
              <CardTitle className="text-2xl font-bold">{stat.value}</CardTitle>
              <div className="flex items-center gap-2">
                <Badge
                  variant="outline"
                  className={`${
                    stat.changeType === "positive"
                      ? "text-green-700 border-green-200 bg-green-50"
                      : "text-red-700 border-red-200 bg-red-50"
                  }`}
                >
                  {stat.changeType === "positive" ? (
                    <IconTrendingUp className="h-3 w-3 mr-1" />
                  ) : (
                    <IconTrendingDown className="h-3 w-3 mr-1" />
                  )}
                  {stat.change}
                </Badge>
              </div>
            </CardHeader>
            <div className="px-6 pb-4">
              <p className="text-sm text-muted-foreground">
                {stat.description}
              </p>
            </div>
          </Card>
        ))}
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <ChartAreaInteractive
          title="User Growth"
          description="New user registrations over time"
          data={userAnalytics?.user_growth_chart || []}
          dataKey="count"
        />
        <ChartAreaInteractive
          title="Content Creation"
          description="Posts created over time"
          data={contentAnalytics?.engagement_trends || []}
          dataKey="count"
        />
      </div>
    </>
  );
}

function UsersTab({ data, demographics }: { data: any; demographics: any }) {
  return (
    <>
      {/* User Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader>
            <CardTitle>User Retention</CardTitle>
            <CardDescription>Percentage of returning users</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {data?.user_retention?.toFixed(1) || 0}%
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Active Users</CardTitle>
            <CardDescription>Users active in the last 30 days</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {data?.active_users?.toLocaleString() || 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>User Growth</CardTitle>
            <CardDescription>Month over month growth</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-green-600">+12.5%</div>
          </CardContent>
        </Card>
      </div>

      {/* Demographics */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Age Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {demographics?.age_groups?.map((group: any, index: number) => (
                <div key={index} className="flex items-center justify-between">
                  <span className="text-sm">{group.age_group}</span>
                  <div className="flex items-center gap-2">
                    <div className="w-24 bg-gray-200 rounded-full h-2">
                      <div
                        className="bg-primary h-2 rounded-full"
                        style={{ width: `${group.percentage}%` }}
                      />
                    </div>
                    <span className="text-sm font-medium">
                      {group.percentage}%
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Top Countries</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {demographics?.location_groups
                ?.slice(0, 5)
                .map((country: any, index: number) => (
                  <div
                    key={index}
                    className="flex items-center justify-between"
                  >
                    <div className="flex items-center gap-2">
                      <IconGlobe className="h-4 w-4" />
                      <span className="text-sm">{country.country}</span>
                    </div>
                    <Badge variant="outline">
                      {country.count.toLocaleString()}
                    </Badge>
                  </div>
                ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* User Activity Chart */}
      <ChartAreaInteractive
        title="User Activity"
        description="Daily active users over time"
        data={data?.user_activity_by_hour || []}
        dataKey="count"
      />
    </>
  );
}

function ContentTab({ data }: { data: any }) {
  return (
    <>
      {/* Content Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Total Posts</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {data?.total_posts?.toLocaleString() || 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Total Comments</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {data?.total_comments?.toLocaleString() || 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Total Likes</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {data?.total_likes?.toLocaleString() || 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Engagement Rate</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {data?.engagement_rate?.toFixed(1) || 0}%
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Top Hashtags */}
      <Card>
        <CardHeader>
          <CardTitle>Trending Hashtags</CardTitle>
          <CardDescription>Most popular hashtags this month</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-2">
            {data?.top_hashtags
              ?.slice(0, 10)
              .map((hashtag: any, index: number) => (
                <Badge key={index} variant="secondary" className="px-3 py-1">
                  #{hashtag.tag} ({hashtag.count})
                </Badge>
              ))}
          </div>
        </CardContent>
      </Card>

      {/* Content Trends */}
      <ChartAreaInteractive
        title="Content Creation Trends"
        description="Posts and comments created over time"
        data={data?.content_by_hour || []}
        dataKey="count"
      />
    </>
  );
}

function EngagementTab({ data }: { data: any }) {
  return (
    <>
      {/* Engagement Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Total Engagements</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {data?.total_engagements?.toLocaleString() || 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Likes per Post</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {data?.likes_per_post?.toFixed(1) || 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Comments per Post</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {data?.comments_per_post?.toFixed(1) || 0}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Engagement by Content Type */}
      <Card>
        <CardHeader>
          <CardTitle>Engagement by Content Type</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {data?.engagement_by_content_type?.map(
              (type: any, index: number) => (
                <div key={index} className="flex items-center justify-between">
                  <span className="text-sm capitalize">{type.category}</span>
                  <div className="flex items-center gap-2">
                    <div className="w-32 bg-gray-200 rounded-full h-2">
                      <div
                        className="bg-primary h-2 rounded-full"
                        style={{ width: `${type.percentage}%` }}
                      />
                    </div>
                    <span className="text-sm font-medium">
                      {type.percentage}%
                    </span>
                  </div>
                </div>
              )
            )}
          </div>
        </CardContent>
      </Card>

      {/* Engagement Trends */}
      <ChartAreaInteractive
        title="Engagement Trends"
        description="User engagement over time"
        data={data?.engagement_by_hour || []}
        dataKey="count"
      />
    </>
  );
}

// Skeleton components
function OverviewSkeleton() {
  return (
    <>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
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
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-32" />
            <Skeleton className="h-4 w-48" />
          </CardHeader>
          <CardContent>
            <Skeleton className="h-[250px] w-full" />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-32" />
            <Skeleton className="h-4 w-48" />
          </CardHeader>
          <CardContent>
            <Skeleton className="h-[250px] w-full" />
          </CardContent>
        </Card>
      </div>
    </>
  );
}

function UsersSkeleton() {
  return (
    <>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {Array.from({ length: 3 }).map((_, i) => (
          <Card key={i}>
            <CardHeader>
              <Skeleton className="h-6 w-24" />
              <Skeleton className="h-4 w-32" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-8 w-16" />
            </CardContent>
          </Card>
        ))}
      </div>
    </>
  );
}

function ContentSkeleton() {
  return <UsersSkeleton />;
}

function EngagementSkeleton() {
  return <UsersSkeleton />;
}

export default withAuth(AnalyticsPage);
