// app/admin/users/page.tsx
"use client";

import { SetStateAction, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { withAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import { User, UserFilter, TableColumn } from "@/types/admin";
import {
  IconUser,
  IconMail,
  IconCalendar,
  IconShield,
  IconBan,
  IconCheck,
  IconX,
} from "@tabler/icons-react";

function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<any>(null);
  const [filters, setFilters] = useState<UserFilter>({ page: 1, limit: 20 });

  // Dialog states
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [showUserDetails, setShowUserDetails] = useState(false);
  const [showStatusDialog, setShowStatusDialog] = useState(false);
  const [showBulkDialog, setShowBulkDialog] = useState(false);
  const [statusAction, setStatusAction] = useState<string>("");
  const [bulkAction, setBulkAction] = useState<string>("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [actionReason, setActionReason] = useState("");

  const router = useRouter();

  // Fetch users
  const fetchUsers = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getUsers(filters);
      setUsers(response.data.data);
      setPagination(response.data.pagination);
    } catch (error: any) {
      console.error("Failed to fetch users:", error);
      setError(error.response?.data?.message || "Failed to load users");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, [filters]);

  // Handle page change
  const handlePageChange = (page: number) => {
    setFilters((prev) => ({ ...prev, page }));
  };

  // Handle sorting
  const handleSort = (column: string, direction: "asc" | "desc") => {
    // Implement sorting logic
    console.log("Sort:", column, direction);
  };

  // Handle filtering
  const handleFilter = (newFilters: Record<string, any>) => {
    setFilters((prev) => ({
      ...prev,
      ...newFilters,
      page: 1, // Reset to first page when filtering
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
      await apiClient.bulkUserAction({
        user_ids: selectedIds,
        action: bulkAction,
        reason: actionReason,
      });

      setShowBulkDialog(false);
      setActionReason("");
      fetchUsers();
    } catch (error: any) {
      console.error("Bulk action failed:", error);
      setError(error.response?.data?.message || "Bulk action failed");
    }
  };

  // Handle user status change
  const handleStatusChange = async (user: User, action: string) => {
    setSelectedUser(user);
    setStatusAction(action);
    setShowStatusDialog(true);
  };

  // Execute status change
  const executeStatusChange = async () => {
    if (!selectedUser) return;

    try {
      switch (statusAction) {
        case "suspend":
          await apiClient.updateUserStatus(selectedUser.id, {
            is_active: true,
            is_suspended: true,
            reason: actionReason,
          });
          break;
        case "activate":
          await apiClient.updateUserStatus(selectedUser.id, {
            is_active: true,
            is_suspended: false,
            reason: actionReason,
          });
          break;
        case "deactivate":
          await apiClient.updateUserStatus(selectedUser.id, {
            is_active: false,
            is_suspended: false,
            reason: actionReason,
          });
          break;
        case "verify":
          await apiClient.verifyUser(selectedUser.id);
          break;
      }

      setShowStatusDialog(false);
      setActionReason("");
      fetchUsers();
    } catch (error: any) {
      console.error("Status change failed:", error);
      setError(error.response?.data?.message || "Status change failed");
    }
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "user",
      label: "User",
      sortable: true,
      render: (_, user: User) => (
        <div className="flex items-center gap-3">
          <Avatar className="h-8 w-8">
            <AvatarImage src={user.profile_picture} alt={user.username} />
            <AvatarFallback>
              {user.first_name?.[0]}
              {user.last_name?.[0]}
            </AvatarFallback>
          </Avatar>
          <div>
            <div className="font-medium">{user.username}</div>
            <div className="text-sm text-muted-foreground">{user.email}</div>
          </div>
        </div>
      ),
    },
    {
      key: "name",
      label: "Name",
      sortable: true,
      render: (_, user: User) =>
        `${user.first_name || ""} ${user.last_name || ""}`.trim() || "-",
    },
    {
      key: "role",
      label: "Role",
      filterable: true,
      render: (value: string) => (
        <Badge variant={value === "admin" ? "default" : "secondary"}>
          {value}
        </Badge>
      ),
    },
    {
      key: "is_verified",
      label: "Verified",
      filterable: true,
      render: (value: boolean) => (
        <Badge variant={value ? "default" : "secondary"}>
          {value ? (
            <IconCheck className="h-3 w-3 mr-1" />
          ) : (
            <IconX className="h-3 w-3 mr-1" />
          )}
          {value ? "Verified" : "Unverified"}
        </Badge>
      ),
    },
    {
      key: "status",
      label: "Status",
      filterable: true,
      render: (_, user: User) => {
        if (user.is_suspended) {
          return <Badge className="bg-red-100 text-red-800">Suspended</Badge>;
        }
        if (!user.is_active) {
          return <Badge className="bg-gray-100 text-gray-800">Inactive</Badge>;
        }
        return <Badge className="bg-green-100 text-green-800">Active</Badge>;
      },
    },
    {
      key: "followers_count",
      label: "Followers",
      sortable: true,
      render: (value: number) => value?.toLocaleString() || "0",
    },
    {
      key: "posts_count",
      label: "Posts",
      sortable: true,
      render: (value: number) => value?.toLocaleString() || "0",
    },
    {
      key: "created_at",
      label: "Joined",
      sortable: true,
      render: (value: string) => new Date(value).toLocaleDateString(),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    {
      label: "Activate Users",
      action: "activate",
      variant: "default" as const,
    },
    {
      label: "Suspend Users",
      action: "suspend",
      variant: "destructive" as const,
    },
    { label: "Verify Users", action: "verify", variant: "default" as const },
    {
      label: "Delete Users",
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
              <h1 className="text-2xl font-bold">Users Management</h1>
              <p className="text-muted-foreground">
                Manage user accounts, permissions, and activities
              </p>
            </div>
          </div>

          <DataTable
            title="All Users"
            description={`Manage ${pagination?.total || 0} registered users`}
            data={users}
            columns={columns}
            loading={loading}
            pagination={pagination}
            searchPlaceholder="Search users by name, email, or username..."
            onPageChange={handlePageChange}
            onSort={handleSort}
            onFilter={handleFilter}
            onRowSelect={handleRowSelect}
            bulkActions={bulkActions}
            onBulkAction={handleBulkAction}
            onRefresh={fetchUsers}
            onExport={() => console.log("Export users")}
          />
        </div>
      </SidebarInset>

      {/* User Status Change Dialog */}
      <Dialog open={showStatusDialog} onOpenChange={setShowStatusDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Change User Status</DialogTitle>
            <DialogDescription>
              You are about to {statusAction} user: {selectedUser?.username}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label htmlFor="reason">Reason (optional)</Label>
              <Textarea
                id="reason"
                placeholder="Enter reason for this action..."
                value={actionReason}
                onChange={(e: { target: { value: SetStateAction<string> } }) =>
                  setActionReason(e.target.value)
                }
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowStatusDialog(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={executeStatusChange}
              variant={
                statusAction === "suspend" || statusAction === "delete"
                  ? "destructive"
                  : "default"
              }
            >
              {statusAction === "suspend" && "Suspend User"}
              {statusAction === "activate" && "Activate User"}
              {statusAction === "deactivate" && "Deactivate User"}
              {statusAction === "verify" && "Verify User"}
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
              You are about to {bulkAction} {selectedIds.length} user(s). This
              action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label htmlFor="bulk-reason">
                Reason (required for destructive actions)
              </Label>
              <Textarea
                id="bulk-reason"
                placeholder="Enter reason for this bulk action..."
                value={actionReason}
                onChange={(e: { target: { value: SetStateAction<string> } }) =>
                  setActionReason(e.target.value)
                }
                required={bulkAction === "suspend" || bulkAction === "delete"}
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
                bulkAction === "suspend" || bulkAction === "delete"
                  ? "destructive"
                  : "default"
              }
              disabled={
                (bulkAction === "suspend" || bulkAction === "delete") &&
                !actionReason.trim()
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

export default withAuth(UsersPage);
