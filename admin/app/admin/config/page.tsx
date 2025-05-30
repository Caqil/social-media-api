// app/admin/config/page.tsx
"use client";

import { useEffect, useState } from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { SiteHeader } from "@/components/site-header";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { withAuth, usePermissions } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import {
  IconSettings,
  IconRefresh,
  IconSave,
  IconFlag,
  IconClock,
  IconShield,
  IconDatabase,
  IconMail,
  IconCloudUpload,
  IconToggleLeft,
  IconToggleRight,
  IconAlertTriangle,
  IconCheck,
  IconX,
  IconHistory,
} from "@tabler/icons-react";

function ConfigPage() {
  const { hasPermission } = usePermissions();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState("general");

  // Configuration states
  const [generalConfig, setGeneralConfig] = useState<any>({});
  const [featureFlags, setFeatureFlags] = useState<any>({});
  const [rateLimits, setRateLimits] = useState<any>({});
  const [configHistory, setConfigHistory] = useState<any[]>([]);

  // Dialog states
  const [showHistoryDialog, setShowHistoryDialog] = useState(false);
  const [showResetDialog, setShowResetDialog] = useState(false);

  // Check if user has permission to manage configuration
  const canManageConfig = hasPermission("config.update");

  const fetchConfiguration = async () => {
    try {
      setLoading(true);
      setError(null);

      const [configResponse, featuresResponse, rateLimitsResponse] =
        await Promise.all([
          apiClient.getConfiguration(),
          apiClient.getFeatureFlags(),
          apiClient.getRateLimits(),
        ]);

      setGeneralConfig(configResponse.data);
      setFeatureFlags(featuresResponse.data);
      setRateLimits(rateLimitsResponse.data);
    } catch (error: any) {
      console.error("Failed to fetch configuration:", error);
      setError(error.response?.data?.message || "Failed to load configuration");
    } finally {
      setLoading(false);
    }
  };

  const fetchConfigHistory = async () => {
    try {
      const response = await apiClient.getConfigurationHistory({ limit: 20 });
      setConfigHistory(response.data);
    } catch (error) {
      console.error("Failed to fetch config history:", error);
    }
  };

  useEffect(() => {
    if (canManageConfig) {
      fetchConfiguration();
    } else {
      setError("You don't have permission to access configuration settings");
      setLoading(false);
    }
  }, [canManageConfig]);

  const handleSaveGeneral = async () => {
    try {
      setSaving(true);
      await apiClient.updateConfiguration(generalConfig);
      setSuccess("General configuration updated successfully");
      setTimeout(() => setSuccess(null), 3000);
    } catch (error: any) {
      setError(
        error.response?.data?.message || "Failed to update configuration"
      );
      setTimeout(() => setError(null), 5000);
    } finally {
      setSaving(false);
    }
  };

  const handleSaveFeatures = async () => {
    try {
      setSaving(true);
      await apiClient.updateFeatureFlags(featureFlags);
      setSuccess("Feature flags updated successfully");
      setTimeout(() => setSuccess(null), 3000);
    } catch (error: any) {
      setError(
        error.response?.data?.message || "Failed to update feature flags"
      );
      setTimeout(() => setError(null), 5000);
    } finally {
      setSaving(false);
    }
  };

  const handleSaveRateLimits = async () => {
    try {
      setSaving(true);
      await apiClient.updateRateLimits(rateLimits);
      setSuccess("Rate limits updated successfully");
      setTimeout(() => setSuccess(null), 3000);
    } catch (error: any) {
      setError(error.response?.data?.message || "Failed to update rate limits");
      setTimeout(() => setError(null), 5000);
    } finally {
      setSaving(false);
    }
  };

  const handleToggleFeature = async (feature: string) => {
    try {
      await apiClient.toggleFeature(feature);
      setFeatureFlags((prev: any) => ({
        ...prev,
        [feature]: !prev[feature],
      }));
      setSuccess(`Feature ${feature} toggled successfully`);
      setTimeout(() => setSuccess(null), 3000);
    } catch (error: any) {
      setError(error.response?.data?.message || "Failed to toggle feature");
      setTimeout(() => setError(null), 5000);
    }
  };

  const handleValidateConfig = async () => {
    try {
      const response = await apiClient.validateConfiguration();
      if (response.data.valid) {
        setSuccess("Configuration is valid");
      } else {
        setError(
          `Configuration validation failed: ${response.data.errors?.join(", ")}`
        );
      }
      setTimeout(() => {
        setSuccess(null);
        setError(null);
      }, 5000);
    } catch (error: any) {
      setError(
        error.response?.data?.message || "Failed to validate configuration"
      );
      setTimeout(() => setError(null), 5000);
    }
  };

  if (!canManageConfig) {
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
              <IconShield className="h-4 w-4" />
              <AlertDescription>
                You don't have permission to access configuration settings.
              </AlertDescription>
            </Alert>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

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
                <Button onClick={fetchConfiguration} className="w-full">
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
              <h1 className="text-2xl font-bold">Configuration</h1>
              <p className="text-muted-foreground">
                Manage system settings and configuration
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button
                onClick={handleValidateConfig}
                variant="outline"
                size="sm"
                disabled={loading}
              >
                <IconCheck className="h-4 w-4 mr-2" />
                Validate
              </Button>
              <Button
                onClick={() => {
                  fetchConfigHistory();
                  setShowHistoryDialog(true);
                }}
                variant="outline"
                size="sm"
              >
                <IconHistory className="h-4 w-4 mr-2" />
                History
              </Button>
              <Button
                onClick={fetchConfiguration}
                variant="outline"
                size="sm"
                disabled={loading}
              >
                <IconRefresh
                  className={`h-4 w-4 mr-2 ${loading ? "animate-spin" : ""}`}
                />
                Refresh
              </Button>
            </div>
          </div>

          {/* Success/Error Messages */}
          {success && (
            <Alert className="border-green-200 bg-green-50">
              <IconCheck className="h-4 w-4 text-green-600" />
              <AlertDescription className="text-green-800">
                {success}
              </AlertDescription>
            </Alert>
          )}

          {error && (
            <Alert variant="destructive">
              <IconAlertTriangle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="general">General Settings</TabsTrigger>
              <TabsTrigger value="features">Feature Flags</TabsTrigger>
              <TabsTrigger value="limits">Rate Limits</TabsTrigger>
            </TabsList>

            {/* General Settings Tab */}
            <TabsContent value="general" className="space-y-4">
              {loading ? (
                <GeneralSkeleton />
              ) : (
                <GeneralSettings
                  config={generalConfig}
                  setConfig={setGeneralConfig}
                  onSave={handleSaveGeneral}
                  saving={saving}
                />
              )}
            </TabsContent>

            {/* Feature Flags Tab */}
            <TabsContent value="features" className="space-y-4">
              {loading ? (
                <FeaturesSkeleton />
              ) : (
                <FeatureFlags
                  features={featureFlags}
                  setFeatures={setFeatureFlags}
                  onSave={handleSaveFeatures}
                  onToggle={handleToggleFeature}
                  saving={saving}
                />
              )}
            </TabsContent>

            {/* Rate Limits Tab */}
            <TabsContent value="limits" className="space-y-4">
              {loading ? (
                <RateLimitsSkeleton />
              ) : (
                <RateLimits
                  limits={rateLimits}
                  setLimits={setRateLimits}
                  onSave={handleSaveRateLimits}
                  saving={saving}
                />
              )}
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>

      {/* Config History Dialog */}
      <Dialog open={showHistoryDialog} onOpenChange={setShowHistoryDialog}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Configuration History</DialogTitle>
            <DialogDescription>
              Recent configuration changes and rollback options
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {configHistory.map((item, index) => (
              <Card key={index}>
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <div>
                      <CardTitle className="text-base">
                        {item.change_type}
                      </CardTitle>
                      <CardDescription>
                        {new Date(item.created_at).toLocaleString()} by{" "}
                        {item.admin_name}
                      </CardDescription>
                    </div>
                    <Badge variant="outline">{item.status}</Badge>
                  </div>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{item.description}</p>
                  {item.changes && (
                    <div className="mt-2 text-xs bg-gray-50 p-2 rounded">
                      {JSON.stringify(item.changes, null, 2)}
                    </div>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowHistoryDialog(false)}
            >
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SidebarProvider>
  );
}

function GeneralSettings({ config, setConfig, onSave, saving }: any) {
  return (
    <div className="space-y-6">
      {/* Application Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <IconSettings className="h-5 w-5" />
            Application Settings
          </CardTitle>
          <CardDescription>
            Basic application configuration and limits
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label>Max Post Length</Label>
              <Input
                type="number"
                value={config.max_post_length || 280}
                onChange={(e) =>
                  setConfig({
                    ...config,
                    max_post_length: parseInt(e.target.value),
                  })
                }
                placeholder="280"
              />
            </div>
            <div>
              <Label>Max File Size (MB)</Label>
              <Input
                type="number"
                value={config.max_file_size || 10}
                onChange={(e) =>
                  setConfig({
                    ...config,
                    max_file_size: parseInt(e.target.value),
                  })
                }
                placeholder="10"
              />
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              checked={config.registration_enabled || false}
              onCheckedChange={(checked: any) =>
                setConfig({ ...config, registration_enabled: checked })
              }
            />
            <Label>Enable User Registration</Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              checked={config.maintenance_mode || false}
              onCheckedChange={(checked: any) =>
                setConfig({ ...config, maintenance_mode: checked })
              }
            />
            <Label>Maintenance Mode</Label>
          </div>
        </CardContent>
      </Card>

      {/* Content Moderation */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <IconShield className="h-5 w-5" />
            Content Moderation
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center space-x-2">
            <Switch
              checked={config.auto_moderation || false}
              onCheckedChange={(checked: any) =>
                setConfig({ ...config, auto_moderation: checked })
              }
            />
            <Label>Enable Auto Moderation</Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              checked={config.profanity_filter || false}
              onCheckedChange={(checked: any) =>
                setConfig({ ...config, profanity_filter: checked })
              }
            />
            <Label>Profanity Filter</Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              checked={config.spam_detection || false}
              onCheckedChange={(checked: any) =>
                setConfig({ ...config, spam_detection: checked })
              }
            />
            <Label>Spam Detection</Label>
          </div>

          <div>
            <Label>Manual Review Threshold</Label>
            <Input
              type="number"
              value={config.manual_review_threshold || 5}
              onChange={(e) =>
                setConfig({
                  ...config,
                  manual_review_threshold: parseInt(e.target.value),
                })
              }
              placeholder="5"
            />
          </div>
        </CardContent>
      </Card>

      {/* Security Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <IconShield className="h-5 w-5" />
            Security Settings
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <Label>Password Min Length</Label>
            <Input
              type="number"
              value={config.password_min_length || 8}
              onChange={(e) =>
                setConfig({
                  ...config,
                  password_min_length: parseInt(e.target.value),
                })
              }
              placeholder="8"
            />
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              checked={config.require_email_verification || false}
              onCheckedChange={(checked: any) =>
                setConfig({ ...config, require_email_verification: checked })
              }
            />
            <Label>Require Email Verification</Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              checked={config.enable_two_factor || false}
              onCheckedChange={(checked: any) =>
                setConfig({ ...config, enable_two_factor: checked })
              }
            />
            <Label>Enable Two-Factor Authentication</Label>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label>Max Login Attempts</Label>
              <Input
                type="number"
                value={config.max_login_attempts || 5}
                onChange={(e) =>
                  setConfig({
                    ...config,
                    max_login_attempts: parseInt(e.target.value),
                  })
                }
                placeholder="5"
              />
            </div>
            <div>
              <Label>Lockout Duration (minutes)</Label>
              <Input
                type="number"
                value={config.lockout_duration || 30}
                onChange={(e) =>
                  setConfig({
                    ...config,
                    lockout_duration: parseInt(e.target.value),
                  })
                }
                placeholder="30"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button onClick={onSave} disabled={saving}>
          {saving ? (
            <>
              <IconRefresh className="h-4 w-4 mr-2 animate-spin" />
              Saving...
            </>
          ) : (
            <>
              <IconSave className="h-4 w-4 mr-2" />
              Save General Settings
            </>
          )}
        </Button>
      </div>
    </div>
  );
}

function FeatureFlags({
  features,
  setFeatures,
  onSave,
  onToggle,
  saving,
}: any) {
  const featureList = [
    {
      key: "stories_enabled",
      label: "Stories",
      description: "Enable user stories feature",
    },
    {
      key: "groups_enabled",
      label: "Groups",
      description: "Enable user groups feature",
    },
    {
      key: "events_enabled",
      label: "Events",
      description: "Enable events feature",
    },
    {
      key: "live_streaming",
      label: "Live Streaming",
      description: "Enable live streaming feature",
    },
    {
      key: "messaging_enabled",
      label: "Messaging",
      description: "Enable direct messaging",
    },
    {
      key: "notifications_enabled",
      label: "Notifications",
      description: "Enable push notifications",
    },
    {
      key: "analytics_enabled",
      label: "Analytics",
      description: "Enable analytics tracking",
    },
    {
      key: "api_access",
      label: "API Access",
      description: "Enable public API access",
    },
  ];

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <IconFlag className="h-5 w-5" />
            Feature Flags
          </CardTitle>
          <CardDescription>
            Enable or disable features across the platform
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {featureList.map((feature) => (
            <div
              key={feature.key}
              className="flex items-center justify-between p-4 border rounded-lg"
            >
              <div>
                <h4 className="font-medium">{feature.label}</h4>
                <p className="text-sm text-muted-foreground">
                  {feature.description}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <Badge
                  variant={features[feature.key] ? "default" : "secondary"}
                >
                  {features[feature.key] ? "Enabled" : "Disabled"}
                </Badge>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onToggle(feature.key)}
                >
                  {features[feature.key] ? (
                    <IconToggleRight className="h-5 w-5 text-green-600" />
                  ) : (
                    <IconToggleLeft className="h-5 w-5 text-gray-400" />
                  )}
                </Button>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button onClick={onSave} disabled={saving}>
          {saving ? (
            <>
              <IconRefresh className="h-4 w-4 mr-2 animate-spin" />
              Saving...
            </>
          ) : (
            <>
              <IconSave className="h-4 w-4 mr-2" />
              Save Feature Flags
            </>
          )}
        </Button>
      </div>
    </div>
  );
}

function RateLimits({ limits, setLimits, onSave, saving }: any) {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <IconClock className="h-5 w-5" />
            Rate Limits
          </CardTitle>
          <CardDescription>
            Configure rate limiting for different actions
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label>Posts per Hour</Label>
              <Input
                type="number"
                value={limits.posts_per_hour || 10}
                onChange={(e) =>
                  setLimits({
                    ...limits,
                    posts_per_hour: parseInt(e.target.value),
                  })
                }
                placeholder="10"
              />
            </div>
            <div>
              <Label>Comments per Hour</Label>
              <Input
                type="number"
                value={limits.comments_per_hour || 50}
                onChange={(e) =>
                  setLimits({
                    ...limits,
                    comments_per_hour: parseInt(e.target.value),
                  })
                }
                placeholder="50"
              />
            </div>
            <div>
              <Label>Messages per Hour</Label>
              <Input
                type="number"
                value={limits.messages_per_hour || 100}
                onChange={(e) =>
                  setLimits({
                    ...limits,
                    messages_per_hour: parseInt(e.target.value),
                  })
                }
                placeholder="100"
              />
            </div>
            <div>
              <Label>Likes per Minute</Label>
              <Input
                type="number"
                value={limits.likes_per_minute || 30}
                onChange={(e) =>
                  setLimits({
                    ...limits,
                    likes_per_minute: parseInt(e.target.value),
                  })
                }
                placeholder="30"
              />
            </div>
            <div>
              <Label>Follows per Hour</Label>
              <Input
                type="number"
                value={limits.follows_per_hour || 20}
                onChange={(e) =>
                  setLimits({
                    ...limits,
                    follows_per_hour: parseInt(e.target.value),
                  })
                }
                placeholder="20"
              />
            </div>
            <div>
              <Label>API Requests per Minute</Label>
              <Input
                type="number"
                value={limits.api_requests_per_minute || 60}
                onChange={(e) =>
                  setLimits({
                    ...limits,
                    api_requests_per_minute: parseInt(e.target.value),
                  })
                }
                placeholder="60"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button onClick={onSave} disabled={saving}>
          {saving ? (
            <>
              <IconRefresh className="h-4 w-4 mr-2 animate-spin" />
              Saving...
            </>
          ) : (
            <>
              <IconSave className="h-4 w-4 mr-2" />
              Save Rate Limits
            </>
          )}
        </Button>
      </div>
    </div>
  );
}

// Skeleton components
function GeneralSkeleton() {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-64" />
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function FeaturesSkeleton() {
  return <GeneralSkeleton />;
}

function RateLimitsSkeleton() {
  return <GeneralSkeleton />;
}

export default withAuth(ConfigPage);
