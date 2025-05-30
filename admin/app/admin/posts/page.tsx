// app/admin/posts/page.tsx
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
import { Post, TableColumn } from "@/types/admin";
import {
  IconMessage,
  IconEye,
  IconEyeOff,
  IconTrash,
  IconHeart,
  IconShare,
  IconCalendar,
  IconMapPin,
  IconPhoto,
  IconVideo,
  IconFile,
  IconFlag,
  IconPin,
} from "@tabler/icons-react";

function PostsPage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<any>(null);
  const [filters, setFilters] = useState<any>({ page: 1, limit: 20 });

  // Dialog states
  const [selectedPost, setSelectedPost] = useState<Post | null>(null);
  const [showPostDetails, setShowPostDetails] = useState(false);
  const [showActionDialog, setShowActionDialog] = useState(false);
  const [showBulkDialog, setShowBulkDialog] = useState(false);
  const [actionType, setActionType] = useState<string>("");
  const [bulkAction, setBulkAction] = useState<string>("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [actionReason, setActionReason] = useState("");

  // Fetch posts
  const fetchPosts = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getPosts(filters);
      setPosts(response.data.data);
      setPagination(response.data.pagination);
    } catch (error: any) {
      console.error("Failed to fetch posts:", error);
      setError(error.response?.data?.message || "Failed to load posts");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPosts();
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
      await apiClient.bulkPostAction({
        post_ids: selectedIds,
        action: bulkAction,
        reason: actionReason,
      });

      setShowBulkDialog(false);
      setActionReason("");
      fetchPosts();
    } catch (error: any) {
      console.error("Bulk action failed:", error);
      setError(error.response?.data?.message || "Bulk action failed");
    }
  };

  // Handle individual post action
  const handlePostAction = (post: Post, action: string) => {
    setSelectedPost(post);
    setActionType(action);
    setShowActionDialog(true);
  };

  // Execute post action
  const executePostAction = async () => {
    if (!selectedPost) return;

    try {
      switch (actionType) {
        case "hide":
          await apiClient.hidePost(selectedPost.id, actionReason);
          break;
        case "delete":
          await apiClient.deletePost(selectedPost.id, actionReason);
          break;
      }

      setShowActionDialog(false);
      setActionReason("");
      fetchPosts();
    } catch (error: any) {
      console.error("Post action failed:", error);
      setError(error.response?.data?.message || "Post action failed");
    }
  };

  // View post details
  const viewPostDetails = (post: Post) => {
    setSelectedPost(post);
    setShowPostDetails(true);
  };

  // Get post type icon
  const getPostTypeIcon = (type: string) => {
    switch (type) {
      case "image":
        return <IconPhoto className="h-4 w-4" />;
      case "video":
        return <IconVideo className="h-4 w-4" />;
      case "poll":
        return <IconMessage className="h-4 w-4" />;
      default:
        return <IconMessage className="h-4 w-4" />;
    }
  };

  // Get privacy level color
  const getPrivacyLevelColor = (level: string) => {
    switch (level) {
      case "public":
        return "bg-green-100 text-green-800";
      case "friends":
        return "bg-blue-100 text-blue-800";
      case "private":
        return "bg-gray-100 text-gray-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "content",
      label: "Post",
      render: (content: string, post: Post) => (
        <div className="max-w-md">
          <div className="flex items-center gap-2 mb-2">
            {getPostTypeIcon(post.type)}
            <Badge variant="outline" className="text-xs">
              {post.type}
            </Badge>
            {post.is_pinned && <IconPin className="h-3 w-3 text-primary" />}
            {post.is_promoted && (
              <Badge variant="default" className="text-xs">
                Promoted
              </Badge>
            )}
          </div>
          <p className="text-sm line-clamp-2 mb-2">{content}</p>
          {post.user && (
            <div className="flex items-center gap-2">
              <Avatar className="h-5 w-5">
                <AvatarImage src={post.user.profile_picture} />
                <AvatarFallback className="text-xs">
                  {post.user.username?.[0]?.toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <span className="text-xs text-muted-foreground">
                {post.user.username}
              </span>
            </div>
          )}
        </div>
      ),
    },
    {
      key: "visibility",
      label: "Privacy",
      filterable: true,
      render: (value: string) => (
        <Badge className={getPrivacyLevelColor(value)}>{value}</Badge>
      ),
    },
    {
      key: "engagement",
      label: "Engagement",
      render: (_, post: Post) => (
        <div className="space-y-1 text-xs">
          <div className="flex items-center gap-1">
            <IconHeart className="h-3 w-3" />
            <span>{post.likes_count || 0}</span>
          </div>
          <div className="flex items-center gap-1">
            <IconMessage className="h-3 w-3" />
            <span>{post.comments_count || 0}</span>
          </div>
          <div className="flex items-center gap-1">
            <IconShare className="h-3 w-3" />
            <span>{post.shares_count || 0}</span>
          </div>
        </div>
      ),
    },
    {
      key: "views_count",
      label: "Views",
      sortable: true,
      render: (value: number) => (
        <div className="flex items-center gap-1">
          <IconEye className="h-3 w-3" />
          <span>{value?.toLocaleString() || 0}</span>
        </div>
      ),
    },
    {
      key: "status",
      label: "Status",
      filterable: true,
      render: (_, post: Post) => {
        if (post.is_hidden) {
          return <Badge variant="destructive">Hidden</Badge>;
        }
        if (post.is_reported) {
          return <Badge variant="secondary">Reported</Badge>;
        }
        return <Badge variant="default">Active</Badge>;
      },
    },
    {
      key: "location",
      label: "Location",
      render: (value: string) =>
        value ? (
          <div className="flex items-center gap-1 text-xs">
            <IconMapPin className="h-3 w-3" />
            <span>{value}</span>
          </div>
        ) : (
          <span className="text-muted-foreground">-</span>
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
      render: (_, post: Post) => (
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewPostDetails(post)}
          >
            <IconEye className="h-3 w-3" />
          </Button>
          {!post.is_hidden && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => handlePostAction(post, "hide")}
            >
              <IconEyeOff className="h-3 w-3" />
            </Button>
          )}
          <Button
            size="sm"
            variant="destructive"
            onClick={() => handlePostAction(post, "delete")}
          >
            <IconTrash className="h-3 w-3" />
          </Button>
        </div>
      ),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    { label: "Hide Posts", action: "hide", variant: "default" as const },
    {
      label: "Delete Posts",
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
              <h1 className="text-2xl font-bold">Posts Management</h1>
              <p className="text-muted-foreground">
                Manage user posts and content moderation
              </p>
            </div>
          </div>

          <DataTable
            title="Posts"
            description={`Manage ${pagination?.total || 0} posts`}
            data={posts}
            columns={columns}
            loading={loading}
            pagination={pagination}
            searchPlaceholder="Search posts by content or user..."
            onPageChange={handlePageChange}
            onFilter={handleFilter}
            onRowSelect={handleRowSelect}
            bulkActions={bulkActions}
            onBulkAction={handleBulkAction}
            onRefresh={fetchPosts}
            onExport={() => console.log("Export posts")}
          />
        </div>
      </SidebarInset>

      {/* Post Details Dialog */}
      <Dialog open={showPostDetails} onOpenChange={setShowPostDetails}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Post Details</DialogTitle>
          </DialogHeader>

          {selectedPost && (
            <Tabs defaultValue="content" className="w-full">
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="content">Content</TabsTrigger>
                <TabsTrigger value="engagement">Engagement</TabsTrigger>
                <TabsTrigger value="metadata">Metadata</TabsTrigger>
              </TabsList>

              <TabsContent value="content" className="space-y-4">
                <Card>
                  <CardHeader>
                    <div className="flex items-center gap-3">
                      {selectedPost.user && (
                        <>
                          <Avatar>
                            <AvatarImage
                              src={selectedPost.user.profile_picture}
                            />
                            <AvatarFallback>
                              {selectedPost.user.username?.[0]?.toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div>
                            <p className="font-medium">
                              {selectedPost.user.username}
                            </p>
                            <p className="text-sm text-muted-foreground">
                              {new Date(
                                selectedPost.created_at
                              ).toLocaleString()}
                            </p>
                          </div>
                        </>
                      )}
                      <div className="ml-auto flex items-center gap-2">
                        {getPostTypeIcon(selectedPost.type)}
                        <Badge variant="outline">{selectedPost.type}</Badge>
                        <Badge
                          className={getPrivacyLevelColor(
                            selectedPost.visibility
                          )}
                        >
                          {selectedPost.visibility}
                        </Badge>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm mb-4">{selectedPost.content}</p>

                    {selectedPost.hashtags &&
                      selectedPost.hashtags.length > 0 && (
                        <div className="mb-4">
                          <p className="text-sm font-medium mb-2">Hashtags:</p>
                          <div className="flex flex-wrap gap-2">
                            {selectedPost.hashtags.map((tag, index) => (
                              <Badge key={index} variant="outline">
                                #{tag}
                              </Badge>
                            ))}
                          </div>
                        </div>
                      )}

                    {selectedPost.mentions &&
                      selectedPost.mentions.length > 0 && (
                        <div className="mb-4">
                          <p className="text-sm font-medium mb-2">Mentions:</p>
                          <div className="flex flex-wrap gap-2">
                            {selectedPost.mentions.map((mention, index) => (
                              <Badge key={index} variant="secondary">
                                @{mention}
                              </Badge>
                            ))}
                          </div>
                        </div>
                      )}

                    {selectedPost.media_urls &&
                      selectedPost.media_urls.length > 0 && (
                        <div className="mb-4">
                          <p className="text-sm font-medium mb-2">Media:</p>
                          <div className="grid grid-cols-2 gap-2">
                            {selectedPost.media_urls.map((url, index) => (
                              <img
                                key={index}
                                src={url}
                                alt={`Post media ${index + 1}`}
                                className="rounded-lg object-cover h-32 w-full"
                              />
                            ))}
                          </div>
                        </div>
                      )}

                    {selectedPost.location && (
                      <div className="flex items-center gap-1 text-sm text-muted-foreground">
                        <IconMapPin className="h-4 w-4" />
                        {selectedPost.location}
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="engagement" className="space-y-4">
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-lg">Likes</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {selectedPost.likes_count || 0}
                      </div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-lg">Comments</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {selectedPost.comments_count || 0}
                      </div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-lg">Shares</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {selectedPost.shares_count || 0}
                      </div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-lg">Views</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {selectedPost.views_count || 0}
                      </div>
                    </CardContent>
                  </Card>
                </div>
              </TabsContent>

              <TabsContent value="metadata" className="space-y-4">
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <label className="font-medium">Post ID</label>
                    <p>{selectedPost.id}</p>
                  </div>
                  <div>
                    <label className="font-medium">Type</label>
                    <p className="capitalize">{selectedPost.type}</p>
                  </div>
                  <div>
                    <label className="font-medium">Visibility</label>
                    <p className="capitalize">{selectedPost.visibility}</p>
                  </div>
                  <div>
                    <label className="font-medium">Created</label>
                    <p>{new Date(selectedPost.created_at).toLocaleString()}</p>
                  </div>
                  <div>
                    <label className="font-medium">Updated</label>
                    <p>{new Date(selectedPost.updated_at).toLocaleString()}</p>
                  </div>
                  <div>
                    <label className="font-medium">Status</label>
                    <div className="flex gap-2">
                      {selectedPost.is_hidden && (
                        <Badge variant="destructive">Hidden</Badge>
                      )}
                      {selectedPost.is_reported && (
                        <Badge variant="secondary">Reported</Badge>
                      )}
                      {selectedPost.is_pinned && (
                        <Badge variant="default">Pinned</Badge>
                      )}
                      {selectedPost.is_promoted && (
                        <Badge variant="default">Promoted</Badge>
                      )}
                      {!selectedPost.is_hidden && !selectedPost.is_reported && (
                        <Badge variant="default">Active</Badge>
                      )}
                    </div>
                  </div>
                </div>
              </TabsContent>
            </Tabs>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowPostDetails(false)}>
              Close
            </Button>
            {selectedPost && !selectedPost.is_hidden && (
              <Button
                variant="outline"
                onClick={() => {
                  setShowPostDetails(false);
                  handlePostAction(selectedPost, "hide");
                }}
              >
                Hide Post
              </Button>
            )}
            {selectedPost && (
              <Button
                variant="destructive"
                onClick={() => {
                  setShowPostDetails(false);
                  handlePostAction(selectedPost, "delete");
                }}
              >
                Delete Post
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
              {actionType === "hide" ? "Hide Post" : "Delete Post"}
            </DialogTitle>
            <DialogDescription>
              {actionType === "hide"
                ? "This post will be hidden from users but remain in the database."
                : "This action cannot be undone. The post will be permanently deleted."}
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
              onClick={executePostAction}
              variant={actionType === "delete" ? "destructive" : "default"}
              disabled={!actionReason.trim()}
            >
              {actionType === "hide" ? "Hide Post" : "Delete Post"}
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
              You are about to {bulkAction} {selectedIds.length} post(s).
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

export default withAuth(PostsPage);
