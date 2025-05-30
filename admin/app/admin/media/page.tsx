// app/admin/media/page.tsx
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
import { withAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import { Media, TableColumn } from "@/types/admin";
import {
  IconPhoto,
  IconVideo,
  IconMusic,
  IconFile,
  IconEye,
  IconDownload,
  IconTrash,
  IconCheck,
  IconX,
  IconClock,
  IconFlag,
  IconDatabase,
  IconCloudUpload,
  IconSettings,
} from "@tabler/icons-react";

function MediaPage() {
  const [media, setMedia] = useState<Media[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<any>(null);
  const [filters, setFilters] = useState<any>({ page: 1, limit: 20 });

  // Storage stats
  const [storageStats, setStorageStats] = useState<any>(null);
  const [mediaStats, setMediaStats] = useState<any>(null);

  // Dialog states
  const [selectedMedia, setSelectedMedia] = useState<Media | null>(null);
  const [showMediaDetails, setShowMediaDetails] = useState(false);
  const [showModerationDialog, setShowModerationDialog] = useState(false);
  const [showCleanupDialog, setShowCleanupDialog] = useState(false);
  const [showBulkDialog, setShowBulkDialog] = useState(false);
  const [moderationAction, setModerationAction] = useState<string>("");
  const [bulkAction, setBulkAction] = useState<string>("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [actionReason, setActionReason] = useState("");

  // Cleanup settings
  const [cleanupDays, setCleanupDays] = useState(30);
  const [cleanupMediaType, setCleanupMediaType] = useState("");

  // Active tab
  const [activeTab, setActiveTab] = useState("media");

  // Fetch media
  const fetchMedia = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getMedia(filters);
      setMedia(response.data.data);
      setPagination(response.data.pagination);
    } catch (error: any) {
      console.error("Failed to fetch media:", error);
      setError(error.response?.data?.message || "Failed to load media");
    } finally {
      setLoading(false);
    }
  };

  // Fetch storage stats
  const fetchStorageStats = async () => {
    try {
      const [storageResponse, mediaResponse] = await Promise.all([
        apiClient.getStorageStats(),
        apiClient.getMediaStats(),
      ]);
      setStorageStats(storageResponse.data);
      setMediaStats(mediaResponse.data);
    } catch (error) {
      console.error("Failed to fetch storage stats:", error);
    }
  };

  useEffect(() => {
    if (activeTab === "media") {
      fetchMedia();
    } else if (activeTab === "storage") {
      fetchStorageStats();
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
      await apiClient.bulkMediaAction({
        media_ids: selectedIds,
        action: bulkAction,
        reason: actionReason,
      });

      setShowBulkDialog(false);
      setActionReason("");
      fetchMedia();
    } catch (error: any) {
      console.error("Bulk action failed:", error);
      setError(error.response?.data?.message || "Bulk action failed");
    }
  };

  // Handle media moderation
  const handleMediaModeration = (media: Media, action: string) => {
    setSelectedMedia(media);
    setModerationAction(action);
    setShowModerationDialog(true);
  };

  // Execute media moderation
  const executeMediaModeration = async () => {
    if (!selectedMedia) return;

    try {
      await apiClient.moderateMedia(selectedMedia.id, {
        action: moderationAction,
        reason: actionReason,
      });

      setShowModerationDialog(false);
      setActionReason("");
      fetchMedia();
    } catch (error: any) {
      console.error("Media moderation failed:", error);
      setError(error.response?.data?.message || "Media moderation failed");
    }
  };

  // Handle storage cleanup
  const handleStorageCleanup = async () => {
    try {
      await apiClient.cleanupStorage({
        older_than_days: cleanupDays,
        media_type: cleanupMediaType || undefined,
      });

      setShowCleanupDialog(false);
      fetchStorageStats();
    } catch (error: any) {
      console.error("Storage cleanup failed:", error);
      setError(error.response?.data?.message || "Storage cleanup failed");
    }
  };

  // View media details
  const viewMediaDetails = (media: Media) => {
    setSelectedMedia(media);
    setShowMediaDetails(true);
  };

  // Get media type icon
  const getMediaTypeIcon = (type: string) => {
    switch (type) {
      case "image":
        return <IconPhoto className="h-4 w-4" />;
      case "video":
        return <IconVideo className="h-4 w-4" />;
      case "audio":
        return <IconMusic className="h-4 w-4" />;
      default:
        return <IconFile className="h-4 w-4" />;
    }
  };

  // Get moderation status color
  const getModerationStatusColor = (status: string) => {
    switch (status) {
      case "approved":
        return "bg-green-100 text-green-800";
      case "rejected":
        return "bg-red-100 text-red-800";
      case "flagged":
        return "bg-yellow-100 text-yellow-800";
      case "pending":
        return "bg-gray-100 text-gray-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  // Format file size
  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "preview",
      label: "Preview",
      render: (_, media: Media) => (
        <div className="w-16 h-16 relative">
          {media.media_type === "image" ? (
            <img
              src={media.thumbnail_url || media.public_url}
              alt={media.original_name}
              className="w-full h-full object-cover rounded-lg"
            />
          ) : (
            <div className="w-full h-full bg-gray-100 rounded-lg flex items-center justify-center">
              {getMediaTypeIcon(media.media_type)}
            </div>
          )}
        </div>
      ),
    },
    {
      key: "filename",
      label: "File",
      render: (_, media: Media) => (
        <div>
          <div className="font-medium text-sm">{media.original_name}</div>
          <div className="text-xs text-muted-foreground">
            {formatFileSize(media.file_size)} • {media.mime_type}
          </div>
        </div>
      ),
    },
    {
      key: "media_type",
      label: "Type",
      filterable: true,
      render: (value: string) => (
        <Badge variant="outline" className="flex items-center gap-1 w-fit">
          {getMediaTypeIcon(value)}
          <span className="capitalize">{value}</span>
        </Badge>
      ),
    },
    {
      key: "dimensions",
      label: "Dimensions",
      render: (_, media: Media) => {
        if (media.width && media.height) {
          return `${media.width} × ${media.height}`;
        }
        if (media.duration) {
          return `${Math.round(media.duration)}s`;
        }
        return "-";
      },
    },
    {
      key: "moderation_status",
      label: "Status",
      filterable: true,
      render: (value: string) => (
        <Badge className={getModerationStatusColor(value)}>
          {value === "approved" && <IconCheck className="h-3 w-3 mr-1" />}
          {value === "rejected" && <IconX className="h-3 w-3 mr-1" />}
          {value === "flagged" && <IconFlag className="h-3 w-3 mr-1" />}
          {value === "pending" && <IconClock className="h-3 w-3 mr-1" />}
          <span className="capitalize">{value}</span>
        </Badge>
      ),
    },
    {
      key: "is_processed",
      label: "Processed",
      render: (value: boolean) => (
        <Badge variant={value ? "default" : "secondary"}>
          {value ? "Yes" : "No"}
        </Badge>
      ),
    },
    {
      key: "created_at",
      label: "Uploaded",
      sortable: true,
      render: (value: string) => (
        <div className="text-sm">{new Date(value).toLocaleDateString()}</div>
      ),
    },
    {
      key: "actions",
      label: "Actions",
      render: (_, media: Media) => (
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewMediaDetails(media)}
          >
            <IconEye className="h-3 w-3" />
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => window.open(media.public_url, "_blank")}
          >
            <IconDownload className="h-3 w-3" />
          </Button>
          {media.moderation_status === "pending" && (
            <>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleMediaModeration(media, "approve")}
              >
                <IconCheck className="h-3 w-3" />
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleMediaModeration(media, "reject")}
              >
                <IconX className="h-3 w-3" />
              </Button>
            </>
          )}
        </div>
      ),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    { label: "Approve Media", action: "approve", variant: "default" as const },
    {
      label: "Reject Media",
      action: "reject",
      variant: "destructive" as const,
    },
    {
      label: "Delete Media",
      action: "delete",
      variant: "destructive" as const,
    },
  ];

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
              <h1 className="text-2xl font-bold">Media Management</h1>
              <p className="text-muted-foreground">
                Manage uploaded media files and storage
              </p>
            </div>
          </div>

          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList>
              <TabsTrigger value="media">Media Files</TabsTrigger>
              <TabsTrigger value="storage">Storage Management</TabsTrigger>
            </TabsList>

            <TabsContent value="media" className="space-y-4">
              <DataTable
                title="Media Files"
                description={`Manage ${pagination?.total || 0} media files`}
                data={media}
                columns={columns}
                loading={loading}
                pagination={pagination}
                searchPlaceholder="Search media by filename..."
                onPageChange={handlePageChange}
                onFilter={handleFilter}
                onRowSelect={handleRowSelect}
                bulkActions={bulkActions}
                onBulkAction={handleBulkAction}
                onRefresh={fetchMedia}
                onExport={() => console.log("Export media")}
              />
            </TabsContent>

            <TabsContent value="storage" className="space-y-4">
              {/* Storage Statistics */}
              {storageStats && (
                <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                  <Card>
                    <CardHeader className="pb-2">
                      <CardDescription>Total Files</CardDescription>
                      <CardTitle className="text-2xl">
                        {storageStats.total_files?.toLocaleString()}
                      </CardTitle>
                    </CardHeader>
                  </Card>
                  <Card>
                    <CardHeader className="pb-2">
                      <CardDescription>Total Storage</CardDescription>
                      <CardTitle className="text-2xl">
                        {storageStats.storage_gb?.toFixed(2)} GB
                      </CardTitle>
                    </CardHeader>
                  </Card>
                  <Card>
                    <CardHeader className="pb-2">
                      <CardDescription>Average File Size</CardDescription>
                      <CardTitle className="text-2xl">
                        {formatFileSize(storageStats.avg_file_size || 0)}
                      </CardTitle>
                    </CardHeader>
                  </Card>
                  <Card>
                    <CardHeader className="pb-2">
                      <CardDescription>Storage Used</CardDescription>
                      <CardTitle className="text-2xl">75%</CardTitle>
                    </CardHeader>
                    <CardContent className="pt-0">
                      <Progress value={75} className="w-full" />
                    </CardContent>
                  </Card>
                </div>
              )}

              {/* Storage by Type */}
              {storageStats?.storage_by_type && (
                <Card>
                  <CardHeader>
                    <CardTitle>Storage by Media Type</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      {storageStats.storage_by_type.map(
                        (item: any, index: number) => (
                          <div
                            key={index}
                            className="flex items-center justify-between"
                          >
                            <div className="flex items-center gap-2">
                              {getMediaTypeIcon(item._id)}
                              <span className="capitalize">{item._id}</span>
                            </div>
                            <div className="text-right">
                              <div className="font-medium">
                                {item.count?.toLocaleString()} files
                              </div>
                              <div className="text-sm text-muted-foreground">
                                {formatFileSize(item.total_size)}
                              </div>
                            </div>
                          </div>
                        )
                      )}
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Storage Cleanup */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <IconSettings className="h-5 w-5" />
                    Storage Cleanup
                  </CardTitle>
                  <CardDescription>
                    Clean up old media files to free up storage space
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Button
                    onClick={() => setShowCleanupDialog(true)}
                    variant="outline"
                  >
                    <IconDatabase className="h-4 w-4 mr-2" />
                    Configure Cleanup
                  </Button>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>

      {/* Media Details Dialog */}
      <Dialog open={showMediaDetails} onOpenChange={setShowMediaDetails}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Media Details</DialogTitle>
          </DialogHeader>

          {selectedMedia && (
            <div className="space-y-4">
              <div className="text-center">
                {selectedMedia.media_type === "image" ? (
                  <img
                    src={selectedMedia.public_url}
                    alt={selectedMedia.original_name}
                    className="max-w-full max-h-64 object-contain mx-auto rounded-lg"
                  />
                ) : selectedMedia.media_type === "video" ? (
                  <video
                    src={selectedMedia.public_url}
                    controls
                    className="max-w-full max-h-64 mx-auto rounded-lg"
                  />
                ) : (
                  <div className="w-32 h-32 bg-gray-100 rounded-lg flex items-center justify-center mx-auto">
                    {getMediaTypeIcon(selectedMedia.media_type)}
                  </div>
                )}
              </div>

              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <label className="font-medium">Filename</label>
                  <p>{selectedMedia.original_name}</p>
                </div>
                <div>
                  <label className="font-medium">File Size</label>
                  <p>{formatFileSize(selectedMedia.file_size)}</p>
                </div>
                <div>
                  <label className="font-medium">MIME Type</label>
                  <p>{selectedMedia.mime_type}</p>
                </div>
                <div>
                  <label className="font-medium">Media Type</label>
                  <p className="capitalize">{selectedMedia.media_type}</p>
                </div>
                {selectedMedia.width && selectedMedia.height && (
                  <div>
                    <label className="font-medium">Dimensions</label>
                    <p>
                      {selectedMedia.width} × {selectedMedia.height}
                    </p>
                  </div>
                )}
                {selectedMedia.duration && (
                  <div>
                    <label className="font-medium">Duration</label>
                    <p>{Math.round(selectedMedia.duration)}s</p>
                  </div>
                )}
                <div>
                  <label className="font-medium">Moderation Status</label>
                  <Badge
                    className={getModerationStatusColor(
                      selectedMedia.moderation_status
                    )}
                  >
                    {selectedMedia.moderation_status}
                  </Badge>
                </div>
                <div>
                  <label className="font-medium">Processed</label>
                  <p>{selectedMedia.is_processed ? "Yes" : "No"}</p>
                </div>
                <div>
                  <label className="font-medium">Storage Provider</label>
                  <p className="capitalize">{selectedMedia.storage_provider}</p>
                </div>
                <div>
                  <label className="font-medium">Uploaded</label>
                  <p>{new Date(selectedMedia.created_at).toLocaleString()}</p>
                </div>
              </div>

              {selectedMedia.moderation_reason && (
                <div>
                  <label className="font-medium">Moderation Reason</label>
                  <p className="text-sm bg-gray-50 p-2 rounded">
                    {selectedMedia.moderation_reason}
                  </p>
                </div>
              )}
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowMediaDetails(false)}
            >
              Close
            </Button>
            <Button
              variant="outline"
              onClick={() =>
                selectedMedia && window.open(selectedMedia.public_url, "_blank")
              }
            >
              <IconDownload className="h-4 w-4 mr-2" />
              Download
            </Button>
            {selectedMedia?.moderation_status === "pending" && (
              <>
                <Button
                  variant="outline"
                  onClick={() => {
                    setShowMediaDetails(false);
                    selectedMedia &&
                      handleMediaModeration(selectedMedia, "approve");
                  }}
                >
                  Approve
                </Button>
                <Button
                  variant="destructive"
                  onClick={() => {
                    setShowMediaDetails(false);
                    selectedMedia &&
                      handleMediaModeration(selectedMedia, "reject");
                  }}
                >
                  Reject
                </Button>
              </>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Moderation Dialog */}
      <Dialog
        open={showModerationDialog}
        onOpenChange={setShowModerationDialog}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {moderationAction === "approve"
                ? "Approve Media"
                : "Reject Media"}
            </DialogTitle>
            <DialogDescription>
              {moderationAction === "approve"
                ? "Approve this media file for public use."
                : "Reject this media file and prevent its use."}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label>Reason (optional)</Label>
              <Textarea
                placeholder="Enter moderation reason..."
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowModerationDialog(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={executeMediaModeration}
              variant={
                moderationAction === "reject" ? "destructive" : "default"
              }
            >
              {moderationAction === "approve" ? "Approve" : "Reject"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Storage Cleanup Dialog */}
      <Dialog open={showCleanupDialog} onOpenChange={setShowCleanupDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Storage Cleanup</DialogTitle>
            <DialogDescription>
              Configure automatic cleanup of old media files
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label>Delete files older than (days)</Label>
              <Input
                type="number"
                value={cleanupDays}
                onChange={(e) => setCleanupDays(parseInt(e.target.value))}
                min="1"
                placeholder="30"
              />
            </div>
            <div>
              <Label>Media Type (optional)</Label>
              <Select
                value={cleanupMediaType}
                onValueChange={setCleanupMediaType}
              >
                <SelectTrigger>
                  <SelectValue placeholder="All media types" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All Types</SelectItem>
                  <SelectItem value="image">Images</SelectItem>
                  <SelectItem value="video">Videos</SelectItem>
                  <SelectItem value="audio">Audio</SelectItem>
                  <SelectItem value="document">Documents</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowCleanupDialog(false)}
            >
              Cancel
            </Button>
            <Button onClick={handleStorageCleanup} variant="destructive">
              Start Cleanup
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
              You are about to {bulkAction} {selectedIds.length} media file(s).
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label>Reason</Label>
              <Textarea
                placeholder="Enter reason for this bulk action..."
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowBulkDialog(false)}>
              Cancel
            </Button>
            <Button
              onClick={executeBulkAction}
              variant={
                bulkAction === "delete" || bulkAction === "reject"
                  ? "destructive"
                  : "default"
              }
            >
              Execute {bulkAction}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SidebarProvider>
  );
}

export default withAuth(MediaPage);
