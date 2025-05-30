// app/admin/system/page.tsx
"use client";

import { useEffect, useState } from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { SiteHeader } from "@/components/site-header";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { withAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import { useAuth } from "@/contexts/auth-context";
import { UserRole } from "@/types/admin";
import {
  IconServer,
  IconDatabase,
  IconActivity,
  IconCpu,
  IconWifi,
  IconRefresh,
  IconTrash,
  IconSettings,
  IconDownload,
  IconUpload,
  IconTools,
  IconAlertTriangle,
  IconCheck,
  IconX,
  IconClock,
  IconShieldCheck,
} from "@tabler/icons-react";
import { HardDriveIcon, MemoryStick } from "lucide-react";

function SystemPage() {
  const { user } = useAuth();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // System data
  const [systemHealth, setSystemHealth] = useState<any>(null);
  const [systemInfo, setSystemInfo] = useState<any>(null);
  const [performanceMetrics, setPerformanceMetrics] = useState<any>(null);
  const [databaseStats, setDatabaseStats] = useState<any>(null);
  const [cacheStats, setCacheStats] = useState<any>(null);
  const [systemLogs, setSystemLogs] = useState<any[]>([]);
  const [backups, setBackups] = useState<any[]>([]);

  // Dialog states
  const [showMaintenanceDialog, setShowMaintenanceDialog] = useState(false);
  const [showBackupDialog, setShowBackupDialog] = useState(false);
  const [showRestoreDialog, setShowRestoreDialog] = useState(false);
  const [showCacheDialog, setShowCacheDialog] = useState(false);

  // Maintenance settings
  const [maintenanceMessage, setMaintenanceMessage] = useState("");
  const [maintenanceDuration, setMaintenanceDuration] = useState("");

  // Backup settings
  const [backupType, setBackupType] = useState("full");
  const [selectedCollections, setSelectedCollections] = useState<string[]>([]);
  const [selectedBackup, setSelectedBackup] = useState("");

  // Cache settings
  const [cacheType, setCacheType] = useState("");
  const [cacheAction, setCacheAction] = useState("");

  // Active tab
  const [activeTab, setActiveTab] = useState("health");

  // Real-time updates
  const [autoRefresh, setAutoRefresh] = useState(true);

  // Check if user is super admin
  const isSuperAdmin = user?.role === UserRole.SUPER_ADMIN;

  // Fetch system data
  const fetchSystemData = async () => {
    try {
      setLoading(true);
      const [healthResponse, infoResponse, metricsResponse] = await Promise.all(
        [
          apiClient.getSystemHealth(),
          apiClient.getSystemInfo(),
          apiClient.getPerformanceMetrics(),
        ]
      );

      setSystemHealth(healthResponse.data);
      setSystemInfo(infoResponse.data);
      setPerformanceMetrics(metricsResponse.data);
    } catch (error: any) {
      console.error("Failed to fetch system data:", error);
      setError(error.response?.data?.message || "Failed to load system data");
    } finally {
      setLoading(false);
    }
  };

  // Fetch database stats
  const fetchDatabaseStats = async () => {
    try {
      const response = await apiClient.getDatabaseStats();
      setDatabaseStats(response.data);
    } catch (error) {
      console.error("Failed to fetch database stats:", error);
    }
  };

  // Fetch cache stats
  const fetchCacheStats = async () => {
    try {
      const response = await apiClient.getCacheStats();
      setCacheStats(response.data);
    } catch (error) {
      console.error("Failed to fetch cache stats:", error);
    }
  };

  // Fetch system logs
  const fetchSystemLogs = async () => {
    try {
      const response = await apiClient.getSystemLogs({ limit: 50 });
      setSystemLogs(response.data);
    } catch (error) {
      console.error("Failed to fetch system logs:", error);
    }
  };

  // Fetch backups
  const fetchBackups = async () => {
    try {
      if (isSuperAdmin) {
        const response = await apiClient.getDatabaseBackups();
        setBackups(response.data);
      }
    } catch (error) {
      console.error("Failed to fetch backups:", error);
    }
  };

  useEffect(() => {
    fetchSystemData();

    if (activeTab === "database") {
      fetchDatabaseStats();
    } else if (activeTab === "cache") {
      fetchCacheStats();
    } else if (activeTab === "logs") {
      fetchSystemLogs();
    } else if (activeTab === "backups") {
      fetchBackups();
    }
  }, [activeTab]);

  // Auto-refresh functionality
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      if (activeTab === "health") {
        fetchSystemData();
      } else if (activeTab === "database") {
        fetchDatabaseStats();
      } else if (activeTab === "cache") {
        fetchCacheStats();
      }
    }, 30000); // Refresh every 30 seconds

    return () => clearInterval(interval);
  }, [activeTab, autoRefresh]);

  // Handle maintenance mode
  const handleMaintenanceMode = async (enable: boolean) => {
    try {
      if (enable) {
        await apiClient.enableMaintenanceMode({
          message: maintenanceMessage,
          duration: maintenanceDuration,
        });
      } else {
        await apiClient.disableMaintenanceMode();
      }

      setShowMaintenanceDialog(false);
      setMaintenanceMessage("");
      setMaintenanceDuration("");
      fetchSystemData();
    } catch (error: any) {
      setError(
        error.response?.data?.message || "Failed to update maintenance mode"
      );
    }
  };

  // Handle cache operations
  const handleCacheOperation = async () => {
    try {
      if (cacheAction === "clear") {
        await apiClient.clearCache(cacheType);
      } else if (cacheAction === "warm") {
        await apiClient.warmCache(cacheType);
      }

      setShowCacheDialog(false);
      setCacheType("");
      setCacheAction("");
      fetchCacheStats();
    } catch (error: any) {
      setError(error.response?.data?.message || "Cache operation failed");
    }
  };

  // Handle database backup
  const handleDatabaseBackup = async () => {
    try {
      await apiClient.backupDatabase({
        backup_type: backupType,
        collections: selectedCollections,
      });

      setShowBackupDialog(false);
      setBackupType("full");
      setSelectedCollections([]);
      fetchBackups();
    } catch (error: any) {
      setError(error.response?.data?.message || "Backup failed");
    }
  };

  // Handle database restore
  const handleDatabaseRestore = async () => {
    try {
      await apiClient.restoreDatabase({
        backup_id: selectedBackup,
        collections: selectedCollections,
      });

      setShowRestoreDialog(false);
      setSelectedBackup("");
      setSelectedCollections([]);
    } catch (error: any) {
      setError(error.response?.data?.message || "Restore failed");
    }
  };

  // Handle database optimization
  const handleDatabaseOptimization = async () => {
    try {
      await apiClient.optimizeDatabase();
      fetchDatabaseStats();
    } catch (error: any) {
      setError(error.response?.data?.message || "Database optimization failed");
    }
  };

  // Get status color
  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case "healthy":
      case "connected":
      case "active":
      case "online":
        return "text-green-600 bg-green-100";
      case "warning":
      case "degraded":
        return "text-yellow-600 bg-yellow-100";
      case "error":
      case "offline":
      case "failed":
        return "text-red-600 bg-red-100";
      default:
        return "text-gray-600 bg-gray-100";
    }
  };

  // Get metric color based on value
  const getMetricColor = (
    value: number,
    thresholds: { warning: number; critical: number }
  ) => {
    if (value >= thresholds.critical) return "text-red-600";
    if (value >= thresholds.warning) return "text-yellow-600";
    return "text-green-600";
  };

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
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">System Management</h1>
              <p className="text-muted-foreground">
                Monitor system health and manage infrastructure
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setAutoRefresh(!autoRefresh)}
              >
                <IconRefresh
                  className={`h-4 w-4 mr-2 ${
                    autoRefresh ? "animate-spin" : ""
                  }`}
                />
                {autoRefresh ? "Auto" : "Manual"}
              </Button>
              {isSuperAdmin && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowMaintenanceDialog(true)}
                >
                  <IconSettings className="h-4 w-4 mr-2" />
                  Maintenance
                </Button>
              )}
            </div>
          </div>

          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="grid w-full grid-cols-6">
              <TabsTrigger value="health">Health</TabsTrigger>
              <TabsTrigger value="performance">Performance</TabsTrigger>
              <TabsTrigger value="database">Database</TabsTrigger>
              <TabsTrigger value="cache">Cache</TabsTrigger>
              <TabsTrigger value="logs">Logs</TabsTrigger>
              {isSuperAdmin && (
                <TabsTrigger value="backups">Backups</TabsTrigger>
              )}
            </TabsList>

            {/* System Health Tab */}
            <TabsContent value="health" className="space-y-4">
              {loading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                  {Array.from({ length: 4 }).map((_, i) => (
                    <Card key={i}>
                      <CardHeader>
                        <Skeleton className="h-4 w-24" />
                        <Skeleton className="h-8 w-16" />
                      </CardHeader>
                    </Card>
                  ))}
                </div>
              ) : systemHealth ? (
                <>
                  {/* Health Overview */}
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>System Status</CardDescription>
                        <div
                          className={`inline-flex px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(
                            systemHealth.status
                          )}`}
                        >
                          <IconServer className="h-4 w-4 mr-2" />
                          {systemHealth.status}
                        </div>
                      </CardHeader>
                    </Card>

                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Database</CardDescription>
                        <div
                          className={`inline-flex px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(
                            systemHealth.database_status
                          )}`}
                        >
                          <IconDatabase className="h-4 w-4 mr-2" />
                          {systemHealth.database_status}
                        </div>
                      </CardHeader>
                    </Card>

                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Cache</CardDescription>
                        <div
                          className={`inline-flex px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(
                            systemHealth.cache_status
                          )}`}
                        >
                          <MemoryStick className="h-4 w-4 mr-2" />
                          {systemHealth.cache_status}
                        </div>
                      </CardHeader>
                    </Card>

                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Storage</CardDescription>
                        <div
                          className={`inline-flex px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(
                            systemHealth.storage_status
                          )}`}
                        >
                          <HardDriveIcon className="h-4 w-4 mr-2" />
                          {systemHealth.storage_status}
                        </div>
                      </CardHeader>
                    </Card>
                  </div>

                  {/* System Metrics */}
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Response Time</CardDescription>
                        <CardTitle className="text-2xl">
                          {systemHealth.response_time}ms
                        </CardTitle>
                      </CardHeader>
                    </Card>

                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Memory Usage</CardDescription>
                        <CardTitle
                          className={`text-2xl ${getMetricColor(
                            systemHealth.memory_usage,
                            { warning: 70, critical: 90 }
                          )}`}
                        >
                          {systemHealth.memory_usage.toFixed(1)}%
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="pt-0">
                        <Progress
                          value={systemHealth.memory_usage}
                          className="w-full"
                        />
                      </CardContent>
                    </Card>

                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>CPU Usage</CardDescription>
                        <CardTitle
                          className={`text-2xl ${getMetricColor(
                            systemHealth.cpu_usage,
                            { warning: 70, critical: 90 }
                          )}`}
                        >
                          {systemHealth.cpu_usage.toFixed(1)}%
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="pt-0">
                        <Progress
                          value={systemHealth.cpu_usage}
                          className="w-full"
                        />
                      </CardContent>
                    </Card>

                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Disk Usage</CardDescription>
                        <CardTitle
                          className={`text-2xl ${getMetricColor(
                            systemHealth.disk_usage,
                            { warning: 80, critical: 95 }
                          )}`}
                        >
                          {systemHealth.disk_usage.toFixed(1)}%
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="pt-0">
                        <Progress
                          value={systemHealth.disk_usage}
                          className="w-full"
                        />
                      </CardContent>
                    </Card>
                  </div>

                  {/* System Alerts */}
                  {systemHealth.alerts && systemHealth.alerts.length > 0 && (
                    <Card>
                      <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                          <IconAlertTriangle className="h-5 w-5" />
                          System Alerts
                        </CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div className="space-y-2">
                          {systemHealth.alerts.map(
                            (alert: any, index: number) => (
                              <Alert
                                key={index}
                                variant={
                                  alert.level === "critical" ||
                                  alert.level === "error"
                                    ? "destructive"
                                    : "default"
                                }
                              >
                                <AlertDescription className="flex items-center justify-between">
                                  <span>{alert.message}</span>
                                  <span className="text-xs text-muted-foreground">
                                    {new Date(alert.timestamp).toLocaleString()}
                                  </span>
                                </AlertDescription>
                              </Alert>
                            )
                          )}
                        </div>
                      </CardContent>
                    </Card>
                  )}
                </>
              ) : null}
            </TabsContent>

            {/* Performance Tab */}
            <TabsContent value="performance" className="space-y-4">
              {performanceMetrics && (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Request Metrics</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div className="flex justify-between">
                        <span>Requests per minute</span>
                        <span className="font-medium">
                          {performanceMetrics.requests_per_minute}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span>Average response time</span>
                        <span className="font-medium">
                          {performanceMetrics.avg_response_time}ms
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span>Error rate</span>
                        <span className="font-medium">
                          {performanceMetrics.error_rate}%
                        </span>
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle>Resource Usage</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div className="flex justify-between">
                        <span>Active connections</span>
                        <span className="font-medium">
                          {performanceMetrics.active_connections}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span>Memory allocated</span>
                        <span className="font-medium">
                          {performanceMetrics.memory_allocated}MB
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span>Goroutines</span>
                        <span className="font-medium">
                          {performanceMetrics.goroutines}
                        </span>
                      </div>
                    </CardContent>
                  </Card>
                </div>
              )}
            </TabsContent>

            {/* Database Tab */}
            <TabsContent value="database" className="space-y-4">
              <div className="flex items-center gap-2 mb-4">
                <Button
                  variant="outline"
                  onClick={fetchDatabaseStats}
                  size="sm"
                >
                  <IconRefresh className="h-4 w-4 mr-2" />
                  Refresh
                </Button>
                {isSuperAdmin && (
                  <Button
                    variant="outline"
                    onClick={handleDatabaseOptimization}
                    size="sm"
                  >
                    <IconTools className="h-4 w-4 mr-2" />
                    Optimize
                  </Button>
                )}
              </div>

              {databaseStats && (
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Collections</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {databaseStats.total_collections}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle>Total Documents</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {databaseStats.total_documents?.toLocaleString()}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle>Database Size</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {databaseStats.database_size_gb?.toFixed(2)} GB
                      </div>
                    </CardContent>
                  </Card>
                </div>
              )}
            </TabsContent>

            {/* Cache Tab */}
            <TabsContent value="cache" className="space-y-4">
              <div className="flex items-center gap-2 mb-4">
                <Button variant="outline" onClick={fetchCacheStats} size="sm">
                  <IconRefresh className="h-4 w-4 mr-2" />
                  Refresh
                </Button>
                {isSuperAdmin && (
                  <Button
                    variant="outline"
                    onClick={() => setShowCacheDialog(true)}
                    size="sm"
                  >
                    <IconSettings className="h-4 w-4 mr-2" />
                    Manage Cache
                  </Button>
                )}
              </div>

              {cacheStats && (
                <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Hit Rate</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {cacheStats.hit_rate}%
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle>Total Keys</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {cacheStats.total_keys?.toLocaleString()}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle>Memory Used</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {cacheStats.memory_used_mb} MB
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle>Connections</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {cacheStats.connected_clients}
                      </div>
                    </CardContent>
                  </Card>
                </div>
              )}
            </TabsContent>

            {/* Logs Tab */}
            <TabsContent value="logs" className="space-y-4">
              <div className="flex items-center gap-2 mb-4">
                <Button variant="outline" onClick={fetchSystemLogs} size="sm">
                  <IconRefresh className="h-4 w-4 mr-2" />
                  Refresh
                </Button>
              </div>

              <Card>
                <CardHeader>
                  <CardTitle>Recent System Logs</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2 max-h-96 overflow-y-auto">
                    {systemLogs.map((log, index) => (
                      <div
                        key={index}
                        className="flex items-start gap-3 p-2 border-b last:border-b-0"
                      >
                        <Badge
                          variant={
                            log.level === "error"
                              ? "destructive"
                              : log.level === "warning"
                              ? "secondary"
                              : "outline"
                          }
                        >
                          {log.level}
                        </Badge>
                        <div className="flex-1">
                          <p className="text-sm">{log.message}</p>
                          <p className="text-xs text-muted-foreground">
                            {new Date(log.timestamp).toLocaleString()}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            {/* Backups Tab (Super Admin Only) */}
            {isSuperAdmin && (
              <TabsContent value="backups" className="space-y-4">
                <div className="flex items-center gap-2 mb-4">
                  <Button
                    variant="outline"
                    onClick={() => setShowBackupDialog(true)}
                    size="sm"
                  >
                    <IconDownload className="h-4 w-4 mr-2" />
                    Create Backup
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => setShowRestoreDialog(true)}
                    size="sm"
                  >
                    <IconUpload className="h-4 w-4 mr-2" />
                    Restore
                  </Button>
                  <Button variant="outline" onClick={fetchBackups} size="sm">
                    <IconRefresh className="h-4 w-4 mr-2" />
                    Refresh
                  </Button>
                </div>

                <Card>
                  <CardHeader>
                    <CardTitle>Database Backups</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      {backups.map((backup, index) => (
                        <div
                          key={index}
                          className="flex items-center justify-between p-3 border rounded-lg"
                        >
                          <div>
                            <p className="font-medium">{backup.name}</p>
                            <p className="text-sm text-muted-foreground">
                              {backup.size} •{" "}
                              {new Date(backup.created_at).toLocaleString()}
                            </p>
                          </div>
                          <div className="flex items-center gap-2">
                            <Badge
                              variant={
                                backup.status === "completed"
                                  ? "default"
                                  : "secondary"
                              }
                            >
                              {backup.status}
                            </Badge>
                            <Button size="sm" variant="outline">
                              <IconDownload className="h-3 w-3" />
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>
            )}
          </Tabs>
        </div>
      </SidebarInset>

      {/* Maintenance Mode Dialog */}
      <Dialog
        open={showMaintenanceDialog}
        onOpenChange={setShowMaintenanceDialog}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Maintenance Mode</DialogTitle>
            <DialogDescription>
              Enable or disable system maintenance mode
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label>Maintenance Message</Label>
              <Textarea
                placeholder="System is under maintenance..."
                value={maintenanceMessage}
                onChange={(e) => setMaintenanceMessage(e.target.value)}
              />
            </div>
            <div>
              <Label>Duration (optional)</Label>
              <Input
                placeholder="2 hours"
                value={maintenanceDuration}
                onChange={(e) => setMaintenanceDuration(e.target.value)}
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowMaintenanceDialog(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={() => handleMaintenanceMode(false)}
              variant="outline"
            >
              Disable Maintenance
            </Button>
            <Button
              onClick={() => handleMaintenanceMode(true)}
              variant="destructive"
            >
              Enable Maintenance
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Cache Management Dialog */}
      <Dialog open={showCacheDialog} onOpenChange={setShowCacheDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Cache Management</DialogTitle>
            <DialogDescription>
              Clear or warm cache for better performance
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label>Action</Label>
              <Select value={cacheAction} onValueChange={setCacheAction}>
                <SelectTrigger>
                  <SelectValue placeholder="Select action" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="clear">Clear Cache</SelectItem>
                  <SelectItem value="warm">Warm Cache</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Cache Type</Label>
              <Select value={cacheType} onValueChange={setCacheType}>
                <SelectTrigger>
                  <SelectValue placeholder="All caches" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All Caches</SelectItem>
                  <SelectItem value="user">User Cache</SelectItem>
                  <SelectItem value="post">Post Cache</SelectItem>
                  <SelectItem value="session">Session Cache</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCacheDialog(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCacheOperation}
              disabled={!cacheAction}
              variant={cacheAction === "clear" ? "destructive" : "default"}
            >
              {cacheAction === "clear" ? "Clear Cache" : "Warm Cache"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Backup Dialog */}
      <Dialog open={showBackupDialog} onOpenChange={setShowBackupDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Database Backup</DialogTitle>
            <DialogDescription>
              Create a backup of the database
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label>Backup Type</Label>
              <Select value={backupType} onValueChange={setBackupType}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="full">Full Backup</SelectItem>
                  <SelectItem value="partial">Partial Backup</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {backupType === "partial" && (
              <div>
                <Label>Collections (optional)</Label>
                <Input
                  placeholder="users,posts,comments"
                  value={selectedCollections.join(",")}
                  onChange={(e) =>
                    setSelectedCollections(
                      e.target.value.split(",").filter(Boolean)
                    )
                  }
                />
              </div>
            )}
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowBackupDialog(false)}
            >
              Cancel
            </Button>
            <Button onClick={handleDatabaseBackup}>Create Backup</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Restore Dialog */}
      <Dialog open={showRestoreDialog} onOpenChange={setShowRestoreDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Restore Database</DialogTitle>
            <DialogDescription>
              Restore database from a backup
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label>Select Backup</Label>
              <Select value={selectedBackup} onValueChange={setSelectedBackup}>
                <SelectTrigger>
                  <SelectValue placeholder="Choose backup" />
                </SelectTrigger>
                <SelectContent>
                  {backups.map((backup) => (
                    <SelectItem key={backup.id} value={backup.id}>
                      {backup.name} -{" "}
                      {new Date(backup.created_at).toLocaleDateString()}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowRestoreDialog(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleDatabaseRestore}
              disabled={!selectedBackup}
              variant="destructive"
            >
              Restore Database
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SidebarProvider>
  );
}

export default withAuth(SystemPage);
