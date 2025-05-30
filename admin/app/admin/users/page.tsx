// admin/app/admin/users/page.tsx - Corrected with proper API calls
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
import { User, UserRole, TableColumn, PaginationMeta } from "@/types/admin";
import {
  IconPlus,
  IconEye,
  IconEdit,
  IconTrash,
  IconDotsVertical,
  IconUser,
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
} from "@tabler/icons-react";

interface UsersPageState {
  users: User[];
  loading: boolean;
  error: string | null;
  pagination: PaginationMeta | undefined; // Changed from null to undefined
  filters: {
    search: string;
    role: string;
    status: string;
    is_verified: string;
    page: number;
    limit: number;
    sort_by?: string;
    sort_order?: "asc" | "desc";
  };
  selectedUsers: string[];
}

interface UserFormData {
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  bio: string;
  role: UserRole;
  is_active: boolean;
  is_verified: boolean;
  password?: string;
}

interface DialogState {
  addUser: boolean;
  viewUser: boolean;
  editUser: boolean;
  deleteUser: boolean;
  suspendUser: boolean;
  selectedUser: User | null;
}

const initialFormData: UserFormData = {
  username: "",
  email: "",
  first_name: "",
  last_name: "",
  bio: "",
  role: UserRole.USER,
  is_active: true,
  is_verified: false,
};

const initialFilters = {
  search: "",
  role: "all",
  status: "all",
  is_verified: "all",
  page: 1,
  limit: 20,
};

function UsersPage() {
  const { user: currentUser } = useAuth();

  const [state, setState] = useState<UsersPageState>({
    users: [],
    loading: true,
    error: null,
    pagination: undefined, // Changed from null to undefined
    filters: initialFilters,
    selectedUsers: [],
  });

  const [dialogs, setDialogs] = useState<DialogState>({
    addUser: false,
    viewUser: false,
    editUser: false,
    deleteUser: false,
    suspendUser: false,
    selectedUser: null,
  });

  const [formData, setFormData] = useState<UserFormData>(initialFormData);
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [suspensionReason, setSuspensionReason] = useState("");

  // Fetch users with proper API parameters
  const fetchUsers = useCallback(async (filters = state.filters) => {
    try {
      setState((prev) => ({ ...prev, loading: true, error: null }));

      const params: any = {
        page: filters.page,
        limit: filters.limit,
      };

      // Add search parameter
      if (filters.search) params.search = filters.search;

      // Add role filter
      if (filters.role && filters.role !== "all") params.role = filters.role;

      // Add status filters
      if (filters.status && filters.status !== "all") {
        if (filters.status === "active") {
          params.is_active = true;
          params.is_suspended = false;
        } else if (filters.status === "inactive") {
          params.is_active = false;
        } else if (filters.status === "suspended") {
          params.is_suspended = true;
        }
      }

      // Add verification filter
      if (filters.is_verified && filters.is_verified !== "all") {
        params.is_verified = filters.is_verified === "true";
      }

      // Add sorting
      if (filters.sort_by) {
        params.sort_by = filters.sort_by;
        params.sort_order = filters.sort_order || "asc";
      }

      console.log("📡 Fetching users with params:", params);
      const response = await apiClient.getUsers(params);

      setState((prev) => ({
        ...prev,
        users: response.data || [],
        pagination: response.pagination || undefined, // Changed from null to undefined
        loading: false,
      }));
    } catch (error: any) {
      console.error("❌ Failed to fetch users:", error);
      setState((prev) => ({
        ...prev,
        loading: false,
        error: error.message || "Failed to fetch users",
      }));
    }
  }, []);

  // Initial load
  useEffect(() => {
    fetchUsers();
  }, []);

  // Handle search with debounce
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (state.filters.search !== undefined) {
        fetchUsers({ ...state.filters, page: 1 });
      }
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [state.filters.search, fetchUsers]);

  // Handle filter changes
  const handleFilterChange = (key: string, value: any) => {
    const newFilters = { ...state.filters, [key]: value, page: 1 };
    setState((prev) => ({ ...prev, filters: newFilters }));

    if (key !== "search") {
      fetchUsers(newFilters);
    }
  };

  // Handle pagination
  const handlePageChange = (page: number) => {
    const newFilters = { ...state.filters, page };
    setState((prev) => ({ ...prev, filters: newFilters }));
    fetchUsers(newFilters);
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
    fetchUsers(newFilters);
  };

  // Handle row selection
  const handleRowSelect = (selectedRows: string[]) => {
    setState((prev) => ({ ...prev, selectedUsers: selectedRows }));
  };

  // Open dialog
  const openDialog = (type: keyof DialogState, user: User | null = null) => {
    setDialogs((prev) => ({ ...prev, [type]: true, selectedUser: user }));

    if (type === "editUser" && user) {
      setFormData({
        username: user.username,
        email: user.email,
        first_name: user.first_name || "",
        last_name: user.last_name || "",
        bio: user.bio || "",
        role: user.role,
        is_active: user.is_active,
        is_verified: user.is_verified,
      });
    }
  };

  // Close dialog
  const closeDialog = (type: keyof DialogState) => {
    setDialogs((prev) => ({ ...prev, [type]: false, selectedUser: null }));
    setFormData(initialFormData);
    setFormError(null);
    setSuspensionReason("");
  };

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormLoading(true);
    setFormError(null);

    try {
      if (dialogs.addUser) {
        // Add new user using createUser method
        const userData = { ...formData };

        // Ensure password is always provided
        if (!userData.password || userData.password.trim() === "") {
          userData.password = generatePassword();
        }

        // Create user data with required password
        const createUserData = {
          username: userData.username,
          email: userData.email,
          password: userData.password, // Now guaranteed to be a string
          first_name: userData.first_name,
          last_name: userData.last_name,
          bio: userData.bio,
          role: userData.role as string, // Convert UserRole enum to string
          is_active: userData.is_active,
          is_verified: userData.is_verified,
        };

        await apiClient.createUser(createUserData);
        closeDialog("addUser");
        fetchUsers();
      } else if (dialogs.editUser && dialogs.selectedUser) {
        // Update existing user using updateUser method
        const { password, ...updateData } = formData;
        const updateUserData = {
          ...updateData,
          role: updateData.role as string, // Convert UserRole enum to string
        };
        await apiClient.updateUser(dialogs.selectedUser.id, updateUserData);
        closeDialog("editUser");
        fetchUsers();
      }
    } catch (error: any) {
      setFormError(error.message || "Operation failed");
    } finally {
      setFormLoading(false);
    }
  };

  // Handle user deletion
  const handleDeleteUser = async () => {
    if (!dialogs.selectedUser) return;

    setFormLoading(true);
    try {
      await apiClient.deleteUser(dialogs.selectedUser.id, "Deleted by admin");
      closeDialog("deleteUser");
      fetchUsers();
    } catch (error: any) {
      setFormError(error.message || "Failed to delete user");
    } finally {
      setFormLoading(false);
    }
  };

  // Handle user suspension
  const handleSuspendUser = async () => {
    if (!dialogs.selectedUser) return;

    setFormLoading(true);
    try {
      await apiClient.updateUserStatus(dialogs.selectedUser.id, {
        is_suspended: !dialogs.selectedUser.is_suspended,
        reason: suspensionReason,
      });
      closeDialog("suspendUser");
      fetchUsers();
    } catch (error: any) {
      setFormError(error.message || "Failed to update user status");
    } finally {
      setFormLoading(false);
    }
  };

  // Handle bulk actions
  const handleBulkAction = async (action: string, selectedIds: string[]) => {
    setFormLoading(true);
    try {
      await apiClient.bulkUserAction({
        user_ids: selectedIds,
        action,
        reason: action.includes("suspend") ? "Bulk action" : undefined,
      });
      setState((prev) => ({ ...prev, selectedUsers: [] }));
      fetchUsers();
    } catch (error: any) {
      setState((prev) => ({
        ...prev,
        error: error.message || "Bulk action failed",
      }));
    } finally {
      setFormLoading(false);
    }
  };

  // Generate random password
  const generatePassword = () => {
    return (
      Math.random().toString(36).slice(-8) +
      Math.random().toString(36).slice(-8)
    );
  };

  // Handle refresh
  const handleRefresh = () => {
    fetchUsers();
  };

  // Handle export
  const handleExport = async () => {
    try {
      await apiClient.exportUsers();
      console.log("✅ Export initiated");
    } catch (error) {
      console.error("❌ Export failed:", error);
    }
  };

  // Format user status
  const formatUserStatus = (user: User) => {
    if (!user) return "Unknown";
    if (user.is_suspended) return "Suspended";
    return user.is_active ? "Active" : "Inactive";
  };

  // Get status badge variant
  const getStatusBadgeVariant = (
    user: User
  ): "default" | "secondary" | "destructive" => {
    if (!user) return "secondary";
    if (user.is_suspended) return "destructive";
    return user.is_active ? "default" : "secondary";
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "profile_picture",
      label: "",
      width: "w-12",
      render: (value: string, user: User) => (
        <Avatar className="h-8 w-8">
          <AvatarImage src={value} alt={user.username || "User"} />
          <AvatarFallback>
            {(user.first_name?.[0] || user.username?.[0] || "U").toUpperCase()}
          </AvatarFallback>
        </Avatar>
      ),
    },
    {
      key: "username",
      label: "Username",
      sortable: true,
      render: (value: string, user: User) => (
        <div className="flex flex-col">
          <span className="font-medium">{value || "Unknown"}</span>
          <span className="text-sm text-muted-foreground">
            {user.first_name || ""} {user.last_name || ""}
          </span>
        </div>
      ),
    },
    {
      key: "email",
      label: "Email",
      sortable: true,
      render: (value: string, user: User) => (
        <div className="flex items-center gap-2">
          <span>{value}</span>
          {user.is_verified && (
            <Badge
              variant="outline"
              className="text-green-600 border-green-200"
            >
              <IconCheck className="h-3 w-3 mr-1" />
              Verified
            </Badge>
          )}
        </div>
      ),
    },
    {
      key: "role",
      label: "Role",
      sortable: true,
      filterable: true,
      render: (value: UserRole) => {
        // Handle undefined/null values
        const roleValue = value || "user";
        return (
          <Badge
            variant={
              roleValue === UserRole.ADMIN || roleValue === UserRole.SUPER_ADMIN
                ? "default"
                : "secondary"
            }
            className={
              roleValue === UserRole.SUPER_ADMIN
                ? "bg-purple-100 text-purple-800"
                : roleValue === UserRole.ADMIN
                ? "bg-blue-100 text-blue-800"
                : roleValue === UserRole.MODERATOR
                ? "bg-green-100 text-green-800"
                : ""
            }
          >
            {roleValue.replace("_", " ").toUpperCase()}
          </Badge>
        );
      },
    },
    {
      key: "status",
      label: "Status",
      sortable: true,
      filterable: true,
      render: (_, user: User) => (
        <Badge variant={getStatusBadgeVariant(user)}>
          {formatUserStatus(user)}
        </Badge>
      ),
    },
    {
      key: "created_at",
      label: "Created",
      sortable: true,
      render: (value: string) => (
        <span className="text-sm text-muted-foreground">
          {value ? new Date(value).toLocaleDateString() : "Unknown"}
        </span>
      ),
    },
    {
      key: "last_active_at",
      label: "Last Active",
      sortable: true,
      render: (value: string) => (
        <span className="text-sm text-muted-foreground">
          {value ? new Date(value).toLocaleDateString() : "Never"}
        </span>
      ),
    },
    {
      key: "actions",
      label: "",
      width: "w-12",
      render: (_, user: User) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
              <IconDotsVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => openDialog("viewUser", user)}>
              <IconEye className="h-4 w-4 mr-2" />
              View Details
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => openDialog("editUser", user)}>
              <IconEdit className="h-4 w-4 mr-2" />
              Edit User
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => openDialog("suspendUser", user)}
              className={
                user.is_suspended ? "text-green-600" : "text-orange-600"
              }
            >
              {user.is_suspended ? (
                <>
                  <IconCheck className="h-4 w-4 mr-2" />
                  Unsuspend
                </>
              ) : (
                <>
                  <IconBan className="h-4 w-4 mr-2" />
                  Suspend
                </>
              )}
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => openDialog("deleteUser", user)}
              className="text-red-600"
            >
              <IconTrash className="h-4 w-4 mr-2" />
              Delete User
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    { label: "Activate Users", action: "activate" },
    { label: "Deactivate Users", action: "deactivate" },
    { label: "Verify Users", action: "verify" },
    {
      label: "Suspend Users",
      action: "suspend",
      variant: "destructive" as const,
    },
    {
      label: "Delete Users",
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
                <div>Failed to load users: {state.error}</div>
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
              <h1 className="text-3xl font-bold">Users</h1>
              <p className="text-muted-foreground">
                Manage user accounts and permissions
              </p>
            </div>
            <Button onClick={() => openDialog("addUser")}>
              <IconUserPlus className="h-4 w-4 mr-2" />
              Add User
            </Button>
          </div>

          {/* Stats Cards */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Total Users
                </CardTitle>
                <IconUsers className="h-4 w-4 text-muted-foreground" />
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
                  Active Users
                </CardTitle>
                <IconUser className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {
                    state.users.filter(
                      (u) => u && u.is_active && !u.is_suspended
                    ).length
                  }
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Verified Users
                </CardTitle>
                <IconCheck className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {state.users.filter((u) => u && u.is_verified).length}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Suspended Users
                </CardTitle>
                <IconUserX className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {state.users.filter((u) => u && u.is_suspended).length}
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
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div>
                  <Label htmlFor="search">Search</Label>
                  <div className="relative">
                    <IconSearch className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="search"
                      placeholder="Search users..."
                      value={state.filters.search}
                      onChange={(e) =>
                        handleFilterChange("search", e.target.value)
                      }
                      className="pl-9"
                    />
                  </div>
                </div>
                <div>
                  <Label htmlFor="role">Role</Label>
                  <Select
                    value={state.filters.role}
                    onValueChange={(value) => handleFilterChange("role", value)}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All roles" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All roles</SelectItem>
                      <SelectItem value="user">User</SelectItem>
                      <SelectItem value="moderator">Moderator</SelectItem>
                      <SelectItem value="admin">Admin</SelectItem>
                      <SelectItem value="super_admin">Super Admin</SelectItem>
                    </SelectContent>
                  </Select>
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
                      <SelectItem value="active">Active</SelectItem>
                      <SelectItem value="inactive">Inactive</SelectItem>
                      <SelectItem value="suspended">Suspended</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="verified">Verification</Label>
                  <Select
                    value={state.filters.is_verified}
                    onValueChange={(value) =>
                      handleFilterChange("is_verified", value)
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All users" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All users</SelectItem>
                      <SelectItem value="true">Verified</SelectItem>
                      <SelectItem value="false">Unverified</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Data Table */}
          <DataTable
            data={state.users}
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
            title="User Management"
            description="View and manage all user accounts"
            emptyMessage="No users found"
            searchPlaceholder="Search users..."
          />
        </div>

        {/* Add/Edit User Dialog */}
        <Dialog
          open={dialogs.addUser || dialogs.editUser}
          onOpenChange={() => {
            if (dialogs.addUser) closeDialog("addUser");
            if (dialogs.editUser) closeDialog("editUser");
          }}
        >
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle>
                {dialogs.addUser ? "Add New User" : "Edit User"}
              </DialogTitle>
              <DialogDescription>
                {dialogs.addUser
                  ? "Create a new user account with the information below."
                  : "Update the user information below."}
              </DialogDescription>
            </DialogHeader>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="first_name">First Name</Label>
                  <Input
                    id="first_name"
                    value={formData.first_name}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        first_name: e.target.value,
                      }))
                    }
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="last_name">Last Name</Label>
                  <Input
                    id="last_name"
                    value={formData.last_name}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        last_name: e.target.value,
                      }))
                    }
                    required
                  />
                </div>
              </div>

              <div>
                <Label htmlFor="username">Username</Label>
                <Input
                  id="username"
                  value={formData.username}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      username: e.target.value,
                    }))
                  }
                  required
                />
              </div>

              <div>
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={formData.email}
                  onChange={(e) =>
                    setFormData((prev) => ({ ...prev, email: e.target.value }))
                  }
                  required
                />
              </div>

              {dialogs.addUser && (
                <div>
                  <Label htmlFor="password">Password</Label>
                  <Input
                    id="password"
                    type="password"
                    value={formData.password || ""}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        password: e.target.value,
                      }))
                    }
                    placeholder="Leave empty to generate random password"
                  />
                </div>
              )}

              <div>
                <Label htmlFor="bio">Bio</Label>
                <Textarea
                  id="bio"
                  value={formData.bio}
                  onChange={(e) =>
                    setFormData((prev) => ({ ...prev, bio: e.target.value }))
                  }
                  rows={3}
                />
              </div>

              <div>
                <Label htmlFor="role">Role</Label>
                <Select
                  value={formData.role}
                  onValueChange={(value: UserRole) =>
                    setFormData((prev) => ({ ...prev, role: value }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={UserRole.USER}>User</SelectItem>
                    <SelectItem value={UserRole.MODERATOR}>
                      Moderator
                    </SelectItem>
                    <SelectItem value={UserRole.ADMIN}>Admin</SelectItem>
                    {currentUser?.role === UserRole.SUPER_ADMIN && (
                      <SelectItem value={UserRole.SUPER_ADMIN}>
                        Super Admin
                      </SelectItem>
                    )}
                  </SelectContent>
                </Select>
              </div>

              <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-2">
                  <Switch
                    id="is_active"
                    checked={formData.is_active}
                    onCheckedChange={(checked) =>
                      setFormData((prev) => ({ ...prev, is_active: checked }))
                    }
                  />
                  <Label htmlFor="is_active">Active</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <Switch
                    id="is_verified"
                    checked={formData.is_verified}
                    onCheckedChange={(checked) =>
                      setFormData((prev) => ({ ...prev, is_verified: checked }))
                    }
                  />
                  <Label htmlFor="is_verified">Verified</Label>
                </div>
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
                  onClick={() => {
                    if (dialogs.addUser) closeDialog("addUser");
                    if (dialogs.editUser) closeDialog("editUser");
                  }}
                >
                  Cancel
                </Button>
                <Button type="submit" disabled={formLoading}>
                  {formLoading && (
                    <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                  )}
                  {dialogs.addUser ? "Create User" : "Update User"}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>

        {/* View User Dialog */}
        <Dialog
          open={dialogs.viewUser}
          onOpenChange={() => closeDialog("viewUser")}
        >
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>User Details</DialogTitle>
              <DialogDescription>
                View detailed information about this user
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedUser && (
              <div className="space-y-6">
                <div className="flex items-center space-x-4">
                  <Avatar className="h-16 w-16">
                    <AvatarImage src={dialogs.selectedUser.profile_picture} />
                    <AvatarFallback>
                      {(
                        dialogs.selectedUser.first_name?.[0] ||
                        dialogs.selectedUser.username?.[0] ||
                        "U"
                      ).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <h3 className="text-lg font-semibold">
                      {dialogs.selectedUser.first_name || ""}{" "}
                      {dialogs.selectedUser.last_name || ""}
                    </h3>
                    <p className="text-muted-foreground">
                      @{dialogs.selectedUser.username || "unknown"}
                    </p>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge
                        variant={getStatusBadgeVariant(dialogs.selectedUser)}
                      >
                        {formatUserStatus(dialogs.selectedUser)}
                      </Badge>
                      {dialogs.selectedUser.is_verified && (
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

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label className="text-sm font-medium">Email</Label>
                    <p className="text-sm text-muted-foreground">
                      {dialogs.selectedUser.email || "No email"}
                    </p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Role</Label>
                    <p className="text-sm text-muted-foreground">
                      {(dialogs.selectedUser.role || "user")
                        .replace("_", " ")
                        .toUpperCase()}
                    </p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Created</Label>
                    <p className="text-sm text-muted-foreground">
                      {dialogs.selectedUser.created_at
                        ? new Date(
                            dialogs.selectedUser.created_at
                          ).toLocaleDateString()
                        : "Unknown"}
                    </p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Last Active</Label>
                    <p className="text-sm text-muted-foreground">
                      {dialogs.selectedUser.last_active_at
                        ? new Date(
                            dialogs.selectedUser.last_active_at
                          ).toLocaleDateString()
                        : "Never"}
                    </p>
                  </div>
                </div>

                {dialogs.selectedUser.bio && (
                  <div>
                    <Label className="text-sm font-medium">Bio</Label>
                    <p className="text-sm text-muted-foreground mt-1">
                      {dialogs.selectedUser.bio}
                    </p>
                  </div>
                )}

                <div className="grid grid-cols-3 gap-4 text-center">
                  <div>
                    <p className="text-2xl font-bold">
                      {dialogs.selectedUser.posts_count || 0}
                    </p>
                    <p className="text-sm text-muted-foreground">Posts</p>
                  </div>
                  <div>
                    <p className="text-2xl font-bold">
                      {dialogs.selectedUser.followers_count || 0}
                    </p>
                    <p className="text-sm text-muted-foreground">Followers</p>
                  </div>
                  <div>
                    <p className="text-2xl font-bold">
                      {dialogs.selectedUser.following_count || 0}
                    </p>
                    <p className="text-sm text-muted-foreground">Following</p>
                  </div>
                </div>
              </div>
            )}

            <DialogFooter>
              <Button variant="outline" onClick={() => closeDialog("viewUser")}>
                Close
              </Button>
              <Button
                onClick={() => {
                  closeDialog("viewUser");
                  if (dialogs.selectedUser)
                    openDialog("editUser", dialogs.selectedUser);
                }}
              >
                <IconEdit className="h-4 w-4 mr-2" />
                Edit User
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Delete User Dialog */}
        <Dialog
          open={dialogs.deleteUser}
          onOpenChange={() => closeDialog("deleteUser")}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete User</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this user? This action cannot be
                undone.
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedUser && (
              <div className="py-4">
                <div className="flex items-center space-x-4 p-4 bg-muted rounded-lg">
                  <Avatar>
                    <AvatarImage src={dialogs.selectedUser.profile_picture} />
                    <AvatarFallback>
                      {(
                        dialogs.selectedUser.first_name?.[0] ||
                        dialogs.selectedUser.username?.[0] ||
                        "U"
                      ).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <p className="font-medium">
                      {dialogs.selectedUser.first_name || ""}{" "}
                      {dialogs.selectedUser.last_name || ""}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      @{dialogs.selectedUser.username || "unknown"}
                    </p>
                  </div>
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
                onClick={() => closeDialog("deleteUser")}
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDeleteUser}
                disabled={formLoading}
              >
                {formLoading && (
                  <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Delete User
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Suspend User Dialog */}
        <Dialog
          open={dialogs.suspendUser}
          onOpenChange={() => closeDialog("suspendUser")}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                {dialogs.selectedUser?.is_suspended
                  ? "Unsuspend User"
                  : "Suspend User"}
              </DialogTitle>
              <DialogDescription>
                {dialogs.selectedUser?.is_suspended
                  ? "Remove suspension from this user account."
                  : "Temporarily restrict this user account."}
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedUser && (
              <div className="space-y-4">
                <div className="flex items-center space-x-4 p-4 bg-muted rounded-lg">
                  <Avatar>
                    <AvatarImage src={dialogs.selectedUser.profile_picture} />
                    <AvatarFallback>
                      {(
                        dialogs.selectedUser.first_name?.[0] ||
                        dialogs.selectedUser.username?.[0] ||
                        "U"
                      ).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <p className="font-medium">
                      {dialogs.selectedUser.first_name || ""}{" "}
                      {dialogs.selectedUser.last_name || ""}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      @{dialogs.selectedUser.username || "unknown"}
                    </p>
                  </div>
                </div>

                {!dialogs.selectedUser.is_suspended && (
                  <div>
                    <Label htmlFor="reason">Reason for suspension</Label>
                    <Textarea
                      id="reason"
                      value={suspensionReason}
                      onChange={(e) => setSuspensionReason(e.target.value)}
                      placeholder="Provide a reason for suspending this user..."
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
                onClick={() => closeDialog("suspendUser")}
              >
                Cancel
              </Button>
              <Button
                variant={
                  dialogs.selectedUser?.is_suspended ? "default" : "destructive"
                }
                onClick={handleSuspendUser}
                disabled={
                  formLoading ||
                  (!dialogs.selectedUser?.is_suspended && !suspensionReason)
                }
              >
                {formLoading && (
                  <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                {dialogs.selectedUser?.is_suspended
                  ? "Unsuspend User"
                  : "Suspend User"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </SidebarInset>
    </SidebarProvider>
  );
}

export default withAuth(UsersPage);
