// app/admin/stories/page.tsx
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
import { Button } from "@/components/ui/button";
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
import { Story, TableColumn } from "@/types/admin";
import {
  IconPhoto,
  IconVideo,
  IconMusic,
  IconEye,
  IconEyeOff,
  IconTrash,
  IconHeart,
  IconClock,
  IconPlay,
  IconPause,
  IconDownload,
  IconCalendar,
} from "@tabler/icons-react";

function StoriesPage() {
  const [stories, setStories] = useState<Story[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<any>(null);
  const [filters, setFilters] = useState<any>({ page: 1, limit: 20 });

  // Dialog states
  const [selectedStory, setSelectedStory] = useState<Story | null>(null);
  const [showStoryDetails, setShowStoryDetails] = useState(false);
  const [showActionDialog, setShowActionDialog] = useState(false);
  const [showBulkDialog, setShowBulkDialog] = useState(false);
  const [actionType, setActionType] = useState<string>("");
  const [bulkAction, setBulkAction] = useState<string>("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [actionReason, setActionReason] = useState("");

  // Fetch stories
  const fetchStories = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getStories(filters);
      setStories(response.data.data);
      setPagination(response.data.pagination);
    } catch (error: any) {
      console.error("Failed to fetch stories:", error);
      setError(error.response?.data?.message || "Failed to load stories");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStories();
  }, [filters]);

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
      await apiClient.bulkStoryAction({
        story_ids: selectedIds,
        action: bulkAction,
        reason: actionReason,
      });

      setShowBulkDialog(false);
      setActionReason("");
      fetchStories();
    } catch (error: any) {
      console.error("Bulk action failed:", error);
      setError(error.response?.data?.message || "Bulk action failed");
    }
  };

  // Handle individual story action
  const handleStoryAction = (story: Story, action: string) => {
    setSelectedStory(story);
    setActionType(action);
    setShowActionDialog(true);
  };

  // Execute story action
  const executeStoryAction = async () => {
    if (!selectedStory) return;

    try {
      switch (actionType) {
        case "hide":
          await apiClient.hideStory(selectedStory.id, actionReason);
          break;
        case "delete":
          await apiClient.deleteStory(selectedStory.id, actionReason);
          break;
      }

      setShowActionDialog(false);
      setActionReason("");
      fetchStories();
    } catch (error: any) {
      console.error("Story action failed:", error);
      setError(error.response?.data?.message || "Story action failed");
    }
  };

  // View story details
  const viewStoryDetails = (story: Story) => {
    setSelectedStory(story);
    setShowStoryDetails(true);
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
        return <IconPhoto className="h-4 w-4" />;
    }
  };

  // Check if story is expired
  const isExpired = (expiresAt: string) => {
    return new Date(expiresAt) < new Date();
  };

  // Format duration
  const formatDuration = (seconds: number) => {
    if (seconds < 60) return `${seconds}s`;
    return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "story",
      label: "Story",
      render: (_, story: Story) => (
        <div className="flex items-center gap-3">
          <div className="relative w-16 h-16">
            {story.media_type === "video" ? (
              <div className="w-full h-full bg-gray-100 rounded-lg flex items-center justify-center">
                <IconVideo className="h-6 w-6 text-gray-400" />
              </div>
            ) : (
              <img
                src={story.media_url}
                alt="Story"
                className="w-full h-full object-cover rounded-lg"
              />
            )}
            {story.media_type === "video" && (
              <div className="absolute inset-0 flex items-center justify-center">
                <IconPlay className="h-4 w-4 text-white" />
              </div>
            )}
            {isExpired(story.expires_at) && (
              <div className="absolute inset-0 bg-black bg-opacity-50 rounded-lg flex items-center justify-center">
                <IconClock className="h-4 w-4 text-white" />
              </div>
            )}
          </div>
          <div>
            <div className="flex items-center gap-2 mb-1">
              {getMediaTypeIcon(story.media_type)}
              <Badge variant="outline" className="text-xs">
                {story.media_type}
              </Badge>
              <Badge variant="secondary" className="text-xs">
                {formatDuration(story.duration)}
              </Badge>
            </div>
            {story.content && (
              <p className="text-sm line-clamp-1 mb-1">{story.content}</p>
            )}
            {story.user && (
              <div className="flex items-center gap-2">
                <Avatar className="h-4 w-4">
                  <AvatarImage src={story.user.profile_picture} />
                  <AvatarFallback className="text-xs">
                    {story.user.username?.[0]?.toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <span className="text-xs text-muted-foreground">
                  {story.user.username}
                </span>
              </div>
            )}
          </div>
        </div>
      ),
    },
    {
      key: "engagement",
      label: "Engagement",
      render: (_, story: Story) => (
        <div className="space-y-1 text-xs">
          <div className="flex items-center gap-1">
            <IconEye className="h-3 w-3" />
            <span>{story.views_count || 0}</span>
          </div>
          <div className="flex items-center gap-1">
            <IconHeart className="h-3 w-3" />
            <span>{story.likes_count || 0}</span>
          </div>
        </div>
      ),
    },
    {
      key: "status",
      label: "Status",
      filterable: true,
      render: (_, story: Story) => {
        if (isExpired(story.expires_at)) {
          return <Badge variant="secondary">Expired</Badge>;
        }
        if (story.is_hidden) {
          return <Badge variant="destructive">Hidden</Badge>;
        }
        return <Badge variant="default">Active</Badge>;
      },
    },
    {
      key: "expires_at",
      label: "Expires",
      sortable: true,
      render: (value: string) => (
        <div className="text-xs">
          <div>{new Date(value).toLocaleDateString()}</div>
          <div className="text-muted-foreground">
            {new Date(value).toLocaleTimeString()}
          </div>
          {isExpired(value) && (
            <Badge variant="outline" className="text-xs mt-1">
              Expired
            </Badge>
          )}
        </div>
      ),
    },
    {
      key: "created_at",
      label: "Created",
      sortable: true,
      render: (value: string) => (
        <div className="text-xs">{new Date(value).toLocaleDateString()}</div>
      ),
    },
    {
      key: "actions",
      label: "Actions",
      render: (_, story: Story) => (
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewStoryDetails(story)}
          >
            <IconEye className="h-3 w-3" />
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => window.open(story.media_url, "_blank")}
          >
            <IconDownload className="h-3 w-3" />
          </Button>
          {!story.is_hidden && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleStoryAction(story, "hide")}
            >
              <IconEyeOff className="h-3 w-3" />
            </Button>
          )}
          <Button
            size="sm"
            variant="destructive"
            onClick={() => handleStoryAction(story, "delete")}
          >
            <IconTrash className="h-3 w-3" />
          </Button>
        </div>
      ),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    { label: "Hide Stories", action: "hide", variant: "default" as const },
    {
      label: "Delete Stories",
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
              <h1 className="text-2xl font-bold">Stories Management</h1>
              <p className="text-muted-foreground">
                Manage user stories and temporary content
              </p>
            </div>
          </div>

          <DataTable
            title="Stories"
            description={`Manage ${pagination?.total || 0} stories`}
            data={stories}
            columns={columns}
            loading={loading}
            pagination={pagination}
            searchPlaceholder="Search stories by content or user..."
            onPageChange={handlePageChange}
            onFilter={handleFilter}
            onRowSelect={handleRowSelect}
            bulkActions={bulkActions}
            onBulkAction={handleBulkAction}
            onRefresh={fetchStories}
            onExport={() => console.log("Export stories")}
          />
        </div>
      </SidebarInset>

      {/* Story Details Dialog */}
      <Dialog open={showStoryDetails} onOpenChange={setShowStoryDetails}>
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Story Details</DialogTitle>
          </DialogHeader>

          {selectedStory && (
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-4">
                    {selectedStory.user && (
                      <>
                        <Avatar>
                          <AvatarImage
                            src={selectedStory.user.profile_picture}
                          />
                          <AvatarFallback>
                            {selectedStory.user.username?.[0]?.toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <p className="font-medium">
                            {selectedStory.user.username}
                          </p>
                          <p className="text-sm text-muted-foreground">
                            {new Date(
                              selectedStory.created_at
                            ).toLocaleString()}
                          </p>
                        </div>
                      </>
                    )}
                    <div className="ml-auto flex items-center gap-2">
                      {getMediaTypeIcon(selectedStory.media_type)}
                      <Badge variant="outline">
                        {selectedStory.media_type}
                      </Badge>
                      <Badge
                        variant={
                          isExpired(selectedStory.expires_at)
                            ? "secondary"
                            : "default"
                        }
                      >
                        {isExpired(selectedStory.expires_at)
                          ? "Expired"
                          : "Active"}
                      </Badge>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  {/* Media Display */}
                  <div className="mb-4">
                    {selectedStory.media_type === "image" ? (
                      <img
                        src={selectedStory.media_url}
                        alt="Story content"
                        className="rounded-lg max-h-96 w-full object-contain"
                      />
                    ) : selectedStory.media_type === "video" ? (
                      <video
                        src={selectedStory.media_url}
                        controls
                        className="rounded-lg max-h-96 w-full"
                      />
                    ) : (
                      <div className="bg-gray-100 rounded-lg p-8 text-center">
                        <IconMusic className="h-12 w-12 mx-auto text-gray-400 mb-2" />
                        <p className="text-gray-600">Audio Story</p>
                        <audio
                          src={selectedStory.media_url}
                          controls
                          className="mt-2"
                        />
                      </div>
                    )}
                  </div>

                  {/* Story Content */}
                  {selectedStory.content && (
                    <div className="mb-4">
                      <h4 className="font-medium mb-2">Content</h4>
                      <p className="text-sm bg-gray-50 p-3 rounded">
                        {selectedStory.content}
                      </p>
                    </div>
                  )}

                  {/* Story Details */}
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <label className="font-medium">Duration</label>
                      <p>{formatDuration(selectedStory.duration)}</p>
                    </div>
                    <div>
                      <label className="font-medium">Views</label>
                      <p>{selectedStory.views_count?.toLocaleString() || 0}</p>
                    </div>
                    <div>
                      <label className="font-medium">Likes</label>
                      <p>{selectedStory.likes_count?.toLocaleString() || 0}</p>
                    </div>
                    <div>
                      <label className="font-medium">Expires</label>
                      <p>
                        {new Date(selectedStory.expires_at).toLocaleString()}
                      </p>
                    </div>
                  </div>

                  {/* Style Information */}
                  {(selectedStory.background_color ||
                    selectedStory.text_color ||
                    selectedStory.font_family) && (
                    <div className="mt-4">
                      <h4 className="font-medium mb-2">Style</h4>
                      <div className="grid grid-cols-3 gap-4 text-sm">
                        {selectedStory.background_color && (
                          <div>
                            <label className="font-medium">Background</label>
                            <div className="flex items-center gap-2">
                              <div
                                className="w-4 h-4 rounded border"
                                style={{
                                  backgroundColor:
                                    selectedStory.background_color,
                                }}
                              />
                              <span>{selectedStory.background_color}</span>
                            </div>
                          </div>
                        )}
                        {selectedStory.text_color && (
                          <div>
                            <label className="font-medium">Text Color</label>
                            <div className="flex items-center gap-2">
                              <div
                                className="w-4 h-4 rounded border"
                                style={{
                                  backgroundColor: selectedStory.text_color,
                                }}
                              />
                              <span>{selectedStory.text_color}</span>
                            </div>
                          </div>
                        )}
                        {selectedStory.font_family && (
                          <div>
                            <label className="font-medium">Font</label>
                            <p>{selectedStory.font_family}</p>
                          </div>
                        )}
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowStoryDetails(false)}
            >
              Close
            </Button>
            <Button
              variant="outline"
              onClick={() =>
                selectedStory && window.open(selectedStory.media_url, "_blank")
              }
            >
              <IconDownload className="h-4 w-4 mr-2" />
              Download
            </Button>
            {selectedStory && !selectedStory.is_hidden && (
              <Button
                variant="outline"
                onClick={() => {
                  setShowStoryDetails(false);
                  handleStoryAction(selectedStory, "hide");
                }}
              >
                Hide Story
              </Button>
            )}
            {selectedStory && (
              <Button
                variant="destructive"
                onClick={() => {
                  setShowStoryDetails(false);
                  handleStoryAction(selectedStory, "delete");
                }}
              >
                Delete Story
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Action Dialog */}
      <Dialog open={showActionDialog} onOpenChange={setShowActionDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {actionType === "hide" ? "Hide Story" : "Delete Story"}
            </DialogTitle>
            <DialogDescription>
              {actionType === "hide"
                ? "This story will be hidden from users but remain in the database."
                : "This action cannot be undone. The story will be permanently deleted."}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">Reason</label>
              <Textarea
                placeholder="Enter reason for this action..."
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
                required
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowActionDialog(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={executeStoryAction}
              variant={actionType === "delete" ? "destructive" : "default"}
              disabled={!actionReason.trim()}
            >
              {actionType === "hide" ? "Hide Story" : "Delete Story"}
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
              You are about to {bulkAction} {selectedIds.length} story(s).
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">Reason</label>
              <Textarea
                placeholder="Enter reason for this bulk action..."
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
                required
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowBulkDialog(false)}>
              Cancel
            </Button>
            <Button
              onClick={executeBulkAction}
              variant={bulkAction === "delete" ? "destructive" : "default"}
              disabled={!actionReason.trim()}
            >
              Execute {bulkAction}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SidebarProvider>
  );
}

export default withAuth(StoriesPage);
