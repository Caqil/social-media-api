// admin/app/admin/posts/page.tsx - Complete Posts Management
"use client";

import { useState, useEffect, useCallback } from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { DataTable } from "@/components/data-table";
import { SiteHeader } from "@/components/site-header";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Textarea } from "@/components/ui/textarea";
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { withAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import {
  Post,
  PrivacyLevel,
  TableColumn,
  PaginationMeta,
  UserResponse,
} from "@/types/admin";
import {
  IconEye,
  IconEyeOff,
  IconTrash,
  IconDotsVertical,
  IconHeart,
  IconMessage,
  IconShare,
  IconLoader2,
  IconAlertCircle,
  IconMessages,
  IconPhoto,
  IconVideo,
  IconFile,
  IconSearch,
  IconRefresh,
  IconDownload,
  IconFlag,
  IconCalendar,
  IconUsers,
  IconGlobe,
  IconLock,
  IconBan,
  IconCheck,
  IconTrendingUp,
  IconActivity,
} from "@tabler/icons-react";

interface PostsPageState {
  posts: Post[];
  loading: boolean;
  error: string | null;
  pagination: PaginationMeta | undefined;
  filters: {
    search: string;
    type: string;
    visibility: string;
    is_reported: string;
    is_hidden: string;
    user_id: string;
    page: number;
    limit: number;
    sort_by?: string;
    sort_order?: "asc" | "desc";
  };
  selectedPosts: string[];
}

interface DialogState {
  viewPost: boolean;
  deletePost: boolean;
  hidePost: boolean;
  selectedPost: Post | null;
}

const initialFilters = {
  search: "",
  type: "all",
  visibility: "all",
  is_reported: "all",
  is_hidden: "all",
  user_id: "",
  page: 1,
  limit: 20,
};

// Utility function to extract user data from post
const getUserFromPost = (post: Post) => {
  const user = post.user || (post as any).author;
  const userId = post.user_id;

  if (user) {
    return {
      displayName: user.first_name
        ? `${user.first_name} ${user.last_name || ""}`.trim()
        : user.username || "Unknown User",
      username: user.username || `user_${userId?.slice(-4) || "unknown"}`,
      profilePicture: user.profile_picture,
      isVerified: user.is_verified || false,
      initials: (
        user.first_name?.[0] ||
        user.username?.[0] ||
        "U"
      ).toUpperCase(),
    };
  }

  // Fallback when no user data is available
  return {
    displayName: userId ? `User ${userId.slice(-8)}` : "Unknown User",
    username: userId ? `user_${userId.slice(-4)}` : "unknown",
    profilePicture: undefined,
    isVerified: false,
    initials: "U",
    userId,
  };
};

function PostsPage() {
  const [state, setState] = useState<PostsPageState>({
    posts: [],
    loading: true,
    error: null,
    pagination: undefined,
    filters: initialFilters,
    selectedPosts: [],
  });

  const [dialogs, setDialogs] = useState<DialogState>({
    viewPost: false,
    deletePost: false,
    hidePost: false,
    selectedPost: null,
  });

  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [actionReason, setActionReason] = useState("");

  // Fetch posts with proper API parameters
  const fetchPosts = useCallback(async (filters = state.filters) => {
    try {
      setState((prev) => ({ ...prev, loading: true, error: null }));

      const params: any = {
        page: filters.page,
        limit: filters.limit,
        include_user: true, // Request user data to be included
      };

      // Add search parameter
      if (filters.search) params.search = filters.search;

      // Add type filter
      if (filters.type && filters.type !== "all") params.type = filters.type;

      // Add visibility filter
      if (filters.visibility && filters.visibility !== "all") {
        params.visibility = filters.visibility;
      }

      // Add reported filter
      if (filters.is_reported && filters.is_reported !== "all") {
        params.is_reported = filters.is_reported === "true";
      }

      // Add hidden filter
      if (filters.is_hidden && filters.is_hidden !== "all") {
        params.is_hidden = filters.is_hidden === "true";
      }

      // Add user filter
      if (filters.user_id) params.user_id = filters.user_id;

      // Add sorting
      if (filters.sort_by) {
        params.sort_by = filters.sort_by;
        params.sort_order = filters.sort_order || "desc";
      }

      console.log("📡 Fetching posts with params:", params);
      const response = await apiClient.getPosts(params);

      // Debug: Check if user data is included
      if (response.data && response.data.length > 0) {
        const firstPost = response.data[0];
        console.log("🔍 Sample post:", {
          id: firstPost.id,
          user_id: firstPost.user_id,
          has_user: !!firstPost.user,
          has_author: !!(firstPost as any).author,
          user_data: firstPost.user || (firstPost as any).author,
        });
      }

      setState((prev) => ({
        ...prev,
        posts: response.data || [],
        pagination: response.pagination || undefined,
        loading: false,
      }));
    } catch (error: any) {
      console.error("❌ Failed to fetch posts:", error);
      setState((prev) => ({
        ...prev,
        loading: false,
        error: error.message || "Failed to fetch posts",
      }));
    }
  }, []);

  // Initial load
  useEffect(() => {
    fetchPosts();
  }, []);

  // Handle search with debounce
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (state.filters.search !== undefined) {
        fetchPosts({ ...state.filters, page: 1 });
      }
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [state.filters.search, fetchPosts]);

  // Handle filter changes
  const handleFilterChange = (key: string, value: any) => {
    const newFilters = { ...state.filters, [key]: value, page: 1 };
    setState((prev) => ({ ...prev, filters: newFilters }));

    if (key !== "search") {
      fetchPosts(newFilters);
    }
  };

  // Handle pagination
  const handlePageChange = (page: number) => {
    const newFilters = { ...state.filters, page };
    setState((prev) => ({ ...prev, filters: newFilters }));
    fetchPosts(newFilters);
  };

  // Handle sort
  const handleSort = (column: string, direction: "asc" | "desc") => {
    const newFilters = {
      ...state.filters,
      sort_by: column,
      sort_order: direction,
      page: 1,
    };
    setState((prev) => ({ ...prev, filters: newFilters }));
    fetchPosts(newFilters);
  };

  // Handle row selection
  const handleRowSelect = (selectedRows: string[]) => {
    setState((prev) => ({ ...prev, selectedPosts: selectedRows }));
  };

  // Open dialog
  const openDialog = (type: keyof DialogState, post: Post | null = null) => {
    setDialogs((prev) => ({ ...prev, [type]: true, selectedPost: post }));
  };

  // Close dialog
  const closeDialog = (type: keyof DialogState) => {
    setDialogs((prev) => ({ ...prev, [type]: false, selectedPost: null }));
    setFormError(null);
    setActionReason("");
  };

  // Handle post deletion
  const handleDeletePost = async () => {
    if (!dialogs.selectedPost) return;

    setFormLoading(true);
    try {
      await apiClient.deletePost(
        dialogs.selectedPost.id,
        actionReason || "Deleted by admin"
      );
      closeDialog("deletePost");
      fetchPosts();
    } catch (error: any) {
      setFormError(error.message || "Failed to delete post");
    } finally {
      setFormLoading(false);
    }
  };

  // Handle post hide/show
  const handleTogglePostVisibility = async () => {
    if (!dialogs.selectedPost) return;

    setFormLoading(true);
    try {
      await apiClient.hidePost(
        dialogs.selectedPost.id,
        actionReason || "Hidden by admin"
      );
      closeDialog("hidePost");
      fetchPosts();
    } catch (error: any) {
      setFormError(error.message || "Failed to update post visibility");
    } finally {
      setFormLoading(false);
    }
  };

  // Handle bulk actions
  const handleBulkAction = async (action: string, selectedIds: string[]) => {
    setFormLoading(true);
    try {
      await apiClient.bulkPostAction({
        post_ids: selectedIds,
        action,
        reason: "Bulk action by admin",
      });
      setState((prev) => ({ ...prev, selectedPosts: [] }));
      fetchPosts();
    } catch (error: any) {
      setState((prev) => ({
        ...prev,
        error: error.message || "Bulk action failed",
      }));
    } finally {
      setFormLoading(false);
    }
  };

  // Handle refresh
  const handleRefresh = () => {
    fetchPosts();
  };

  // Handle export
  const handleExport = async () => {
    try {
      await apiClient.exportPosts();
      console.log("✅ Export initiated");
    } catch (error) {
      console.error("❌ Export failed:", error);
    }
  };

  // Format post type icon
  const getPostTypeIcon = (type: string) => {
    switch (type?.toLowerCase()) {
      case "image":
        return <IconPhoto className="h-4 w-4" />;
      case "video":
        return <IconVideo className="h-4 w-4" />;
      case "poll":
        return <IconActivity className="h-4 w-4" />;
      case "shared":
        return <IconShare className="h-4 w-4" />;
      default:
        return <IconMessage className="h-4 w-4" />;
    }
  };

  // Format visibility icon
  const getVisibilityIcon = (visibility: PrivacyLevel) => {
    switch (visibility) {
      case PrivacyLevel.PUBLIC:
        return <IconGlobe className="h-3 w-3 mr-1" />;
      case PrivacyLevel.FRIENDS:
        return <IconUsers className="h-3 w-3 mr-1" />;
      case PrivacyLevel.PRIVATE:
        return <IconLock className="h-3 w-3 mr-1" />;
      default:
        return <IconGlobe className="h-3 w-3 mr-1" />;
    }
  };

  // Format post status
  const formatPostStatus = (post: Post) => {
    if (post.is_hidden) return "Hidden";
    if (post.is_reported) return "Reported";
    return "Active";
  };

  // Get status badge variant
  const getStatusBadgeVariant = (
    post: Post
  ): "default" | "secondary" | "destructive" => {
    if (post.is_hidden) return "destructive";
    if (post.is_reported) return "secondary";
    return "default";
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "user",
      label: "Author",
      render: (_, post: Post) => {
        const userInfo = getUserFromPost(post);

        return (
          <div className="flex items-center gap-3">
            <Avatar className="h-8 w-8">
              <AvatarImage
                src={userInfo.profilePicture}
                alt={userInfo.username}
              />
              <AvatarFallback>{userInfo.initials}</AvatarFallback>
            </Avatar>
            <div className="min-w-0">
              <div className="font-medium flex items-center gap-2">
                {userInfo.displayName}
                {userInfo.isVerified && (
                  <div className="inline-block w-3 h-3 bg-blue-500 rounded-full flex items-center justify-center">
                    <div className="w-1.5 h-1.5 bg-white rounded-full" />
                  </div>
                )}
              </div>
              <div className="text-sm text-muted-foreground">
                @{userInfo.username}
                {!(post.user || (post as any).author) && post.user_id && (
                  <span
                    className="ml-1 text-xs text-orange-500"
                    title="User data not loaded"
                  >
                    (No user data)
                  </span>
                )}
              </div>
            </div>
          </div>
        );
      },
    },
    {
      key: "content",
      label: "Content",
      render: (content: string, post: Post) => (
        <div className="max-w-md">
          <div className="flex items-center gap-2 mb-1">
            {getPostTypeIcon(post.type)}
            <span className="text-xs text-muted-foreground capitalize">
              {post.type || "text"}
            </span>
          </div>
          <div className="text-sm line-clamp-2">{content || "No content"}</div>
          {post.media_urls && post.media_urls.length > 0 && (
            <div className="text-xs text-muted-foreground mt-1">
              +{post.media_urls.length} media file
              {post.media_urls.length > 1 ? "s" : ""}
            </div>
          )}
        </div>
      ),
    },
    {
      key: "visibility",
      label: "Visibility",
      sortable: true,
      filterable: true,
      render: (visibility: PrivacyLevel) => (
        <Badge variant="outline" className="capitalize">
          {getVisibilityIcon(visibility)}
          {visibility || "public"}
        </Badge>
      ),
    },
    {
      key: "engagement",
      label: "Engagement",
      render: (_, post: Post) => (
        <div className="text-sm space-y-1">
          <div className="flex items-center gap-4">
            <span className="flex items-center gap-1">
              <IconHeart className="h-3 w-3" />
              {(post.likes_count || 0).toLocaleString()}
            </span>
            <span className="flex items-center gap-1">
              <IconMessage className="h-3 w-3" />
              {(post.comments_count || 0).toLocaleString()}
            </span>
            <span className="flex items-center gap-1">
              <IconShare className="h-3 w-3" />
              {(post.shares_count || 0).toLocaleString()}
            </span>
          </div>
          {post.views_count !== undefined && (
            <div className="text-muted-foreground flex items-center gap-1">
              <IconEye className="h-3 w-3" />
              {post.views_count.toLocaleString()} views
            </div>
          )}
        </div>
      ),
    },
    {
      key: "status",
      label: "Status",
      sortable: true,
      filterable: true,
      render: (_, post: Post) => (
        <div className="space-y-1">
          <Badge variant={getStatusBadgeVariant(post)}>
            {formatPostStatus(post)}
          </Badge>
          {post.is_reported && (
            <div className="flex items-center gap-1 text-xs text-orange-600">
              <IconFlag className="h-3 w-3" />
              Reported
            </div>
          )}
        </div>
      ),
    },
    {
      key: "created_at",
      label: "Created",
      sortable: true,
      render: (value: string) => (
        <div className="flex items-center gap-2 text-sm">
          <IconCalendar className="h-4 w-4 text-muted-foreground" />
          <div>
            <div>
              {value ? new Date(value).toLocaleDateString() : "Unknown"}
            </div>
            <div className="text-xs text-muted-foreground">
              {value ? new Date(value).toLocaleTimeString() : ""}
            </div>
          </div>
        </div>
      ),
    },
    {
      key: "actions",
      label: "",
      width: "w-12",
      render: (_, post: Post) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
              <IconDotsVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => openDialog("viewPost", post)}>
              <IconEye className="h-4 w-4 mr-2" />
              View Details
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => openDialog("hidePost", post)}
              className={post.is_hidden ? "text-green-600" : "text-orange-600"}
            >
              {post.is_hidden ? (
                <>
                  <IconEye className="h-4 w-4 mr-2" />
                  Show Post
                </>
              ) : (
                <>
                  <IconEyeOff className="h-4 w-4 mr-2" />
                  Hide Post
                </>
              )}
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => openDialog("deletePost", post)}
              className="text-red-600"
            >
              <IconTrash className="h-4 w-4 mr-2" />
              Delete Post
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    { label: "Show Posts", action: "show" },
    { label: "Hide Posts", action: "hide" },
    {
      label: "Delete Posts",
      action: "delete",
      variant: "destructive" as const,
    },
  ];

  if (state.error && !state.loading) {
    return (
      <SidebarProvider>
        <AppSidebar variant="inset" />
        <SidebarInset>
          <SiteHeader />
          <div className="flex h-screen items-center justify-center p-6">
            <Alert variant="destructive" className="max-w-md">
              <IconAlertCircle className="h-5 w-5" />
              <AlertDescription className="space-y-4">
                <div>Failed to load posts: {state.error}</div>
                <Button
                  onClick={handleRefresh}
                  variant="outline"
                  className="w-full"
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
    <SidebarProvider>
      <AppSidebar variant="inset" />
      <SidebarInset>
        <SiteHeader />

        <div className="flex flex-1 flex-col gap-4 p-4 md:gap-6 md:p-6">
          {/* Header */}
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold">Posts</h1>
              <p className="text-muted-foreground">
                Manage user posts and content moderation
              </p>
            </div>
          </div>

          {/* Stats Cards */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Total Posts
                </CardTitle>
                <IconMessages className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {state.pagination?.total?.toLocaleString() || "0"}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Active Posts
                </CardTitle>
                <IconCheck className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {
                    state.posts.filter(
                      (p) => p && !p.is_hidden && !p.is_reported
                    ).length
                  }
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Reported Posts
                </CardTitle>
                <IconFlag className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {state.posts.filter((p) => p && p.is_reported).length}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Hidden Posts
                </CardTitle>
                <IconEyeOff className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {state.posts.filter((p) => p && p.is_hidden).length}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Filters */}
          <Card>
            <CardHeader>
              <CardTitle>Filters</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
                <div>
                  <Label htmlFor="search">Search</Label>
                  <div className="relative">
                    <IconSearch className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="search"
                      placeholder="Search posts..."
                      value={state.filters.search}
                      onChange={(e) =>
                        handleFilterChange("search", e.target.value)
                      }
                      className="pl-9"
                    />
                  </div>
                </div>
                <div>
                  <Label htmlFor="type">Type</Label>
                  <Select
                    value={state.filters.type}
                    onValueChange={(value) => handleFilterChange("type", value)}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All types" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All types</SelectItem>
                      <SelectItem value="text">Text</SelectItem>
                      <SelectItem value="image">Image</SelectItem>
                      <SelectItem value="video">Video</SelectItem>
                      <SelectItem value="poll">Poll</SelectItem>
                      <SelectItem value="shared">Shared</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="visibility">Visibility</Label>
                  <Select
                    value={state.filters.visibility}
                    onValueChange={(value) =>
                      handleFilterChange("visibility", value)
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All visibility" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All visibility</SelectItem>
                      <SelectItem value="public">Public</SelectItem>
                      <SelectItem value="friends">Friends</SelectItem>
                      <SelectItem value="private">Private</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="reported">Reported</Label>
                  <Select
                    value={state.filters.is_reported}
                    onValueChange={(value) =>
                      handleFilterChange("is_reported", value)
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All posts" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All posts</SelectItem>
                      <SelectItem value="false">Not reported</SelectItem>
                      <SelectItem value="true">Reported</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="hidden">Hidden</Label>
                  <Select
                    value={state.filters.is_hidden}
                    onValueChange={(value) =>
                      handleFilterChange("is_hidden", value)
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All posts" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All posts</SelectItem>
                      <SelectItem value="false">Visible</SelectItem>
                      <SelectItem value="true">Hidden</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Data Table */}
          <DataTable
            data={state.posts}
            columns={columns}
            loading={state.loading}
            pagination={state.pagination}
            onPageChange={handlePageChange}
            onSort={handleSort}
            onRowSelect={handleRowSelect}
            bulkActions={bulkActions}
            onBulkAction={handleBulkAction}
            onRefresh={handleRefresh}
            onExport={handleExport}
            title="Posts Management"
            description="View and manage user posts"
            emptyMessage="No posts found"
            searchPlaceholder="Search posts..."
          />
        </div>

        {/* View Post Dialog */}
        <Dialog
          open={dialogs.viewPost}
          onOpenChange={() => closeDialog("viewPost")}
        >
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Post Details</DialogTitle>
              <DialogDescription>
                View detailed information about this post
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedPost && (
              <div className="space-y-6">
                {/* Post Author */}
                <div className="flex items-center space-x-4">
                  <Avatar className="h-12 w-12">
                    <AvatarImage
                      src={getUserFromPost(dialogs.selectedPost).profilePicture}
                    />
                    <AvatarFallback>
                      {getUserFromPost(dialogs.selectedPost).initials}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <h3 className="font-semibold">
                      {getUserFromPost(dialogs.selectedPost).displayName}
                    </h3>
                    <p className="text-sm text-muted-foreground">
                      @{getUserFromPost(dialogs.selectedPost).username}
                    </p>
                  </div>
                  <div className="ml-auto space-y-1">
                    <Badge
                      variant={getStatusBadgeVariant(dialogs.selectedPost)}
                    >
                      {formatPostStatus(dialogs.selectedPost)}
                    </Badge>
                    <Badge variant="outline" className="capitalize">
                      {getVisibilityIcon(dialogs.selectedPost.visibility)}
                      {dialogs.selectedPost.visibility || "public"}
                    </Badge>
                  </div>
                </div>

                {/* Post Content */}
                <div className="space-y-4">
                  <div>
                    <Label className="text-sm font-medium">Content</Label>
                    <div className="mt-1 p-3 bg-muted rounded-lg">
                      <div className="flex items-center gap-2 mb-2">
                        {getPostTypeIcon(dialogs.selectedPost.type)}
                        <span className="text-sm text-muted-foreground capitalize">
                          {dialogs.selectedPost.type || "text"} post
                        </span>
                      </div>
                      <p className="text-sm whitespace-pre-wrap">
                        {dialogs.selectedPost.content || "No content"}
                      </p>
                    </div>
                  </div>

                  {/* Media */}
                  {dialogs.selectedPost.media_urls &&
                    dialogs.selectedPost.media_urls.length > 0 && (
                      <div>
                        <Label className="text-sm font-medium">Media</Label>
                        <div className="mt-1 grid grid-cols-2 gap-2">
                          {dialogs.selectedPost.media_urls
                            .slice(0, 4)
                            .map((url, index) => (
                              <div
                                key={index}
                                className="aspect-square bg-muted rounded-lg flex items-center justify-center"
                              >
                                <IconFile className="h-8 w-8 text-muted-foreground" />
                              </div>
                            ))}
                        </div>
                        {dialogs.selectedPost.media_urls.length > 4 && (
                          <p className="text-xs text-muted-foreground mt-2">
                            +{dialogs.selectedPost.media_urls.length - 4} more
                            files
                          </p>
                        )}
                      </div>
                    )}

                  {/* Hashtags */}
                  {dialogs.selectedPost.hashtags &&
                    dialogs.selectedPost.hashtags.length > 0 && (
                      <div>
                        <Label className="text-sm font-medium">Hashtags</Label>
                        <div className="mt-1 flex flex-wrap gap-1">
                          {dialogs.selectedPost.hashtags.map((tag, index) => (
                            <Badge
                              key={index}
                              variant="secondary"
                              className="text-xs"
                            >
                              #{tag}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}

                  {/* Location */}
                  {dialogs.selectedPost.location && (
                    <div>
                      <Label className="text-sm font-medium">Location</Label>
                      <p className="text-sm text-muted-foreground mt-1">
                        {dialogs.selectedPost.location}
                      </p>
                    </div>
                  )}
                </div>

                {/* Post Stats */}
                <div className="grid grid-cols-4 gap-4 text-center">
                  <div>
                    <p className="text-2xl font-bold">
                      {(dialogs.selectedPost.likes_count || 0).toLocaleString()}
                    </p>
                    <p className="text-sm text-muted-foreground">Likes</p>
                  </div>
                  <div>
                    <p className="text-2xl font-bold">
                      {(
                        dialogs.selectedPost.comments_count || 0
                      ).toLocaleString()}
                    </p>
                    <p className="text-sm text-muted-foreground">Comments</p>
                  </div>
                  <div>
                    <p className="text-2xl font-bold">
                      {(
                        dialogs.selectedPost.shares_count || 0
                      ).toLocaleString()}
                    </p>
                    <p className="text-sm text-muted-foreground">Shares</p>
                  </div>
                  <div>
                    <p className="text-2xl font-bold">
                      {(dialogs.selectedPost.views_count || 0).toLocaleString()}
                    </p>
                    <p className="text-sm text-muted-foreground">Views</p>
                  </div>
                </div>

                {/* Post Metadata */}
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <Label className="text-sm font-medium">Created</Label>
                    <p className="text-muted-foreground">
                      {dialogs.selectedPost.created_at
                        ? new Date(
                            dialogs.selectedPost.created_at
                          ).toLocaleString()
                        : "Unknown"}
                    </p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Updated</Label>
                    <p className="text-muted-foreground">
                      {dialogs.selectedPost.updated_at
                        ? new Date(
                            dialogs.selectedPost.updated_at
                          ).toLocaleString()
                        : "Never"}
                    </p>
                  </div>
                </div>
              </div>
            )}

            <DialogFooter>
              <Button variant="outline" onClick={() => closeDialog("viewPost")}>
                Close
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Delete Post Dialog */}
        <Dialog
          open={dialogs.deletePost}
          onOpenChange={() => closeDialog("deletePost")}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete Post</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this post? This action cannot be
                undone.
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedPost && (
              <div className="space-y-4">
                <div className="p-4 bg-muted rounded-lg">
                  <div className="flex items-center gap-3 mb-2">
                    <Avatar className="h-8 w-8">
                      <AvatarImage
                        src={
                          getUserFromPost(dialogs.selectedPost).profilePicture
                        }
                      />
                      <AvatarFallback>
                        {getUserFromPost(dialogs.selectedPost).initials}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="font-medium text-sm">
                        {getUserFromPost(dialogs.selectedPost).displayName}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {dialogs.selectedPost.created_at
                          ? new Date(
                              dialogs.selectedPost.created_at
                            ).toLocaleDateString()
                          : "Unknown date"}
                      </p>
                    </div>
                  </div>
                  <p className="text-sm line-clamp-3">
                    {dialogs.selectedPost.content || "No content"}
                  </p>
                </div>

                <div>
                  <Label htmlFor="reason">Reason for deletion (optional)</Label>
                  <Textarea
                    id="reason"
                    value={actionReason}
                    onChange={(e) => setActionReason(e.target.value)}
                    placeholder="Provide a reason for deleting this post..."
                  />
                </div>
              </div>
            )}

            {formError && (
              <Alert variant="destructive">
                <IconAlertCircle className="h-4 w-4" />
                <AlertDescription>{formError}</AlertDescription>
              </Alert>
            )}

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => closeDialog("deletePost")}
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDeletePost}
                disabled={formLoading}
              >
                {formLoading && (
                  <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Delete Post
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Hide/Show Post Dialog */}
        <Dialog
          open={dialogs.hidePost}
          onOpenChange={() => closeDialog("hidePost")}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                {dialogs.selectedPost?.is_hidden ? "Show Post" : "Hide Post"}
              </DialogTitle>
              <DialogDescription>
                {dialogs.selectedPost?.is_hidden
                  ? "Make this post visible to users again."
                  : "Hide this post from users while keeping it in the system."}
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedPost && (
              <div className="space-y-4">
                <div className="p-4 bg-muted rounded-lg">
                  <div className="flex items-center gap-3 mb-2">
                    <Avatar className="h-8 w-8">
                      <AvatarImage
                        src={
                          getUserFromPost(dialogs.selectedPost).profilePicture
                        }
                      />
                      <AvatarFallback>
                        {getUserFromPost(dialogs.selectedPost).initials}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="font-medium text-sm">
                        {getUserFromPost(dialogs.selectedPost).displayName}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {dialogs.selectedPost.created_at
                          ? new Date(
                              dialogs.selectedPost.created_at
                            ).toLocaleDateString()
                          : "Unknown date"}
                      </p>
                    </div>
                  </div>
                  <p className="text-sm line-clamp-3">
                    {dialogs.selectedPost.content || "No content"}
                  </p>
                </div>

                {!dialogs.selectedPost.is_hidden && (
                  <div>
                    <Label htmlFor="reason">Reason for hiding</Label>
                    <Textarea
                      id="reason"
                      value={actionReason}
                      onChange={(e) => setActionReason(e.target.value)}
                      placeholder="Provide a reason for hiding this post..."
                      required
                    />
                  </div>
                )}
              </div>
            )}

            {formError && (
              <Alert variant="destructive">
                <IconAlertCircle className="h-4 w-4" />
                <AlertDescription>{formError}</AlertDescription>
              </Alert>
            )}

            <DialogFooter>
              <Button variant="outline" onClick={() => closeDialog("hidePost")}>
                Cancel
              </Button>
              <Button
                onClick={handleTogglePostVisibility}
                disabled={
                  formLoading ||
                  (!dialogs.selectedPost?.is_hidden && !actionReason)
                }
              >
                {formLoading && (
                  <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                {dialogs.selectedPost?.is_hidden ? "Show Post" : "Hide Post"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </SidebarInset>
    </SidebarProvider>
  );
}

export default withAuth(PostsPage);

// Export for use in routing
export { PostsPage };
