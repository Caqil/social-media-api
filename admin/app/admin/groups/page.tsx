// admin/app/admin/groups/page.tsx
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
  CardFooter,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { withAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import {
  Group,
  GroupType,
  TableColumn,
  PaginationMeta,
  UserResponse,
} from "@/types/admin";
import {
  IconEye,
  IconUserCheck,
  IconTrash,
  IconDotsVertical,
  IconLoader2,
  IconAlertCircle,
  IconUsers,
  IconUsersGroup,
  IconSearch,
  IconRefresh,
  IconDownload,
  IconCalendar,
  IconGlobe,
  IconLock,
  IconEyeOff,
  IconCheck,
  IconX,
  IconUserPlus,
  IconCertificate,
  IconSettings,
  IconCategory,
  IconTags,
} from "@tabler/icons-react";

// Define the state interface for the Groups page
interface GroupsPageState {
  groups: Group[];
  loading: boolean;
  error: string | null;
  pagination: PaginationMeta | undefined;
  filters: {
    search: string;
    type: string;
    category: string;
    status: string;
    is_verified: string;
    page: number;
    limit: number;
    sort_by?: string;
    sort_order?: "asc" | "desc";
  };
  selectedGroups: string[];
}

// Define the dialog state interface
interface DialogState {
  viewGroup: boolean;
  viewMembers: boolean;
  deleteGroup: boolean;
  updateStatus: boolean;
  selectedGroup: Group | null;
}

// Initial filters
const initialFilters = {
  search: "",
  type: "all",
  category: "all",
  status: "all",
  is_verified: "all",
  page: 1,
  limit: 20,
};

function GroupsPage() {
  // State management
  const [state, setState] = useState<GroupsPageState>({
    groups: [],
    loading: true,
    error: null,
    pagination: undefined,
    filters: initialFilters,
    selectedGroups: [],
  });

  // Dialog state
  const [dialogs, setDialogs] = useState<DialogState>({
    viewGroup: false,
    viewMembers: false,
    deleteGroup: false,
    updateStatus: false,
    selectedGroup: null,
  });

  // Members state for the selected group
  const [members, setMembers] = useState<any[]>([]);
  const [membersLoading, setMembersLoading] = useState(false);
  const [membersPagination, setMembersPagination] = useState<
    PaginationMeta | undefined
  >(undefined);

  // Form state
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [actionReason, setActionReason] = useState("");

  // Process groups to ensure they have creator data
  const processGroupsWithCreatorData = async (
    groups: Group[]
  ): Promise<Group[]> => {
    if (!groups || groups.length === 0) return [];

    // Log the structure of the first group to help debugging
    if (groups.length > 0) {
      console.log("🔍 Sample group structure:", {
        id: groups[0].id,
        name: groups[0].name,
        created_by: groups[0].created_by,
        has_creator: !!groups[0].creator,
        creator_fields: groups[0].creator ? Object.keys(groups[0].creator) : [],
        other_potential_creator_fields: {
          has_owner: !!(groups[0] as any).owner,
          has_admin: !!(groups[0] as any).admin,
          has_created_by_user: !!(groups[0] as any).created_by_user,
        },
      });
    }

    // First, try to normalize any existing creator data that might be under different field names
    const normalizedGroups = groups.map((group) => {
      // If there's a created_by ID but no creator object
      if (group.created_by && !group.creator) {
        // Check for creator data under different field names
        const creatorData =
          (group as any).owner ||
          (group as any).admin ||
          (group as any).created_by_user ||
          (group as any).creator_user;

        if (creatorData) {
          group.creator = creatorData;
          console.log(
            `🔄 Found creator data under different field for group ${group.id}`
          );
        }
      }
      return group;
    });

    // Find groups without creator data
    const groupsWithoutCreatorData = normalizedGroups.filter(
      (group) => group.created_by && (!group.creator || !group.creator.username)
    );

    // If all groups have creator data, return as is
    if (groupsWithoutCreatorData.length === 0) {
      return normalizedGroups;
    }

    console.log(
      `⚠️ Found ${groupsWithoutCreatorData.length} groups without complete creator data`
    );

    // Get unique creator IDs
    const creatorIds = [
      ...new Set(groupsWithoutCreatorData.map((group) => group.created_by)),
    ];

    // Fetch creator data for these groups
    try {
      // Log the creator IDs we're fetching
      console.log(`🔄 Fetching creator data for IDs: ${creatorIds.join(", ")}`);

      const response = await apiClient.getUsersByIds(creatorIds);

      if (response.data && Array.isArray(response.data)) {
        // Log the users we retrieved
        console.log(
          `✅ Retrieved ${response.data.length} user records for creators`
        );

        // Create a map of user ID to user data
        const usersMap = new Map();
        response.data.forEach((user) => {
          usersMap.set(user.id, user);
        });

        // Update groups with fetched creator data
        return normalizedGroups.map((group) => {
          if (group.created_by && (!group.creator || !group.creator.username)) {
            const userData = usersMap.get(group.created_by);
            if (userData) {
              group.creator = userData;
              console.log(`✅ Added creator data to group ${group.id}`);
            } else {
              console.log(
                `⚠️ Couldn't find user data for creator ID ${group.created_by}`
              );
            }
          }
          return group;
        });
      } else {
        console.log(`⚠️ No user data returned for creator IDs`);
      }
    } catch (error) {
      console.error("❌ Failed to fetch creator data:", error);
    }

    return normalizedGroups;
  };

  // Fetch groups
  // Updated fetchGroups function with proper filter handling

  const fetchGroups = useCallback(async (filters = state.filters) => {
    try {
      setState((prev) => ({ ...prev, loading: true, error: null }));

      // Debug the filters being applied
      console.log("🔍 Applying filters:", filters);

      const params: any = {
        page: filters.page,
        limit: filters.limit,
        include_creator: true,
        expand: "creator,stats",
      };

      // Add search filter
      if (filters.search) params.search = filters.search;

      // Add type filter - ensure it's in the format the API expects
      if (filters.type && filters.type !== "all") {
        // The API might expect 'type' in a specific format
        // It could be lowercase, uppercase, or an enum value
        params.type = filters.type.toLowerCase();

        // Debug the type parameter
        console.log(`🔍 Setting type parameter to: ${params.type}`);
      }

      // Add category filter
      if (filters.category && filters.category !== "all") {
        params.category = filters.category;

        // Debug the category parameter
        console.log(`🔍 Setting category parameter to: ${params.category}`);
      }

      // Handle status filter - convert to is_active parameter
      if (filters.status && filters.status !== "all") {
        if (filters.status === "active") {
          params.is_active = true;
        } else if (filters.status === "inactive") {
          params.is_active = false;
        }

        // Debug the status parameter
        console.log(`🔍 Setting is_active parameter to: ${params.is_active}`);
      }

      // Handle verification filter - ensure boolean conversion
      if (filters.is_verified && filters.is_verified !== "all") {
        // Convert string 'true'/'false' to boolean
        params.is_verified = filters.is_verified === "true";

        // Debug the verification parameter
        console.log(
          `🔍 Setting is_verified parameter to: ${params.is_verified}`
        );
      }

      // Add sorting
      if (filters.sort_by) {
        params.sort_by = filters.sort_by;
        params.sort_order = filters.sort_order || "desc";
      }

      console.log("📡 Fetching groups with params:", params);
      const response = await apiClient.getGroups(params);

      // Check response structure for debugging
      if (response.data && Array.isArray(response.data)) {
        console.log(`✅ Received ${response.data.length} groups from API`);

        // Check if filters were applied correctly
        if (filters.type !== "all" && response.data.length > 0) {
          const typeCounts = countByProperty(response.data, "type");
          console.log("📊 Group types in response:", typeCounts);
        }

        if (filters.category !== "all" && response.data.length > 0) {
          const categoryCounts = countByProperty(response.data, "category");
          console.log("📊 Group categories in response:", categoryCounts);
        }
      } else {
        console.warn("⚠️ Response data is not an array or is empty");
      }

      // Process groups to ensure all have creator data
      let groupsWithCreatorData = response.data || [];
      groupsWithCreatorData = await processGroupsWithCreatorData(
        groupsWithCreatorData
      );

      setState((prev) => ({
        ...prev,
        groups: groupsWithCreatorData,
        pagination: response.pagination || undefined,
        loading: false,
      }));
    } catch (error: any) {
      console.error("❌ Failed to fetch groups:", error);
      setState((prev) => ({
        ...prev,
        loading: false,
        error: error.message || "Failed to fetch groups",
      }));
    }
  }, []);

  // Helper function to count occurrences of property values
  const countByProperty = (array: any[], property: string) => {
    return array.reduce((acc, item) => {
      const value = item[property];
      acc[value] = (acc[value] || 0) + 1;
      return acc;
    }, {});
  };

  // Fetch group members
  const fetchGroupMembers = async (groupId: string, page: number = 1) => {
    if (!groupId) return;

    try {
      setMembersLoading(true);

      const response = await apiClient.getGroupMembers(groupId, {
        page,
        limit: 10,
        include_user: true,
      });

      setMembers(response.data || []);
      setMembersPagination(response.pagination);
    } catch (error: any) {
      console.error("❌ Failed to fetch group members:", error);
      setMembers([]);
    } finally {
      setMembersLoading(false);
    }
  };

  // Initial load
  useEffect(() => {
    fetchGroups();
  }, [fetchGroups]);

  // Handle search with debounce
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (state.filters.search !== undefined) {
        fetchGroups({ ...state.filters, page: 1 });
      }
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [state.filters.search, fetchGroups]);

  // Handle filter changes
  const handleFilterChange = (key: string, value: any) => {
    const newFilters = { ...state.filters, [key]: value, page: 1 };
    setState((prev) => ({ ...prev, filters: newFilters }));

    if (key !== "search") {
      fetchGroups(newFilters);
    }
  };

  // Handle pagination
  const handlePageChange = (page: number) => {
    const newFilters = { ...state.filters, page };
    setState((prev) => ({ ...prev, filters: newFilters }));
    fetchGroups(newFilters);
  };

  // Handle members pagination
  const handleMembersPageChange = (page: number) => {
    if (dialogs.selectedGroup) {
      fetchGroupMembers(dialogs.selectedGroup.id, page);
    }
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
    fetchGroups(newFilters);
  };

  // Handle row selection
  const handleRowSelect = (selectedRows: string[]) => {
    setState((prev) => ({ ...prev, selectedGroups: selectedRows }));
  };

  // Open dialog
  const openDialog = (type: keyof DialogState, group: Group | null = null) => {
    setDialogs((prev) => ({ ...prev, [type]: true, selectedGroup: group }));

    // If opening members dialog, fetch members
    if (type === "viewMembers" && group) {
      fetchGroupMembers(group.id);
    }
  };

  // Close dialog
  const closeDialog = (type: keyof DialogState) => {
    setDialogs((prev) => ({ ...prev, [type]: false }));

    // Only clear selectedGroup when closing all dialogs
    if (
      Object.keys(dialogs).filter(
        (key) => key !== type && dialogs[key as keyof DialogState]
      ).length === 0
    ) {
      setDialogs((prev) => ({ ...prev, selectedGroup: null }));
    }

    setFormError(null);
    setActionReason("");

    // Clear members data when closing the members dialog
    if (type === "viewMembers") {
      setMembers([]);
      setMembersPagination(undefined);
    }
  };

  // Handle group deletion
  const handleDeleteGroup = async () => {
    if (!dialogs.selectedGroup) return;

    setFormLoading(true);
    try {
      await apiClient.deleteGroup(
        dialogs.selectedGroup.id,
        actionReason || "Deleted by admin"
      );
      closeDialog("deleteGroup");
      fetchGroups();
    } catch (error: any) {
      setFormError(error.message || "Failed to delete group");
    } finally {
      setFormLoading(false);
    }
  };

  // Handle group status update
  const handleUpdateGroupStatus = async () => {
    if (!dialogs.selectedGroup) return;

    setFormLoading(true);
    try {
      await apiClient.updateGroupStatus(dialogs.selectedGroup.id, {
        is_active: !dialogs.selectedGroup.is_active,
        reason:
          actionReason ||
          `${
            dialogs.selectedGroup.is_active ? "Deactivated" : "Activated"
          } by admin`,
      });
      closeDialog("updateStatus");
      fetchGroups();
    } catch (error: any) {
      setFormError(error.message || "Failed to update group status");
    } finally {
      setFormLoading(false);
    }
  };

  // Handle bulk actions
  const handleBulkAction = async (action: string, selectedIds: string[]) => {
    setFormLoading(true);
    try {
      await apiClient.bulkGroupAction({
        group_ids: selectedIds,
        action,
        reason: "Bulk action by admin",
      });
      setState((prev) => ({ ...prev, selectedGroups: [] }));
      fetchGroups();
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
    fetchGroups();
  };

  // Format group type with icon
  const getGroupTypeIcon = (type: GroupType) => {
    switch (type) {
      case GroupType.PUBLIC:
        return <IconGlobe className="h-3 w-3 mr-1" />;
      case GroupType.PRIVATE:
        return <IconUsers className="h-3 w-3 mr-1" />;
      case GroupType.SECRET:
        return <IconLock className="h-3 w-3 mr-1" />;
      default:
        return <IconGlobe className="h-3 w-3 mr-1" />;
    }
  };

  // Format group status
  const formatGroupStatus = (group: Group) => {
    if (!group.is_active) return "Inactive";
    return "Active";
  };

  // Get status badge variant
  const getStatusBadgeVariant = (
    group: Group
  ): "default" | "secondary" | "destructive" => {
    if (!group.is_active) return "destructive";
    return "default";
  };

  // Table columns configuration
  const columns: TableColumn[] = [
    {
      key: "avatar",
      label: "",
      width: "w-12",
      render: (value: string, group: Group) => (
        <Avatar className="h-8 w-8">
          <AvatarImage src={group.avatar || value} alt={group.name} />
          <AvatarFallback>{group.name?.charAt(0) || "G"}</AvatarFallback>
        </Avatar>
      ),
    },
    {
      key: "name",
      label: "Name",
      sortable: true,
      render: (value: string, group: Group) => (
        <div className="flex flex-col">
          <div className="flex items-center gap-2">
            <span className="font-medium">{value}</span>
            {group.is_verified && (
              <Badge
                variant="outline"
                className="text-blue-600 border-blue-200"
              >
                <IconCheck className="h-3 w-3 mr-1" />
                Verified
              </Badge>
            )}
          </div>
          <span className="text-sm text-muted-foreground">
            {group.description?.substring(0, 50)}
            {group.description && group.description.length > 50 ? "..." : ""}
          </span>
        </div>
      ),
    },
    {
      key: "creator",
      label: "Creator",
      render: (_, group: Group) => {
        // First, try to get creator from the group object
        const creator = group.creator;

        // If creator is available with username, display normally
        if (creator && creator.username) {
          return (
            <div className="flex items-center gap-2">
              <Avatar className="h-6 w-6">
                <AvatarImage
                  src={creator.profile_picture}
                  alt={creator.username}
                />
                <AvatarFallback>
                  {(
                    creator.first_name?.[0] ||
                    creator.username?.[0] ||
                    "U"
                  ).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0">
                <div className="text-sm font-medium truncate">
                  {creator.first_name && creator.last_name
                    ? `${creator.first_name} ${creator.last_name}`
                    : creator.first_name || creator.username}
                </div>
                <div className="text-xs text-muted-foreground">
                  @{creator.username}
                </div>
              </div>
            </div>
          );
        }

        // If we have creator with partial data, display what we have
        if (creator) {
          const displayName =
            creator.first_name || creator.email || creator.id || "Unknown User";

          return (
            <div className="flex items-center gap-2">
              <Avatar className="h-6 w-6">
                <AvatarImage src={creator.profile_picture} />
                <AvatarFallback>
                  {(creator.first_name?.[0] || "U").toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0">
                <div className="text-sm font-medium truncate">
                  {displayName}
                </div>
                <div className="text-xs text-muted-foreground">
                  <Button
                    variant="link"
                    className="h-auto p-0 text-xs text-blue-500"
                    onClick={() => fetchGroups()}
                  >
                    Refresh data
                  </Button>
                </div>
              </div>
            </div>
          );
        }

        // If we at least have the creator ID, show a minimal display
        if (group.created_by) {
          return (
            <div className="flex items-center gap-2">
              <Avatar className="h-6 w-6">
                <AvatarFallback>U</AvatarFallback>
              </Avatar>
              <div className="min-w-0">
                <div className="text-sm font-medium truncate">
                  Creator {group.created_by.slice(-5)}
                </div>
                <div className="text-xs text-muted-foreground">
                  <Button
                    variant="link"
                    className="h-auto p-0 text-xs text-blue-500"
                    onClick={() => fetchGroups()}
                  >
                    Load creator
                  </Button>
                </div>
              </div>
            </div>
          );
        }

        // Complete fallback if no creator data available at all
        return (
          <div className="flex items-center gap-2">
            <Avatar className="h-6 w-6">
              <AvatarFallback>U</AvatarFallback>
            </Avatar>
            <div className="text-sm text-muted-foreground">
              <span
                onClick={() => fetchGroups()}
                className="cursor-pointer hover:underline"
              >
                Unknown Creator
              </span>
            </div>
          </div>
        );
      },
    },
    {
      key: "type",
      label: "Type",
      sortable: true,
      filterable: true,
      render: (value: GroupType) => (
        <Badge variant="outline" className="capitalize">
          {getGroupTypeIcon(value)}
          {value?.toLowerCase() || "public"}
        </Badge>
      ),
    },
    {
      key: "category",
      label: "Category",
      sortable: true,
      filterable: true,
      render: (value: string) => (
        <Badge variant="secondary" className="capitalize">
          <IconCategory className="h-3 w-3 mr-1" />
          {value || "Uncategorized"}
        </Badge>
      ),
    },
    {
      key: "members_count",
      label: "Members",
      sortable: true,
      render: (value: number) => (
        <div className="flex items-center gap-2">
          <IconUsers className="h-4 w-4 text-muted-foreground" />
          <span className="font-medium">{value?.toLocaleString() || "0"}</span>
        </div>
      ),
    },
    {
      key: "status",
      label: "Status",
      sortable: true,
      filterable: true,
      render: (_, group: Group) => (
        <Badge variant={getStatusBadgeVariant(group)}>
          {formatGroupStatus(group)}
        </Badge>
      ),
    },
    {
      key: "created_at",
      label: "Created",
      sortable: true,
      render: (value: string) => (
        <div className="flex items-center gap-2 text-sm">
          <IconCalendar className="h-4 w-4 text-muted-foreground" />
          <span>
            {value ? new Date(value).toLocaleDateString() : "Unknown"}
          </span>
        </div>
      ),
    },
    {
      key: "actions",
      label: "",
      width: "w-12",
      render: (_, group: Group) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
              <IconDotsVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => openDialog("viewGroup", group)}>
              <IconEye className="h-4 w-4 mr-2" />
              View Details
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => openDialog("viewMembers", group)}>
              <IconUserCheck className="h-4 w-4 mr-2" />
              View Members
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => openDialog("updateStatus", group)}
              className={
                !group.is_active ? "text-green-600" : "text-orange-600"
              }
            >
              {!group.is_active ? (
                <>
                  <IconCheck className="h-4 w-4 mr-2" />
                  Activate
                </>
              ) : (
                <>
                  <IconX className="h-4 w-4 mr-2" />
                  Deactivate
                </>
              )}
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => openDialog("deleteGroup", group)}
              className="text-red-600"
            >
              <IconTrash className="h-4 w-4 mr-2" />
              Delete Group
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    { label: "Activate Groups", action: "activate" },
    { label: "Deactivate Groups", action: "deactivate" },
    { label: "Verify Groups", action: "verify" },
    {
      label: "Delete Groups",
      action: "delete",
      variant: "destructive" as const,
    },
  ];

  // Error state
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
                <div>Failed to load groups: {state.error}</div>
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
              <h1 className="text-3xl font-bold">Groups</h1>
              <p className="text-muted-foreground">
                Manage groups and their members
              </p>
            </div>
          </div>

          {/* Stats Cards */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Total Groups
                </CardTitle>
                <IconUsersGroup className="h-4 w-4 text-muted-foreground" />
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
                  Active Groups
                </CardTitle>
                <IconCheck className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {state.groups.filter((g) => g.is_active).length}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Verified Groups
                </CardTitle>
                <IconCertificate className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {state.groups.filter((g) => g.is_verified).length}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Public Groups
                </CardTitle>
                <IconGlobe className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {
                    state.groups.filter((g) => g.type === GroupType.PUBLIC)
                      .length
                  }
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
              <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
                <div>
                  <Label htmlFor="search">Search</Label>
                  <div className="relative">
                    <IconSearch className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="search"
                      placeholder="Search groups..."
                      value={state.filters.search}
                      onChange={(e) =>
                        handleFilterChange("search", e.target.value)
                      }
                      className="pl-9"
                    />
                  </div>
                </div>
                <div>
                  <Label htmlFor="type">Type</Label>
                  <Select
                    value={state.filters.type}
                    onValueChange={(value) => handleFilterChange("type", value)}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All types" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All types</SelectItem>
                      <SelectItem value="public">Public</SelectItem>
                      <SelectItem value="private">Private</SelectItem>
                      <SelectItem value="secret">Secret</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="category">Category</Label>
                  <Select
                    value={state.filters.category}
                    onValueChange={(value) =>
                      handleFilterChange("category", value)
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All categories" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All categories</SelectItem>
                      <SelectItem value="sports">Sports</SelectItem>
                      <SelectItem value="entertainment">
                        Entertainment
                      </SelectItem>
                      <SelectItem value="education">Education</SelectItem>
                      <SelectItem value="business">Business</SelectItem>
                      <SelectItem value="technology">Technology</SelectItem>
                      <SelectItem value="lifestyle">Lifestyle</SelectItem>
                      <SelectItem value="gaming">Gaming</SelectItem>
                      <SelectItem value="other">Other</SelectItem>
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
                      <SelectValue placeholder="All groups" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All groups</SelectItem>
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
            data={state.groups}
            columns={columns}
            loading={state.loading}
            pagination={state.pagination}
            onPageChange={handlePageChange}
            onSort={handleSort}
            onRowSelect={handleRowSelect}
            bulkActions={bulkActions}
            onBulkAction={handleBulkAction}
            onRefresh={handleRefresh}
            title="Groups Management"
            description="View and manage all groups"
            emptyMessage="No groups found"
            searchPlaceholder="Search groups..."
          />
        </div>

        {/* View Group Dialog */}
        <Dialog
          open={dialogs.viewGroup}
          onOpenChange={() => closeDialog("viewGroup")}
        >
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Group Details</DialogTitle>
              <DialogDescription>
                View detailed information about this group
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedGroup && (
              <div className="space-y-6">
                {/* Group Header */}
                <div className="flex items-center space-x-4">
                  <Avatar className="h-16 w-16">
                    <AvatarImage src={dialogs.selectedGroup.avatar} />
                    <AvatarFallback>
                      {dialogs.selectedGroup.name?.charAt(0) || "G"}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <h3 className="text-lg font-semibold flex items-center gap-2">
                      {dialogs.selectedGroup.name}
                      {dialogs.selectedGroup.is_verified && (
                        <Badge
                          variant="outline"
                          className="text-blue-600 border-blue-200"
                        >
                          <IconCheck className="h-3 w-3 mr-1" />
                          Verified
                        </Badge>
                      )}
                    </h3>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge
                        variant={getStatusBadgeVariant(dialogs.selectedGroup)}
                      >
                        {formatGroupStatus(dialogs.selectedGroup)}
                      </Badge>
                      <Badge variant="outline" className="capitalize">
                        {getGroupTypeIcon(dialogs.selectedGroup.type)}
                        {dialogs.selectedGroup.type?.toLowerCase() || "public"}
                      </Badge>
                    </div>
                  </div>
                  <div className="ml-auto text-center">
                    <p className="text-2xl font-bold">
                      {dialogs.selectedGroup.members_count?.toLocaleString() ||
                        "0"}
                    </p>
                    <p className="text-sm text-muted-foreground">Members</p>
                  </div>
                </div>

                <Separator />

                {/* Group Info */}
                <div className="space-y-4">
                  {/* Description */}
                  {dialogs.selectedGroup.description && (
                    <div>
                      <Label className="text-sm font-medium">Description</Label>
                      <p className="text-sm text-muted-foreground mt-1">
                        {dialogs.selectedGroup.description}
                      </p>
                    </div>
                  )}

                  {/* Group Details */}
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label className="text-sm font-medium">Category</Label>
                      <p className="text-sm text-muted-foreground capitalize">
                        {dialogs.selectedGroup.category || "Uncategorized"}
                      </p>
                    </div>
                    {dialogs.selectedGroup.location && (
                      <div>
                        <Label className="text-sm font-medium">Location</Label>
                        <p className="text-sm text-muted-foreground">
                          {dialogs.selectedGroup.location}
                        </p>
                      </div>
                    )}
                    {dialogs.selectedGroup.website && (
                      <div>
                        <Label className="text-sm font-medium">Website</Label>
                        <p className="text-sm text-muted-foreground">
                          <a
                            href={dialogs.selectedGroup.website}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-600 hover:underline"
                          >
                            {dialogs.selectedGroup.website}
                          </a>
                        </p>
                      </div>
                    )}
                    <div>
                      <Label className="text-sm font-medium">Created</Label>
                      <p className="text-sm text-muted-foreground">
                        {dialogs.selectedGroup.created_at
                          ? new Date(
                              dialogs.selectedGroup.created_at
                            ).toLocaleString()
                          : "Unknown"}
                      </p>
                    </div>
                  </div>

                  {/* Creator */}
                  <div>
                    <Label className="text-sm font-medium">Created By</Label>
                    <div className="flex items-center gap-3 mt-1">
                      <Avatar className="h-8 w-8">
                        <AvatarImage
                          src={dialogs.selectedGroup.creator?.profile_picture}
                          alt={
                            dialogs.selectedGroup.creator?.username || "Creator"
                          }
                        />
                        <AvatarFallback>
                          {(
                            dialogs.selectedGroup.creator?.first_name?.[0] ||
                            dialogs.selectedGroup.creator?.username?.[0] ||
                            "U"
                          ).toUpperCase()}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <p className="font-medium">
                          {dialogs.selectedGroup.creator?.first_name &&
                          dialogs.selectedGroup.creator?.last_name
                            ? `${dialogs.selectedGroup.creator.first_name} ${dialogs.selectedGroup.creator.last_name}`
                            : dialogs.selectedGroup.creator?.username ||
                              "Unknown Creator"}
                        </p>
                        {dialogs.selectedGroup.creator?.username && (
                          <p className="text-sm text-muted-foreground">
                            @{dialogs.selectedGroup.creator.username}
                          </p>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Tags */}
                  {dialogs.selectedGroup.tags &&
                    dialogs.selectedGroup.tags.length > 0 && (
                      <div>
                        <Label className="text-sm font-medium">Tags</Label>
                        <div className="flex flex-wrap gap-2 mt-2">
                          {dialogs.selectedGroup.tags.map((tag, index) => (
                            <Badge key={index} variant="secondary">
                              <IconTags className="h-3 w-3 mr-1" />
                              {tag}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}

                  {/* Group Settings */}
                  {dialogs.selectedGroup.settings && (
                    <div>
                      <Label className="text-sm font-medium">Settings</Label>
                      <div className="grid grid-cols-2 gap-4 mt-2">
                        <div className="flex items-center justify-between p-2 bg-muted rounded-md">
                          <span className="text-sm">Require Approval</span>
                          <Badge
                            variant={
                              dialogs.selectedGroup.settings.require_approval
                                ? "default"
                                : "secondary"
                            }
                          >
                            {dialogs.selectedGroup.settings.require_approval
                              ? "Yes"
                              : "No"}
                          </Badge>
                        </div>
                        <div className="flex items-center justify-between p-2 bg-muted rounded-md">
                          <span className="text-sm">Member Posts</span>
                          <Badge
                            variant={
                              dialogs.selectedGroup.settings.allow_member_posts
                                ? "default"
                                : "secondary"
                            }
                          >
                            {dialogs.selectedGroup.settings.allow_member_posts
                              ? "Allowed"
                              : "Restricted"}
                          </Badge>
                        </div>
                        <div className="flex items-center justify-between p-2 bg-muted rounded-md">
                          <span className="text-sm">Member Invites</span>
                          <Badge
                            variant={
                              dialogs.selectedGroup.settings
                                .allow_member_invites
                                ? "default"
                                : "secondary"
                            }
                          >
                            {dialogs.selectedGroup.settings.allow_member_invites
                              ? "Allowed"
                              : "Restricted"}
                          </Badge>
                        </div>
                        <div className="flex items-center justify-between p-2 bg-muted rounded-md">
                          <span className="text-sm">Content Moderation</span>
                          <Badge
                            variant={
                              dialogs.selectedGroup.settings.content_moderation
                                ? "default"
                                : "secondary"
                            }
                          >
                            {dialogs.selectedGroup.settings.content_moderation
                              ? "Enabled"
                              : "Disabled"}
                          </Badge>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => closeDialog("viewGroup")}
              >
                Close
              </Button>
              <Button
                onClick={() => {
                  closeDialog("viewGroup");
                  openDialog("viewMembers", dialogs.selectedGroup);
                }}
              >
                <IconUsers className="h-4 w-4 mr-2" />
                View Members
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* View Members Dialog */}
        <Dialog
          open={dialogs.viewMembers}
          onOpenChange={() => closeDialog("viewMembers")}
        >
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Group Members</DialogTitle>
              <DialogDescription>
                {dialogs.selectedGroup && (
                  <span>Members of {dialogs.selectedGroup.name}</span>
                )}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              {membersLoading ? (
                <div className="space-y-4">
                  {Array.from({ length: 5 }).map((_, i) => (
                    <div key={i} className="flex items-center gap-4">
                      <div className="h-10 w-10 rounded-full bg-muted animate-pulse" />
                      <div className="space-y-2 flex-1">
                        <div className="h-4 w-32 bg-muted animate-pulse rounded" />
                        <div className="h-3 w-24 bg-muted animate-pulse rounded" />
                      </div>
                      <div className="h-6 w-16 bg-muted animate-pulse rounded" />
                    </div>
                  ))}
                </div>
              ) : members.length === 0 ? (
                <div className="py-8 text-center">
                  <IconUsers className="h-12 w-12 mx-auto text-muted-foreground/50 mb-3" />
                  <p className="text-muted-foreground">No members found</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {members.map((member, index) => {
                    const user = member.user;

                    return (
                      <div
                        key={index}
                        className="flex items-center justify-between p-3 bg-muted/40 rounded-lg"
                      >
                        <div className="flex items-center gap-3">
                          <Avatar className="h-10 w-10">
                            <AvatarImage
                              src={user?.profile_picture}
                              alt={user?.username || "Member"}
                            />
                            <AvatarFallback>
                              {(
                                user?.first_name?.[0] ||
                                user?.username?.[0] ||
                                "U"
                              ).toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div>
                            <p className="font-medium">
                              {user?.first_name && user?.last_name
                                ? `${user.first_name} ${user.last_name}`
                                : user?.username || "Unknown User"}
                            </p>
                            {user?.username && (
                              <p className="text-sm text-muted-foreground">
                                @{user.username}
                              </p>
                            )}
                          </div>
                        </div>
                        <Badge className="capitalize">
                          {member.role || "member"}
                        </Badge>
                      </div>
                    );
                  })}
                </div>
              )}

              {/* Pagination for members */}
              {membersPagination && membersPagination.total_pages > 1 && (
                <div className="flex items-center justify-center gap-2 mt-4">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      handleMembersPageChange(
                        membersPagination.current_page - 1
                      )
                    }
                    disabled={!membersPagination.has_previous}
                  >
                    Previous
                  </Button>
                  <span className="text-sm text-muted-foreground">
                    Page {membersPagination.current_page} of{" "}
                    {membersPagination.total_pages}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      handleMembersPageChange(
                        membersPagination.current_page + 1
                      )
                    }
                    disabled={!membersPagination.has_next}
                  >
                    Next
                  </Button>
                </div>
              )}
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => closeDialog("viewMembers")}
              >
                Close
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Delete Group Dialog */}
        <Dialog
          open={dialogs.deleteGroup}
          onOpenChange={() => closeDialog("deleteGroup")}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete Group</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this group? This action cannot
                be undone.
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedGroup && (
              <div className="space-y-4">
                <div className="p-4 bg-muted rounded-lg">
                  <div className="flex items-center gap-3 mb-2">
                    <Avatar className="h-8 w-8">
                      <AvatarImage
                        src={dialogs.selectedGroup.avatar}
                        alt={dialogs.selectedGroup.name}
                      />
                      <AvatarFallback>
                        {dialogs.selectedGroup.name?.charAt(0) || "G"}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="font-medium">
                        {dialogs.selectedGroup.name}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {dialogs.selectedGroup.members_count || 0} members
                      </p>
                    </div>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {dialogs.selectedGroup.description?.substring(0, 100)}
                    {dialogs.selectedGroup.description &&
                    dialogs.selectedGroup.description.length > 100
                      ? "..."
                      : ""}
                  </p>
                </div>

                <div>
                  <Label htmlFor="reason">Reason for deletion (optional)</Label>
                  <Textarea
                    id="reason"
                    value={actionReason}
                    onChange={(e) => setActionReason(e.target.value)}
                    placeholder="Provide a reason for deleting this group..."
                  />
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
                onClick={() => closeDialog("deleteGroup")}
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDeleteGroup}
                disabled={formLoading}
              >
                {formLoading && (
                  <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Delete Group
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Update Status Dialog */}
        <Dialog
          open={dialogs.updateStatus}
          onOpenChange={() => closeDialog("updateStatus")}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                {dialogs.selectedGroup?.is_active
                  ? "Deactivate Group"
                  : "Activate Group"}
              </DialogTitle>
              <DialogDescription>
                {dialogs.selectedGroup?.is_active
                  ? "This will make the group inactive and hide it from users."
                  : "This will make the group active and visible to users."}
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedGroup && (
              <div className="space-y-4">
                <div className="p-4 bg-muted rounded-lg">
                  <div className="flex items-center gap-3 mb-2">
                    <Avatar className="h-8 w-8">
                      <AvatarImage
                        src={dialogs.selectedGroup.avatar}
                        alt={dialogs.selectedGroup.name}
                      />
                      <AvatarFallback>
                        {dialogs.selectedGroup.name?.charAt(0) || "G"}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="font-medium">
                        {dialogs.selectedGroup.name}
                      </p>
                      <div className="flex items-center gap-2">
                        <Badge
                          variant={getStatusBadgeVariant(dialogs.selectedGroup)}
                        >
                          {formatGroupStatus(dialogs.selectedGroup)}
                        </Badge>
                        <Badge variant="outline" className="capitalize">
                          {getGroupTypeIcon(dialogs.selectedGroup.type)}
                          {dialogs.selectedGroup.type?.toLowerCase() ||
                            "public"}
                        </Badge>
                      </div>
                    </div>
                  </div>
                </div>

                <div>
                  <Label htmlFor="reason">Reason (optional)</Label>
                  <Textarea
                    id="reason"
                    value={actionReason}
                    onChange={(e) => setActionReason(e.target.value)}
                    placeholder={`Provide a reason for ${
                      dialogs.selectedGroup.is_active
                        ? "deactivating"
                        : "activating"
                    } this group...`}
                  />
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
                onClick={() => closeDialog("updateStatus")}
              >
                Cancel
              </Button>
              <Button
                variant={
                  dialogs.selectedGroup?.is_active ? "destructive" : "default"
                }
                onClick={handleUpdateGroupStatus}
                disabled={formLoading}
              >
                {formLoading && (
                  <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                {dialogs.selectedGroup?.is_active
                  ? "Deactivate Group"
                  : "Activate Group"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </SidebarInset>
    </SidebarProvider>
  );
}

export default withAuth(GroupsPage);
