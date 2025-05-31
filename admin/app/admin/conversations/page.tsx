// admin/app/admin/conversations/page.tsx - Complete Conversations Management Page
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Progress } from "@/components/ui/progress";
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from "recharts";
import { withAuth, useAuth } from "@/contexts/auth-context";
import { apiClient } from "@/lib/api-client";
import { TableColumn, PaginationMeta } from "@/types/admin";
import {
  IconEye,
  IconTrash,
  IconDotsVertical,
  IconUser,
  IconMail,
  IconCalendar,
  IconMessage,
  IconLoader2,
  IconAlertCircle,
  IconSearch,
  IconFilter,
  IconRefresh,
  IconDownload,
  IconMessageCircle,
  IconUsers,
  IconClock,
  IconFile,
  IconVideo,
  IconMusic,
  IconPaperclip,
  IconX,
  IconCheck,
  IconBan,
  IconArchive,
  IconVolume,
  IconChartBar,
  IconFlag,
  IconShield,
  IconSettings,
} from "@tabler/icons-react";
import { ImageIcon, VolumeXIcon } from "lucide-react";

// Types
interface Conversation {
  id: string;
  type: "direct" | "group";
  title?: string;
  avatar?: string;
  participant_ids: string[];
  last_message_id?: string;
  last_message_at?: string;
  is_archived: boolean;
  is_muted: boolean;
  unread_count: number;
  created_at: string;
  participants?: Array<{
    id: string;
    username: string;
    first_name?: string;
    last_name?: string;
    profile_picture?: string;
  }>;
  last_message?: {
    id: string;
    content?: string;
    content_type: string;
    created_at: string;
    sender?: {
      id: string;
      username: string;
    };
  };
}

interface Message {
  id: string;
  content?: string;
  content_type: "text" | "image" | "video" | "audio" | "file" | "location";
  media_url?: string;
  file_name?: string;
  file_size?: number;
  is_read: boolean;
  read_at?: string;
  is_edited: boolean;
  edited_at?: string;
  created_at: string;
  sender?: {
    id: string;
    username: string;
    first_name?: string;
    last_name?: string;
    profile_picture?: string;
  };
}

interface ConversationAnalytics {
  conversation_id: string;
  conversation_type: string;
  participant_count: number;
  message_statistics: Array<{
    total_messages: number;
    text_messages: number;
    media_messages: number;
    read_messages: number;
    avg_length: number;
  }>;
  activity_by_day: Array<{
    _id: string;
    message_count: number;
  }>;
  participant_activity: Array<{
    _id: string;
    username: string;
    message_count: number;
    last_message: string;
  }>;
}

interface ConversationReport {
  id: string;
  reason: string;
  status: string;
  created_at: string;
  reporter: {
    username: string;
  };
  message_content: string;
}

interface ConversationsPageState {
  conversations: Conversation[];
  messages: Message[];
  analytics: ConversationAnalytics | null;
  reports: ConversationReport[];
  loading: boolean;
  messagesLoading: boolean;
  analyticsLoading: boolean;
  reportsLoading: boolean;
  error: string | null;
  pagination: PaginationMeta | undefined;
  filters: {
    search: string;
    type: string;
    is_archived: string;
    is_muted: string;
    date_from: string;
    date_to: string;
    page: number;
    limit: number;
    sort_by?: string;
    sort_order?: "asc" | "desc";
  };
  selectedConversations: string[];
  activeTab: "conversations" | "messages" | "analytics" | "reports";
  selectedConversationId: string | null;
}

interface DialogState {
  viewConversation: boolean;
  viewMessages: boolean;
  viewAnalytics: boolean;
  viewReports: boolean;
  deleteConversation: boolean;
  selectedConversation: Conversation | null;
}

const initialFilters = {
  search: "",
  type: "all",
  is_archived: "all",
  is_muted: "all",
  date_from: "",
  date_to: "",
  page: 1,
  limit: 20,
};

function ConversationsPage() {
  const { user: currentUser } = useAuth();

  const [state, setState] = useState<ConversationsPageState>({
    conversations: [],
    messages: [],
    analytics: null,
    reports: [],
    loading: true,
    messagesLoading: false,
    analyticsLoading: false,
    reportsLoading: false,
    error: null,
    pagination: undefined,
    filters: initialFilters,
    selectedConversations: [],
    activeTab: "conversations",
    selectedConversationId: null,
  });

  const [dialogs, setDialogs] = useState<DialogState>({
    viewConversation: false,
    viewMessages: false,
    viewAnalytics: false,
    viewReports: false,
    deleteConversation: false,
    selectedConversation: null,
  });

  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [deleteReason, setDeleteReason] = useState("");
  const [deleteMessages, setDeleteMessages] = useState(false);

  // Fetch conversations
  const fetchConversations = useCallback(async (filters = state.filters) => {
    try {
      setState((prev) => ({ ...prev, loading: true, error: null }));

      const params: any = {
        page: filters.page,
        limit: filters.limit,
      };

      if (filters.search) params.search = filters.search;
      if (filters.type && filters.type !== "all") params.type = filters.type;
      if (filters.is_archived && filters.is_archived !== "all") {
        params.is_archived = filters.is_archived === "true";
      }
      if (filters.is_muted && filters.is_muted !== "all") {
        params.is_muted = filters.is_muted === "true";
      }
      if (filters.date_from) params.date_from = filters.date_from;
      if (filters.date_to) params.date_to = filters.date_to;
      if (filters.sort_by) {
        params.sort_by = filters.sort_by;
        params.sort_order = filters.sort_order || "desc";
      }

      console.log("📡 Fetching conversations with params:", params);
      const response = await apiClient.getConversations(params);

      const validConversations = (response.data || []).filter(
        (conv: any) => conv && conv.id && typeof conv.id === "string"
      );

      setState((prev) => ({
        ...prev,
        conversations: validConversations,
        pagination: response.pagination || undefined,
        loading: false,
      }));
    } catch (error: any) {
      console.error("❌ Failed to fetch conversations:", error);
      setState((prev) => ({
        ...prev,
        loading: false,
        conversations: [],
        error: error.message || "Failed to fetch conversations",
      }));
    }
  }, []);

  // Fetch conversation messages
  const fetchConversationMessages = useCallback(
    async (conversationId: string) => {
      try {
        setState((prev) => ({ ...prev, messagesLoading: true }));

        const response = await apiClient.getConversationMessages(
          conversationId,
          {
            page: 1,
            limit: 100,
          }
        );

        setState((prev) => ({
          ...prev,
          messages: response.data || [],
          messagesLoading: false,
        }));
      } catch (error: any) {
        console.error("❌ Failed to fetch conversation messages:", error);
        setState((prev) => ({
          ...prev,
          messages: [],
          messagesLoading: false,
        }));
      }
    },
    []
  );

  // Fetch conversation analytics
  const fetchConversationAnalytics = useCallback(
    async (conversationId: string) => {
      try {
        setState((prev) => ({ ...prev, analyticsLoading: true }));

        const response = await apiClient.getConversationAnalytics(
          conversationId
        );

        setState((prev) => ({
          ...prev,
          analytics: response.data,
          analyticsLoading: false,
        }));
      } catch (error: any) {
        console.error("❌ Failed to fetch conversation analytics:", error);
        setState((prev) => ({
          ...prev,
          analytics: null,
          analyticsLoading: false,
        }));
      }
    },
    []
  );

  // Fetch conversation reports
  const fetchConversationReports = useCallback(
    async (conversationId: string) => {
      try {
        setState((prev) => ({ ...prev, reportsLoading: true }));

        const response = await apiClient.getConversationReports(conversationId);

        setState((prev) => ({
          ...prev,
          reports: response.data || [],
          reportsLoading: false,
        }));
      } catch (error: any) {
        console.error("❌ Failed to fetch conversation reports:", error);
        setState((prev) => ({
          ...prev,
          reports: [],
          reportsLoading: false,
        }));
      }
    },
    []
  );

  // Initial load
  useEffect(() => {
    fetchConversations();
  }, []);

  // Handle search with debounce
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (state.filters.search !== undefined) {
        fetchConversations({ ...state.filters, page: 1 });
      }
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [state.filters.search, fetchConversations]);

  // Handle filter changes
  const handleFilterChange = (key: string, value: any) => {
    const newFilters = { ...state.filters, [key]: value, page: 1 };
    setState((prev) => ({ ...prev, filters: newFilters }));

    if (key !== "search") {
      fetchConversations(newFilters);
    }
  };

  // Handle pagination
  const handlePageChange = (page: number) => {
    const newFilters = { ...state.filters, page };
    setState((prev) => ({ ...prev, filters: newFilters }));
    fetchConversations(newFilters);
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
    fetchConversations(newFilters);
  };

  // Handle row selection
  const handleRowSelect = (selectedRows: string[]) => {
    setState((prev) => ({ ...prev, selectedConversations: selectedRows }));
  };

  // Open dialog
  const openDialog = (
    type: keyof DialogState,
    conversation: Conversation | null = null
  ) => {
    setDialogs((prev) => ({
      ...prev,
      [type]: true,
      selectedConversation: conversation,
    }));

    if (conversation) {
      setState((prev) => ({
        ...prev,
        selectedConversationId: conversation.id,
      }));

      // Load additional data based on dialog type
      if (type === "viewMessages") {
        fetchConversationMessages(conversation.id);
      } else if (type === "viewAnalytics") {
        fetchConversationAnalytics(conversation.id);
      } else if (type === "viewReports") {
        fetchConversationReports(conversation.id);
      }
    }
  };

  // Close dialog
  const closeDialog = (type: keyof DialogState) => {
    setDialogs((prev) => ({
      ...prev,
      [type]: false,
      selectedConversation: null,
    }));
    setFormError(null);
    setDeleteReason("");
    setDeleteMessages(false);
  };

  // Handle conversation deletion
  const handleDeleteConversation = async () => {
    if (!dialogs.selectedConversation) return;

    setFormLoading(true);
    try {
      await apiClient.deleteConversation(
        dialogs.selectedConversation.id,
        deleteReason || "Deleted by admin",
        deleteMessages
      );
      closeDialog("deleteConversation");
      fetchConversations();
    } catch (error: any) {
      setFormError(error.message || "Failed to delete conversation");
    } finally {
      setFormLoading(false);
    }
  };

  // Handle bulk actions
  const handleBulkAction = async (action: string, selectedIds: string[]) => {
    setFormLoading(true);
    try {
      await apiClient.bulkConversationAction({
        conversation_ids: selectedIds,
        action,
        reason: deleteReason || "Bulk action by admin",
      });
      setState((prev) => ({ ...prev, selectedConversations: [] }));
      fetchConversations();
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
    fetchConversations();
  };

  // Handle export
  const handleExport = async () => {
    try {
      const csvContent = state.conversations.map((conversation) => ({
        id: conversation.id,
        type: conversation.type,
        title: conversation.title || "",
        participant_count: conversation.participant_ids?.length || 0,
        is_archived: conversation.is_archived,
        is_muted: conversation.is_muted,
        unread_count: conversation.unread_count,
        created_at: conversation.created_at,
        last_activity: conversation.last_message_at || "",
      }));

      console.log("✅ Export data prepared", csvContent);
    } catch (error) {
      console.error("❌ Export failed:", error);
    }
  };

  // Handle tab change with data loading
  const handleTabChange = (
    tab: "conversations" | "messages" | "analytics" | "reports"
  ) => {
    setState((prev) => ({ ...prev, activeTab: tab }));

    if (state.selectedConversationId) {
      if (tab === "messages" && state.messages.length === 0) {
        fetchConversationMessages(state.selectedConversationId);
      } else if (tab === "analytics" && !state.analytics) {
        fetchConversationAnalytics(state.selectedConversationId);
      } else if (tab === "reports" && state.reports.length === 0) {
        fetchConversationReports(state.selectedConversationId);
      }
    }
  };

  // Format conversation title
  const formatConversationTitle = (conversation: Conversation) => {
    if (conversation.title) return conversation.title;
    if (conversation.type === "group")
      return `Group Chat (${
        conversation.participant_ids?.length || 0
      } members)`;
    return "Direct Message";
  };

  // Get content type icon
  const getContentTypeIcon = (type: string) => {
    switch (type) {
      case "image":
        return <ImageIcon className="h-4 w-4" />;
      case "video":
        return <IconVideo className="h-4 w-4" />;
      case "audio":
        return <IconMusic className="h-4 w-4" />;
      case "file":
        return <IconFile className="h-4 w-4" />;
      case "location":
        return <IconUser className="h-4 w-4" />;
      default:
        return <IconMessage className="h-4 w-4" />;
    }
  };

  // Truncate content
  const truncateContent = (
    content: string | undefined,
    maxLength: number = 50
  ) => {
    if (!content) return "No content";
    return content.length > maxLength
      ? content.substring(0, maxLength) + "..."
      : content;
  };

  // Conversations table columns
  const conversationColumns: TableColumn[] = [
    {
      key: "title",
      label: "Conversation",
      sortable: true,
      render: (_, conversation: Conversation) => (
        <div className="flex items-center gap-3">
          <div className="relative">
            {conversation.type === "group" ? (
              <div className="flex -space-x-2">
                {(conversation.participants || [])
                  .filter((participant) => participant && participant.id)
                  .slice(0, 3)
                  .map((participant, index) => (
                    <Avatar
                      key={`avatar-${participant.id || index}`}
                      className="h-8 w-8 border-2 border-background"
                    >
                      <AvatarImage src={participant.profile_picture} />
                      <AvatarFallback>
                        {(
                          participant.first_name?.[0] ||
                          participant.username?.[0] ||
                          "U"
                        ).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                  ))}
                {(conversation.participants?.length || 0) > 3 && (
                  <div className="h-8 w-8 rounded-full bg-muted border-2 border-background flex items-center justify-center text-xs">
                    +{(conversation.participants?.length || 0) - 3}
                  </div>
                )}
              </div>
            ) : (
              <Avatar className="h-8 w-8">
                <AvatarImage
                  src={conversation.participants?.[0]?.profile_picture}
                />
                <AvatarFallback>
                  {(
                    conversation.participants?.[0]?.first_name?.[0] ||
                    conversation.participants?.[0]?.username?.[0] ||
                    "U"
                  ).toUpperCase()}
                </AvatarFallback>
              </Avatar>
            )}
          </div>
          <div className="flex flex-col">
            <span className="font-medium">
              {formatConversationTitle(conversation)}
            </span>
            <span className="text-sm text-muted-foreground">
              {conversation.type === "group"
                ? `${conversation.participant_ids?.length || 0} members`
                : conversation.participants?.[0]?.username || "Direct message"}
            </span>
          </div>
          <div className="flex gap-1">
            {conversation.is_archived && (
              <Badge variant="secondary" className="text-xs">
                <IconArchive className="h-3 w-3 mr-1" />
                Archived
              </Badge>
            )}
            {conversation.is_muted && (
              <Badge variant="outline" className="text-xs">
                <VolumeXIcon className="h-3 w-3 mr-1" />
                Muted
              </Badge>
            )}
          </div>
        </div>
      ),
    },
    {
      key: "type",
      label: "Type",
      filterable: true,
      render: (value: string) => (
        <Badge variant={value === "group" ? "default" : "secondary"}>
          {value === "group" ? "Group" : "Direct"}
        </Badge>
      ),
    },
    {
      key: "last_message",
      label: "Last Message",
      render: (_, conversation: Conversation) => (
        <div className="max-w-xs">
          {conversation.last_message ? (
            <div className="flex flex-col">
              <div className="flex items-center gap-2">
                {getContentTypeIcon(conversation.last_message.content_type)}
                <span className="text-sm">
                  {conversation.last_message.content_type === "text"
                    ? truncateContent(conversation.last_message.content)
                    : `${conversation.last_message.content_type} message`}
                </span>
              </div>
              <span className="text-xs text-muted-foreground">
                {conversation.last_message.sender?.username || "Unknown"} •{" "}
                {new Date(
                  conversation.last_message.created_at
                ).toLocaleTimeString()}
              </span>
            </div>
          ) : (
            <span className="text-sm text-muted-foreground">No messages</span>
          )}
        </div>
      ),
    },
    {
      key: "unread_count",
      label: "Unread",
      sortable: true,
      render: (value: number) =>
        value > 0 ? (
          <Badge variant="destructive">{value}</Badge>
        ) : (
          <span className="text-muted-foreground">-</span>
        ),
    },
    {
      key: "last_message_at",
      label: "Last Activity",
      sortable: true,
      render: (value: string) =>
        value ? (
          <div className="flex flex-col">
            <span className="text-sm">
              {new Date(value).toLocaleDateString()}
            </span>
            <span className="text-xs text-muted-foreground">
              {new Date(value).toLocaleTimeString()}
            </span>
          </div>
        ) : (
          <span className="text-muted-foreground">Never</span>
        ),
    },
    {
      key: "actions",
      label: "",
      width: "w-12",
      render: (_, conversation: Conversation) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
              <IconDotsVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              onClick={() => openDialog("viewConversation", conversation)}
            >
              <IconEye className="h-4 w-4 mr-2" />
              View Details
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => openDialog("viewMessages", conversation)}
            >
              <IconMessage className="h-4 w-4 mr-2" />
              View Messages
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => openDialog("viewAnalytics", conversation)}
            >
              <IconChartBar className="h-4 w-4 mr-2" />
              Analytics
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => openDialog("viewReports", conversation)}
            >
              <IconFlag className="h-4 w-4 mr-2" />
              Reports
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => openDialog("deleteConversation", conversation)}
              className="text-red-600"
            >
              <IconTrash className="h-4 w-4 mr-2" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  // Messages table columns for sub-messages view
  const messageColumns: TableColumn[] = [
    {
      key: "sender",
      label: "Sender",
      render: (_, message: Message) => (
        <div className="flex items-center gap-2">
          <Avatar className="h-6 w-6">
            <AvatarImage src={message.sender?.profile_picture} />
            <AvatarFallback>
              {(
                message.sender?.first_name?.[0] ||
                message.sender?.username?.[0] ||
                "U"
              ).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <span className="text-sm font-medium">
            {message.sender?.username || "Unknown"}
          </span>
        </div>
      ),
    },
    {
      key: "content",
      label: "Content",
      render: (_, message: Message) => (
        <div className="flex items-center gap-2 max-w-md">
          {getContentTypeIcon(message.content_type)}
          <span className="text-sm">
            {message.content_type === "text"
              ? truncateContent(message.content, 100)
              : message.file_name || `${message.content_type} message`}
          </span>
        </div>
      ),
    },
    {
      key: "is_read",
      label: "Status",
      render: (value: boolean, message: Message) => (
        <div className="flex flex-col gap-1">
          <Badge variant={value ? "default" : "secondary"}>
            {value ? "Read" : "Unread"}
          </Badge>
          {message.is_edited && (
            <Badge variant="outline" className="text-xs">
              Edited
            </Badge>
          )}
        </div>
      ),
    },
    {
      key: "created_at",
      label: "Sent",
      render: (value: string) => (
        <div className="flex flex-col">
          <span className="text-sm">
            {new Date(value).toLocaleDateString()}
          </span>
          <span className="text-xs text-muted-foreground">
            {new Date(value).toLocaleTimeString()}
          </span>
        </div>
      ),
    },
  ];

  // Bulk actions
  const bulkActions = [
    {
      label: "Archive Conversations",
      action: "archive",
      variant: "default" as const,
    },
    {
      label: "Unarchive Conversations",
      action: "unarchive",
      variant: "default" as const,
    },
    {
      label: "Mute Conversations",
      action: "mute",
      variant: "default" as const,
    },
    {
      label: "Unmute Conversations",
      action: "unmute",
      variant: "default" as const,
    },
    {
      label: "Delete Conversations",
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
                <div>Failed to load conversations: {state.error}</div>
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
              <h1 className="text-3xl font-bold">Conversations</h1>
              <p className="text-muted-foreground">
                Manage conversations, messages, and communication analytics
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" onClick={handleRefresh}>
                <IconRefresh className="h-4 w-4 mr-2" />
                Refresh
              </Button>
              <Button variant="outline" onClick={handleExport}>
                <IconDownload className="h-4 w-4 mr-2" />
                Export
              </Button>
            </div>
          </div>

          {/* Filters */}
          <Card>
            <CardHeader>
              <CardTitle>Filters</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <div>
                  <Label htmlFor="search">Search</Label>
                  <div className="relative">
                    <IconSearch className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="search"
                      placeholder="Search conversations..."
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
                      <SelectItem value="direct">Direct Messages</SelectItem>
                      <SelectItem value="group">Group Chats</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="is_archived">Archived Status</Label>
                  <Select
                    value={state.filters.is_archived}
                    onValueChange={(value) =>
                      handleFilterChange("is_archived", value)
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All conversations" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All conversations</SelectItem>
                      <SelectItem value="false">Active</SelectItem>
                      <SelectItem value="true">Archived</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="is_muted">Muted Status</Label>
                  <Select
                    value={state.filters.is_muted}
                    onValueChange={(value) =>
                      handleFilterChange("is_muted", value)
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All conversations" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All conversations</SelectItem>
                      <SelectItem value="false">Not Muted</SelectItem>
                      <SelectItem value="true">Muted</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Main Data Table */}
          <DataTable
            data={state.conversations}
            columns={conversationColumns}
            loading={state.loading}
            pagination={state.pagination}
            onPageChange={handlePageChange}
            onSort={handleSort}
            onRowSelect={handleRowSelect}
            bulkActions={bulkActions}
            onBulkAction={handleBulkAction}
            onRefresh={handleRefresh}
            onExport={handleExport}
            title="Conversation Management"
            description="View and manage all conversations and messaging"
            emptyMessage="No conversations found"
            searchPlaceholder="Search conversations..."
          />
        </div>

        {/* View Conversation Dialog */}
        <Dialog
          open={dialogs.viewConversation}
          onOpenChange={() => closeDialog("viewConversation")}
        >
          <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Conversation Details</DialogTitle>
              <DialogDescription>
                Detailed view of conversation information and activity
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedConversation && (
              <div className="space-y-6">
                {/* Conversation Header */}
                <div className="flex items-center space-x-4">
                  <div className="relative">
                    {dialogs.selectedConversation.type === "group" ? (
                      <div className="flex -space-x-2">
                        {(dialogs.selectedConversation.participants || [])
                          .filter(
                            (participant) => participant && participant.id
                          )
                          .slice(0, 3)
                          .map((participant, index) => (
                            <Avatar
                              key={`modal-avatar-${participant.id || index}`}
                              className="h-12 w-12 border-2 border-background"
                            >
                              <AvatarImage src={participant.profile_picture} />
                              <AvatarFallback>
                                {(
                                  participant.first_name?.[0] ||
                                  participant.username?.[0] ||
                                  "U"
                                ).toUpperCase()}
                              </AvatarFallback>
                            </Avatar>
                          ))}
                      </div>
                    ) : (
                      <Avatar className="h-12 w-12">
                        <AvatarImage
                          src={
                            dialogs.selectedConversation.participants?.[0]
                              ?.profile_picture
                          }
                        />
                        <AvatarFallback>
                          {(
                            dialogs.selectedConversation.participants?.[0]
                              ?.first_name?.[0] ||
                            dialogs.selectedConversation.participants?.[0]
                              ?.username?.[0] ||
                            "U"
                          ).toUpperCase()}
                        </AvatarFallback>
                      </Avatar>
                    )}
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold">
                      {formatConversationTitle(dialogs.selectedConversation)}
                    </h3>
                    <p className="text-muted-foreground">
                      {dialogs.selectedConversation.type === "group"
                        ? `${
                            dialogs.selectedConversation.participant_ids
                              ?.length || 0
                          } members`
                        : "Direct conversation"}
                    </p>
                    <div className="flex gap-2 mt-2">
                      <Badge
                        variant={
                          dialogs.selectedConversation.type === "group"
                            ? "default"
                            : "secondary"
                        }
                      >
                        {dialogs.selectedConversation.type === "group"
                          ? "Group Chat"
                          : "Direct Message"}
                      </Badge>
                      {dialogs.selectedConversation.is_archived && (
                        <Badge variant="secondary">
                          <IconArchive className="h-3 w-3 mr-1" />
                          Archived
                        </Badge>
                      )}
                      {dialogs.selectedConversation.is_muted && (
                        <Badge variant="outline">
                          <VolumeXIcon className="h-3 w-3 mr-1" />
                          Muted
                        </Badge>
                      )}
                    </div>
                  </div>
                </div>

                <Separator />

                {/* Quick Stats */}
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <Card>
                    <CardContent className="p-4">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-sm text-muted-foreground">
                            Participants
                          </p>
                          <p className="text-2xl font-bold">
                            {dialogs.selectedConversation.participant_ids
                              ?.length || 0}
                          </p>
                        </div>
                        <IconUsers className="h-8 w-8 text-muted-foreground" />
                      </div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardContent className="p-4">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-sm text-muted-foreground">
                            Unread Messages
                          </p>
                          <p className="text-2xl font-bold">
                            {dialogs.selectedConversation.unread_count}
                          </p>
                        </div>
                        <IconMessage className="h-8 w-8 text-muted-foreground" />
                      </div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardContent className="p-4">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-sm text-muted-foreground">
                            Created
                          </p>
                          <p className="text-sm font-medium">
                            {new Date(
                              dialogs.selectedConversation.created_at
                            ).toLocaleDateString()}
                          </p>
                        </div>
                        <IconCalendar className="h-8 w-8 text-muted-foreground" />
                      </div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardContent className="p-4">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-sm text-muted-foreground">
                            Last Activity
                          </p>
                          <p className="text-sm font-medium">
                            {dialogs.selectedConversation.last_message_at
                              ? new Date(
                                  dialogs.selectedConversation.last_message_at
                                ).toLocaleDateString()
                              : "No activity"}
                          </p>
                        </div>
                        <IconClock className="h-8 w-8 text-muted-foreground" />
                      </div>
                    </CardContent>
                  </Card>
                </div>

                {/* Participants List */}
                {dialogs.selectedConversation.participants &&
                  dialogs.selectedConversation.participants.length > 0 && (
                    <div>
                      <Label className="text-sm font-medium">
                        Participants
                      </Label>
                      <div className="mt-2 grid grid-cols-1 md:grid-cols-2 gap-2">
                        {dialogs.selectedConversation.participants
                          .filter(
                            (participant) => participant && participant.id
                          )
                          .map((participant, index) => (
                            <div
                              key={`participant-${participant.id || index}`}
                              className="flex items-center gap-3 p-3 bg-muted rounded-lg"
                            >
                              <Avatar className="h-8 w-8">
                                <AvatarImage
                                  src={participant.profile_picture}
                                />
                                <AvatarFallback className="text-xs">
                                  {(
                                    participant.first_name?.[0] ||
                                    participant.username?.[0] ||
                                    "U"
                                  ).toUpperCase()}
                                </AvatarFallback>
                              </Avatar>
                              <div>
                                <span className="text-sm font-medium">
                                  {participant.username || "Unknown User"}
                                </span>
                                <span className="text-xs text-muted-foreground ml-2">
                                  {participant.first_name}{" "}
                                  {participant.last_name}
                                </span>
                              </div>
                            </div>
                          ))}
                      </div>
                    </div>
                  )}

                {/* Last Message */}
                {dialogs.selectedConversation.last_message && (
                  <div>
                    <Label className="text-sm font-medium">Last Message</Label>
                    <div className="mt-2 p-4 bg-muted rounded-lg">
                      <div className="flex items-center gap-2 mb-2">
                        <span className="text-sm font-medium">
                          {dialogs.selectedConversation.last_message.sender
                            ?.username || "Unknown"}
                        </span>
                        <span className="text-xs text-muted-foreground">
                          {new Date(
                            dialogs.selectedConversation.last_message.created_at
                          ).toLocaleString()}
                        </span>
                        {getContentTypeIcon(
                          dialogs.selectedConversation.last_message.content_type
                        )}
                      </div>
                      <p className="text-sm">
                        {dialogs.selectedConversation.last_message.content ||
                          `${dialogs.selectedConversation.last_message.content_type} message`}
                      </p>
                    </div>
                  </div>
                )}
              </div>
            )}

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => closeDialog("viewConversation")}
              >
                Close
              </Button>
              <Button
                onClick={() => {
                  closeDialog("viewConversation");
                  if (dialogs.selectedConversation)
                    openDialog("viewMessages", dialogs.selectedConversation);
                }}
              >
                <IconMessage className="h-4 w-4 mr-2" />
                View Messages
              </Button>
              <Button
                onClick={() => {
                  closeDialog("viewConversation");
                  if (dialogs.selectedConversation)
                    openDialog("viewAnalytics", dialogs.selectedConversation);
                }}
              >
                <IconChartBar className="h-4 w-4 mr-2" />
                Analytics
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* View Messages Dialog */}
        <Dialog
          open={dialogs.viewMessages}
          onOpenChange={() => closeDialog("viewMessages")}
        >
          <DialogContent className="max-w-6xl max-h-[80vh]">
            <DialogHeader>
              <DialogTitle>Conversation Messages</DialogTitle>
              <DialogDescription>
                All messages in this conversation
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              {state.messagesLoading ? (
                <div className="flex items-center justify-center p-8">
                  <IconLoader2 className="h-8 w-8 animate-spin" />
                  <span className="ml-2">Loading messages...</span>
                </div>
              ) : state.messages.length > 0 ? (
                <div className="max-h-96 overflow-y-auto border rounded-lg">
                  <DataTable
                    data={state.messages}
                    columns={messageColumns}
                    loading={false}
                    showPagination={false}
                    title="Messages in Conversation"
                    description={`${state.messages.length} messages found`}
                    emptyMessage="No messages found"
                  />
                </div>
              ) : (
                <div className="text-center p-8">
                  <IconMessage className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                  <p className="text-muted-foreground">No messages found</p>
                </div>
              )}
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => closeDialog("viewMessages")}
              >
                Close
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Analytics Dialog */}
        <Dialog
          open={dialogs.viewAnalytics}
          onOpenChange={() => closeDialog("viewAnalytics")}
        >
          <DialogContent className="max-w-6xl max-h-[80vh]">
            <DialogHeader>
              <DialogTitle>Conversation Analytics</DialogTitle>
              <DialogDescription>
                Detailed analytics and insights for this conversation
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-6">
              {state.analyticsLoading ? (
                <div className="flex items-center justify-center p-8">
                  <IconLoader2 className="h-8 w-8 animate-spin" />
                  <span className="ml-2">Loading analytics...</span>
                </div>
              ) : state.analytics ? (
                <>
                  {/* Message Statistics */}
                  {state.analytics.message_statistics?.length > 0 && (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                      <Card>
                        <CardContent className="p-4">
                          <div className="text-center">
                            <p className="text-2xl font-bold">
                              {state.analytics.message_statistics[0]
                                ?.total_messages || 0}
                            </p>
                            <p className="text-sm text-muted-foreground">
                              Total Messages
                            </p>
                          </div>
                        </CardContent>
                      </Card>
                      <Card>
                        <CardContent className="p-4">
                          <div className="text-center">
                            <p className="text-2xl font-bold">
                              {state.analytics.message_statistics[0]
                                ?.text_messages || 0}
                            </p>
                            <p className="text-sm text-muted-foreground">
                              Text Messages
                            </p>
                          </div>
                        </CardContent>
                      </Card>
                      <Card>
                        <CardContent className="p-4">
                          <div className="text-center">
                            <p className="text-2xl font-bold">
                              {state.analytics.message_statistics[0]
                                ?.media_messages || 0}
                            </p>
                            <p className="text-sm text-muted-foreground">
                              Media Messages
                            </p>
                          </div>
                        </CardContent>
                      </Card>
                      <Card>
                        <CardContent className="p-4">
                          <div className="text-center">
                            <p className="text-2xl font-bold">
                              {Math.round(
                                state.analytics.message_statistics[0]
                                  ?.avg_length || 0
                              )}
                            </p>
                            <p className="text-sm text-muted-foreground">
                              Avg Length
                            </p>
                          </div>
                        </CardContent>
                      </Card>
                    </div>
                  )}

                  {/* Activity Chart */}
                  {state.analytics.activity_by_day?.length > 0 && (
                    <Card>
                      <CardHeader>
                        <CardTitle>Message Activity (Last 30 Days)</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <ResponsiveContainer width="100%" height={300}>
                          <LineChart data={state.analytics.activity_by_day}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis dataKey="_id" />
                            <YAxis />
                            <Tooltip />
                            <Line
                              type="monotone"
                              dataKey="message_count"
                              stroke="#8884d8"
                              strokeWidth={2}
                            />
                          </LineChart>
                        </ResponsiveContainer>
                      </CardContent>
                    </Card>
                  )}

                  {/* Participant Activity */}
                  {state.analytics.participant_activity?.length > 0 && (
                    <Card>
                      <CardHeader>
                        <CardTitle>Participant Activity</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <ResponsiveContainer width="100%" height={300}>
                          <BarChart data={state.analytics.participant_activity}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis dataKey="username" />
                            <YAxis />
                            <Tooltip />
                            <Bar dataKey="message_count" fill="#8884d8" />
                          </BarChart>
                        </ResponsiveContainer>
                      </CardContent>
                    </Card>
                  )}
                </>
              ) : (
                <div className="text-center p-8">
                  <IconChartBar className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                  <p className="text-muted-foreground">
                    No analytics data available
                  </p>
                </div>
              )}
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => closeDialog("viewAnalytics")}
              >
                Close
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Reports Dialog */}
        <Dialog
          open={dialogs.viewReports}
          onOpenChange={() => closeDialog("viewReports")}
        >
          <DialogContent className="max-w-4xl max-h-[80vh]">
            <DialogHeader>
              <DialogTitle>Conversation Reports</DialogTitle>
              <DialogDescription>
                Reports and moderation issues for this conversation
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              {state.reportsLoading ? (
                <div className="flex items-center justify-center p-8">
                  <IconLoader2 className="h-8 w-8 animate-spin" />
                  <span className="ml-2">Loading reports...</span>
                </div>
              ) : state.reports.length > 0 ? (
                <div className="space-y-4">
                  {state.reports.map((report, index) => (
                    <Card key={`report-${report.id || index}`}>
                      <CardContent className="p-4">
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            <div className="flex items-center gap-2 mb-2">
                              <Badge
                                variant={
                                  report.status === "resolved"
                                    ? "default"
                                    : report.status === "pending"
                                    ? "secondary"
                                    : "destructive"
                                }
                              >
                                {report.status}
                              </Badge>
                              <span className="text-sm text-muted-foreground">
                                {report.reason}
                              </span>
                            </div>
                            <p className="text-sm mb-2">
                              <strong>Reported by:</strong>{" "}
                              {report.reporter?.username}
                            </p>
                            <p className="text-sm text-muted-foreground">
                              <strong>Message:</strong>{" "}
                              {truncateContent(report.message_content, 100)}
                            </p>
                          </div>
                          <div className="text-right">
                            <p className="text-xs text-muted-foreground">
                              {new Date(report.created_at).toLocaleString()}
                            </p>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              ) : (
                <div className="text-center p-8">
                  <IconFlag className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                  <p className="text-muted-foreground">No reports found</p>
                </div>
              )}
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => closeDialog("viewReports")}
              >
                Close
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Delete Conversation Dialog */}
        <Dialog
          open={dialogs.deleteConversation}
          onOpenChange={() => closeDialog("deleteConversation")}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete Conversation</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this conversation? This action
                cannot be undone.
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedConversation && (
              <div className="space-y-4">
                <div className="p-4 bg-muted rounded-lg">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="relative">
                      {dialogs.selectedConversation.type === "group" ? (
                        <div className="flex -space-x-1">
                          {(dialogs.selectedConversation.participants || [])
                            .slice(0, 2)
                            .map((participant, index) => (
                              <Avatar
                                key={`delete-avatar-${participant.id || index}`}
                                className="h-6 w-6 border border-background"
                              >
                                <AvatarImage
                                  src={participant.profile_picture}
                                />
                                <AvatarFallback className="text-xs">
                                  {(
                                    participant.first_name?.[0] ||
                                    participant.username?.[0] ||
                                    "U"
                                  ).toUpperCase()}
                                </AvatarFallback>
                              </Avatar>
                            ))}
                        </div>
                      ) : (
                        <Avatar className="h-6 w-6">
                          <AvatarImage
                            src={
                              dialogs.selectedConversation.participants?.[0]
                                ?.profile_picture
                            }
                          />
                          <AvatarFallback className="text-xs">
                            {(
                              dialogs.selectedConversation.participants?.[0]
                                ?.first_name?.[0] ||
                              dialogs.selectedConversation.participants?.[0]
                                ?.username?.[0] ||
                              "U"
                            ).toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                      )}
                    </div>
                    <div>
                      <p className="font-medium">
                        {formatConversationTitle(dialogs.selectedConversation)}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        {dialogs.selectedConversation.type === "group"
                          ? `${
                              dialogs.selectedConversation.participant_ids
                                ?.length || 0
                            } participants`
                          : "Direct message"}
                      </p>
                    </div>
                  </div>
                </div>

                <div>
                  <Label htmlFor="deleteReason">
                    Reason for deletion (optional)
                  </Label>
                  <Textarea
                    id="deleteReason"
                    value={deleteReason}
                    onChange={(e) => setDeleteReason(e.target.value)}
                    placeholder="Provide a reason for deleting this conversation..."
                  />
                </div>

                <div className="flex items-center space-x-2">
                  <Switch
                    id="deleteMessages"
                    checked={deleteMessages}
                    onCheckedChange={setDeleteMessages}
                  />
                  <Label htmlFor="deleteMessages" className="text-sm">
                    Also delete all messages in this conversation
                  </Label>
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
                onClick={() => closeDialog("deleteConversation")}
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDeleteConversation}
                disabled={formLoading}
              >
                {formLoading && (
                  <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Delete Conversation
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </SidebarInset>
    </SidebarProvider>
  );
}

export default withAuth(ConversationsPage);
