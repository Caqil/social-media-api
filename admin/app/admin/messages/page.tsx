// admin/app/admin/messages/page.tsx - Fixed Messages Management Page
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
  IconPhoto,
} from "@tabler/icons-react";

// Types
interface Message {
  id: string;
  conversation_id: string;
  sender_id: string;
  content?: string;
  content_type: "text" | "image" | "video" | "audio" | "file" | "location";
  media_url?: string;
  file_name?: string;
  file_size?: number;
  is_read: boolean;
  read_at?: string;
  is_edited: boolean;
  edited_at?: string;
  reply_to_id?: string;
  created_at: string;
  updated_at: string;
  sender?: {
    id: string;
    username: string;
    email: string;
    first_name?: string;
    last_name?: string;
    profile_picture?: string;
    is_verified: boolean;
  };
  conversation?: {
    id: string;
    type: "direct" | "group";
    title?: string;
    participant_count?: number;
  };
}

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
  last_message?: Message;
}

interface MessagesPageState {
  messages: Message[];
  conversations: Conversation[];
  loading: boolean;
  conversationsLoading: boolean;
  error: string | null;
  conversationsError: string | null;
  pagination: PaginationMeta | undefined;
  filters: {
    search: string;
    conversation_id: string;
    content_type: string;
    is_read: string;
    date_from: string;
    date_to: string;
    page: number;
    limit: number;
    sort_by?: string;
    sort_order?: "asc" | "desc";
  };
  selectedMessages: string[];
  activeTab: "messages" | "conversations";
}

interface DialogState {
  viewMessage: boolean;
  viewConversation: boolean;
  deleteMessage: boolean;
  deleteConversation: boolean;
  selectedMessage: Message | null;
  selectedConversation: Conversation | null;
}

const initialFilters = {
  search: "",
  conversation_id: "all",
  content_type: "all",
  is_read: "all",
  date_from: "",
  date_to: "",
  page: 1,
  limit: 20,
};

// Utility function to validate ObjectID format
const isValidObjectId = (id: string): boolean => {
  return /^[0-9a-fA-F]{24}$/.test(id);
};

// Utility function to safely get conversation ID
const getSafeConversationId = (conversationId: string): string | null => {
  if (
    !conversationId ||
    conversationId === "all" ||
    conversationId === "undefined"
  ) {
    return null;
  }

  if (!isValidObjectId(conversationId)) {
    console.warn(`Invalid conversation ID format: ${conversationId}`);
    return null;
  }

  return conversationId;
};

function MessagesPage() {
  const { user: currentUser } = useAuth();

  const [state, setState] = useState<MessagesPageState>({
    messages: [],
    conversations: [],
    loading: true,
    conversationsLoading: false,
    error: null,
    conversationsError: null,
    pagination: undefined,
    filters: initialFilters,
    selectedMessages: [],
    activeTab: "messages",
  });

  const [dialogs, setDialogs] = useState<DialogState>({
    viewMessage: false,
    viewConversation: false,
    deleteMessage: false,
    deleteConversation: false,
    selectedMessage: null,
    selectedConversation: null,
  });

  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [deleteReason, setDeleteReason] = useState("");

  // Fetch messages with improved error handling
  const fetchMessages = useCallback(async (filters = state.filters) => {
    try {
      setState((prev) => ({ ...prev, loading: true, error: null }));

      const params: any = {
        page: filters.page,
        limit: filters.limit,
      };

      // Add search parameter
      if (filters.search) params.search = filters.search;

      // Add conversation filter with validation
      const safeConversationId = getSafeConversationId(filters.conversation_id);
      if (safeConversationId) {
        params.conversation_id = safeConversationId;
      }

      // Add content type filter
      if (filters.content_type && filters.content_type !== "all") {
        params.content_type = filters.content_type;
      }

      // Add read status filter
      if (filters.is_read && filters.is_read !== "all") {
        params.is_read = filters.is_read;
      }

      // Add date range filters
      if (filters.date_from) params.date_from = filters.date_from;
      if (filters.date_to) params.date_to = filters.date_to;

      // Add sorting
      if (filters.sort_by) {
        params.sort_by = filters.sort_by;
        params.sort_order = filters.sort_order || "desc";
      }

      console.log("📡 Fetching messages with params:", params);
      const response = await apiClient.getMessages(params);

      // Enhanced data validation
      let validMessages: Message[] = [];
      if (response.data && Array.isArray(response.data)) {
        validMessages = response.data.filter((message: any) => {
          if (!message || !message.id || typeof message.id !== "string") {
            console.warn("Invalid message object:", message);
            return false;
          }
          return true;
        });
      } else if (response.success && response.data) {
        // Handle single object response
        console.warn("Expected array but got single object:", response.data);
      }

      setState((prev) => ({
        ...prev,
        messages: validMessages,
        pagination: response.pagination || undefined,
        loading: false,
      }));
    } catch (error: any) {
      console.error("❌ Failed to fetch messages:", error);
      setState((prev) => ({
        ...prev,
        loading: false,
        messages: [],
        error:
          error.message ||
          "Failed to fetch messages. Please check your filters and try again.",
      }));
    }
  }, []);

  // Fetch conversations with improved error handling
  const fetchConversations = useCallback(async () => {
    try {
      setState((prev) => ({
        ...prev,
        conversationsLoading: true,
        conversationsError: null,
      }));

      console.log("📡 Fetching conversations...");
      const response = await apiClient.getConversations({
        limit: 100,
      });

      // Enhanced data validation for conversations
      let validConversations: Conversation[] = [];
      if (response.data && Array.isArray(response.data)) {
        validConversations = response.data.filter((conv: any) => {
          if (!conv || !conv.id || typeof conv.id !== "string") {
            console.warn("Invalid conversation object:", conv);
            return false;
          }

          // Validate ObjectID format
          if (!isValidObjectId(conv.id)) {
            console.warn(`Invalid conversation ID format: ${conv.id}`);
            return false;
          }

          return true;
        });
      }

      setState((prev) => ({
        ...prev,
        conversations: validConversations,
        conversationsLoading: false,
      }));
    } catch (error: any) {
      console.error("❌ Failed to fetch conversations:", error);
      setState((prev) => ({
        ...prev,
        conversations: [],
        conversationsLoading: false,
        conversationsError: error.message || "Failed to fetch conversations",
      }));
    }
  }, []);

  // Initial load with better error handling
  useEffect(() => {
    const loadData = async () => {
      try {
        // Load conversations first (they're needed for filters)
        await fetchConversations();
        // Then load messages
        await fetchMessages();
      } catch (error) {
        console.error("Failed to load initial data:", error);
      }
    };

    loadData();
  }, []);

  // Handle search with debounce
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (state.filters.search !== undefined) {
        fetchMessages({ ...state.filters, page: 1 });
      }
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [state.filters.search, fetchMessages]);

  // Handle filter changes
  const handleFilterChange = (key: string, value: any) => {
    console.log(`🔄 Filter change: ${key} = ${value}`);

    // Validate conversation ID if that's what's being changed
    if (key === "conversation_id" && value !== "all") {
      const safeId = getSafeConversationId(value);
      if (!safeId && value !== "all") {
        console.warn(`Invalid conversation ID format: ${value}`);
        return; // Don't update filter with invalid ID
      }
    }

    const newFilters = { ...state.filters, [key]: value, page: 1 };
    setState((prev) => ({ ...prev, filters: newFilters }));

    if (key !== "search") {
      fetchMessages(newFilters);
    }
  };

  // Handle pagination
  const handlePageChange = (page: number) => {
    const newFilters = { ...state.filters, page };
    setState((prev) => ({ ...prev, filters: newFilters }));
    fetchMessages(newFilters);
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
    fetchMessages(newFilters);
  };

  // Handle row selection
  const handleRowSelect = (selectedRows: string[]) => {
    setState((prev) => ({ ...prev, selectedMessages: selectedRows }));
  };

  // Open dialog
  const openDialog = (
    type: keyof DialogState,
    item: Message | Conversation | null = null
  ) => {
    if (type === "viewMessage" || type === "deleteMessage") {
      setDialogs((prev) => ({
        ...prev,
        [type]: true,
        selectedMessage: item as Message,
      }));
    } else if (type === "viewConversation" || type === "deleteConversation") {
      setDialogs((prev) => ({
        ...prev,
        [type]: true,
        selectedConversation: item as Conversation,
      }));
    }
  };

  // Close dialog
  const closeDialog = (type: keyof DialogState) => {
    setDialogs((prev) => ({
      ...prev,
      [type]: false,
      selectedMessage: null,
      selectedConversation: null,
    }));
    setFormError(null);
    setDeleteReason("");
  };

  // Handle message deletion
  const handleDeleteMessage = async () => {
    if (!dialogs.selectedMessage) return;

    setFormLoading(true);
    try {
      await apiClient.deleteMessage(
        dialogs.selectedMessage.id,
        deleteReason || "Deleted by admin"
      );
      closeDialog("deleteMessage");
      fetchMessages();
    } catch (error: any) {
      setFormError(error.message || "Failed to delete message");
    } finally {
      setFormLoading(false);
    }
  };

  // Handle bulk actions
  const handleBulkAction = async (action: string, selectedIds: string[]) => {
    setFormLoading(true);
    try {
      await apiClient.bulkMessageAction({
        message_ids: selectedIds,
        action,
        reason: deleteReason || "Bulk action by admin",
      });
      setState((prev) => ({ ...prev, selectedMessages: [] }));
      fetchMessages();
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
    if (state.activeTab === "messages") {
      fetchMessages();
    } else {
      fetchConversations();
    }
  };

  // Handle export
  const handleExport = async () => {
    try {
      const csvContent = state.messages.map((message) => ({
        id: message.id,
        sender: message.sender?.username || "Unknown",
        content: message.content || "",
        type: message.content_type,
        conversation_type: message.conversation?.type || "",
        created_at: message.created_at,
        is_read: message.is_read,
      }));

      console.log("✅ Export data prepared", csvContent);
    } catch (error) {
      console.error("❌ Export failed:", error);
    }
  };

  // Format content type
  const formatContentType = (type: string) => {
    const typeMap: Record<string, string> = {
      text: "Text",
      image: "Image",
      video: "Video",
      audio: "Audio",
      file: "File",
      location: "Location",
    };
    return typeMap[type] || type;
  };

  // Get content type icon
  const getContentTypeIcon = (type: string) => {
    switch (type) {
      case "image":
        return <IconPhoto className="h-4 w-4" />;
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

  // Get content type badge variant
  const getContentTypeBadgeVariant = (
    type: string
  ): "default" | "secondary" | "destructive" => {
    switch (type) {
      case "image":
      case "video":
        return "default";
      case "audio":
      case "file":
        return "secondary";
      default:
        return "secondary";
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

  // Format conversation title with safety checks
  const formatConversationTitle = (conversation: Conversation) => {
    if (!conversation) return "Unknown Conversation";

    if (conversation.title) return conversation.title;
    if (conversation.type === "group")
      return `Group Chat (${
        conversation.participant_ids?.length || 0
      } members)`;
    return "Direct Message";
  };

  // Messages table columns configuration
  const messageColumns: TableColumn[] = [
    {
      key: "sender",
      label: "Sender",
      sortable: true,
      render: (_, message: Message) => (
        <div className="flex items-center gap-2">
          <Avatar className="h-8 w-8">
            <AvatarImage
              src={message.sender?.profile_picture}
              alt={message.sender?.username || "User"}
            />
            <AvatarFallback>
              {(
                message.sender?.first_name?.[0] ||
                message.sender?.username?.[0] ||
                "U"
              ).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="flex flex-col">
            <span className="font-medium">
              {message.sender?.username || "Unknown User"}
            </span>
            <span className="text-sm text-muted-foreground">
              {message.sender?.first_name} {message.sender?.last_name}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "content",
      label: "Content",
      render: (_, message: Message) => (
        <div className="flex items-center gap-2 max-w-xs">
          {getContentTypeIcon(message.content_type)}
          <div className="flex flex-col">
            <span className="text-sm">
              {message.content_type === "text"
                ? truncateContent(message.content)
                : message.file_name || formatContentType(message.content_type)}
            </span>
            {message.file_size && (
              <span className="text-xs text-muted-foreground">
                {(message.file_size / 1024 / 1024).toFixed(2)} MB
              </span>
            )}
          </div>
        </div>
      ),
    },
    {
      key: "content_type",
      label: "Type",
      filterable: true,
      render: (value: string) => (
        <Badge variant={getContentTypeBadgeVariant(value)}>
          {formatContentType(value)}
        </Badge>
      ),
    },
    {
      key: "conversation",
      label: "Conversation",
      render: (_, message: Message) => (
        <div className="flex items-center gap-2">
          <IconMessageCircle className="h-4 w-4 text-muted-foreground" />
          <div className="flex flex-col">
            <span className="text-sm font-medium">
              {message.conversation?.type === "group"
                ? "Group Chat"
                : "Direct Message"}
            </span>
            {message.conversation?.title && (
              <span className="text-xs text-muted-foreground">
                {truncateContent(message.conversation.title, 30)}
              </span>
            )}
          </div>
        </div>
      ),
    },
    {
      key: "is_read",
      label: "Status",
      sortable: true,
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
      sortable: true,
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
    {
      key: "actions",
      label: "",
      width: "w-12",
      render: (_, message: Message) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
              <IconDotsVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              onClick={() => openDialog("viewMessage", message)}
            >
              <IconEye className="h-4 w-4 mr-2" />
              View Details
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => openDialog("deleteMessage", message)}
              className="text-red-600"
            >
              <IconTrash className="h-4 w-4 mr-2" />
              Delete Message
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

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
              <span className="text-sm">
                {truncateContent(conversation.last_message.content)}
              </span>
              <span className="text-xs text-muted-foreground">
                {conversation.last_message.sender?.username || "Unknown"}
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
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => openDialog("deleteConversation", conversation)}
              className="text-red-600"
            >
              <IconTrash className="h-4 w-4 mr-2" />
              Delete Conversation
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];

  // Bulk actions configuration
  const bulkActions = [
    {
      label: "Delete Messages",
      action: "delete",
      variant: "destructive" as const,
    },
    { label: "Mark as Read", action: "mark_read" },
    { label: "Mark as Unread", action: "mark_unread" },
  ];

  // Error display component
  if ((state.error || state.conversationsError) && !state.loading) {
    return (
      <SidebarProvider>
        <AppSidebar variant="inset" />
        <SidebarInset>
          <SiteHeader />
          <div className="flex h-screen items-center justify-center p-6">
            <Alert variant="destructive" className="max-w-md">
              <IconAlertCircle className="h-5 w-5" />
              <AlertDescription className="space-y-4">
                <div>{state.error || state.conversationsError}</div>
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
              <h1 className="text-3xl font-bold">Messages</h1>
              <p className="text-muted-foreground">
                Manage messages and conversations
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

          {/* Show loading state */}
          {state.conversationsLoading && (
            <div className="flex items-center gap-2 text-muted-foreground">
              <IconLoader2 className="h-4 w-4 animate-spin" />
              <span>Loading conversations...</span>
            </div>
          )}

          {/* Show conversations error */}
          {state.conversationsError && (
            <Alert variant="destructive">
              <IconAlertCircle className="h-4 w-4" />
              <AlertDescription>{state.conversationsError}</AlertDescription>
            </Alert>
          )}

          {/* Tabs */}
          <Tabs
            value={state.activeTab}
            onValueChange={(value) =>
              setState((prev) => ({
                ...prev,
                activeTab: value as "messages" | "conversations",
              }))
            }
            className="w-full"
          >
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="messages">Messages</TabsTrigger>
              <TabsTrigger value="conversations">Conversations</TabsTrigger>
            </TabsList>

            <TabsContent value="messages" className="space-y-4">
              {/* Filters */}
              <Card>
                <CardHeader>
                  <CardTitle>Filters</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
                    <div>
                      <Label htmlFor="search">Search</Label>
                      <div className="relative">
                        <IconSearch className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                        <Input
                          id="search"
                          placeholder="Search messages..."
                          value={state.filters.search}
                          onChange={(e) =>
                            handleFilterChange("search", e.target.value)
                          }
                          className="pl-9"
                        />
                      </div>
                    </div>
                    <div>
                      <Label htmlFor="conversation">Conversation</Label>
                      <Select
                        value={state.filters.conversation_id}
                        onValueChange={(value) =>
                          handleFilterChange("conversation_id", value)
                        }
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="All conversations" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="all">All conversations</SelectItem>
                          {state.conversations
                            .filter(
                              (conversation) =>
                                conversation &&
                                conversation.id &&
                                isValidObjectId(conversation.id)
                            )
                            .map((conversation, index) => (
                              <SelectItem
                                key={`conversation-${conversation.id || index}`}
                                value={conversation.id}
                              >
                                {formatConversationTitle(conversation)}
                              </SelectItem>
                            ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <div>
                      <Label htmlFor="content_type">Content Type</Label>
                      <Select
                        value={state.filters.content_type}
                        onValueChange={(value) =>
                          handleFilterChange("content_type", value)
                        }
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="All types" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="all">All types</SelectItem>
                          <SelectItem value="text">Text</SelectItem>
                          <SelectItem value="image">Image</SelectItem>
                          <SelectItem value="video">Video</SelectItem>
                          <SelectItem value="audio">Audio</SelectItem>
                          <SelectItem value="file">File</SelectItem>
                          <SelectItem value="location">Location</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div>
                      <Label htmlFor="is_read">Read Status</Label>
                      <Select
                        value={state.filters.is_read}
                        onValueChange={(value) =>
                          handleFilterChange("is_read", value)
                        }
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="All messages" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="all">All messages</SelectItem>
                          <SelectItem value="true">Read</SelectItem>
                          <SelectItem value="false">Unread</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div>
                      <Label htmlFor="date_from">Date From</Label>
                      <Input
                        id="date_from"
                        type="date"
                        value={state.filters.date_from}
                        onChange={(e) =>
                          handleFilterChange("date_from", e.target.value)
                        }
                      />
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Messages Data Table */}
              <DataTable
                data={state.messages}
                columns={messageColumns}
                loading={state.loading}
                pagination={state.pagination}
                onPageChange={handlePageChange}
                onSort={handleSort}
                onRowSelect={handleRowSelect}
                bulkActions={bulkActions}
                onBulkAction={handleBulkAction}
                onRefresh={handleRefresh}
                onExport={handleExport}
                title="Message Management"
                description="View and manage all messages"
                emptyMessage="No messages found"
                searchPlaceholder="Search messages..."
              />
            </TabsContent>

            <TabsContent value="conversations" className="space-y-4">
              {/* Conversations Data Table */}
              <DataTable
                data={state.conversations}
                columns={conversationColumns}
                loading={state.conversationsLoading}
                onRefresh={() => fetchConversations()}
                title="Conversation Management"
                description="View and manage all conversations"
                emptyMessage="No conversations found"
                searchPlaceholder="Search conversations..."
              />
            </TabsContent>
          </Tabs>
        </div>

        {/* Dialogs remain the same as in original code... */}
        {/* View Message Dialog */}
        <Dialog
          open={dialogs.viewMessage}
          onOpenChange={() => closeDialog("viewMessage")}
        >
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Message Details</DialogTitle>
              <DialogDescription>
                View detailed information about this message
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedMessage && (
              <div className="space-y-6">
                <div className="flex items-center space-x-4">
                  <Avatar className="h-12 w-12">
                    <AvatarImage
                      src={dialogs.selectedMessage.sender?.profile_picture}
                    />
                    <AvatarFallback>
                      {(
                        dialogs.selectedMessage.sender?.first_name?.[0] ||
                        dialogs.selectedMessage.sender?.username?.[0] ||
                        "U"
                      ).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <h3 className="text-lg font-semibold">
                      {dialogs.selectedMessage.sender?.first_name}{" "}
                      {dialogs.selectedMessage.sender?.last_name}
                    </h3>
                    <p className="text-muted-foreground">
                      @{dialogs.selectedMessage.sender?.username || "unknown"}
                    </p>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge
                        variant={
                          dialogs.selectedMessage.is_read
                            ? "default"
                            : "secondary"
                        }
                      >
                        {dialogs.selectedMessage.is_read ? "Read" : "Unread"}
                      </Badge>
                      {dialogs.selectedMessage.is_edited && (
                        <Badge variant="outline">Edited</Badge>
                      )}
                    </div>
                  </div>
                </div>

                <Separator />

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label className="text-sm font-medium">Content Type</Label>
                    <div className="flex items-center gap-2">
                      {getContentTypeIcon(dialogs.selectedMessage.content_type)}
                      <span className="text-sm">
                        {formatContentType(
                          dialogs.selectedMessage.content_type
                        )}
                      </span>
                    </div>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Conversation</Label>
                    <p className="text-sm text-muted-foreground">
                      {dialogs.selectedMessage.conversation?.type === "group"
                        ? "Group Chat"
                        : "Direct Message"}
                    </p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Sent</Label>
                    <p className="text-sm text-muted-foreground">
                      {new Date(
                        dialogs.selectedMessage.created_at
                      ).toLocaleString()}
                    </p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Read At</Label>
                    <p className="text-sm text-muted-foreground">
                      {dialogs.selectedMessage.read_at
                        ? new Date(
                            dialogs.selectedMessage.read_at
                          ).toLocaleString()
                        : "Not read"}
                    </p>
                  </div>
                </div>

                {dialogs.selectedMessage.content && (
                  <div>
                    <Label className="text-sm font-medium">Content</Label>
                    <div className="mt-2 p-3 bg-muted rounded-lg">
                      <p className="text-sm whitespace-pre-wrap">
                        {dialogs.selectedMessage.content}
                      </p>
                    </div>
                  </div>
                )}

                {dialogs.selectedMessage.media_url && (
                  <div>
                    <Label className="text-sm font-medium">Media</Label>
                    <div className="mt-2 p-3 bg-muted rounded-lg">
                      <div className="flex items-center gap-2">
                        <IconPaperclip className="h-4 w-4" />
                        <span className="text-sm">
                          {dialogs.selectedMessage.file_name || "Media file"}
                        </span>
                        {dialogs.selectedMessage.file_size && (
                          <span className="text-xs text-muted-foreground">
                            (
                            {(
                              dialogs.selectedMessage.file_size /
                              1024 /
                              1024
                            ).toFixed(2)}{" "}
                            MB)
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                )}

                {dialogs.selectedMessage.is_edited && (
                  <div>
                    <Label className="text-sm font-medium">Edit History</Label>
                    <p className="text-sm text-muted-foreground mt-1">
                      Last edited:{" "}
                      {dialogs.selectedMessage.edited_at
                        ? new Date(
                            dialogs.selectedMessage.edited_at
                          ).toLocaleString()
                        : "Unknown"}
                    </p>
                  </div>
                )}
              </div>
            )}

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => closeDialog("viewMessage")}
              >
                Close
              </Button>
              <Button
                variant="destructive"
                onClick={() => {
                  closeDialog("viewMessage");
                  if (dialogs.selectedMessage)
                    openDialog("deleteMessage", dialogs.selectedMessage);
                }}
              >
                <IconTrash className="h-4 w-4 mr-2" />
                Delete Message
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Delete Message Dialog */}
        <Dialog
          open={dialogs.deleteMessage}
          onOpenChange={() => closeDialog("deleteMessage")}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete Message</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this message? This action cannot
                be undone.
              </DialogDescription>
            </DialogHeader>

            {dialogs.selectedMessage && (
              <div className="space-y-4">
                <div className="p-4 bg-muted rounded-lg">
                  <div className="flex items-center gap-3 mb-2">
                    <Avatar className="h-8 w-8">
                      <AvatarImage
                        src={dialogs.selectedMessage.sender?.profile_picture}
                      />
                      <AvatarFallback>
                        {(
                          dialogs.selectedMessage.sender?.first_name?.[0] ||
                          dialogs.selectedMessage.sender?.username?.[0] ||
                          "U"
                        ).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="font-medium">
                        {dialogs.selectedMessage.sender?.username || "Unknown"}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        {new Date(
                          dialogs.selectedMessage.created_at
                        ).toLocaleString()}
                      </p>
                    </div>
                  </div>
                  <p className="text-sm">
                    {dialogs.selectedMessage.content ||
                      `${formatContentType(
                        dialogs.selectedMessage.content_type
                      )} message`}
                  </p>
                </div>

                <div>
                  <Label htmlFor="deleteReason">
                    Reason for deletion (optional)
                  </Label>
                  <Textarea
                    id="deleteReason"
                    value={deleteReason}
                    onChange={(e) => setDeleteReason(e.target.value)}
                    placeholder="Provide a reason for deleting this message..."
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
                onClick={() => closeDialog("deleteMessage")}
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDeleteMessage}
                disabled={formLoading}
              >
                {formLoading && (
                  <IconLoader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Delete Message
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </SidebarInset>
    </SidebarProvider>
  );
}

export default withAuth(MessagesPage);
