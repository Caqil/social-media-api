// app/admin/comments/page.tsx
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
import { withAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import { Comment, TableColumn } from "@/types/admin";
import {
  IconMessage,
  IconEye,
  IconEyeOff,
  IconTrash,
  IconUser,
  IconCalendar,
  IconThumbUp,
} from "@tabler/icons-react";

function CommentsPage() {
  const [comments, setComments] = useState<Comment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<any>(null);
  const [filters, setFilters] = useState<any>({ page: 1, limit: 20 });

  // Dialog states
  const [selectedComment, setSelectedComment] = useState<Comment | null>(null);
  const [showCommentDetails, setShowCommentDetails] = useState(false);
  const [showActionDialog, setShowActionDialog] = useState(false);
  const [showBulkDialog, setShowBulkDialog] = useState(false);
  const [actionType, setActionType] = useState<string>("");
  const [bulkAction, setBulkAction] = useState<string>("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [actionReason, setActionReason] = useState("");

  // Fetch comments
  const fetchComments = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getComments(filters);
      setComments(response.data.data);
      setPagination(response.data.pagination);
    } catch (error: any) {
      console.error("Failed to fetch comments:", error);
      setError(error.response?.data?.message || "Failed to load comments");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchComments();
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
      await apiClient.bulkCommentAction({
        comment_ids: selectedIds,
        action: bulkAction,
        reason: actionReason,
      });

      setShowBulkDialog(false);
      setActionReason("");
      fetchComments();
    } catch (error: any) {
      console.error("Bulk action failed:", error);
      setError(error.response?.data?.message || "Bulk action failed");
    }
  };

  // Handle individual comment action
  const handleCommentAction = (comment: Comment, action: string) => {
    setSelectedComment(comment);
    setActionType(action);
    setShowActionDialog(true);
  };

  // Execute comment action
  const executeCommentAction = async () => {
    if (!selectedComment) return;

    try {
      switch (actionType) {
        case "hide":
          await apiClient.hideComment(selectedComment.id, actionReason);
          break;
        case "delete":
          await apiClient.deleteComment(selectedComment.id, actionReason);
          break;
      }

      setShowActionDialog(false);
      setActionReason("");
      fetchComments();
    } catch (error: any) {
      console.error("Comment action failed:", error);
      setError(error.response?.data?.message || "Comment action failed");
    }
  };

  // View comment details
  const viewCommentDetails = (comment: Comment) => {
    setSelectedComment(comment);
    setShowCommentDetails(true);
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "content",
      label: "Comment",
      render: (content: string, comment: Comment) => (
        <div className="max-w-md">
          <p className="text-sm line-clamp-2">{content}</p>
          {comment.user && (
            <div className="flex items-center gap-2 mt-1">
              <Avatar className="h-5 w-5">
                <AvatarImage src={comment.user.profile_picture} />
                <AvatarFallback className="text-xs">
                  {comment.user.username?.[0]?.toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <span className="text-xs text-muted-foreground">
                {comment.user.username}
              </span>
            </div>
          )}
        </div>
      ),
    },
    {
      key: "post",
      label: "Post",
      render: (_, comment: Comment) => (
        <div className="text-xs text-muted-foreground">
          Post ID: {comment.post_id?.slice(-8)}
        </div>
      ),
    },
    {
      key: "likes_count",
      label: "Likes",
      sortable: true,
      render: (value: number) => (
        <div className="flex items-center gap-1">
          <IconThumbUp className="h-3 w-3" />
          <span>{value || 0}</span>
        </div>
      ),
    },
    {
      key: "replies_count",
      label: "Replies",
      sortable: true,
      render: (value: number) => (
        <Badge variant="outline">{value || 0} replies</Badge>
      ),
    },
    {
      key: "status",
      label: "Status",
      filterable: true,
      render: (_, comment: Comment) => {
        if (comment.is_hidden) {
          return <Badge variant="destructive">Hidden</Badge>;
        }
        if (comment.is_reported) {
          return <Badge variant="secondary">Reported</Badge>;
        }
        return <Badge variant="default">Active</Badge>;
      },
    },
    {
      key: "depth",
      label: "Depth",
      render: (value: number) => (
        <Badge variant="outline">Level {value || 0}</Badge>
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
      render: (_, comment: Comment) => (
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewCommentDetails(comment)}
          >
            <IconEye className="h-3 w-3" />
          </Button>
          {!comment.is_hidden && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleCommentAction(comment, "hide")}
            >
              <IconEyeOff className="h-3 w-3" />
            </Button>
          )}
          <Button
            size="sm"
            variant="destructive"
            onClick={() => handleCommentAction(comment, "delete")}
          >
            <IconTrash className="h-3 w-3" />
          </Button>
        </div>
      ),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    { label: "Hide Comments", action: "hide", variant: "default" as const },
    {
      label: "Delete Comments",
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
              <h1 className="text-2xl font-bold">Comments Management</h1>
              <p className="text-muted-foreground">
                Manage user comments and moderation
              </p>
            </div>
          </div>

          <DataTable
            title="Comments"
            description={`Manage ${pagination?.total || 0} comments`}
            data={comments}
            columns={columns}
            loading={loading}
            pagination={pagination}
            searchPlaceholder="Search comments by content or user..."
            onPageChange={handlePageChange}
            onFilter={handleFilter}
            onRowSelect={handleRowSelect}
            bulkActions={bulkActions}
            onBulkAction={handleBulkAction}
            onRefresh={fetchComments}
            onExport={() => console.log("Export comments")}
          />
        </div>
      </SidebarInset>

      {/* Comment Details Dialog */}
      <Dialog open={showCommentDetails} onOpenChange={setShowCommentDetails}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Comment Details</DialogTitle>
          </DialogHeader>

          {selectedComment && (
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-3">
                    {selectedComment.user && (
                      <>
                        <Avatar>
                          <AvatarImage
                            src={selectedComment.user.profile_picture}
                          />
                          <AvatarFallback>
                            {selectedComment.user.username?.[0]?.toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <p className="font-medium">
                            {selectedComment.user.username}
                          </p>
                          <p className="text-sm text-muted-foreground">
                            {new Date(
                              selectedComment.created_at
                            ).toLocaleString()}
                          </p>
                        </div>
                      </>
                    )}
                  </div>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{selectedComment.content}</p>

                  <div className="flex items-center gap-4 mt-4 text-sm text-muted-foreground">
                    <div className="flex items-center gap-1">
                      <IconThumbUp className="h-4 w-4" />
                      {selectedComment.likes_count} likes
                    </div>
                    <div className="flex items-center gap-1">
                      <IconMessage className="h-4 w-4" />
                      {selectedComment.replies_count} replies
                    </div>
                    <div className="flex items-center gap-1">
                      <IconCalendar className="h-4 w-4" />
                      Level {selectedComment.depth}
                    </div>
                  </div>

                  {selectedComment.media_urls &&
                    selectedComment.media_urls.length > 0 && (
                      <div className="mt-4">
                        <p className="text-sm font-medium mb-2">Media:</p>
                        <div className="grid grid-cols-2 gap-2">
                          {selectedComment.media_urls.map((url, index) => (
                            <img
                              key={index}
                              src={url}
                              alt={`Comment media ${index + 1}`}
                              className="rounded-lg object-cover h-20 w-full"
                            />
                          ))}
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
              onClick={() => setShowCommentDetails(false)}
            >
              Close
            </Button>
            {selectedComment && !selectedComment.is_hidden && (
              <Button
                variant="outline"
                onClick={() => {
                  setShowCommentDetails(false);
                  handleCommentAction(selectedComment, "hide");
                }}
              >
                Hide Comment
              </Button>
            )}
            {selectedComment && (
              <Button
                variant="destructive"
                onClick={() => {
                  setShowCommentDetails(false);
                  handleCommentAction(selectedComment, "delete");
                }}
              >
                Delete Comment
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
              {actionType === "hide" ? "Hide Comment" : "Delete Comment"}
            </DialogTitle>
            <DialogDescription>
              {actionType === "hide"
                ? "This comment will be hidden from users but remain in the database."
                : "This action cannot be undone. The comment will be permanently deleted."}
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
              onClick={executeCommentAction}
              variant={actionType === "delete" ? "destructive" : "default"}
              disabled={!actionReason.trim()}
            >
              {actionType === "hide" ? "Hide Comment" : "Delete Comment"}
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
              You are about to {bulkAction} {selectedIds.length} comment(s).
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

export default withAuth(CommentsPage);
