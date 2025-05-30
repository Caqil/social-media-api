"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { SiteHeader } from "@/components/site-header";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { DataTable } from "@/components/data-table";
import { withAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import { requestManager } from "@/lib/request-manager";
import {
  User,
  UserResponse,
  UserFilter,
  PaginationMeta,
  UserRole,
} from "@/types/admin";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  IconAlertCircle,
  IconCheck,
  IconDots,
  IconEye,
  IconFilter,
  IconLoader2,
  IconLock,
  IconMail,
  IconPencil,
  IconRefresh,
  IconSearch,
  IconShield,
  IconTrash,
  IconUser,
  IconUserCheck,
  IconUserOff,
  IconUsers,
  IconX,
} from "@tabler/icons-react";

// Type for API errors
interface ApiError {
  response?: {
    data?: {
      message?: string;
    };
  };
  message?: string;
}

// Cache key for users data
const USERS_CACHE_KEY = "admin-users-list";

function UsersPage() {
  // State for users data
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<PaginationMeta | null>(null);

  // State for filters and search
  const [filters, setFilters] = useState<Partial<UserFilter>>({
    page: 1,
    limit: 10,
    search: "",
  });

  // State for selected user and actions
  const [selectedUser, setSelectedUser] = useState<UserResponse | null>(null);
  const [showUserDetails, setShowUserDetails] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [showSuspendConfirm, setShowSuspendConfirm] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [selectedRows, setSelectedRows] = useState<string[]>([]);
  const [showAdvancedFilters, setShowAdvancedFilters] = useState(false);

  // Fetch users data with requestManager
  const fetchUsers = useCallback(
    async (forceRefresh = false) => {
      try {
        setLoading(true);
        setError(null);

        // Create cache key with filters
        const cacheKey = `${USERS_CACHE_KEY}-${JSON.stringify(filters)}`;

        if (forceRefresh) {
          requestManager.clearCache(cacheKey);
        }

        const response = await requestManager.request(
          cacheKey,
          () => apiClient.getUsers(filters),
          {
            cache: !forceRefresh,
            cacheDuration: 30000, // 30 seconds
          }
        );

        if (response.data && Array.isArray(response.data)) {
          setUsers(response.data);
          setPagination(response.pagination || null);
        } else {
          setError("Invalid response format");
        }
      } catch (err) {
        console.error("Failed to fetch users:", err);

        const error = err as ApiError;
        const errorMessage =
          error.response?.data?.message ||
          error.message ||
          "Failed to load users";

        setError(errorMessage);
      } finally {
        setLoading(false);
      }
    },
    [filters]
  );

  // Fetch user details
  const fetchUserDetails = useCallback(async (userId: string) => {
    try {
      setActionLoading(true);

      const response = await apiClient.getUser(userId);

      if (response.data) {
        setSelectedUser(response.data as UserResponse);
        setShowUserDetails(true);
      }
    } catch (err) {
      console.error("Failed to fetch user details:", err);

      const error = err as ApiError;
      const errorMessage =
        error.response?.data?.message ||
        error.message ||
        "Failed to load user details";

      setError(errorMessage);
    } finally {
      setActionLoading(false);
    }
  }, []);

  // Initial data load
  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  // Handle search
  const handleSearch = (value: string) => {
    setFilters({
      ...filters,
      search: value,
      page: 1, // Reset to first page on new search
    });
  };

  // Handle filter change
  const handleFilterChange = (key: string, value: any) => {
    if (key === "role") {
      setFilters({
        ...filters,
        [key]: value as UserRole,
        page: 1, // Reset to first page on filter change
      });
    } else {
      setFilters({
        ...filters,
        [key]: value,
        page: 1, // Reset to first page on filter change
      });
    }
  };

  // Handle page change
  const handlePageChange = (page: number) => {
    setFilters({
      ...filters,
      page,
    });
  };

  // Handle user actions
  const handleViewUser = (userId: string) => {
    fetchUserDetails(userId);
  };

  const handleSuspendUser = async () => {
    if (!selectedUser) return;

    try {
      setActionLoading(true);

      // Check if the method exists in your API client
      // If not, use a mock implementation
      if (typeof apiClient.updateUserStatus === "function") {
        await apiClient.updateUserStatus(selectedUser.id, {
          is_active: true,
          is_suspended: true,
          reason: "Administrative action",
        });
      } else {
        console.log(`Would suspend user: ${selectedUser.id}`);
        await new Promise((resolve) => setTimeout(resolve, 1000)); // Simulate API call
      }

      // Refresh users list
      fetchUsers(true);
      setShowSuspendConfirm(false);
      setSelectedUser(null);
    } catch (err) {
      console.error("Failed to suspend user:", err);

      const error = err as ApiError;
      const errorMessage =
        error.response?.data?.message ||
        error.message ||
        "Failed to suspend user";

      setError(errorMessage);
    } finally {
      setActionLoading(false);
    }
  };

  const handleDeleteUser = async () => {
    if (!selectedUser) return;

    try {
      setActionLoading(true);

      // Since deleteUser doesn't exist, we'll use a more generic approach
      // You may need to implement this method in your API client
      /*
      await apiClient.deleteUser(selectedUser.id, {
        reason: "Administrative action"
      });
      */

      // For now, let's use a mock implementation
      console.log(`Would delete user: ${selectedUser.id}`);
      await new Promise((resolve) => setTimeout(resolve, 1000)); // Simulate API call

      // Refresh users list
      fetchUsers(true);
      setShowDeleteConfirm(false);
      setSelectedUser(null);
    } catch (err) {
      console.error("Failed to delete user:", err);

      const error = err as ApiError;
      const errorMessage =
        error.response?.data?.message ||
        error.message ||
        "Failed to delete user";

      setError(errorMessage);
    } finally {
      setActionLoading(false);
    }
  };

  // Handle bulk actions
  const handleBulkAction = async (action: string, userIds: string[]) => {
    try {
      setActionLoading(true);

      // Check if the method exists in your API client
      // If not, use a mock implementation
      if (typeof apiClient.bulkUserAction === "function") {
        await apiClient.bulkUserAction({
          user_ids: userIds,
          action,
          reason: "Administrative action",
        });
      } else {
        console.log(`Would perform bulk action "${action}" on users:`, userIds);
        await new Promise((resolve) => setTimeout(resolve, 1000)); // Simulate API call
      }

      // Refresh users list
      fetchUsers(true);
      setSelectedRows([]);
    } catch (err) {
      console.error(`Failed to ${action} users:`, err);

      const error = err as ApiError;
      const errorMessage =
        error.response?.data?.message ||
        error.message ||
        `Failed to ${action} users`;

      setError(errorMessage);
    } finally {
      setActionLoading(false);
    }
  };

  // Clear all filters
  const clearFilters = () => {
    setFilters({
      page: 1,
      limit: 10,
      search: "",
    });
    setShowAdvancedFilters(false);
  };

  // Format role for display
  const formatRole = (role: string | undefined | null): string => {
    if (!role) {
      return "user"; // Fallback value for undefined or null
    }
    return role.replace("_", " ").replace(/\b\w/g, (l) => l.toUpperCase());
  };

  // Define table columns
  const columns = useMemo(
    () => [
      {
        key: "profile",
        label: "User",
        width: "auto",
        render: (_: unknown, user: UserResponse) => (
          <div className="flex items-center gap-3">
            <Avatar className="h-9 w-9">
              <AvatarImage
                src={user.profile_picture || ""}
                alt={user.username}
              />
              <AvatarFallback className="bg-primary/10 text-primary">
                {user.first_name?.[0] || user.username?.[0] || "U"}
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-col">
              <span className="font-medium">
                {user.first_name
                  ? `${user.first_name} ${user.last_name || ""}`
                  : user.username}
              </span>
              <span className="text-muted-foreground text-xs">
                {user.email}
              </span>
            </div>
          </div>
        ),
      },
      {
        key: "role",
        label: "Role",
        width: "120px",
        filterable: true,
        sortable: true,
        render: (role: string) => (
          <Badge
            variant={
              role === "admin" || role === "super_admin" ? "default" : "outline"
            }
          >
            {formatRole(role)}
          </Badge>
        ),
      },
      {
        key: "is_verified",
        label: "Verified",
        width: "100px",
        filterable: true,
        render: (verified: boolean) => (
          <div className="flex justify-center">
            {verified ? (
              <Badge variant="default" className="bg-green-100 text-green-800">
                <IconCheck className="h-3 w-3 mr-1" />
                Yes
              </Badge>
            ) : (
              <Badge variant="outline" className="bg-amber-50 text-amber-800">
                <IconX className="h-3 w-3 mr-1" />
                No
              </Badge>
            )}
          </div>
        ),
      },
      {
        key: "status",
        label: "Status",
        width: "120px",
        filterable: true,
        render: (_: any, user: UserResponse) => {
          if (user.is_suspended) {
            return (
              <Badge variant="destructive">
                <IconUserOff className="h-3 w-3 mr-1" />
                Suspended
              </Badge>
            );
          } else if (!user.is_active) {
            return (
              <Badge variant="outline" className="bg-gray-100 text-gray-800">
                Inactive
              </Badge>
            );
          } else {
            return (
              <Badge variant="outline" className="bg-green-50 text-green-800">
                <IconUserCheck className="h-3 w-3 mr-1" />
                Active
              </Badge>
            );
          }
        },
      },
      {
        key: "created_at",
        label: "Joined",
        width: "150px",
        sortable: true,
        render: (date: string) => (
          <span className="text-muted-foreground text-sm">
            {new Date(date).toLocaleDateString()}
          </span>
        ),
      },
      {
        key: "role",
        label: "Role",
        width: "120px",
        filterable: true,
        sortable: true,
        render: (role: string | undefined | null) => (
          <Badge
            variant={
              role === "admin" || role === "super_admin" ? "default" : "outline"
            }
          >
            {formatRole(role)}
          </Badge>
        ),
      },
    ],
    []
  );

  // Table bulk actions
  const bulkActions = [
    {
      label: "Verify Users",
      action: "verify",
      icon: IconUserCheck,
    },
    {
      label: "Suspend Users",
      action: "suspend",
      icon: IconLock,
    },
    {
      label: "Delete Users",
      action: "delete",
      variant: "destructive" as const,
      icon: IconTrash,
    },
  ];

  // Render error message
  if (error && !loading) {
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
          <div className="flex h-screen items-center justify-center p-6">
            <Alert variant="destructive" className="max-w-md">
              <IconAlertCircle className="h-5 w-5 mr-2" />
              <AlertDescription className="space-y-4">
                <div className="text-lg font-medium">Failed to load users</div>
                <div>{error}</div>
                <Button
                  onClick={() => fetchUsers(true)}
                  className="w-full"
                  disabled={loading}
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

        {/* Loading progress bar */}
        {loading && (
          <div className="fixed top-0 left-0 right-0 z-50">
            <div
              className="h-1 bg-primary transition-all duration-300 ease-in-out"
              style={{ width: "70%" }}
            />
          </div>
        )}

        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            <div className="flex flex-col gap-4 p-4 md:gap-6 md:p-6">
              {/* Page Header */}
              <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <h1 className="text-2xl font-bold tracking-tight">Users</h1>
                  <p className="text-muted-foreground">
                    Manage user accounts, permissions and status
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <Button>
                    <IconUser className="h-4 w-4 mr-2" />
                    Add New User
                  </Button>
                </div>
              </div>

              {/* Users Table Card */}
              <Card className="overflow-hidden">
                <CardHeader className="bg-muted/50 pb-3">
                  <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                      <CardTitle className="text-lg">User Accounts</CardTitle>
                      <CardDescription>
                        {pagination
                          ? `Showing ${
                              (pagination.current_page - 1) *
                                pagination.per_page +
                              1
                            } to ${Math.min(
                              pagination.current_page * pagination.per_page,
                              pagination.total
                            )} of ${pagination.total} users`
                          : "Loading users..."}
                      </CardDescription>
                    </div>

                    {/* Search and Filters */}
                    <div className="flex flex-wrap items-center gap-2">
                      <div className="relative max-w-sm">
                        <IconSearch className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                        <Input
                          placeholder="Search users..."
                          className="pl-9 w-full md:w-60"
                          value={filters.search || ""}
                          onChange={(e) => handleSearch(e.target.value)}
                        />
                      </div>

                      <Button
                        variant={showAdvancedFilters ? "default" : "outline"}
                        size="sm"
                        onClick={() =>
                          setShowAdvancedFilters(!showAdvancedFilters)
                        }
                      >
                        <IconFilter className="h-4 w-4 mr-2" />
                        Filters
                        {(filters.role ||
                          filters.is_active !== undefined ||
                          filters.is_verified !== undefined) && (
                          <Badge variant="secondary" className="ml-2">
                            {[
                              filters.role ? 1 : 0,
                              filters.is_active !== undefined ? 1 : 0,
                              filters.is_verified !== undefined ? 1 : 0,
                            ].reduce((a, b) => a + b, 0)}
                          </Badge>
                        )}
                      </Button>

                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => fetchUsers(true)}
                        disabled={loading}
                      >
                        <IconRefresh
                          className={`h-4 w-4 ${loading ? "animate-spin" : ""}`}
                        />
                      </Button>
                    </div>
                  </div>

                  {/* Advanced Filters */}
                  {showAdvancedFilters && (
                    <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4 mt-4 p-4 bg-background rounded-lg border">
                      <div className="space-y-2">
                        <Label>Role</Label>
                        <Select
                          value={filters.role || ""}
                          onValueChange={(value: string) => {
                            if (value === "") {
                              const newFilters = { ...filters };
                              delete newFilters.role;
                              setFilters(newFilters);
                            } else {
                              handleFilterChange("role", value as UserRole);
                            }
                          }}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="All roles" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="">All roles</SelectItem>
                            <SelectItem value="user">User</SelectItem>
                            <SelectItem value="moderator">Moderator</SelectItem>
                            <SelectItem value="admin">Admin</SelectItem>
                            <SelectItem value="super_admin">
                              Super Admin
                            </SelectItem>
                          </SelectContent>
                        </Select>
                      </div>

                      <div className="space-y-2">
                        <Label>Status</Label>
                        <Select
                          value={
                            filters.is_active === undefined
                              ? ""
                              : filters.is_active
                              ? "active"
                              : "inactive"
                          }
                          onValueChange={(value) => {
                            if (value === "") {
                              const newFilters = { ...filters };
                              delete newFilters.is_active;
                              setFilters(newFilters);
                            } else {
                              handleFilterChange(
                                "is_active",
                                value === "active"
                              );
                            }
                          }}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Any status" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="">Any status</SelectItem>
                            <SelectItem value="active">Active</SelectItem>
                            <SelectItem value="inactive">Inactive</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>

                      <div className="space-y-2">
                        <Label>Verification</Label>
                        <Select
                          value={
                            filters.is_verified === undefined
                              ? ""
                              : filters.is_verified
                              ? "verified"
                              : "unverified"
                          }
                          onValueChange={(value) => {
                            if (value === "") {
                              const newFilters = { ...filters };
                              delete newFilters.is_verified;
                              setFilters(newFilters);
                            } else {
                              handleFilterChange(
                                "is_verified",
                                value === "verified"
                              );
                            }
                          }}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Any verification" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="">Any verification</SelectItem>
                            <SelectItem value="verified">Verified</SelectItem>
                            <SelectItem value="unverified">
                              Unverified
                            </SelectItem>
                          </SelectContent>
                        </Select>
                      </div>

                      <div className="space-y-2 md:col-span-3 lg:col-span-1">
                        <Label>&nbsp;</Label>
                        <div className="flex items-center gap-2">
                          <Button
                            variant="outline"
                            className="flex-1"
                            onClick={clearFilters}
                          >
                            <IconX className="h-4 w-4 mr-2" />
                            Clear Filters
                          </Button>
                        </div>
                      </div>
                    </div>
                  )}
                </CardHeader>

                <CardContent className="p-0">
                  <DataTable
                    data={users}
                    columns={columns}
                    loading={loading}
                    pagination={pagination || undefined}
                    onPageChange={handlePageChange}
                    onRowSelect={setSelectedRows}
                    bulkActions={bulkActions}
                    onBulkAction={handleBulkAction}
                    emptyMessage="No users found matching your criteria"
                    showRefresh={false} // We handle this separately
                    onFilter={() => {}} // We handle filters separately
                  />
                </CardContent>
              </Card>

              {/* Quick Stats */}
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium">
                      Total Users
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-center justify-between">
                      <div className="text-2xl font-bold">
                        {loading ? (
                          <Skeleton className="h-8 w-16" />
                        ) : (
                          pagination?.total.toLocaleString() || 0
                        )}
                      </div>
                      <div className="rounded-full bg-blue-100 p-2 text-blue-600">
                        <IconUsers className="h-4 w-4" />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium">
                      Active Users
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-center justify-between">
                      <div className="text-2xl font-bold">
                        {loading ? (
                          <Skeleton className="h-8 w-16" />
                        ) : (
                          users
                            .filter((u) => u.is_active)
                            .length.toLocaleString()
                        )}
                      </div>
                      <div className="rounded-full bg-green-100 p-2 text-green-600">
                        <IconUserCheck className="h-4 w-4" />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium">
                      Admins
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-center justify-between">
                      <div className="text-2xl font-bold">
                        {loading ? (
                          <Skeleton className="h-8 w-16" />
                        ) : (
                          users
                            .filter(
                              (u) =>
                                u.role === "admin" || u.role === "super_admin"
                            )
                            .length.toLocaleString()
                        )}
                      </div>
                      <div className="rounded-full bg-purple-100 p-2 text-purple-600">
                        <IconShield className="h-4 w-4" />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium">
                      Unverified Users
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-center justify-between">
                      <div className="text-2xl font-bold">
                        {loading ? (
                          <Skeleton className="h-8 w-16" />
                        ) : (
                          users
                            .filter((u) => !u.is_verified)
                            .length.toLocaleString()
                        )}
                      </div>
                      <div className="rounded-full bg-amber-100 p-2 text-amber-600">
                        <IconMail className="h-4 w-4" />
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          </div>
        </div>
      </SidebarInset>

      {/* User Details Dialog */}
      {selectedUser && (
        <Dialog open={showUserDetails} onOpenChange={setShowUserDetails}>
          <DialogContent className="max-w-3xl">
            <DialogHeader>
              <DialogTitle>User Details</DialogTitle>
              <DialogDescription>
                Detailed information about {selectedUser.username}
              </DialogDescription>
            </DialogHeader>

            <div className="mt-4">
              <Tabs defaultValue="profile">
                <TabsList className="mb-4">
                  <TabsTrigger value="profile">Profile</TabsTrigger>
                  <TabsTrigger value="activity">Activity</TabsTrigger>
                  <TabsTrigger value="security">Security</TabsTrigger>
                </TabsList>

                <TabsContent value="profile" className="space-y-4">
                  <div className="flex flex-col md:flex-row gap-6">
                    <div className="flex flex-col items-center">
                      <Avatar className="h-24 w-24">
                        <AvatarImage
                          src={selectedUser.profile_picture || ""}
                          alt={selectedUser.username}
                        />
                        <AvatarFallback className="text-2xl bg-primary/10 text-primary">
                          {selectedUser.first_name?.[0] ||
                            selectedUser.username?.[0] ||
                            "U"}
                        </AvatarFallback>
                      </Avatar>

                      <div className="mt-4 text-center">
                        <h3 className="font-semibold text-lg">
                          {selectedUser.first_name
                            ? `${selectedUser.first_name} ${
                                selectedUser.last_name || ""
                              }`
                            : selectedUser.username}
                        </h3>
                        <p className="text-muted-foreground text-sm">
                          {selectedUser.email}
                        </p>

                        <div className="mt-2">
                          <Badge
                            variant={
                              selectedUser.role === "admin" ||
                              selectedUser.role === "super_admin"
                                ? "default"
                                : "outline"
                            }
                          >
                            {formatRole(selectedUser.role)}
                          </Badge>
                        </div>
                      </div>

                      <div className="mt-4 flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setShowSuspendConfirm(true)}
                        >
                          <IconLock className="h-4 w-4 mr-1" />
                          Suspend
                        </Button>
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => setShowDeleteConfirm(true)}
                        >
                          <IconTrash className="h-4 w-4 mr-1" />
                          Delete
                        </Button>
                      </div>
                    </div>

                    <div className="flex-1 space-y-4">
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                          <Label className="text-muted-foreground">
                            Username
                          </Label>
                          <div className="font-medium">
                            {selectedUser.username}
                          </div>
                        </div>

                        <div>
                          <Label className="text-muted-foreground">Email</Label>
                          <div className="font-medium">
                            {selectedUser.email}
                          </div>
                        </div>

                        <div>
                          <Label className="text-muted-foreground">
                            Status
                          </Label>
                          <div>
                            {selectedUser.is_suspended ? (
                              <Badge variant="destructive">Suspended</Badge>
                            ) : !selectedUser.is_active ? (
                              <Badge variant="outline">Inactive</Badge>
                            ) : (
                              <Badge
                                variant="outline"
                                className="bg-green-50 text-green-800"
                              >
                                Active
                              </Badge>
                            )}
                          </div>
                        </div>

                        <div>
                          <Label className="text-muted-foreground">
                            Verified
                          </Label>
                          <div>
                            {selectedUser.is_verified ? (
                              <Badge
                                variant="outline"
                                className="bg-green-50 text-green-800"
                              >
                                Yes
                              </Badge>
                            ) : (
                              <Badge
                                variant="outline"
                                className="bg-amber-50 text-amber-800"
                              >
                                No
                              </Badge>
                            )}
                          </div>
                        </div>

                        <div>
                          <Label className="text-muted-foreground">
                            Joined Date
                          </Label>
                          <div className="font-medium">
                            {new Date(
                              selectedUser.created_at
                            ).toLocaleDateString()}
                          </div>
                        </div>

                        <div>
                          <Label className="text-muted-foreground">
                            Last Active
                          </Label>
                          <div className="font-medium">
                            {selectedUser.last_active_at
                              ? new Date(
                                  selectedUser.last_active_at
                                ).toLocaleString()
                              : "Never"}
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </TabsContent>

                <TabsContent value="activity">
                  <div className="rounded-md border">
                    <div className="p-4 text-center text-muted-foreground">
                      User activity data will be displayed here
                    </div>
                  </div>
                </TabsContent>

                <TabsContent value="security">
                  <div className="rounded-md border">
                    <div className="p-4 text-center text-muted-foreground">
                      User security settings will be displayed here
                    </div>
                  </div>
                </TabsContent>
              </Tabs>
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setShowUserDetails(false)}
              >
                Close
              </Button>
              <Button>
                <IconPencil className="h-4 w-4 mr-2" />
                Edit User
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      {/* Delete Confirmation Dialog */}
      <Dialog open={showDeleteConfirm} onOpenChange={setShowDeleteConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete User</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this user? This action cannot be
              undone.
            </DialogDescription>
          </DialogHeader>

          <div className="mt-4 p-4 bg-destructive/10 rounded-md">
            <p className="text-sm font-medium">User to be deleted:</p>
            <p className="text-sm">
              {selectedUser?.username} ({selectedUser?.email})
            </p>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowDeleteConfirm(false)}
              disabled={actionLoading}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDeleteUser}
              disabled={actionLoading}
            >
              {actionLoading ? (
                <>
                  <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                  Deleting...
                </>
              ) : (
                <>
                  <IconTrash className="h-4 w-4 mr-2" />
                  Delete User
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Suspend Confirmation Dialog */}
      <Dialog open={showSuspendConfirm} onOpenChange={setShowSuspendConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Suspend User</DialogTitle>
            <DialogDescription>
              Are you sure you want to suspend this user? They will be unable to
              access their account.
            </DialogDescription>
          </DialogHeader>

          <div className="mt-4 p-4 bg-amber-50 rounded-md">
            <p className="text-sm font-medium">User to be suspended:</p>
            <p className="text-sm">
              {selectedUser?.username} ({selectedUser?.email})
            </p>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowSuspendConfirm(false)}
              disabled={actionLoading}
            >
              Cancel
            </Button>
            <Button
              variant="default"
              onClick={handleSuspendUser}
              disabled={actionLoading}
            >
              {actionLoading ? (
                <>
                  <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                  Suspending...
                </>
              ) : (
                <>
                  <IconLock className="h-4 w-4 mr-2" />
                  Suspend User
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SidebarProvider>
  );
}

export default withAuth(UsersPage);
