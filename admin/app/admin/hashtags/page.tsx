// app/admin/hashtags/page.tsx
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
import { Hashtag, TableColumn } from "@/types/admin";
import {
  IconHash,
  IconEye,
  IconBan,
  IconCheck,
  IconX,
  IconTrendingUp,
  IconTrendingDown,
  IconCalendar,
  IconFlag,
  IconShield,
  IconTrash,
  IconRefresh,
} from "@tabler/icons-react";

function HashtagsPage() {
  const [hashtags, setHashtags] = useState<Hashtag[]>([]);
  const [trendingHashtags, setTrendingHashtags] = useState<Hashtag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<any>(null);
  const [filters, setFilters] = useState<any>({ page: 1, limit: 20 });

  // Dialog states
  const [selectedHashtag, setSelectedHashtag] = useState<Hashtag | null>(null);
  const [showHashtagDetails, setShowHashtagDetails] = useState(false);
  const [showActionDialog, setShowActionDialog] = useState(false);
  const [showBulkDialog, setShowBulkDialog] = useState(false);
  const [actionType, setActionType] = useState<string>("");
  const [bulkAction, setBulkAction] = useState<string>("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [actionReason, setActionReason] = useState("");

  // Active tab
  const [activeTab, setActiveTab] = useState("hashtags");

  // Fetch hashtags
  const fetchHashtags = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getHashtags(filters);
      setHashtags(response.data.data);
      setPagination(response.data.pagination);
    } catch (error: any) {
      console.error("Failed to fetch hashtags:", error);
      setError(error.response?.data?.message || "Failed to load hashtags");
    } finally {
      setLoading(false);
    }
  };

  // Fetch trending hashtags
  const fetchTrendingHashtags = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getTrendingHashtags(50);
      setTrendingHashtags(response.data);
    } catch (error: any) {
      console.error("Failed to fetch trending hashtags:", error);
      setError(
        error.response?.data?.message || "Failed to load trending hashtags"
      );
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === "hashtags") {
      fetchHashtags();
    } else if (activeTab === "trending") {
      fetchTrendingHashtags();
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
      await apiClient.bulkHashtagAction({
        hashtag_ids: selectedIds,
        action: bulkAction,
        reason: actionReason,
      });

      setShowBulkDialog(false);
      setActionReason("");
      if (activeTab === "hashtags") {
        fetchHashtags();
      } else {
        fetchTrendingHashtags();
      }
    } catch (error: any) {
      console.error("Bulk action failed:", error);
      setError(error.response?.data?.message || "Bulk action failed");
    }
  };

  // Handle individual hashtag action
  const handleHashtagAction = (hashtag: Hashtag, action: string) => {
    setSelectedHashtag(hashtag);
    setActionType(action);
    setShowActionDialog(true);
  };

  // Execute hashtag action
  const executeHashtagAction = async () => {
    if (!selectedHashtag) return;

    try {
      switch (actionType) {
        case "block":
          await apiClient.blockHashtag(selectedHashtag.id, actionReason);
          break;
        case "unblock":
          await apiClient.unblockHashtag(selectedHashtag.id);
          break;
        case "delete":
          await apiClient.deleteHashtag(selectedHashtag.id, actionReason);
          break;
      }

      setShowActionDialog(false);
      setActionReason("");
      if (activeTab === "hashtags") {
        fetchHashtags();
      } else {
        fetchTrendingHashtags();
      }
    } catch (error: any) {
      console.error("Hashtag action failed:", error);
      setError(error.response?.data?.message || "Hashtag action failed");
    }
  };

  // View hashtag details
  const viewHashtagDetails = async (hashtag: Hashtag) => {
    try {
      const response = await apiClient.getHashtag(hashtag.id);
      setSelectedHashtag(response.data);
      setShowHashtagDetails(true);
    } catch (error) {
      console.error("Failed to fetch hashtag details:", error);
    }
  };

  // Get trending indicator
  const getTrendingIndicator = (trendingScore: number) => {
    if (trendingScore > 50) {
      return <IconTrendingUp className="h-4 w-4 text-green-600" />;
    } else if (trendingScore > 20) {
      return <IconTrendingUp className="h-4 w-4 text-yellow-600" />;
    } else {
      return <IconTrendingDown className="h-4 w-4 text-gray-400" />;
    }
  };

  // Format large numbers
  const formatNumber = (num: number) => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + "M";
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + "K";
    }
    return num.toString();
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "hashtag",
      label: "Hashtag",
      render: (_, hashtag: Hashtag) => (
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2">
            <IconHash className="h-4 w-4 text-primary" />
            <span className="font-medium">#{hashtag.tag}</span>
          </div>
          <div className="flex items-center gap-1">
            {getTrendingIndicator(hashtag.trending_score)}
            <span className="text-xs text-muted-foreground">
              {hashtag.trending_score.toFixed(1)}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "usage",
      label: "Usage",
      sortable: true,
      render: (_, hashtag: Hashtag) => (
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <span className="font-medium">
              {formatNumber(hashtag.usage_count)}
            </span>
            <Badge variant="outline" className="text-xs">
              Recent
            </Badge>
          </div>
          <div className="text-xs text-muted-foreground">
            {formatNumber(hashtag.total_usage)} total
          </div>
        </div>
      ),
    },
    {
      key: "category",
      label: "Category",
      filterable: true,
      render: (value: string) =>
        value ? (
          <Badge variant="secondary">{value}</Badge>
        ) : (
          <span className="text-muted-foreground">Uncategorized</span>
        ),
    },
    {
      key: "status",
      label: "Status",
      filterable: true,
      render: (_, hashtag: Hashtag) => {
        if (hashtag.is_blocked) {
          return (
            <Badge variant="destructive">
              <IconBan className="h-3 w-3 mr-1" />
              Blocked
            </Badge>
          );
        }
        return (
          <Badge variant="default">
            <IconCheck className="h-3 w-3 mr-1" />
            Active
          </Badge>
        );
      },
    },
    {
      key: "last_used_at",
      label: "Last Used",
      sortable: true,
      render: (value: string) =>
        value ? (
          <div className="text-xs">{new Date(value).toLocaleDateString()}</div>
        ) : (
          <span className="text-muted-foreground">Never</span>
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
      render: (_, hashtag: Hashtag) => (
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewHashtagDetails(hashtag)}
          >
            <IconEye className="h-3 w-3" />
          </Button>
          {hashtag.is_blocked ? (
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleHashtagAction(hashtag, "unblock")}
            >
              <IconCheck className="h-3 w-3" />
            </Button>
          ) : (
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleHashtagAction(hashtag, "block")}
            >
              <IconBan className="h-3 w-3" />
            </Button>
          )}
          <Button
            size="sm"
            variant="destructive"
            onClick={() => handleHashtagAction(hashtag, "delete")}
          >
            <IconTrash className="h-3 w-3" />
          </Button>
        </div>
      ),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    {
      label: "Block Hashtags",
      action: "block",
      variant: "destructive" as const,
    },
    {
      label: "Unblock Hashtags",
      action: "unblock",
      variant: "default" as const,
    },
    {
      label: "Delete Hashtags",
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
              <h1 className="text-2xl font-bold">Hashtags Management</h1>
              <p className="text-muted-foreground">
                Manage hashtags and trending topics
              </p>
            </div>
          </div>

          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList>
              <TabsTrigger value="hashtags">All Hashtags</TabsTrigger>
              <TabsTrigger value="trending">Trending</TabsTrigger>
            </TabsList>

            <TabsContent value="hashtags" className="space-y-4">
              <DataTable
                title="Hashtags"
                description={`Manage ${pagination?.total || 0} hashtags`}
                data={hashtags}
                columns={columns}
                loading={loading}
                pagination={pagination}
                searchPlaceholder="Search hashtags..."
                onPageChange={handlePageChange}
                onFilter={handleFilter}
                onRowSelect={handleRowSelect}
                bulkActions={bulkActions}
                onBulkAction={handleBulkAction}
                onRefresh={fetchHashtags}
                onExport={() => console.log("Export hashtags")}
              />
            </TabsContent>

            <TabsContent value="trending" className="space-y-4">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h3 className="text-lg font-semibold">Trending Hashtags</h3>
                  <p className="text-sm text-muted-foreground">
                    Most popular hashtags in the last 24 hours
                  </p>
                </div>
                <Button
                  onClick={fetchTrendingHashtags}
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

              {loading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {Array.from({ length: 6 }).map((_, i) => (
                    <Card key={i}>
                      <CardHeader>
                        <div className="h-4 bg-gray-200 rounded animate-pulse" />
                        <div className="h-3 bg-gray-200 rounded animate-pulse w-2/3" />
                      </CardHeader>
                    </Card>
                  ))}
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {trendingHashtags.slice(0, 20).map((hashtag, index) => (
                    <Card
                      key={hashtag.id}
                      className={`cursor-pointer hover:shadow-md transition-shadow ${
                        index < 3 ? "border-primary/20 bg-primary/5" : ""
                      }`}
                      onClick={() => viewHashtagDetails(hashtag)}
                    >
                      <CardHeader className="pb-2">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <span className="text-lg font-bold text-muted-foreground">
                              #{index + 1}
                            </span>
                            <div className="flex items-center gap-1">
                              <IconHash className="h-4 w-4 text-primary" />
                              <span className="font-semibold">
                                #{hashtag.tag}
                              </span>
                            </div>
                          </div>
                          {getTrendingIndicator(hashtag.trending_score)}
                        </div>
                        <div className="flex items-center gap-4 text-sm text-muted-foreground">
                          <span>{formatNumber(hashtag.usage_count)} uses</span>
                          <span>
                            Score: {hashtag.trending_score.toFixed(1)}
                          </span>
                        </div>
                      </CardHeader>
                      {hashtag.description && (
                        <CardContent className="pt-0">
                          <p className="text-sm text-muted-foreground line-clamp-2">
                            {hashtag.description}
                          </p>
                        </CardContent>
                      )}
                    </Card>
                  ))}
                </div>
              )}
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>

      {/* Hashtag Details Dialog */}
      <Dialog open={showHashtagDetails} onOpenChange={setShowHashtagDetails}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <IconHash className="h-5 w-5 text-primary" />#
              {selectedHashtag?.tag}
            </DialogTitle>
          </DialogHeader>

          {selectedHashtag && (
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div>
                      <CardTitle className="flex items-center gap-2">
                        #{selectedHashtag.tag}
                        {selectedHashtag.is_blocked && (
                          <Badge variant="destructive">
                            <IconBan className="h-3 w-3 mr-1" />
                            Blocked
                          </Badge>
                        )}
                      </CardTitle>
                      {selectedHashtag.description && (
                        <CardDescription className="mt-2">
                          {selectedHashtag.description}
                        </CardDescription>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      {getTrendingIndicator(selectedHashtag.trending_score)}
                      <Badge variant="outline">
                        Score: {selectedHashtag.trending_score.toFixed(1)}
                      </Badge>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <Card>
                      <CardHeader className="pb-2">
                        <CardTitle className="text-lg">Usage Count</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div className="text-2xl font-bold">
                          {selectedHashtag.usage_count.toLocaleString()}
                        </div>
                        <p className="text-sm text-muted-foreground">
                          Recent usage
                        </p>
                      </CardContent>
                    </Card>
                    <Card>
                      <CardHeader className="pb-2">
                        <CardTitle className="text-lg">Total Usage</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div className="text-2xl font-bold">
                          {selectedHashtag.total_usage.toLocaleString()}
                        </div>
                        <p className="text-sm text-muted-foreground">
                          All time
                        </p>
                      </CardContent>
                    </Card>
                  </div>

                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <label className="font-medium">Category</label>
                      <p>{selectedHashtag.category || "Uncategorized"}</p>
                    </div>
                    <div>
                      <label className="font-medium">Trending Score</label>
                      <p>{selectedHashtag.trending_score.toFixed(2)}</p>
                    </div>
                    <div>
                      <label className="font-medium">Created</label>
                      <p>
                        {new Date(
                          selectedHashtag.created_at
                        ).toLocaleDateString()}
                      </p>
                    </div>
                    <div>
                      <label className="font-medium">Last Used</label>
                      <p>
                        {selectedHashtag.last_used_at
                          ? new Date(
                              selectedHashtag.last_used_at
                            ).toLocaleDateString()
                          : "Never"}
                      </p>
                    </div>
                  </div>

                  {selectedHashtag.is_blocked &&
                    selectedHashtag.block_reason && (
                      <div>
                        <label className="font-medium">Block Reason</label>
                        <p className="text-sm bg-red-50 p-3 rounded border border-red-200">
                          {selectedHashtag.block_reason}
                        </p>
                      </div>
                    )}
                </CardContent>
              </Card>

              {/* Recent Posts with this hashtag */}
              {selectedHashtag.recent_posts &&
                selectedHashtag.recent_posts.length > 0 && (
                  <Card>
                    <CardHeader>
                      <CardTitle>Recent Posts</CardTitle>
                      <CardDescription>
                        Latest posts using this hashtag
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-3">
                        {selectedHashtag.recent_posts
                          .slice(0, 3)
                          .map((post: any, index: number) => (
                            <div key={index} className="border rounded p-3">
                              <p className="text-sm line-clamp-2">
                                {post.content}
                              </p>
                              <div className="flex items-center gap-2 mt-2 text-xs text-muted-foreground">
                                <span>by @{post.user?.username}</span>
                                <span>•</span>
                                <span>
                                  {new Date(
                                    post.created_at
                                  ).toLocaleDateString()}
                                </span>
                              </div>
                            </div>
                          ))}
                      </div>
                    </CardContent>
                  </Card>
                )}
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowHashtagDetails(false)}
            >
              Close
            </Button>
            {selectedHashtag && (
              <>
                {selectedHashtag.is_blocked ? (
                  <Button
                    onClick={() => {
                      setShowHashtagDetails(false);
                      handleHashtagAction(selectedHashtag, "unblock");
                    }}
                  >
                    Unblock Hashtag
                  </Button>
                ) : (
                  <Button
                    variant="outline"
                    onClick={() => {
                      setShowHashtagDetails(false);
                      handleHashtagAction(selectedHashtag, "block");
                    }}
                  >
                    Block Hashtag
                  </Button>
                )}
              </>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Action Dialog */}
      <Dialog open={showActionDialog} onOpenChange={setShowActionDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {actionType === "block" && "Block Hashtag"}
              {actionType === "unblock" && "Unblock Hashtag"}
              {actionType === "delete" && "Delete Hashtag"}
            </DialogTitle>
            <DialogDescription>
              {actionType === "block" &&
                "This hashtag will be blocked and won't appear in search results or trending lists."}
              {actionType === "unblock" &&
                "This hashtag will be unblocked and can appear in search results and trending lists again."}
              {actionType === "delete" &&
                "This action cannot be undone. The hashtag will be permanently deleted."}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {actionType !== "unblock" && (
              <div>
                <label className="text-sm font-medium">Reason</label>
                <Textarea
                  placeholder="Enter reason for this action..."
                  value={actionReason}
                  onChange={(e) => setActionReason(e.target.value)}
                  required
                />
              </div>
            )}
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowActionDialog(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={executeHashtagAction}
              variant={
                actionType === "delete" || actionType === "block"
                  ? "destructive"
                  : "default"
              }
              disabled={actionType !== "unblock" && !actionReason.trim()}
            >
              {actionType === "block" && "Block Hashtag"}
              {actionType === "unblock" && "Unblock Hashtag"}
              {actionType === "delete" && "Delete Hashtag"}
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
              You are about to {bulkAction} {selectedIds.length} hashtag(s).
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {bulkAction !== "unblock" && (
              <div>
                <label className="text-sm font-medium">Reason</label>
                <Textarea
                  placeholder="Enter reason for this bulk action..."
                  value={actionReason}
                  onChange={(e) => setActionReason(e.target.value)}
                  required
                />
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowBulkDialog(false)}>
              Cancel
            </Button>
            <Button
              onClick={executeBulkAction}
              variant={
                bulkAction === "delete" || bulkAction === "block"
                  ? "destructive"
                  : "default"
              }
              disabled={bulkAction !== "unblock" && !actionReason.trim()}
            >
              Execute {bulkAction}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SidebarProvider>
  );
}

export default withAuth(HashtagsPage);
