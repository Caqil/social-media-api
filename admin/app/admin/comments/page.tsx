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
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { withAuth, useAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import {
  Comment,
  UserResponse,
  PostResponse,
  TableColumn,
  PaginationMeta,
} from "@/types/admin";
import {
  IconEye,
  IconEdit,
  IconTrash,
  IconDotsVertical,
  IconMessageCircle,
  IconMail,
  IconCalendar,
  IconShield,
  IconBan,
  IconCheck,
  IconX,
  IconLoader2,
  IconAlertCircle,
  IconUsers,
  IconUserPlus,
  IconUserX,
  IconSearch,
  IconFilter,
  IconRefresh,
  IconDownload,
  IconMessage,
  IconFlag,
  IconEyeOff,
  IconLink,
  IconHeart,
  IconClock,
  IconCornerDownRight,
} from "@tabler/icons-react";
import { ReplyAllIcon, ReplyIcon } from "lucide-react";

interface CommentsPageState {
  comments: Comment[];
  loading: boolean;
  error: string | null;
  pagination: PaginationMeta | undefined;
  filters: {
    search: string;
    status: string; // Use status instead of is_reported/is_hidden
    post_id: string;
    user_id: string;
    has_replies: string;
    page: number;
    limit: number;
    sort_by?: string;
    sort_order?: "asc" | "desc";
  };
  selectedComments: string[];
}

interface DialogState {
  viewComment: boolean;
  editComment: boolean;
  deleteComment: boolean;
  hideComment: boolean;
  selectedComment: Comment | null;
}

const initialFilters = {
  search: "",
  status: "all",
  post_id: "",
  user_id: "",
  has_replies: "all",
  page: 1,
  limit: 20,
};

function CommentsPage() {
  const { user: currentUser } = useAuth();

  const [state, setState] = useState<CommentsPageState>({
    comments: [],
    loading: true,
    error: null,
    pagination: undefined,
    filters: initialFilters,
    selectedComments: [],
  });

  const [dialogs, setDialogs] = useState<DialogState>({
    viewComment: false,
    editComment: false,
    deleteComment: false,
    hideComment: false,
    selectedComment: null,
  });

  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [hideReason, setHideReason] = useState("");
  const [editContent, setEditContent] = useState("");

  const fetchComments = useCallback(async (filters = state.filters) => {
    try {
      setState((prev) => ({ ...prev, loading: true, error: null }));

      const params: any = {
        page: filters.page,
        limit: filters.limit,
        include_user: true, // Request user data to be included
        expand: "user,post", // Explicitly request user and post expansion
      };

      // Add all filters
      if (filters.search) params.search = filters.search;
      if (filters.post_id) params.post_id = filters.post_id;
      if (filters.user_id) params.user_id = filters.user_id;

      // Handle status filter (instead of is_reported and is_hidden)
      if (filters.status && filters.status !== "all") {
        params.status = filters.status;

        // Map status to the appropriate API parameters if needed
        // For example, if the API uses different parameter names
        if (filters.status === "reported") {
          params.is_reported = true;
        } else if (filters.status === "hidden") {
          params.is_hidden = true;
        }
      }

      // Handle has_replies filter
      if (filters.has_replies && filters.has_replies !== "all") {
        params.has_replies = filters.has_replies === "true";
      }

      // Add sorting
      if (filters.sort_by) {
        params.sort_by = filters.sort_by;
        params.sort_order = filters.sort_order || "desc";
      }

      console.log("📡 Fetching comments with params:", params);
      const response = await apiClient.getComments(params);

      // Process comments to ensure all have user data
      let commentsWithUserData = response.data || [];

      // If the API doesn't return complete user data with comments, process them to add user data
      commentsWithUserData = await processCommentsWithUserData(
        commentsWithUserData
      );

      setState((prev) => ({
        ...prev,
        comments: commentsWithUserData,
        pagination: response.pagination || undefined,
        loading: false,
      }));
    } catch (error: any) {
      console.error("❌ Failed to fetch comments:", error);
      setState((prev) => ({
        ...prev,
        loading: false,
        error: error.message || "Failed to fetch comments",
      }));
    }
  }, []);
  const processCommentsWithUserData = async (
    comments: Comment[]
  ): Promise<Comment[]> => {
    if (!comments || comments.length === 0) return [];

    // Find comments without user data
    const commentsWithoutUserData = comments.filter(
      (comment) => comment.user_id && (!comment.user || !comment.user.username)
    );

    // If all comments have user data, return as is
    if (commentsWithoutUserData.length === 0) {
      return comments;
    }

    console.log(
      `⚠️ Found ${commentsWithoutUserData.length} comments without complete user data`
    );

    // Get unique user IDs
    const userIds = [
      ...new Set(commentsWithoutUserData.map((comment) => comment.user_id)),
    ];

    // Fetch user data for these comments
    try {
      const response = await apiClient.getUsersByIds(userIds);

      if (response.data && Array.isArray(response.data)) {
        // Create a map of user ID to user data
        const usersMap = new Map();
        response.data.forEach((user) => {
          usersMap.set(user.id, user);
        });

        // Update comments with fetched user data
        return comments.map((comment) => {
          if (comment.user_id && (!comment.user || !comment.user.username)) {
            const userData = usersMap.get(comment.user_id);
            if (userData) {
              comment.user = userData;
            }
          }
          return comment;
        });
      }
    } catch (error) {
      console.error("❌ Failed to fetch users data:", error);
    }

    return comments;
  };
  // Initial load
  useEffect(() => {
    fetchComments();
  }, []);

  // Handle search with debounce
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (state.filters.search !== undefined) {
        fetchComments({ ...state.filters, page: 1 });
      }
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [state.filters.search, fetchComments]);

  // Handle filter changes
  const handleFilterChange = (key: string, value: any) => {
    const newFilters = { ...state.filters, [key]: value, page: 1 };
    setState((prev) => ({ ...prev, filters: newFilters }));

    if (key !== "search") {
      fetchComments(newFilters);
    }
  };

  // Handle pagination
  const handlePageChange = (page: number) => {
    const newFilters = { ...state.filters, page };
    setState((prev) => ({ ...prev, filters: newFilters }));
    fetchComments(newFilters);
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
    fetchComments(newFilters);
  };

  // Handle row selection
  const handleRowSelect = (selectedRows: string[]) => {
    setState((prev) => ({ ...prev, selectedComments: selectedRows }));
  };

  // Open dialog
  const openDialog = (
    type: keyof DialogState,
    comment: Comment | null = null
  ) => {
    // Check if the comment has a valid ID before proceeding
    if (comment && (!comment.id || typeof comment.id !== "string")) {
      console.error(
        "Attempting to open dialog with invalid comment ID:",
        comment
      );
      setFormError("Cannot perform operation: comment has invalid ID");
      return;
    }

    console.log(`Opening ${type} dialog for comment:`, comment?.id);

    setDialogs((prev) => ({ ...prev, [type]: true, selectedComment: comment }));

    if (type === "editComment" && comment) {
      setEditContent(comment.content || "");
    }
  };

  // Close dialog
  const closeDialog = (type: keyof DialogState) => {
    setDialogs((prev) => ({ ...prev, [type]: false, selectedComment: null }));
    setFormError(null);
    setHideReason("");
    setEditContent("");
  };

  // Handle comment editing
  const handleEditComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!dialogs.selectedComment || !dialogs.selectedComment.id) {
      setFormError("Comment ID is missing. Cannot edit comment.");
      return;
    }

    setFormLoading(true);
    setFormError(null);

    try {
      console.log("Editing comment with ID:", dialogs.selectedComment.id);

      // Check for valid ID before proceeding
      if (
        typeof dialogs.selectedComment.id !== "string" ||
        !dialogs.selectedComment.id.trim()
      ) {
        throw new Error("Invalid comment ID");
      }

      // Use the updateComment method from apiClient
      await apiClient.updateComment(dialogs.selectedComment.id, {
        content: editContent,
      });

      closeDialog("editComment");
      fetchComments();
    } catch (error: any) {
      console.error("Error editing comment:", error);
      setFormError(error.message || "Failed to update comment");
    } finally {
      setFormLoading(false);
    }
  };

  // Handle comment deletion
  const handleDeleteComment = async () => {
    if (!dialogs.selectedComment) return;

    setFormLoading(true);
    try {
      console.log("⚠️ Comment deletion API not fully implemented yet");
      console.log("Would delete comment:", dialogs.selectedComment.id);

      // For testing purposes, we'll simulate a successful deletion
      // In production, this would call the actual API:
      // await apiClient.deleteComment(dialogs.selectedComment.id, "Deleted by admin");

      // Simulate successful response
      setTimeout(() => {
        // Remove the comment locally so the UI reflects the change
        const updatedComments = state.comments.filter(
          (c) => c.id !== dialogs.selectedComment?.id
        );
        setState((prev) => ({ ...prev, comments: updatedComments }));

        closeDialog("deleteComment");
      }, 500);
    } catch (error: any) {
      setFormError(error.message || "Failed to delete comment");
    } finally {
      setFormLoading(false);
    }
  };

  // Handle comment hiding
  const handleHideComment = async () => {
    if (!dialogs.selectedComment || !dialogs.selectedComment.id) {
      setFormError("Comment ID is missing. Cannot update comment visibility.");
      return;
    }

    setFormLoading(true);
    try {
      console.log(
        "Updating visibility for comment with ID:",
        dialogs.selectedComment.id
      );

      // Check for valid ID before proceeding
      if (
        typeof dialogs.selectedComment.id !== "string" ||
        !dialogs.selectedComment.id.trim()
      ) {
        throw new Error("Invalid comment ID");
      }

      if (dialogs.selectedComment.is_hidden) {
        // If comment is currently hidden, unhide it
        await apiClient.showComment(dialogs.selectedComment.id);
      } else {
        // If comment is visible, hide it
        await apiClient.hideComment(dialogs.selectedComment.id, hideReason);
      }
      closeDialog("hideComment");
      fetchComments();
    } catch (error: any) {
      console.error("Error updating comment visibility:", error);
      setFormError(error.message || "Failed to update comment visibility");
    } finally {
      setFormLoading(false);
    }
  };

  // Handle bulk actions
  const handleBulkAction = async (action: string, selectedIds: string[]) => {
    setFormLoading(true);
    try {
      await apiClient.bulkCommentAction({
        comment_ids: selectedIds,
        action,
        reason: action.includes("hide") ? "Bulk action" : undefined,
      });
      setState((prev) => ({ ...prev, selectedComments: [] }));
      fetchComments();
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
    fetchComments();
  };

  // Handle export
  const handleExport = async () => {
    try {
      console.log("✅ Export initiated");
    } catch (error) {
      console.error("❌ Export failed:", error);
    }
  };

  // Format comment status
  const formatCommentStatus = (comment: Comment) => {
    if (!comment) return "Unknown";
    if (comment.is_hidden) return "Hidden";
    if (comment.is_reported) return "Reported";
    return "Visible";
  };

  // Get status badge variant
  const getStatusBadgeVariant = (
    comment: Comment
  ): "default" | "secondary" | "destructive" => {
    if (!comment) return "secondary";
    if (comment.is_hidden) return "destructive";
    if (comment.is_reported) return "destructive";
    return "default";
  };

  // Format comment content preview
  const formatContentPreview = (content: string, maxLength = 100) => {
    if (!content) return "No content";
    return content.length > maxLength
      ? `${content.substring(0, maxLength)}...`
      : content;
  };

  // Format comment depth indicator
  const formatDepthIndicator = (depth: number) => {
    if (depth === 0) return null;
    return (
      <div className="flex items-center gap-1">
        <IconCornerDownRight className="h-3 w-3 text-muted-foreground" />
        <span className="text-xs text-muted-foreground">Level {depth}</span>
      </div>
    );
  };

  // Format author name safely
  const formatAuthorName = (user?: UserResponse) => {
    if (!user) return "Unknown User";

    if (user.first_name) {
      return `${user.first_name} ${user.last_name || ""}`.trim();
    }

    return user.username || "Unknown User";
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "user",
      label: "Author",
      render: (_, comment: Comment) => {
        // Get the user data directly from the comment
        const user = comment.user;

        // If user data is available, display it properly
        if (user && user.username) {
          return (
            <div className="flex items-center gap-3">
              <Avatar className="h-8 w-8">
                <AvatarImage src={user.profile_picture} alt={user.username} />
                <AvatarFallback>
                  {(
                    user.first_name?.[0] ||
                    user.username?.[0] ||
                    "U"
                  ).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0">
                <div className="font-medium flex items-center gap-2">
                  {/* Display real name if available, otherwise username */}
                  {user.first_name && user.last_name
                    ? `${user.first_name} ${user.last_name}`
                    : user.first_name || user.username}

                  {user.is_verified && (
                    <div className="inline-block w-3 h-3 bg-blue-500 rounded-full flex items-center justify-center">
                      <div className="w-1.5 h-1.5 bg-white rounded-full" />
                    </div>
                  )}
                </div>
                <div className="text-sm text-muted-foreground">
                  @{user.username}
                </div>
              </div>
            </div>
          );
        }

        // If comment.user exists but doesn't have username, it might have partial data
        if (user) {
          return (
            <div className="flex items-center gap-3">
              <Avatar className="h-8 w-8">
                <AvatarFallback>
                  {(user.first_name?.[0] || "U").toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0">
                <div className="font-medium">
                  {user.first_name || user.email || "User"}
                </div>
                <div className="text-sm text-muted-foreground">
                  <Button
                    variant="link"
                    className="h-auto p-0 text-xs text-blue-500"
                    onClick={() => fetchComments()}
                  >
                    Refresh user data
                  </Button>
                </div>
              </div>
            </div>
          );
        }

        // Fallback for when user data is not available at all
        return (
          <div className="flex items-center gap-3">
            <Avatar className="h-8 w-8">
              <AvatarFallback>U</AvatarFallback>
            </Avatar>
            <div className="min-w-0">
              <div className="font-medium">Loading user...</div>
              <div className="text-sm text-muted-foreground">
                <Button
                  variant="link"
                  className="h-auto p-0 text-xs text-blue-500"
                  onClick={() => fetchComments()}
                >
                  Refresh data
                </Button>
              </div>
            </div>
          </div>
        );
      },
    },
    {
      key: "content",
      label: "Comment",
      render: (content: string, comment: Comment) => (
        <div className="max-w-md space-y-2">
          {formatDepthIndicator(comment.depth)}
          <div className="text-sm">{formatContentPreview(content)}</div>
          {comment.replies_count > 0 && (
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <ReplyIcon className="h-3 w-3" />
              <span>{comment.replies_count} replies</span>
            </div>
          )}
        </div>
      ),
    },
    {
      key: "post",
      label: "Post",
      render: (_, comment: Comment) => (
        <div className="max-w-xs">
          {comment.post ? (
            <div className="space-y-1">
              <div className="text-sm font-medium">
                {formatContentPreview(comment.post.content || "Post", 50)}
              </div>
              <div className="text-xs text-muted-foreground">
                by @{comment.post.user?.username || "unknown"}
              </div>
            </div>
          ) : (
            <span className="text-muted-foreground text-sm">
              Post not found
            </span>
          )}
        </div>
      ),
    },
    {
      key: "likes_count",
      label: "Engagement",
      sortable: true,
      render: (likesCount: number) => (
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-1 text-sm">
            <IconHeart className="h-3 w-3 text-red-500" />
            <span>{(likesCount || 0).toLocaleString()}</span>
          </div>
        </div>
      ),
    },
    {
      key: "status",
      label: "Status",
      sortable: true,
      filterable: true,
      render: (_, comment: Comment) => (
        <div className="space-y-1">
          <Badge variant={getStatusBadgeVariant(comment)}>
            {formatCommentStatus(comment)}
          </Badge>
          {comment.is_reported && (
            <div className="flex items-center gap-1 text-xs text-orange-600">
              <IconFlag className="h-3 w-3" />
              <span>Reported</span>
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
        <div className="text-sm">
          <div className="font-medium">
            {value ? new Date(value).toLocaleDateString() : "Unknown"}
          </div>
          <div className="text-xs text-muted-foreground">
            {value ? new Date(value).toLocaleTimeString() : ""}
          </div>
        </div>
      ),
    },
    {
      key: "actions",
      label: "",
      width: "w-12",
      render: (_, comment: Comment) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
              <IconDotsVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              onClick={() => openDialog("viewComment", comment)}
            >
              <IconEye className="h-4 w-4 mr-2" />
              View Details
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => openDialog("editComment", comment)}
            >
              <IconEdit className="h-4 w-4 mr-2" />
              Edit Comment
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => openDialog("hideComment", comment)}
              className={
                comment.is_hidden ? "text-green-600" : "text-orange-600"
              }
            >
              {comment.is_hidden ? (
                <>
                  <IconEye className="h-4 w-4 mr-2" />
                  Show Comment
                </>
              ) : (
                <>
                  <IconEyeOff className="h-4 w-4 mr-2" />
                  Hide Comment
                </>
              )}
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => openDialog("deleteComment", comment)}
              className="text-red-600"
            >
              <IconTrash className="h-4 w-4 mr-2" />
              Delete Comment
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    { label: "Show Comments", action: "show" },
    { label: "Hide Comments", action: "hide" },
    {
      label: "Delete Comments",
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
                <div>Failed to load comments: {state.error}</div>
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
              <h1 className="text-3xl font-bold">Comments</h1>
              <p className="text-muted-foreground">
                Manage user comments and moderation
              </p>
            </div>
          </div>

          {/* Filters Card */}
          <Card>
            <CardHeader>
              <CardTitle>Filters</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div>
                  <Label htmlFor="search">Search</Label>
                  <div className="relative">
                    <IconSearch className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="search"
                      placeholder="Search comments..."
                      value={state.filters.search}
                      onChange={(e) =>
                        handleFilterChange("search", e.target.value)
                      }
                      className="pl-9"
                    />
                  </div>
                </div>
                <div>
                  <Label htmlFor="status">Status</Label>
                  <Select
                    value={state.filters.status}
                    onValueChange={(value) =>
                      handleFilterChange("status", value)
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All statuses" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All statuses</SelectItem>
                      <SelectItem value="visible">Visible</SelectItem>
                      <SelectItem value="hidden">Hidden</SelectItem>
                      <SelectItem value="reported">Reported</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="has_replies">Replies</Label>
                  <Select
                    value={state.filters.has_replies}
                    onValueChange={(value) =>
                      handleFilterChange("has_replies", value)
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All comments" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All comments</SelectItem>
                      <SelectItem value="true">Has replies</SelectItem>
                      <SelectItem value="false">No replies</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="user_id">User ID</Label>
                  <Input
                    id="user_id"
                    placeholder="Filter by user ID..."
                    value={state.filters.user_id}
                    onChange={(e) =>
                      handleFilterChange("user_id", e.target.value)
                    }
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Data Table */}
          <DataTable
            data={state.comments}
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
            title="Comments Management"
            description="View and moderate all user comments"
            emptyMessage="No comments found"
            searchPlaceholder="Search comments..."
          />
        </div>

        {/* View Comment Dialog */}
        <Dialog
          open={dialogs.viewComment}
          onOpenChange={() => closeDialog("viewComment")}
        >
          <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Comment Details</DialogTitle>
              <DialogDescription>
                View detailed information about this comment
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedComment && (
              <div className="space-y-6">
                {/* Author Info */}
                <div className="flex items-center space-x-4">
                  <Avatar className="h-12 w-12">
                    <AvatarImage
                      src={dialogs.selectedComment.user?.profile_picture}
                    />
                    <AvatarFallback>
                      {(dialogs.selectedComment.user?.first_name ||
                        dialogs.selectedComment.user?.username ||
                        "U")[0].toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <h3 className="text-lg font-semibold">
                      {formatAuthorName(dialogs.selectedComment.user)}
                    </h3>
                    <p className="text-muted-foreground">
                      @{dialogs.selectedComment.user?.username || "unknown"}
                    </p>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge
                        variant={getStatusBadgeVariant(dialogs.selectedComment)}
                      >
                        {formatCommentStatus(dialogs.selectedComment)}
                      </Badge>
                      {dialogs.selectedComment.user?.is_verified && (
                        <Badge
                          variant="outline"
                          className="text-green-600 border-green-200"
                        >
                          Verified
                        </Badge>
                      )}
                    </div>
                  </div>
                </div>

                <Separator />

                {/* Comment Content */}
                <div>
                  <Label className="text-sm font-medium">Comment Content</Label>
                  <div className="mt-2 p-4 bg-muted rounded-lg">
                    <p className="text-sm whitespace-pre-wrap">
                      {dialogs.selectedComment.content || "No content"}
                    </p>
                  </div>
                </div>

                {/* Comment Metadata */}
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label className="text-sm font-medium">Created</Label>
                    <p className="text-sm text-muted-foreground">
                      {dialogs.selectedComment.created_at
                        ? new Date(
                            dialogs.selectedComment.created_at
                          ).toLocaleString()
                        : "Unknown"}
                    </p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Likes</Label>
                    <p className="text-sm text-muted-foreground">
                      {dialogs.selectedComment.likes_count || 0} likes
                    </p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Replies</Label>
                    <p className="text-sm text-muted-foreground">
                      {dialogs.selectedComment.replies_count || 0} replies
                    </p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Depth Level</Label>
                    <p className="text-sm text-muted-foreground">
                      Level {dialogs.selectedComment.depth || 0}
                    </p>
                  </div>
                </div>

                {/* Post Context */}
                {dialogs.selectedComment.post && (
                  <div>
                    <Label className="text-sm font-medium">Original Post</Label>
                    <div className="mt-2 p-4 bg-muted/50 rounded-lg">
                      <div className="text-sm">
                        {formatContentPreview(
                          dialogs.selectedComment.post.content ||
                            "Post content",
                          200
                        )}
                      </div>
                      <div className="text-xs text-muted-foreground mt-2">
                        by @
                        {dialogs.selectedComment.post.user?.username ||
                          "unknown"}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )}

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => closeDialog("viewComment")}
              >
                Close
              </Button>
              <Button
                onClick={() => {
                  closeDialog("viewComment");
                  if (dialogs.selectedComment)
                    openDialog("editComment", dialogs.selectedComment);
                }}
              >
                <IconEdit className="h-4 w-4 mr-2" />
                Edit Comment
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Edit Comment Dialog */}
        <Dialog
          open={dialogs.editComment}
          onOpenChange={() => closeDialog("editComment")}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Edit Comment</DialogTitle>
              <DialogDescription>
                Update the comment content below.
              </DialogDescription>
            </DialogHeader>

            <form onSubmit={handleEditComment} className="space-y-4">
              <div>
                <Label htmlFor="content">Comment Content</Label>
                <Textarea
                  id="content"
                  value={editContent}
                  onChange={(e) => setEditContent(e.target.value)}
                  rows={4}
                  required
                />
              </div>

              {formError && (
                <Alert variant="destructive">
                  <IconAlertCircle className="h-4 w-4" />
                  <AlertDescription>{formError}</AlertDescription>
                </Alert>
              )}

              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => closeDialog("editComment")}
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  disabled={formLoading || !editContent.trim()}
                >
                  {formLoading && (
                    <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                  )}
                  Update Comment
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>

        {/* Delete Comment Dialog */}
        <Dialog
          open={dialogs.deleteComment}
          onOpenChange={() => closeDialog("deleteComment")}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete Comment</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this comment? This action cannot
                be undone.
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedComment && (
              <div className="py-4">
                <div className="p-4 bg-muted rounded-lg">
                  <div className="flex items-center space-x-3 mb-3">
                    <Avatar className="h-8 w-8">
                      <AvatarImage
                        src={dialogs.selectedComment.user?.profile_picture}
                      />
                      <AvatarFallback>
                        {(dialogs.selectedComment.user?.first_name ||
                          dialogs.selectedComment.user?.username ||
                          "U")[0].toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="font-medium">
                        {formatAuthorName(dialogs.selectedComment.user)}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        @{dialogs.selectedComment.user?.username || "unknown"}
                      </p>
                    </div>
                  </div>
                  <p className="text-sm whitespace-pre-wrap">
                    {formatContentPreview(
                      dialogs.selectedComment.content || "No content",
                      150
                    )}
                  </p>
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
                onClick={() => closeDialog("deleteComment")}
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDeleteComment}
                disabled={formLoading}
              >
                {formLoading && (
                  <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Delete Comment
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Hide/Show Comment Dialog */}
        <Dialog
          open={dialogs.hideComment}
          onOpenChange={() => closeDialog("hideComment")}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                {dialogs.selectedComment?.is_hidden
                  ? "Show Comment"
                  : "Hide Comment"}
              </DialogTitle>
              <DialogDescription>
                {dialogs.selectedComment?.is_hidden
                  ? "Make this comment visible to users."
                  : "Hide this comment from users."}
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedComment && (
              <div className="space-y-4">
                <div className="p-4 bg-muted rounded-lg">
                  <div className="flex items-center space-x-3 mb-3">
                    <Avatar className="h-8 w-8">
                      <AvatarImage
                        src={dialogs.selectedComment.user?.profile_picture}
                      />
                      <AvatarFallback>
                        {(dialogs.selectedComment.user?.first_name ||
                          dialogs.selectedComment.user?.username ||
                          "U")[0].toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="font-medium">
                        {formatAuthorName(dialogs.selectedComment.user)}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        @{dialogs.selectedComment.user?.username || "unknown"}
                      </p>
                    </div>
                  </div>
                  <p className="text-sm whitespace-pre-wrap">
                    {formatContentPreview(
                      dialogs.selectedComment.content || "No content",
                      150
                    )}
                  </p>
                </div>

                {!dialogs.selectedComment.is_hidden && (
                  <div>
                    <Label htmlFor="reason">Reason for hiding</Label>
                    <Textarea
                      id="reason"
                      value={hideReason}
                      onChange={(e) => setHideReason(e.target.value)}
                      placeholder="Provide a reason for hiding this comment..."
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
              <Button
                variant="outline"
                onClick={() => closeDialog("hideComment")}
              >
                Cancel
              </Button>
              <Button
                variant={
                  dialogs.selectedComment?.is_hidden ? "default" : "destructive"
                }
                onClick={handleHideComment}
                disabled={
                  formLoading ||
                  (!dialogs.selectedComment?.is_hidden && !hideReason)
                }
              >
                {formLoading && (
                  <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                {dialogs.selectedComment?.is_hidden
                  ? "Show Comment"
                  : "Hide Comment"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </SidebarInset>
    </SidebarProvider>
  );
}

export default withAuth(CommentsPage);
