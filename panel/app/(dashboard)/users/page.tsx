"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { UsersTable } from "@/components/users/users-table";
import { UserDetailDialog } from "@/components/users/user-detail-dialog";
import { UserFormDialog } from "@/components/users/user-form-dialog";
import { UserFilters } from "@/components/users/user-filters";
import { BulkActions } from "@/components/users/bulk-actions";
import { UserStats } from "@/components/users/user-stats";
import {
  Search,
  Plus,
  Filter,
  Download,
  Upload,
  Users,
  UserCheck,
  UserX,
  Shield,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";

export default function UsersPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedUsers, setSelectedUsers] = useState<string[]>([]);
  const [showUserDetail, setShowUserDetail] = useState(false);
  const [showUserForm, setShowUserForm] = useState(false);
  const [showFilters, setShowFilters] = useState(false);
  const [selectedUser, setSelectedUser] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState("all");

  const handleUserSelect = (userId: string) => {
    setSelectedUser(userId);
    setShowUserDetail(true);
  };

  const handleCreateUser = () => {
    setSelectedUser(null);
    setShowUserForm(true);
  };

  const handleEditUser = (userId: string) => {
    setSelectedUser(userId);
    setShowUserForm(true);
  };

  const handleBulkAction = (action: string, userIds: string[]) => {
    console.log(`Bulk action: ${action} for users:`, userIds);
    // Implement bulk actions logic
  };

  const handleExport = () => {
    console.log("Exporting users data...");
    // Implement export functionality
  };

  const handleImport = () => {
    console.log("Importing users data...");
    // Implement import functionality
  };

  // Mock user counts for tabs
  const userCounts = {
    all: 12485,
    active: 11203,
    suspended: 892,
    pending: 390,
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Users</h1>
          <p className="text-muted-foreground">
            Manage and monitor all users on your platform
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleImport}>
            <Upload className="mr-2 h-4 w-4" />
            Import
          </Button>
          <Button variant="outline" size="sm" onClick={handleExport}>
            <Download className="mr-2 h-4 w-4" />
            Export
          </Button>
          <Button onClick={handleCreateUser}>
            <Plus className="mr-2 h-4 w-4" />
            Add User
          </Button>
        </div>
      </div>

      {/* Stats Overview */}
      <UserStats />

      {/* Search and Filters */}
      <Card>
        <CardHeader>
          <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div className="flex flex-1 items-center gap-4">
              <div className="relative flex-1 max-w-md">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder="Search users by name, email, or username..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10"
                />
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowFilters(!showFilters)}
              >
                <Filter className="mr-2 h-4 w-4" />
                Filters
              </Button>
            </div>

            {selectedUsers.length > 0 && (
              <BulkActions
                selectedCount={selectedUsers.length}
                onAction={handleBulkAction}
                selectedUsers={selectedUsers}
              />
            )}
          </div>

          {showFilters && (
            <div className="mt-4">
              <UserFilters />
            </div>
          )}
        </CardHeader>

        <CardContent>
          {/* User Tabs */}
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList>
              <TabsTrigger value="all" className="flex items-center gap-2">
                <Users className="h-4 w-4" />
                All Users
                <Badge variant="secondary">
                  {userCounts.all.toLocaleString()}
                </Badge>
              </TabsTrigger>
              <TabsTrigger value="active" className="flex items-center gap-2">
                <UserCheck className="h-4 w-4" />
                Active
                <Badge variant="secondary">
                  {userCounts.active.toLocaleString()}
                </Badge>
              </TabsTrigger>
              <TabsTrigger
                value="suspended"
                className="flex items-center gap-2"
              >
                <UserX className="h-4 w-4" />
                Suspended
                <Badge variant="secondary">
                  {userCounts.suspended.toLocaleString()}
                </Badge>
              </TabsTrigger>
              <TabsTrigger value="pending" className="flex items-center gap-2">
                <Shield className="h-4 w-4" />
                Pending
                <Badge variant="secondary">
                  {userCounts.pending.toLocaleString()}
                </Badge>
              </TabsTrigger>
            </TabsList>

            <div className="mt-6">
              <TabsContent value="all" className="space-y-4">
                <UsersTable
                  searchQuery={searchQuery}
                  filterStatus="all"
                  selectedUsers={selectedUsers}
                  onUserSelect={handleUserSelect}
                  onUserEdit={handleEditUser}
                  onSelectionChange={setSelectedUsers}
                />
              </TabsContent>

              <TabsContent value="active" className="space-y-4">
                <UsersTable
                  searchQuery={searchQuery}
                  filterStatus="active"
                  selectedUsers={selectedUsers}
                  onUserSelect={handleUserSelect}
                  onUserEdit={handleEditUser}
                  onSelectionChange={setSelectedUsers}
                />
              </TabsContent>

              <TabsContent value="suspended" className="space-y-4">
                <UsersTable
                  searchQuery={searchQuery}
                  filterStatus="suspended"
                  selectedUsers={selectedUsers}
                  onUserSelect={handleUserSelect}
                  onUserEdit={handleEditUser}
                  onSelectionChange={setSelectedUsers}
                />
              </TabsContent>

              <TabsContent value="pending" className="space-y-4">
                <UsersTable
                  searchQuery={searchQuery}
                  filterStatus="pending"
                  selectedUsers={selectedUsers}
                  onUserSelect={handleUserSelect}
                  onUserEdit={handleEditUser}
                  onSelectionChange={setSelectedUsers}
                />
              </TabsContent>
            </div>
          </Tabs>
        </CardContent>
      </Card>

      {/* Dialogs */}
      <UserDetailDialog
        open={showUserDetail}
        onOpenChange={setShowUserDetail}
        userId={selectedUser}
        onEdit={handleEditUser}
      />

      <UserFormDialog
        open={showUserForm}
        onOpenChange={setShowUserForm}
        userId={selectedUser}
        mode={selectedUser ? "edit" : "create"}
      />
    </div>
  );
}
