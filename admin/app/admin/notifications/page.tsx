// app/admin/notifications/page.tsx
"use client";

import { useEffect, useState } from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { SiteHeader } from "@/components/site-header";
import { DataTable } from "@/components/data-table";
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
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { withAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import { Notification, TableColumn } from "@/types/admin";
import {
  IconBell,
  IconEye,
  IconTrash,
  IconSend,
  IconUser,
  IconUsers,
  IconHeart,
  IconMessage,
  IconUserPlus,
  IconAt,
  IconSettings,
  IconFlag,
  IconCheck,
  IconClock,
  IconMail,
  IconX,
} from "@tabler/icons-react";

function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [pagination, setPagination] = useState<any>(null);
  const [filters, setFilters] = useState<any>({ page: 1, limit: 20 });
  const [notificationStats, setNotificationStats] = useState<any>(null);

  // Dialog states
  const [selectedNotification, setSelectedNotification] =
    useState<Notification | null>(null);
  const [showNotificationDetails, setShowNotificationDetails] = useState(false);
  const [showSendDialog, setShowSendDialog] = useState(false);
  const [showBroadcastDialog, setShowBroadcastDialog] = useState(false);
  const [showBulkDialog, setShowBulkDialog] = useState(false);
  const [bulkAction, setBulkAction] = useState<string>("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);

  // Send notification form
  const [sendForm, setSendForm] = useState({
    user_ids: "",
    title: "",
    message: "",
    type: "system",
    data: "",
  });

  // Broadcast notification form
  const [broadcastForm, setBroadcastForm] = useState({
    title: "",
    message: "",
    type: "system",
    target_audience: "all",
    data: "",
  });

  // Active tab
  const [activeTab, setActiveTab] = useState("notifications");

  // Fetch notifications
  const fetchNotifications = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getNotifications(filters);
      setNotifications(response.data.data);
      setPagination(response.data.pagination);
    } catch (error: any) {
      console.error("Failed to fetch notifications:", error);
      setError(error.response?.data?.message || "Failed to load notifications");
    } finally {
      setLoading(false);
    }
  };

  // Fetch notification stats
  const fetchNotificationStats = async () => {
    try {
      const response = await apiClient.getNotificationStats();
      setNotificationStats(response.data);
    } catch (error) {
      console.error("Failed to fetch notification stats:", error);
    }
  };

  useEffect(() => {
    if (activeTab === "notifications") {
      fetchNotifications();
    } else if (activeTab === "stats") {
      fetchNotificationStats();
    }
  }, [filters, activeTab]);

  // Handle page change
  const handlePageChange = (page: number) => {
    setFilters((prev: any) => ({ ...prev, page }));
  };

  // Handle filtering
  const handleFilter = (newFilters: Record<string, any>) => {
    setFilters((prev: any) => ({
      ...prev,
      ...newFilters,
      page: 1,
    }));
  };

  // Handle row selection
  const handleRowSelect = (selectedRows: string[]) => {
    setSelectedIds(selectedRows);
  };

  // Handle bulk actions
  const handleBulkAction = (action: string, ids: string[]) => {
    setBulkAction(action);
    setSelectedIds(ids);
    setShowBulkDialog(true);
  };

  // Execute bulk action
  const executeBulkAction = async () => {
    try {
      await apiClient.bulkNotificationAction({
        notification_ids: selectedIds,
        action: bulkAction,
      });

      setShowBulkDialog(false);
      fetchNotifications();
      setSuccess("Bulk action completed successfully");
      setTimeout(() => setSuccess(null), 3000);
    } catch (error: any) {
      console.error("Bulk action failed:", error);
      setError(error.response?.data?.message || "Bulk action failed");
      setTimeout(() => setError(null), 5000);
    }
  };

  // View notification details
  const viewNotificationDetails = (notification: Notification) => {
    setSelectedNotification(notification);
    setShowNotificationDetails(true);
  };

  // Handle send notification
  const handleSendNotification = async () => {
    try {
      const userIds = sendForm.user_ids
        .split(",")
        .map((id) => id.trim())
        .filter(Boolean);

      await apiClient.sendNotificationToUsers({
        user_ids: userIds,
        title: sendForm.title,
        message: sendForm.message,
        type: sendForm.type,
        data: sendForm.data ? JSON.parse(sendForm.data) : undefined,
      });

      setShowSendDialog(false);
      setSendForm({
        user_ids: "",
        title: "",
        message: "",
        type: "system",
        data: "",
      });
      setSuccess("Notification sent successfully");
      setTimeout(() => setSuccess(null), 3000);
    } catch (error: any) {
      setError(error.response?.data?.message || "Failed to send notification");
      setTimeout(() => setError(null), 5000);
    }
  };

  // Handle broadcast notification
  const handleBroadcastNotification = async () => {
    try {
      await apiClient.broadcastNotification({
        title: broadcastForm.title,
        message: broadcastForm.message,
        type: broadcastForm.type,
        target_audience: broadcastForm.target_audience,
        data: broadcastForm.data ? JSON.parse(broadcastForm.data) : undefined,
      });

      setShowBroadcastDialog(false);
      setBroadcastForm({
        title: "",
        message: "",
        type: "system",
        target_audience: "all",
        data: "",
      });
      setSuccess("Broadcast sent successfully");
      setTimeout(() => setSuccess(null), 3000);
    } catch (error: any) {
      setError(error.response?.data?.message || "Failed to send broadcast");
      setTimeout(() => setError(null), 5000);
    }
  };

  // Get notification type icon
  const getNotificationTypeIcon = (type: string) => {
    switch (type) {
      case "like":
        return <IconHeart className="h-4 w-4" />;
      case "comment":
        return <IconMessage className="h-4 w-4" />;
      case "follow":
        return <IconUserPlus className="h-4 w-4" />;
      case "mention":
        return <IconAt className="h-4 w-4" />;
      case "system":
        return <IconSettings className="h-4 w-4" />;
      case "admin":
        return <IconFlag className="h-4 w-4" />;
      default:
        return <IconBell className="h-4 w-4" />;
    }
  };

  // Get notification type color
  const getNotificationTypeColor = (type: string) => {
    switch (type) {
      case "like":
        return "text-red-600 bg-red-100";
      case "comment":
        return "text-blue-600 bg-blue-100";
      case "follow":
        return "text-green-600 bg-green-100";
      case "mention":
        return "text-purple-600 bg-purple-100";
      case "system":
        return "text-orange-600 bg-orange-100";
      case "admin":
        return "text-indigo-600 bg-indigo-100";
      default:
        return "text-gray-600 bg-gray-100";
    }
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "notification",
      label: "Notification",
      render: (_, notification: Notification) => (
        <div className="max-w-md">
          <div className="flex items-center gap-2 mb-2">
            <div
              className={`p-1 rounded ${getNotificationTypeColor(
                notification.type
              )}`}
            >
              {getNotificationTypeIcon(notification.type)}
            </div>
            <Badge variant="outline" className="text-xs">
              {notification.type}
            </Badge>
            {notification.is_read && (
              <Badge variant="secondary" className="text-xs">
                Read
              </Badge>
            )}
          </div>

          <h4 className="font-medium mb-1">{notification.title}</h4>
          <p className="text-sm text-muted-foreground line-clamp-2 mb-2">
            {notification.message}
          </p>

          {(notification.user || notification.sender) && (
            <div className="flex items-center gap-2">
              <Avatar className="h-4 w-4">
                <AvatarImage
                  src={
                    notification.user?.profile_picture ||
                    notification.sender?.profile_picture
                  }
                />
                <AvatarFallback className="text-xs">
                  {(notification.user?.username ||
                    notification.sender?.username)?.[0]?.toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <span className="text-xs text-muted-foreground">
                to {notification.user?.username || "user"}
                {notification.sender && (
                  <> from {notification.sender.username}</>
                )}
              </span>
            </div>
          )}
        </div>
      ),
    },
    {
      key: "type",
      label: "Type",
      filterable: true,
      render: (value: string) => (
        <Badge className={getNotificationTypeColor(value)}>
          {getNotificationTypeIcon(value)}
          <span className="ml-1 capitalize">{value}</span>
        </Badge>
      ),
    },
    {
      key: "status",
      label: "Status",
      filterable: true,
      render: (_, notification: Notification) =>
        notification.is_read ? (
          <Badge variant="default">
            <IconCheck className="h-3 w-3 mr-1" />
            Read
          </Badge>
        ) : (
          <Badge variant="secondary">
            <IconClock className="h-3 w-3 mr-1" />
            Unread
          </Badge>
        ),
    },
    {
      key: "created_at",
      label: "Sent",
      sortable: true,
      render: (value: string) => (
        <div className="text-xs">{new Date(value).toLocaleString()}</div>
      ),
    },
    {
      key: "read_at",
      label: "Read At",
      render: (value: string) =>
        value ? (
          <div className="text-xs">{new Date(value).toLocaleString()}</div>
        ) : (
          <span className="text-muted-foreground">Unread</span>
        ),
    },
    {
      key: "actions",
      label: "Actions",
      render: (_, notification: Notification) => (
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewNotificationDetails(notification)}
          >
            <IconEye className="h-3 w-3" />
          </Button>
          <Button
            size="sm"
            variant="destructive"
            onClick={() => handleDeleteNotification(notification)}
          >
            <IconTrash className="h-3 w-3" />
          </Button>
        </div>
      ),
    },
  ];

  // Handle delete notification
  const handleDeleteNotification = async (notification: Notification) => {
    try {
      await apiClient.deleteNotification(notification.id);
      fetchNotifications();
      setSuccess("Notification deleted successfully");
      setTimeout(() => setSuccess(null), 3000);
    } catch (error: any) {
      setError(
        error.response?.data?.message || "Failed to delete notification"
      );
      setTimeout(() => setError(null), 5000);
    }
  };

  // Bulk actions configuration
  const bulkActions = [
    {
      label: "Delete Notifications",
      action: "delete",
      variant: "destructive" as const,
    },
  ];

  if (error && !loading) {
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
              <h1 className="text-2xl font-bold">Notifications Management</h1>
              <p className="text-muted-foreground">
                Manage notifications and send messages to users
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button
                onClick={() => setShowSendDialog(true)}
                variant="outline"
                size="sm"
              >
                <IconSend className="h-4 w-4 mr-2" />
                Send to Users
              </Button>
              <Button onClick={() => setShowBroadcastDialog(true)} size="sm">
                <IconMail className="h-4 w-4 mr-2" />
                Broadcast
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
              <IconX className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList>
              <TabsTrigger value="notifications">Notifications</TabsTrigger>
              <TabsTrigger value="stats">Statistics</TabsTrigger>
            </TabsList>

            <TabsContent value="notifications" className="space-y-4">
              <DataTable
                title="Notifications"
                description={`Manage ${pagination?.total || 0} notifications`}
                data={notifications}
                columns={columns}
                loading={loading}
                pagination={pagination}
                searchPlaceholder="Search notifications by title or message..."
                onPageChange={handlePageChange}
                onFilter={handleFilter}
                onRowSelect={handleRowSelect}
                bulkActions={bulkActions}
                onBulkAction={handleBulkAction}
                onRefresh={fetchNotifications}
                onExport={() => console.log("Export notifications")}
              />
            </TabsContent>

            <TabsContent value="stats" className="space-y-4">
              {notificationStats && (
                <>
                  {/* Notification Statistics */}
                  <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Total Sent</CardDescription>
                        <CardTitle className="text-2xl">
                          {notificationStats.total_sent?.toLocaleString() || 0}
                        </CardTitle>
                      </CardHeader>
                    </Card>
                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Read Rate</CardDescription>
                        <CardTitle className="text-2xl">
                          {notificationStats.read_rate?.toFixed(1) || 0}%
                        </CardTitle>
                      </CardHeader>
                    </Card>
                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Pending</CardDescription>
                        <CardTitle className="text-2xl">
                          {notificationStats.pending?.toLocaleString() || 0}
                        </CardTitle>
                      </CardHeader>
                    </Card>
                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Failed</CardDescription>
                        <CardTitle className="text-2xl">
                          {notificationStats.failed?.toLocaleString() || 0}
                        </CardTitle>
                      </CardHeader>
                    </Card>
                  </div>

                  {/* Notifications by Type */}
                  <Card>
                    <CardHeader>
                      <CardTitle>Notifications by Type</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        {notificationStats.by_type?.map(
                          (item: any, index: number) => (
                            <div
                              key={index}
                              className="flex items-center justify-between"
                            >
                              <div className="flex items-center gap-2">
                                <div
                                  className={`p-1 rounded ${getNotificationTypeColor(
                                    item.type
                                  )}`}
                                >
                                  {getNotificationTypeIcon(item.type)}
                                </div>
                                <span className="capitalize">{item.type}</span>
                              </div>
                              <div className="flex items-center gap-2">
                                <div className="w-24 bg-gray-200 rounded-full h-2">
                                  <div
                                    className="bg-primary h-2 rounded-full"
                                    style={{ width: `${item.percentage}%` }}
                                  />
                                </div>
                                <Badge variant="outline">{item.count}</Badge>
                              </div>
                            </div>
                          )
                        )}
                      </div>
                    </CardContent>
                  </Card>
                </>
              )}
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>

      {/* Notification Details Dialog */}
      <Dialog
        open={showNotificationDetails}
        onOpenChange={setShowNotificationDetails}
      >
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Notification Details</DialogTitle>
          </DialogHeader>

          {selectedNotification && (
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div
                        className={`p-2 rounded ${getNotificationTypeColor(
                          selectedNotification.type
                        )}`}
                      >
                        {getNotificationTypeIcon(selectedNotification.type)}
                      </div>
                      <div>
                        <CardTitle>{selectedNotification.title}</CardTitle>
                        <CardDescription>
                          {new Date(
                            selectedNotification.created_at
                          ).toLocaleString()}
                        </CardDescription>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge
                        className={getNotificationTypeColor(
                          selectedNotification.type
                        )}
                      >
                        {selectedNotification.type}
                      </Badge>
                      {selectedNotification.is_read ? (
                        <Badge variant="default">Read</Badge>
                      ) : (
                        <Badge variant="secondary">Unread</Badge>
                      )}
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <h4 className="font-medium mb-2">Message</h4>
                    <p className="text-sm bg-gray-50 p-3 rounded">
                      {selectedNotification.message}
                    </p>
                  </div>

                  {selectedNotification.data && (
                    <div>
                      <h4 className="font-medium mb-2">Additional Data</h4>
                      <pre className="text-xs bg-gray-50 p-3 rounded overflow-auto">
                        {JSON.stringify(selectedNotification.data, null, 2)}
                      </pre>
                    </div>
                  )}

                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <label className="font-medium">Type</label>
                      <p className="capitalize">{selectedNotification.type}</p>
                    </div>
                    <div>
                      <label className="font-medium">Status</label>
                      <p>{selectedNotification.is_read ? "Read" : "Unread"}</p>
                    </div>
                    <div>
                      <label className="font-medium">Sent</label>
                      <p>
                        {new Date(
                          selectedNotification.created_at
                        ).toLocaleString()}
                      </p>
                    </div>
                    {selectedNotification.read_at && (
                      <div>
                        <label className="font-medium">Read At</label>
                        <p>
                          {new Date(
                            selectedNotification.read_at
                          ).toLocaleString()}
                        </p>
                      </div>
                    )}
                  </div>

                  {selectedNotification.action_url && (
                    <div>
                      <label className="font-medium">Action URL</label>
                      <p className="text-sm text-primary">
                        {selectedNotification.action_url}
                      </p>
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowNotificationDetails(false)}
            >
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Send Notification Dialog */}
      <Dialog open={showSendDialog} onOpenChange={setShowSendDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Send Notification to Users</DialogTitle>
            <DialogDescription>
              Send a notification to specific users
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label>User IDs (comma-separated)</Label>
              <Input
                placeholder="user1,user2,user3"
                value={sendForm.user_ids}
                onChange={(e) =>
                  setSendForm({ ...sendForm, user_ids: e.target.value })
                }
              />
            </div>
            <div>
              <Label>Type</Label>
              <Select
                value={sendForm.type}
                onValueChange={(value) =>
                  setSendForm({ ...sendForm, type: value })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="system">System</SelectItem>
                  <SelectItem value="admin">Admin</SelectItem>
                  <SelectItem value="announcement">Announcement</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Title</Label>
              <Input
                placeholder="Notification title"
                value={sendForm.title}
                onChange={(e) =>
                  setSendForm({ ...sendForm, title: e.target.value })
                }
              />
            </div>
            <div>
              <Label>Message</Label>
              <Textarea
                placeholder="Notification message"
                value={sendForm.message}
                onChange={(e) =>
                  setSendForm({ ...sendForm, message: e.target.value })
                }
              />
            </div>
            <div>
              <Label>Additional Data (JSON, optional)</Label>
              <Textarea
                placeholder='{"key": "value"}'
                value={sendForm.data}
                onChange={(e) =>
                  setSendForm({ ...sendForm, data: e.target.value })
                }
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowSendDialog(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleSendNotification}
              disabled={
                !sendForm.user_ids || !sendForm.title || !sendForm.message
              }
            >
              <IconSend className="h-4 w-4 mr-2" />
              Send Notification
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Broadcast Dialog */}
      <Dialog open={showBroadcastDialog} onOpenChange={setShowBroadcastDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Broadcast Notification</DialogTitle>
            <DialogDescription>
              Send a notification to all users or specific audience
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label>Target Audience</Label>
              <Select
                value={broadcastForm.target_audience}
                onValueChange={(value) =>
                  setBroadcastForm({ ...broadcastForm, target_audience: value })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Users</SelectItem>
                  <SelectItem value="active">Active Users</SelectItem>
                  <SelectItem value="new">New Users</SelectItem>
                  <SelectItem value="premium">Premium Users</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Type</Label>
              <Select
                value={broadcastForm.type}
                onValueChange={(value) =>
                  setBroadcastForm({ ...broadcastForm, type: value })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="system">System</SelectItem>
                  <SelectItem value="admin">Admin</SelectItem>
                  <SelectItem value="announcement">Announcement</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Title</Label>
              <Input
                placeholder="Broadcast title"
                value={broadcastForm.title}
                onChange={(e) =>
                  setBroadcastForm({ ...broadcastForm, title: e.target.value })
                }
              />
            </div>
            <div>
              <Label>Message</Label>
              <Textarea
                placeholder="Broadcast message"
                value={broadcastForm.message}
                onChange={(e) =>
                  setBroadcastForm({
                    ...broadcastForm,
                    message: e.target.value,
                  })
                }
              />
            </div>
            <div>
              <Label>Additional Data (JSON, optional)</Label>
              <Textarea
                placeholder='{"key": "value"}'
                value={broadcastForm.data}
                onChange={(e) =>
                  setBroadcastForm({ ...broadcastForm, data: e.target.value })
                }
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowBroadcastDialog(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleBroadcastNotification}
              disabled={!broadcastForm.title || !broadcastForm.message}
            >
              <IconMail className="h-4 w-4 mr-2" />
              Send Broadcast
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Bulk Action Dialog */}
      <Dialog open={showBulkDialog} onOpenChange={setShowBulkDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Bulk Action Confirmation</DialogTitle>
            <DialogDescription>
              You are about to {bulkAction} {selectedIds.length}{" "}
              notification(s).
            </DialogDescription>
          </DialogHeader>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowBulkDialog(false)}>
              Cancel
            </Button>
            <Button onClick={executeBulkAction} variant="destructive">
              Execute {bulkAction}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SidebarProvider>
  );
}

export default withAuth(NotificationsPage);
