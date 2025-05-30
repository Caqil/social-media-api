// app/admin/reports/page.tsx
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
import { Report, TableColumn } from "@/types/admin";
import {
  IconFlag,
  IconEye,
  IconCheck,
  IconX,
  IconClock,
  IconShield,
  IconUser,
  IconMessage,
  IconUsersGroup,
  IconCalendarEvent,
  IconAlertTriangle,
  IconMail,
} from "@tabler/icons-react";

function ReportsPage() {
  const [reports, setReports] = useState<Report[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<any>(null);
  const [filters, setFilters] = useState<any>({ page: 1, limit: 20 });
  const [reportStats, setReportStats] = useState<any>(null);

  // Dialog states
  const [selectedReport, setSelectedReport] = useState<Report | null>(null);
  const [showReportDetails, setShowReportDetails] = useState(false);
  const [showActionDialog, setShowActionDialog] = useState(false);
  const [showBulkDialog, setShowBulkDialog] = useState(false);
  const [actionType, setActionType] = useState<string>("");
  const [bulkAction, setBulkAction] = useState<string>("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [actionReason, setActionReason] = useState("");
  const [resolution, setResolution] = useState("");

  // Active tab
  const [activeTab, setActiveTab] = useState("reports");

  // Fetch reports
  const fetchReports = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getReports(filters);
      setReports(response.data.data);
      setPagination(response.data.pagination);
    } catch (error: any) {
      console.error("Failed to fetch reports:", error);
      setError(error.response?.data?.message || "Failed to load reports");
    } finally {
      setLoading(false);
    }
  };

  // Fetch report stats
  const fetchReportStats = async () => {
    try {
      const response = await apiClient.getReportStats();
      setReportStats(response.data);
    } catch (error) {
      console.error("Failed to fetch report stats:", error);
    }
  };

  useEffect(() => {
    if (activeTab === "reports") {
      fetchReports();
    } else if (activeTab === "stats") {
      fetchReportStats();
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
      await apiClient.bulkReportAction({
        report_ids: selectedIds,
        action: bulkAction,
        resolution: resolution,
        reason: actionReason,
      });

      setShowBulkDialog(false);
      setActionReason("");
      setResolution("");
      fetchReports();
    } catch (error: any) {
      console.error("Bulk action failed:", error);
      setError(error.response?.data?.message || "Bulk action failed");
    }
  };

  // Handle individual report action
  const handleReportAction = (report: Report, action: string) => {
    setSelectedReport(report);
    setActionType(action);
    setShowActionDialog(true);
  };

  // Execute report action
  const executeReportAction = async () => {
    if (!selectedReport) return;

    try {
      switch (actionType) {
        case "resolve":
          await apiClient.resolveReport(selectedReport.id, {
            resolution: resolution,
            note: actionReason,
          });
          break;
        case "reject":
          await apiClient.rejectReport(selectedReport.id, {
            note: actionReason,
          });
          break;
        case "assign":
          // Implement assign logic if needed
          break;
      }

      setShowActionDialog(false);
      setActionReason("");
      setResolution("");
      fetchReports();
    } catch (error: any) {
      console.error("Report action failed:", error);
      setError(error.response?.data?.message || "Report action failed");
    }
  };

  // View report details
  const viewReportDetails = (report: Report) => {
    setSelectedReport(report);
    setShowReportDetails(true);
  };

  // Get report reason icon
  const getReportReasonIcon = (reason: string) => {
    switch (reason) {
      case "spam":
        return <IconMail className="h-4 w-4" />;
      case "harassment":
        return <IconShield className="h-4 w-4" />;
      case "hate_speech":
        return <IconAlertTriangle className="h-4 w-4" />;
      case "inappropriate_content":
        return <IconFlag className="h-4 w-4" />;
      default:
        return <IconFlag className="h-4 w-4" />;
    }
  };

  // Get target type icon
  const getTargetTypeIcon = (type: string) => {
    switch (type) {
      case "user":
        return <IconUser className="h-4 w-4" />;
      case "post":
        return <IconMessage className="h-4 w-4" />;
      case "comment":
        return <IconMessage className="h-4 w-4" />;
      case "group":
        return <IconUsersGroup className="h-4 w-4" />;
      case "event":
        return <IconCalendarEvent className="h-4 w-4" />;
      default:
        return <IconFlag className="h-4 w-4" />;
    }
  };

  // Get status color
  const getStatusColor = (status: string) => {
    switch (status) {
      case "pending":
        return "bg-yellow-100 text-yellow-800";
      case "reviewing":
        return "bg-blue-100 text-blue-800";
      case "resolved":
        return "bg-green-100 text-green-800";
      case "rejected":
        return "bg-red-100 text-red-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  // Get priority color
  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case "critical":
        return "bg-red-100 text-red-800";
      case "high":
        return "bg-orange-100 text-orange-800";
      case "medium":
        return "bg-yellow-100 text-yellow-800";
      case "low":
        return "bg-green-100 text-green-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "report",
      label: "Report",
      render: (_, report: Report) => (
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            {getReportReasonIcon(report.reason)}
            <Badge variant="outline" className="text-xs">
              {report.reason.replace("_", " ")}
            </Badge>
            <Badge className={getPriorityColor(report.priority)}>
              {report.priority}
            </Badge>
          </div>
          <p className="text-sm line-clamp-2">{report.description}</p>
          {report.reporter && (
            <div className="flex items-center gap-2">
              <Avatar className="h-4 w-4">
                <AvatarImage src={report.reporter.profile_picture} />
                <AvatarFallback className="text-xs">
                  {report.reporter.username?.[0]?.toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <span className="text-xs text-muted-foreground">
                by {report.reporter.username}
              </span>
            </div>
          )}
        </div>
      ),
    },
    {
      key: "target_type",
      label: "Target",
      filterable: true,
      render: (value: string, report: Report) => (
        <div className="flex items-center gap-2">
          {getTargetTypeIcon(value)}
          <div>
            <Badge variant="secondary" className="text-xs">
              {value}
            </Badge>
            {report.target_user && (
              <p className="text-xs text-muted-foreground mt-1">
                @{report.target_user.username}
              </p>
            )}
          </div>
        </div>
      ),
    },
    {
      key: "status",
      label: "Status",
      filterable: true,
      render: (value: string) => (
        <Badge className={getStatusColor(value)}>
          {value === "pending" && <IconClock className="h-3 w-3 mr-1" />}
          {value === "reviewing" && <IconEye className="h-3 w-3 mr-1" />}
          {value === "resolved" && <IconCheck className="h-3 w-3 mr-1" />}
          {value === "rejected" && <IconX className="h-3 w-3 mr-1" />}
          <span className="capitalize">{value}</span>
        </Badge>
      ),
    },
    {
      key: "assigned_to",
      label: "Assigned",
      render: (_, report: Report) =>
        report.assigned_admin ? (
          <div className="flex items-center gap-2">
            <Avatar className="h-5 w-5">
              <AvatarFallback className="text-xs">
                {report.assigned_admin.username?.[0]?.toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <span className="text-xs">{report.assigned_admin.username}</span>
          </div>
        ) : (
          <span className="text-xs text-muted-foreground">Unassigned</span>
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
      render: (_, report: Report) => (
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewReportDetails(report)}
          >
            <IconEye className="h-3 w-3" />
          </Button>
          {report.status === "pending" && (
            <>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleReportAction(report, "resolve")}
              >
                <IconCheck className="h-3 w-3" />
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleReportAction(report, "reject")}
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
    {
      label: "Resolve Reports",
      action: "resolve",
      variant: "default" as const,
    },
    {
      label: "Reject Reports",
      action: "reject",
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
              <h1 className="text-2xl font-bold">Reports Management</h1>
              <p className="text-muted-foreground">
                Review and moderate user reports
              </p>
            </div>
          </div>

          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList>
              <TabsTrigger value="reports">Reports</TabsTrigger>
              <TabsTrigger value="stats">Statistics</TabsTrigger>
            </TabsList>

            <TabsContent value="reports" className="space-y-4">
              <DataTable
                title="Reports"
                description={`Manage ${pagination?.total || 0} reports`}
                data={reports}
                columns={columns}
                loading={loading}
                pagination={pagination}
                searchPlaceholder="Search reports by reason or description..."
                onPageChange={handlePageChange}
                onFilter={handleFilter}
                onRowSelect={handleRowSelect}
                bulkActions={bulkActions}
                onBulkAction={handleBulkAction}
                onRefresh={fetchReports}
                onExport={() => console.log("Export reports")}
              />
            </TabsContent>

            <TabsContent value="stats" className="space-y-4">
              {reportStats && (
                <>
                  {/* Report Statistics */}
                  <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Total Reports</CardDescription>
                        <CardTitle className="text-2xl">
                          {reportStats.total_reports?.toLocaleString() || 0}
                        </CardTitle>
                      </CardHeader>
                    </Card>
                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Pending</CardDescription>
                        <CardTitle className="text-2xl">
                          {reportStats.pending_reports?.toLocaleString() || 0}
                        </CardTitle>
                      </CardHeader>
                    </Card>
                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Resolved</CardDescription>
                        <CardTitle className="text-2xl">
                          {reportStats.resolved_reports?.toLocaleString() || 0}
                        </CardTitle>
                      </CardHeader>
                    </Card>
                    <Card>
                      <CardHeader className="pb-2">
                        <CardDescription>Resolution Rate</CardDescription>
                        <CardTitle className="text-2xl">
                          {reportStats.resolution_rate?.toFixed(1) || 0}%
                        </CardTitle>
                      </CardHeader>
                    </Card>
                  </div>

                  {/* Reports by Reason */}
                  <Card>
                    <CardHeader>
                      <CardTitle>Reports by Reason</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        {reportStats.reports_by_reason?.map(
                          (item: any, index: number) => (
                            <div
                              key={index}
                              className="flex items-center justify-between"
                            >
                              <div className="flex items-center gap-2">
                                {getReportReasonIcon(item.reason)}
                                <span className="capitalize">
                                  {item.reason.replace("_", " ")}
                                </span>
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

      {/* Report Details Dialog */}
      <Dialog open={showReportDetails} onOpenChange={setShowReportDetails}>
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Report Details</DialogTitle>
          </DialogHeader>

          {selectedReport && (
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      {getReportReasonIcon(selectedReport.reason)}
                      <div>
                        <CardTitle className="flex items-center gap-2">
                          {selectedReport.reason.replace("_", " ")}
                          <Badge
                            className={getPriorityColor(
                              selectedReport.priority
                            )}
                          >
                            {selectedReport.priority}
                          </Badge>
                        </CardTitle>
                        <CardDescription>
                          Reported on{" "}
                          {new Date(
                            selectedReport.created_at
                          ).toLocaleDateString()}
                        </CardDescription>
                      </div>
                    </div>
                    <Badge className={getStatusColor(selectedReport.status)}>
                      {selectedReport.status}
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <h4 className="font-medium mb-2">Description</h4>
                    <p className="text-sm bg-gray-50 p-3 rounded">
                      {selectedReport.description}
                    </p>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <h4 className="font-medium mb-2">Reporter</h4>
                      {selectedReport.reporter && (
                        <div className="flex items-center gap-2">
                          <Avatar>
                            <AvatarImage
                              src={selectedReport.reporter.profile_picture}
                            />
                            <AvatarFallback>
                              {selectedReport.reporter.username?.[0]?.toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div>
                            <p className="font-medium">
                              {selectedReport.reporter.username}
                            </p>
                            <p className="text-sm text-muted-foreground">
                              {selectedReport.reporter.email}
                            </p>
                          </div>
                        </div>
                      )}
                    </div>

                    <div>
                      <h4 className="font-medium mb-2">Target</h4>
                      <div className="flex items-center gap-2">
                        {getTargetTypeIcon(selectedReport.target_type)}
                        <div>
                          <Badge variant="secondary">
                            {selectedReport.target_type}
                          </Badge>
                          {selectedReport.target_user && (
                            <p className="text-sm text-muted-foreground mt-1">
                              @{selectedReport.target_user.username}
                            </p>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>

                  {selectedReport.evidence_urls &&
                    selectedReport.evidence_urls.length > 0 && (
                      <div>
                        <h4 className="font-medium mb-2">Evidence</h4>
                        <div className="grid grid-cols-2 gap-2">
                          {selectedReport.evidence_urls.map((url, index) => (
                            <img
                              key={index}
                              src={url}
                              alt={`Evidence ${index + 1}`}
                              className="rounded-lg object-cover h-24 w-full"
                            />
                          ))}
                        </div>
                      </div>
                    )}

                  {selectedReport.resolution && (
                    <div>
                      <h4 className="font-medium mb-2">Resolution</h4>
                      <p className="text-sm bg-green-50 p-3 rounded">
                        {selectedReport.resolution}
                      </p>
                      {selectedReport.resolution_note && (
                        <p className="text-sm text-muted-foreground mt-2">
                          Note: {selectedReport.resolution_note}
                        </p>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowReportDetails(false)}
            >
              Close
            </Button>
            {selectedReport && selectedReport.status === "pending" && (
              <>
                <Button
                  variant="outline"
                  onClick={() => {
                    setShowReportDetails(false);
                    handleReportAction(selectedReport, "reject");
                  }}
                >
                  Reject Report
                </Button>
                <Button
                  onClick={() => {
                    setShowReportDetails(false);
                    handleReportAction(selectedReport, "resolve");
                  }}
                >
                  Resolve Report
                </Button>
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
              {actionType === "resolve" ? "Resolve Report" : "Reject Report"}
            </DialogTitle>
            <DialogDescription>
              {actionType === "resolve"
                ? "Mark this report as resolved and provide a resolution."
                : "Reject this report and provide a reason."}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {actionType === "resolve" && (
              <div>
                <label className="text-sm font-medium">Resolution</label>
                <Select value={resolution} onValueChange={setResolution}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select resolution" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="content_removed">
                      Content Removed
                    </SelectItem>
                    <SelectItem value="user_warned">User Warned</SelectItem>
                    <SelectItem value="user_suspended">
                      User Suspended
                    </SelectItem>
                    <SelectItem value="no_action">
                      No Action Required
                    </SelectItem>
                    <SelectItem value="policy_updated">
                      Policy Updated
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}
            <div>
              <label className="text-sm font-medium">
                {actionType === "resolve" ? "Note" : "Reason"}
              </label>
              <Textarea
                placeholder={`Enter ${
                  actionType === "resolve" ? "note" : "reason"
                }...`}
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
              onClick={executeReportAction}
              variant={actionType === "reject" ? "destructive" : "default"}
              disabled={
                !actionReason.trim() ||
                (actionType === "resolve" && !resolution)
              }
            >
              {actionType === "resolve" ? "Resolve Report" : "Reject Report"}
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
              You are about to {bulkAction} {selectedIds.length} report(s).
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {bulkAction === "resolve" && (
              <div>
                <label className="text-sm font-medium">Resolution</label>
                <Select value={resolution} onValueChange={setResolution}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select resolution" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="content_removed">
                      Content Removed
                    </SelectItem>
                    <SelectItem value="user_warned">User Warned</SelectItem>
                    <SelectItem value="user_suspended">
                      User Suspended
                    </SelectItem>
                    <SelectItem value="no_action">
                      No Action Required
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}
            <div>
              <label className="text-sm font-medium">
                {bulkAction === "resolve" ? "Note" : "Reason"}
              </label>
              <Textarea
                placeholder={`Enter ${
                  bulkAction === "resolve" ? "note" : "reason"
                } for bulk action...`}
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
              variant={bulkAction === "reject" ? "destructive" : "default"}
              disabled={
                !actionReason.trim() ||
                (bulkAction === "resolve" && !resolution)
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

export default withAuth(ReportsPage);
