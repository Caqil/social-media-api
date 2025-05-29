"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { StatsCards } from "@/components/dashboard/stats-cards";
import { UserGrowthChart } from "@/components/dashboard/user-growth-chart";
import { EngagementChart } from "@/components/dashboard/engagement-chart";
import { RecentActivity } from "@/components/dashboard/recent-activity";
import { PopularContent } from "@/components/dashboard/popular-content";
import { SystemAlerts } from "@/components/dashboard/system-alerts";
import { Button } from "@/components/ui/button";
import { RefreshCw, Download, Calendar } from "lucide-react";
import { Badge } from "@/components/ui/badge";

export default function DashboardPage() {
  const [lastUpdated, setLastUpdated] = useState<Date>(new Date());
  const [isRefreshing, setIsRefreshing] = useState(false);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1000));
    setLastUpdated(new Date());
    setIsRefreshing(false);
  };

  const handleExport = () => {
    // Implement export functionality
    console.log("Exporting dashboard data...");
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-muted-foreground">
            Welcome back! Here's what's happening with your social media
            platform.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Badge variant="outline" className="text-xs">
            <Calendar className="mr-1 h-3 w-3" />
            Last updated: {lastUpdated.toLocaleTimeString()}
          </Badge>

          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={isRefreshing}
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`}
            />
            Refresh
          </Button>

          <Button variant="outline" size="sm" onClick={handleExport}>
            <Download className="mr-2 h-4 w-4" />
            Export
          </Button>
        </div>
      </div>

      {/* System Alerts */}
      <SystemAlerts />

      {/* Stats Cards */}
      <StatsCards />

      {/* Charts Section */}
      <div className="grid gap-6 md:grid-cols-2">
        <UserGrowthChart />
        <EngagementChart />
      </div>

      {/* Content Section */}
      <div className="grid gap-6 md:grid-cols-2">
        <RecentActivity />
        <PopularContent />
      </div>

      {/* Additional Metrics */}
      <div className="grid gap-6 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Server Status</CardTitle>
            <Badge variant="default" className="bg-green-100 text-green-800">
              Healthy
            </Badge>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>CPU Usage</span>
                <span>45%</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Memory Usage</span>
                <span>62%</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Disk Usage</span>
                <span>78%</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              API Performance
            </CardTitle>
            <Badge variant="outline">24h avg</Badge>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>Response Time</span>
                <span>145ms</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Success Rate</span>
                <span>99.8%</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Requests/min</span>
                <span>1,234</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Storage</CardTitle>
            <Badge variant="outline">Total</Badge>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>Media Files</span>
                <span>1.2TB</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Database</span>
                <span>450GB</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Backups</span>
                <span>800GB</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
