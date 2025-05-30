// app/admin/users/page.tsx - Simple Users Management
"use client";

import { useEffect, useState } from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { SiteHeader } from "@/components/site-header";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { withAuth, useAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import {
  IconRefresh,
  IconSearch,
  IconUsers,
  IconShield,
  IconShieldCheck,
} from "@tabler/icons-react";

interface User {
  id: string;
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
  role: string;
  is_verified: boolean;
  is_active: boolean;
  created_at: string;
  last_active_at?: string;
}

interface PaginatedUsers {
  data: User[];
  pagination: {
    current_page: number;
    per_page: number;
    total: number;
    total_pages: number;
    has_next: boolean;
    has_previous: boolean;
  };
}

function UsersPage() {
  const { user } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [pagination, setPagination] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [currentPage, setCurrentPage] = useState(1);

  const fetchUsers = async (page = 1, search = "") => {
    try {
      setLoading(true);
      setError(null);

      console.log(`🔄 Fetching users (page ${page})...`);

      const params = {
        page,
        limit: 20,
        ...(search && { search }),
      };

      const response = await apiClient.getUsers(params);

      if (response.success && response.data) {
        const paginatedData = response as unknown as PaginatedUsers;
        setUsers(paginatedData.data);
        setPagination(paginatedData.pagination);
        console.log(`✅ Fetched ${paginatedData.data.length} users`);
      } else {
        throw new Error(response.message || "Failed to fetch users");
      }
    } catch (error: any) {
      console.error("❌ Failed to fetch users:", error);
      const errorMessage =
        error.response?.data?.message ||
        error.message ||
        "Failed to load users";
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers(currentPage, searchQuery);
  }, [currentPage]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentPage(1);
    fetchUsers(1, searchQuery);
  };

  const handleRefresh = () => {
    fetchUsers(currentPage, searchQuery);
  };

  const getRoleBadgeVariant = (role: string) => {
    switch (role) {
      case "super_admin":
        return "destructive";
      case "admin":
        return "default";
      case "moderator":
        return "secondary";
      default:
        return "outline";
    }
  };

  const getRoleIcon = (role: string) => {
    switch (role) {
      case "super_admin":
        return IconShieldCheck;
      case "admin":
      case "moderator":
        return IconShield;
      default:
        return IconUsers;
    }
  };

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
          <div className="flex h-screen items-center justify-center">
            <Alert variant="destructive" className="max-w-md">
              <AlertDescription className="space-y-4">
                <div>{error}</div>
                <Button onClick={handleRefresh} className="w-full">
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
        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
              {/* Header */}
              <div className="px-4 lg:px-6">
                <div className="flex items-center justify-between">
                  <div>
                    <h1 className="text-2xl font-bold">Users Management</h1>
                    <p className="text-muted-foreground">
                      Manage platform users and their permissions
                    </p>
                  </div>
                  <Button
                    onClick={handleRefresh}
                    variant="outline"
                    size="sm"
                    disabled={loading}
                  >
                    <IconRefresh
                      className={`h-4 w-4 mr-2 ${
                        loading ? "animate-spin" : ""
                      }`}
                    />
                    Refresh
                  </Button>
                </div>
              </div>

              {/* Search and Filters */}
              <div className="px-4 lg:px-6">
                <Card>
                  <CardHeader>
                    <CardTitle>Search Users</CardTitle>
                    <CardDescription>
                      Find users by username, email, or name
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <form onSubmit={handleSearch} className="flex gap-2">
                      <div className="flex-1">
                        <Input
                          type="text"
                          placeholder="Search users..."
                          value={searchQuery}
                          onChange={(e) => setSearchQuery(e.target.value)}
                          disabled={loading}
                        />
                      </div>
                      <Button type="submit" disabled={loading}>
                        <IconSearch className="h-4 w-4 mr-2" />
                        Search
                      </Button>
                    </form>
                  </CardContent>
                </Card>
              </div>

              {/* Users Table */}
              <div className="px-4 lg:px-6">
                <Card>
                  <CardHeader>
                    <CardTitle>
                      Users {pagination && `(${pagination.total} total)`}
                    </CardTitle>
                    <CardDescription>
                      {pagination &&
                        `Showing ${
                          (pagination.current_page - 1) * pagination.per_page +
                          1
                        }-${Math.min(
                          pagination.current_page * pagination.per_page,
                          pagination.total
                        )} of ${pagination.total} users`}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    {loading ? (
                      <UsersTableSkeleton />
                    ) : (
                      <>
                        <Table>
                          <TableHeader>
                            <TableRow>
                              <TableHead>User</TableHead>
                              <TableHead>Role</TableHead>
                              <TableHead>Status</TableHead>
                              <TableHead>Joined</TableHead>
                              <TableHead>Last Active</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {users.map((user) => {
                              const RoleIcon = getRoleIcon(user.role);
                              return (
                                <TableRow key={user.id}>
                                  <TableCell>
                                    <div>
                                      <div className="font-medium flex items-center gap-2">
                                        {user.first_name} {user.last_name}
                                        {user.is_verified && (
                                          <IconShieldCheck className="h-4 w-4 text-blue-500" />
                                        )}
                                      </div>
                                      <div className="text-sm text-muted-foreground">
                                        @{user.username} • {user.email}
                                      </div>
                                    </div>
                                  </TableCell>
                                  <TableCell>
                                    <Badge
                                      variant={getRoleBadgeVariant(user.role)}
                                      className="flex items-center gap-1 w-fit"
                                    >
                                      <RoleIcon className="h-3 w-3" />
                                      {user.role.replace("_", " ")}
                                    </Badge>
                                  </TableCell>
                                  <TableCell>
                                    <Badge
                                      variant={
                                        user.is_active ? "default" : "secondary"
                                      }
                                    >
                                      {user.is_active ? "Active" : "Inactive"}
                                    </Badge>
                                  </TableCell>
                                  <TableCell>
                                    {new Date(
                                      user.created_at
                                    ).toLocaleDateString()}
                                  </TableCell>
                                  <TableCell>
                                    {user.last_active_at
                                      ? new Date(
                                          user.last_active_at
                                        ).toLocaleDateString()
                                      : "Never"}
                                  </TableCell>
                                </TableRow>
                              );
                            })}
                          </TableBody>
                        </Table>

                        {/* Pagination */}
                        {pagination && pagination.total_pages > 1 && (
                          <div className="flex items-center justify-between mt-4">
                            <div className="text-sm text-muted-foreground">
                              Page {pagination.current_page} of{" "}
                              {pagination.total_pages}
                            </div>
                            <div className="flex gap-2">
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() =>
                                  setCurrentPage(Math.max(1, currentPage - 1))
                                }
                                disabled={!pagination.has_previous || loading}
                              >
                                Previous
                              </Button>
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => setCurrentPage(currentPage + 1)}
                                disabled={!pagination.has_next || loading}
                              >
                                Next
                              </Button>
                            </div>
                          </div>
                        )}
                      </>
                    )}
                  </CardContent>
                </Card>
              </div>
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

function UsersTableSkeleton() {
  return (
    <div className="space-y-4">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="flex items-center space-x-4">
          <Skeleton className="h-10 w-10 rounded-full" />
          <div className="space-y-2 flex-1">
            <Skeleton className="h-4 w-48" />
            <Skeleton className="h-3 w-32" />
          </div>
          <Skeleton className="h-6 w-16" />
          <Skeleton className="h-6 w-16" />
          <Skeleton className="h-4 w-20" />
          <Skeleton className="h-4 w-20" />
        </div>
      ))}
    </div>
  );
}

export default withAuth(UsersPage);
