// app/admin/groups/page.tsx
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { withAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import { Group, GroupMember, TableColumn } from "@/types/admin";
import {
  IconUsers,
  IconEye,
  IconBan,
  IconCheck,
  IconX,
  IconShield,
  IconGlobe,
  IconLock,
  IconEyeOff,
  IconCalendar,
  IconMapPin,
} from "@tabler/icons-react";

function GroupsPage() {
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<any>(null);
  const [filters, setFilters] = useState<any>({ page: 1, limit: 20 });

  // Dialog states
  const [selectedGroup, setSelectedGroup] = useState<Group | null>(null);
  const [showGroupDetails, setShowGroupDetails] = useState(false);
  const [showMembersDialog, setShowMembersDialog] = useState(false);
  const [showActionDialog, setShowActionDialog] = useState(false);
  const [showBulkDialog, setShowBulkDialog] = useState(false);
  const [actionType, setActionType] = useState<string>("");
  const [bulkAction, setBulkAction] = useState<string>("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [actionReason, setActionReason] = useState("");

  // Members data
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [membersLoading, setMembersLoading] = useState(false);
  const [membersPagination, setMembersPagination] = useState<any>(null);

  // Fetch groups
  const fetchGroups = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getGroups(filters);
      setGroups(response.data);
      setPagination(response.pagination);
    } catch (error: any) {
      console.error("Failed to fetch groups:", error);
      setError(error.response?.data?.message || "Failed to load groups");
    } finally {
      setLoading(false);
    }
  };

  // Fetch group members
  const fetchGroupMembers = async (groupId: string) => {
    try {
      setMembersLoading(true);
      const response = await apiClient.getGroupMembers(groupId);
      setMembers(response.data);
      setMembersPagination(response.pagination);
    } catch (error: any) {
      console.error("Failed to fetch group members:", error);
    } finally {
      setMembersLoading(false);
    }
  };

  useEffect(() => {
    fetchGroups();
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
      await apiClient.bulkGroupAction({
        group_ids: selectedIds,
        action: bulkAction,
        reason: actionReason,
      });

      setShowBulkDialog(false);
      setActionReason("");
      fetchGroups();
    } catch (error: any) {
      console.error("Bulk action failed:", error);
      setError(error.response?.data?.message || "Bulk action failed");
    }
  };

  // Handle individual group action
  const handleGroupAction = (group: Group, action: string) => {
    setSelectedGroup(group);
    setActionType(action);
    setShowActionDialog(true);
  };

  // Execute group action
  const executeGroupAction = async () => {
    if (!selectedGroup) return;

    try {
      switch (actionType) {
        case "suspend":
          await apiClient.updateGroupStatus(selectedGroup.id, {
            is_active: false,
            reason: actionReason,
          });
          break;
        case "activate":
          await apiClient.updateGroupStatus(selectedGroup.id, {
            is_active: true,
            reason: actionReason,
          });
          break;
        case "delete":
          await apiClient.deleteGroup(selectedGroup.id, actionReason);
          break;
      }

      setShowActionDialog(false);
      setActionReason("");
      fetchGroups();
    } catch (error: any) {
      console.error("Group action failed:", error);
      setError(error.response?.data?.message || "Group action failed");
    }
  };

  // View group details
  const viewGroupDetails = async (group: Group) => {
    try {
      const response = await apiClient.getGroup(group.id);
      setSelectedGroup(response.data);
      setShowGroupDetails(true);
    } catch (error) {
      console.error("Failed to fetch group details:", error);
    }
  };

  // View group members
  const viewGroupMembers = (group: Group) => {
    setSelectedGroup(group);
    fetchGroupMembers(group.id);
    setShowMembersDialog(true);
  };

  // Get group type icon
  const getGroupTypeIcon = (type: string) => {
    switch (type) {
      case "public":
        return <IconGlobe className="h-4 w-4" />;
      case "private":
        return <IconLock className="h-4 w-4" />;
      case "secret":
        return <IconEyeOff className="h-4 w-4" />;
      default:
        return <IconUsers className="h-4 w-4" />;
    }
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "group",
      label: "Group",
      render: (_, group: Group) => (
        <div className="flex items-center gap-3">
          <Avatar className="h-10 w-10">
            <AvatarImage src={group.avatar} alt={group.name} />
            <AvatarFallback>{group.name?.[0]?.toUpperCase()}</AvatarFallback>
          </Avatar>
          <div>
            <div className="font-medium">{group.name}</div>
            <div className="text-sm text-muted-foreground line-clamp-1">
              {group.description}
            </div>
          </div>
        </div>
      ),
    },
    {
      key: "type",
      label: "Type",
      filterable: true,
      render: (value: string) => (
        <Badge variant="outline" className="flex items-center gap-1 w-fit">
          {getGroupTypeIcon(value)}
          <span className="capitalize">{value}</span>
        </Badge>
      ),
    },
    {
      key: "category",
      label: "Category",
      filterable: true,
      render: (value: string) => <Badge variant="secondary">{value}</Badge>,
    },
    {
      key: "members_count",
      label: "Members",
      sortable: true,
      render: (value: number, group: Group) => (
        <div
          className="flex items-center gap-1 cursor-pointer hover:text-primary"
          onClick={() => viewGroupMembers(group)}
        >
          <IconUsers className="h-4 w-4" />
          <span>{value?.toLocaleString() || 0}</span>
        </div>
      ),
    },
    {
      key: "posts_count",
      label: "Posts",
      sortable: true,
      render: (value: number) => value?.toLocaleString() || 0,
    },
    {
      key: "is_verified",
      label: "Verified",
      filterable: true,
      render: (value: boolean) => (
        <Badge variant={value ? "default" : "secondary"}>
          {value ? (
            <>
              <IconCheck className="h-3 w-3 mr-1" />
              Verified
            </>
          ) : (
            <>
              <IconX className="h-3 w-3 mr-1" />
              Unverified
            </>
          )}
        </Badge>
      ),
    },
    {
      key: "status",
      label: "Status",
      filterable: true,
      render: (_, group: Group) => (
        <Badge variant={group.is_active ? "default" : "destructive"}>
          {group.is_active ? "Active" : "Suspended"}
        </Badge>
      ),
    },
    {
      key: "created_at",
      label: "Created",
      sortable: true,
      render: (value: string) => (
        <div className="text-sm">{new Date(value).toLocaleDateString()}</div>
      ),
    },
    {
      key: "actions",
      label: "Actions",
      render: (_, group: Group) => (
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewGroupDetails(group)}
          >
            <IconEye className="h-3 w-3" />
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewGroupMembers(group)}
          >
            <IconUsers className="h-3 w-3" />
          </Button>
          {group.is_active ? (
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleGroupAction(group, "suspend")}
            >
              <IconBan className="h-3 w-3" />
            </Button>
          ) : (
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleGroupAction(group, "activate")}
            >
              <IconCheck className="h-3 w-3" />
            </Button>
          )}
        </div>
      ),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    {
      label: "Activate Groups",
      action: "activate",
      variant: "default" as const,
    },
    {
      label: "Suspend Groups",
      action: "suspend",
      variant: "destructive" as const,
    },
    {
      label: "Delete Groups",
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
              <h1 className="text-2xl font-bold">Groups Management</h1>
              <p className="text-muted-foreground">
                Manage community groups and their members
              </p>
            </div>
          </div>

          <DataTable
            title="Groups"
            description={`Manage ${pagination?.total || 0} groups`}
            data={groups}
            columns={columns}
            loading={loading}
            pagination={pagination}
            searchPlaceholder="Search groups by name or description..."
            onPageChange={handlePageChange}
            onFilter={handleFilter}
            onRowSelect={handleRowSelect}
            bulkActions={bulkActions}
            onBulkAction={handleBulkAction}
            onRefresh={fetchGroups}
            onExport={() => console.log("Export groups")}
          />
        </div>
      </SidebarInset>

      {/* Group Details Dialog */}
      <Dialog open={showGroupDetails} onOpenChange={setShowGroupDetails}>
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Group Details</DialogTitle>
          </DialogHeader>

          {selectedGroup && (
            <Tabs defaultValue="details" className="w-full">
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="details">Details</TabsTrigger>
                <TabsTrigger value="settings">Settings</TabsTrigger>
                <TabsTrigger value="stats">Statistics</TabsTrigger>
              </TabsList>

              <TabsContent value="details" className="space-y-4">
                <Card>
                  <CardHeader>
                    <div className="flex items-center gap-4">
                      <Avatar className="h-16 w-16">
                        <AvatarImage src={selectedGroup.avatar} />
                        <AvatarFallback className="text-2xl">
                          {selectedGroup.name?.[0]?.toUpperCase()}
                        </AvatarFallback>
                      </Avatar>
                      <div className="flex-1">
                        <CardTitle className="flex items-center gap-2">
                          {selectedGroup.name}
                          {selectedGroup.is_verified && (
                            <IconCheck className="h-5 w-5 text-primary" />
                          )}
                        </CardTitle>
                        <CardDescription>
                          {selectedGroup.description}
                        </CardDescription>
                        <div className="flex items-center gap-4 mt-2 text-sm">
                          <div className="flex items-center gap-1">
                            {getGroupTypeIcon(selectedGroup.type)}
                            <span className="capitalize">
                              {selectedGroup.type}
                            </span>
                          </div>
                          <div className="flex items-center gap-1">
                            <IconUsers className="h-4 w-4" />
                            {selectedGroup.members_count} members
                          </div>
                          {selectedGroup.location && (
                            <div className="flex items-center gap-1">
                              <IconMapPin className="h-4 w-4" />
                              {selectedGroup.location}
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div>
                      <h4 className="font-medium mb-2">Category</h4>
                      <Badge>{selectedGroup.category}</Badge>
                    </div>

                    {selectedGroup.tags && selectedGroup.tags.length > 0 && (
                      <div>
                        <h4 className="font-medium mb-2">Tags</h4>
                        <div className="flex flex-wrap gap-2">
                          {selectedGroup.tags.map((tag, index) => (
                            <Badge key={index} variant="outline">
                              #{tag}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}

                    {selectedGroup.website && (
                      <div>
                        <h4 className="font-medium mb-2">Website</h4>
                        <a
                          href={selectedGroup.website}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                        >
                          {selectedGroup.website}
                        </a>
                      </div>
                    )}

                    {selectedGroup.rules && selectedGroup.rules.length > 0 && (
                      <div>
                        <h4 className="font-medium mb-2">Rules</h4>
                        <ol className="list-decimal list-inside space-y-1">
                          {selectedGroup.rules.map((rule, index) => (
                            <li key={index} className="text-sm">
                              {rule}
                            </li>
                          ))}
                        </ol>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="settings" className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle>Group Settings</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="text-sm font-medium">
                          Require Approval
                        </label>
                        <p className="text-sm text-muted-foreground">
                          {selectedGroup.settings?.require_approval
                            ? "Yes"
                            : "No"}
                        </p>
                      </div>
                      <div>
                        <label className="text-sm font-medium">
                          Member Posts
                        </label>
                        <p className="text-sm text-muted-foreground">
                          {selectedGroup.settings?.allow_member_posts
                            ? "Allowed"
                            : "Restricted"}
                        </p>
                      </div>
                      <div>
                        <label className="text-sm font-medium">
                          Member Invites
                        </label>
                        <p className="text-sm text-muted-foreground">
                          {selectedGroup.settings?.allow_member_invites
                            ? "Allowed"
                            : "Restricted"}
                        </p>
                      </div>
                      <div>
                        <label className="text-sm font-medium">
                          Posting Permissions
                        </label>
                        <p className="text-sm text-muted-foreground capitalize">
                          {selectedGroup.settings?.posting_permissions?.replace(
                            "_",
                            " "
                          )}
                        </p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="stats" className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">Members</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {selectedGroup.members_count?.toLocaleString()}
                      </div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">Posts</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {selectedGroup.posts_count?.toLocaleString()}
                      </div>
                    </CardContent>
                  </Card>
                </div>

                <div className="text-sm text-muted-foreground space-y-1">
                  <p>
                    Created:{" "}
                    {new Date(selectedGroup.created_at).toLocaleDateString()}
                  </p>
                  <p>
                    Last Updated:{" "}
                    {new Date(selectedGroup.updated_at).toLocaleDateString()}
                  </p>
                  <p>
                    Status: {selectedGroup.is_active ? "Active" : "Suspended"}
                  </p>
                  {selectedGroup.is_verified && <p>✓ Verified Group</p>}
                </div>
              </TabsContent>
            </Tabs>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowGroupDetails(false)}
            >
              Close
            </Button>
            {selectedGroup && (
              <>
                <Button
                  variant="outline"
                  onClick={() => viewGroupMembers(selectedGroup)}
                >
                  View Members
                </Button>
                {selectedGroup.is_active ? (
                  <Button
                    variant="outline"
                    onClick={() => {
                      setShowGroupDetails(false);
                      handleGroupAction(selectedGroup, "suspend");
                    }}
                  >
                    Suspend Group
                  </Button>
                ) : (
                  <Button
                    onClick={() => {
                      setShowGroupDetails(false);
                      handleGroupAction(selectedGroup, "activate");
                    }}
                  >
                    Activate Group
                  </Button>
                )}
              </>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Group Members Dialog */}
      <Dialog open={showMembersDialog} onOpenChange={setShowMembersDialog}>
        <DialogContent className="max-w-4xl max-h-[80vh]">
          <DialogHeader>
            <DialogTitle>
              {selectedGroup?.name} - Members ({selectedGroup?.members_count})
            </DialogTitle>
          </DialogHeader>

          <div className="space-y-4 max-h-96 overflow-y-auto">
            {membersLoading ? (
              <div className="text-center py-8">Loading members...</div>
            ) : members.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No members found
              </div>
            ) : (
              <div className="space-y-2">
                {members.map((member) => (
                  <div
                    key={member.id}
                    className="flex items-center justify-between p-3 border rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      {member.user && (
                        <>
                          <Avatar>
                            <AvatarImage src={member.user.profile_picture} />
                            <AvatarFallback>
                              {member.user.username?.[0]?.toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div>
                            <p className="font-medium">
                              {member.user.username}
                            </p>
                            <p className="text-sm text-muted-foreground">
                              Joined{" "}
                              {new Date(member.joined_at).toLocaleDateString()}
                            </p>
                          </div>
                        </>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge
                        variant={
                          member.role === "admin"
                            ? "default"
                            : member.role === "moderator"
                            ? "secondary"
                            : "outline"
                        }
                      >
                        {member.role === "admin" && (
                          <IconShield className="h-3 w-3 mr-1" />
                        )}
                        <span className="capitalize">{member.role}</span>
                      </Badge>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowMembersDialog(false)}
            >
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Action Dialog */}
      <Dialog open={showActionDialog} onOpenChange={setShowActionDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {actionType === "suspend"
                ? "Suspend Group"
                : actionType === "activate"
                ? "Activate Group"
                : "Delete Group"}
            </DialogTitle>
            <DialogDescription>
              {actionType === "delete"
                ? "This action cannot be undone. The group and all its content will be permanently deleted."
                : `This will ${actionType} the group: ${selectedGroup?.name}`}
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
              onClick={executeGroupAction}
              variant={
                actionType === "delete" || actionType === "suspend"
                  ? "destructive"
                  : "default"
              }
              disabled={!actionReason.trim()}
            >
              {actionType === "suspend" && "Suspend Group"}
              {actionType === "activate" && "Activate Group"}
              {actionType === "delete" && "Delete Group"}
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
              You are about to {bulkAction} {selectedIds.length} group(s).
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
              variant={
                bulkAction === "delete" || bulkAction === "suspend"
                  ? "destructive"
                  : "default"
              }
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

export default withAuth(GroupsPage);
