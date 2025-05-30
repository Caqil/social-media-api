// app/admin/messages/page.tsx
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
import { Message, Conversation, TableColumn } from "@/types/admin";
import {
  IconMessage,
  IconEye,
  IconTrash,
  IconUsers,
  IconUser,
  IconPhoto,
  IconVideo,
  IconMusic,
  IconFile,
  IconMapPin,
  IconClock,
  IconCheck,
  IconX,
  IconEdit,
  IconArchive,
  IconVolume,
  IconVolumeX,
} from "@tabler/icons-react";

function MessagesPage() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<any>(null);
  const [conversationsPagination, setConversationsPagination] =
    useState<any>(null);
  const [filters, setFilters] = useState<any>({ page: 1, limit: 20 });
  const [conversationFilters, setConversationFilters] = useState<any>({
    page: 1,
    limit: 20,
  });

  // Dialog states
  const [selectedMessage, setSelectedMessage] = useState<Message | null>(null);
  const [selectedConversation, setSelectedConversation] =
    useState<Conversation | null>(null);
  const [showMessageDetails, setShowMessageDetails] = useState(false);
  const [showConversationDetails, setShowConversationDetails] = useState(false);
  const [showActionDialog, setShowActionDialog] = useState(false);
  const [showBulkDialog, setShowBulkDialog] = useState(false);
  const [actionType, setActionType] = useState<string>("");
  const [bulkAction, setBulkAction] = useState<string>("");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [actionReason, setActionReason] = useState("");

  // Active tab
  const [activeTab, setActiveTab] = useState("messages");

  // Fetch messages
  const fetchMessages = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getMessages(filters);
      setMessages(response.data.data);
      setPagination(response.data.pagination);
    } catch (error: any) {
      console.error("Failed to fetch messages:", error);
      setError(error.response?.data?.message || "Failed to load messages");
    } finally {
      setLoading(false);
    }
  };

  // Fetch conversations
  const fetchConversations = async () => {
    try {
      setLoading(true);
      const response = await apiClient.getConversations(conversationFilters);
      setConversations(response.data.data);
      setConversationsPagination(response.data.pagination);
    } catch (error: any) {
      console.error("Failed to fetch conversations:", error);
      setError(error.response?.data?.message || "Failed to load conversations");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === "messages") {
      fetchMessages();
    } else if (activeTab === "conversations") {
      fetchConversations();
    }
  }, [filters, conversationFilters, activeTab]);

  // Handle page change
  const handlePageChange = (page: number) => {
    if (activeTab === "messages") {
      setFilters((prev: any) => ({ ...prev, page }));
    } else {
      setConversationFilters((prev: any) => ({ ...prev, page }));
    }
  };

  // Handle filtering
  const handleFilter = (newFilters: Record<string, any>) => {
    if (activeTab === "messages") {
      setFilters((prev: any) => ({
        ...prev,
        ...newFilters,
        page: 1,
      }));
    } else {
      setConversationFilters((prev: any) => ({
        ...prev,
        ...newFilters,
        page: 1,
      }));
    }
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
      await apiClient.bulkMessageAction({
        message_ids: selectedIds,
        action: bulkAction,
        reason: actionReason,
      });

      setShowBulkDialog(false);
      setActionReason("");
      if (activeTab === "messages") {
        fetchMessages();
      } else {
        fetchConversations();
      }
    } catch (error: any) {
      console.error("Bulk action failed:", error);
      setError(error.response?.data?.message || "Bulk action failed");
    }
  };

  // Handle individual message action
  const handleMessageAction = (message: Message, action: string) => {
    setSelectedMessage(message);
    setActionType(action);
    setShowActionDialog(true);
  };

  // Execute message action
  const executeMessageAction = async () => {
    if (!selectedMessage) return;

    try {
      switch (actionType) {
        case "delete":
          await apiClient.deleteMessage(selectedMessage.id, actionReason);
          break;
      }

      setShowActionDialog(false);
      setActionReason("");
      if (activeTab === "messages") {
        fetchMessages();
      } else {
        fetchConversations();
      }
    } catch (error: any) {
      console.error("Message action failed:", error);
      setError(error.response?.data?.message || "Message action failed");
    }
  };

  // View message details
  const viewMessageDetails = (message: Message) => {
    setSelectedMessage(message);
    setShowMessageDetails(true);
  };

  // View conversation details
  const viewConversationDetails = async (conversation: Conversation) => {
    try {
      const response = await apiClient.getConversation(conversation.id);
      setSelectedConversation(response.data);
      setShowConversationDetails(true);
    } catch (error) {
      console.error("Failed to fetch conversation details:", error);
    }
  };

  // Get message type icon
  const getMessageTypeIcon = (type: string) => {
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
        return <IconMapPin className="h-4 w-4" />;
      default:
        return <IconMessage className="h-4 w-4" />;
    }
  };

  // Format file size
  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  // Messages table columns
  const messageColumns: TableColumn[] = [
    {
      key: "message",
      label: "Message",
      render: (_, message: Message) => (
        <div className="max-w-md">
          <div className="flex items-center gap-2 mb-2">
            {getMessageTypeIcon(message.message_type)}
            <Badge variant="outline" className="text-xs">
              {message.message_type}
            </Badge>
            {message.is_read && (
              <Badge variant="secondary" className="text-xs">
                Read
              </Badge>
            )}
            {message.is_edited && (
              <Badge variant="outline" className="text-xs">
                Edited
              </Badge>
            )}
          </div>

          {message.content && (
            <p className="text-sm line-clamp-2 mb-2">{message.content}</p>
          )}

          {message.file_name && (
            <div className="text-xs text-muted-foreground mb-2">
              {message.file_name} ({formatFileSize(message.file_size || 0)})
            </div>
          )}

          {message.sender && (
            <div className="flex items-center gap-2">
              <Avatar className="h-4 w-4">
                <AvatarImage src={message.sender.profile_picture} />
                <AvatarFallback className="text-xs">
                  {message.sender.username?.[0]?.toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <span className="text-xs text-muted-foreground">
                from {message.sender.username}
              </span>
              {message.recipient && (
                <>
                  <span className="text-xs text-muted-foreground">to</span>
                  <span className="text-xs text-muted-foreground">
                    {message.recipient.username}
                  </span>
                </>
              )}
            </div>
          )}
        </div>
      ),
    },
    {
      key: "conversation_type",
      label: "Type",
      render: (_, message: Message) => (
        <Badge variant="outline">
          {message.conversation?.type === "group" ? "Group" : "Direct"}
        </Badge>
      ),
    },
    {
      key: "status",
      label: "Status",
      render: (_, message: Message) => (
        <div className="space-y-1">
          {message.is_read ? (
            <Badge variant="default">
              <IconCheck className="h-3 w-3 mr-1" />
              Read
            </Badge>
          ) : (
            <Badge variant="secondary">
              <IconClock className="h-3 w-3 mr-1" />
              Unread
            </Badge>
          )}
          {message.is_edited && (
            <div>
              <Badge variant="outline" className="text-xs">
                <IconEdit className="h-3 w-3 mr-1" />
                Edited
              </Badge>
            </div>
          )}
        </div>
      ),
    },
    {
      key: "created_at",
      label: "Sent",
      sortable: true,
      render: (value: string) => (
        <div className="text-xs">{new Date(value).toLocaleString()}</div>
      ),
    },
    {
      key: "actions",
      label: "Actions",
      render: (_, message: Message) => (
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewMessageDetails(message)}
          >
            <IconEye className="h-3 w-3" />
          </Button>
          <Button
            size="sm"
            variant="destructive"
            onClick={() => handleMessageAction(message, "delete")}
          >
            <IconTrash className="h-3 w-3" />
          </Button>
        </div>
      ),
    },
  ];

  // Conversations table columns
  const conversationColumns: TableColumn[] = [
    {
      key: "conversation",
      label: "Conversation",
      render: (_, conversation: Conversation) => (
        <div className="flex items-center gap-3">
          <Avatar className="h-10 w-10">
            <AvatarImage src={conversation.avatar} />
            <AvatarFallback>
              {conversation.type === "group" ? (
                <IconUsers className="h-4 w-4" />
              ) : (
                <IconUser className="h-4 w-4" />
              )}
            </AvatarFallback>
          </Avatar>
          <div>
            <div className="font-medium">
              {conversation.title || `${conversation.type} conversation`}
            </div>
            <div className="text-sm text-muted-foreground">
              {conversation.participants?.length || 0} participant(s)
            </div>
            {conversation.last_message && (
              <div className="text-xs text-muted-foreground line-clamp-1">
                Last:{" "}
                {conversation.last_message.content ||
                  `${conversation.last_message.message_type} message`}
              </div>
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
        <Badge variant="outline" className="flex items-center gap-1 w-fit">
          {value === "group" ? (
            <IconUsers className="h-3 w-3" />
          ) : (
            <IconUser className="h-3 w-3" />
          )}
          <span className="capitalize">{value}</span>
        </Badge>
      ),
    },
    {
      key: "unread_count",
      label: "Unread",
      render: (value: number) =>
        value > 0 ? (
          <Badge variant="destructive">{value}</Badge>
        ) : (
          <Badge variant="outline">0</Badge>
        ),
    },
    {
      key: "status",
      label: "Status",
      render: (_, conversation: Conversation) => (
        <div className="space-y-1">
          {conversation.is_archived && (
            <Badge variant="secondary">
              <IconArchive className="h-3 w-3 mr-1" />
              Archived
            </Badge>
          )}
          {conversation.is_muted && (
            <Badge variant="outline">
              <IconVolumeX className="h-3 w-3 mr-1" />
              Muted
            </Badge>
          )}
          {!conversation.is_archived && !conversation.is_muted && (
            <Badge variant="default">
              <IconVolume className="h-3 w-3 mr-1" />
              Active
            </Badge>
          )}
        </div>
      ),
    },
    {
      key: "last_message_at",
      label: "Last Activity",
      sortable: true,
      render: (value: string) =>
        value ? (
          <div className="text-xs">{new Date(value).toLocaleString()}</div>
        ) : (
          <span className="text-muted-foreground">No messages</span>
        ),
    },
    {
      key: "actions",
      label: "Actions",
      render: (_, conversation: Conversation) => (
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="outline"
            onClick={() => viewConversationDetails(conversation)}
          >
            <IconEye className="h-3 w-3" />
          </Button>
        </div>
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
              <h1 className="text-2xl font-bold">Messages Management</h1>
              <p className="text-muted-foreground">
                Manage user messages and conversations
              </p>
            </div>
          </div>

          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList>
              <TabsTrigger value="messages">Messages</TabsTrigger>
              <TabsTrigger value="conversations">Conversations</TabsTrigger>
            </TabsList>

            <TabsContent value="messages" className="space-y-4">
              <DataTable
                title="Messages"
                description={`Manage ${pagination?.total || 0} messages`}
                data={messages}
                columns={messageColumns}
                loading={loading}
                pagination={pagination}
                searchPlaceholder="Search messages by content..."
                onPageChange={handlePageChange}
                onFilter={handleFilter}
                onRowSelect={handleRowSelect}
                bulkActions={bulkActions}
                onBulkAction={handleBulkAction}
                onRefresh={fetchMessages}
                onExport={() => console.log("Export messages")}
              />
            </TabsContent>

            <TabsContent value="conversations" className="space-y-4">
              <DataTable
                title="Conversations"
                description={`Manage ${
                  conversationsPagination?.total || 0
                } conversations`}
                data={conversations}
                columns={conversationColumns}
                loading={loading}
                pagination={conversationsPagination}
                searchPlaceholder="Search conversations..."
                onPageChange={handlePageChange}
                onFilter={handleFilter}
                onRefresh={fetchConversations}
                onExport={() => console.log("Export conversations")}
              />
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>

      {/* Message Details Dialog */}
      <Dialog open={showMessageDetails} onOpenChange={setShowMessageDetails}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Message Details</DialogTitle>
          </DialogHeader>

          {selectedMessage && (
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-3">
                    {selectedMessage.sender && (
                      <>
                        <Avatar>
                          <AvatarImage
                            src={selectedMessage.sender.profile_picture}
                          />
                          <AvatarFallback>
                            {selectedMessage.sender.username?.[0]?.toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <p className="font-medium">
                            {selectedMessage.sender.username}
                          </p>
                          <p className="text-sm text-muted-foreground">
                            {new Date(
                              selectedMessage.created_at
                            ).toLocaleString()}
                          </p>
                        </div>
                      </>
                    )}
                    <div className="ml-auto flex items-center gap-2">
                      {getMessageTypeIcon(selectedMessage.message_type)}
                      <Badge variant="outline">
                        {selectedMessage.message_type}
                      </Badge>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  {selectedMessage.content && (
                    <div className="mb-4">
                      <h4 className="font-medium mb-2">Content</h4>
                      <p className="text-sm bg-gray-50 p-3 rounded">
                        {selectedMessage.content}
                      </p>
                    </div>
                  )}

                  {selectedMessage.media_url && (
                    <div className="mb-4">
                      <h4 className="font-medium mb-2">Media</h4>
                      {selectedMessage.message_type === "image" ? (
                        <img
                          src={selectedMessage.media_url}
                          alt="Message media"
                          className="rounded-lg max-h-64 object-contain"
                        />
                      ) : selectedMessage.message_type === "video" ? (
                        <video
                          src={selectedMessage.media_url}
                          controls
                          className="rounded-lg max-h-64"
                        />
                      ) : selectedMessage.message_type === "audio" ? (
                        <audio src={selectedMessage.media_url} controls />
                      ) : (
                        <div className="flex items-center gap-2 p-3 border rounded">
                          <IconFile className="h-4 w-4" />
                          <span>{selectedMessage.file_name}</span>
                          <Badge variant="outline">
                            {formatFileSize(selectedMessage.file_size || 0)}
                          </Badge>
                        </div>
                      )}
                    </div>
                  )}

                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <label className="font-medium">Message Type</label>
                      <p className="capitalize">
                        {selectedMessage.message_type}
                      </p>
                    </div>
                    <div>
                      <label className="font-medium">Status</label>
                      <p>{selectedMessage.is_read ? "Read" : "Unread"}</p>
                    </div>
                    {selectedMessage.is_edited && (
                      <div>
                        <label className="font-medium">Edited</label>
                        <p>
                          {new Date(
                            selectedMessage.edited_at!
                          ).toLocaleString()}
                        </p>
                      </div>
                    )}
                    {selectedMessage.read_at && (
                      <div>
                        <label className="font-medium">Read At</label>
                        <p>
                          {new Date(selectedMessage.read_at).toLocaleString()}
                        </p>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowMessageDetails(false)}
            >
              Close
            </Button>
            {selectedMessage && (
              <Button
                variant="destructive"
                onClick={() => {
                  setShowMessageDetails(false);
                  handleMessageAction(selectedMessage, "delete");
                }}
              >
                Delete Message
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Conversation Details Dialog */}
      <Dialog
        open={showConversationDetails}
        onOpenChange={setShowConversationDetails}
      >
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Conversation Details</DialogTitle>
          </DialogHeader>

          {selectedConversation && (
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-4">
                    <Avatar className="h-12 w-12">
                      <AvatarImage src={selectedConversation.avatar} />
                      <AvatarFallback>
                        {selectedConversation.type === "group" ? (
                          <IconUsers className="h-6 w-6" />
                        ) : (
                          <IconUser className="h-6 w-6" />
                        )}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1">
                      <CardTitle>
                        {selectedConversation.title ||
                          `${selectedConversation.type} conversation`}
                      </CardTitle>
                      <CardDescription>
                        {selectedConversation.type === "group"
                          ? "Group Chat"
                          : "Direct Message"}
                      </CardDescription>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant="outline">
                        {selectedConversation.type}
                      </Badge>
                      {selectedConversation.is_archived && (
                        <Badge variant="secondary">Archived</Badge>
                      )}
                      {selectedConversation.is_muted && (
                        <Badge variant="outline">Muted</Badge>
                      )}
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <h4 className="font-medium mb-2">Participants</h4>
                    <div className="space-y-2">
                      {selectedConversation.participants?.map(
                        (participant, index) => (
                          <div key={index} className="flex items-center gap-2">
                            <Avatar className="h-6 w-6">
                              <AvatarImage src={participant.profile_picture} />
                              <AvatarFallback className="text-xs">
                                {participant.username?.[0]?.toUpperCase()}
                              </AvatarFallback>
                            </Avatar>
                            <span className="text-sm">
                              {participant.username}
                            </span>
                          </div>
                        )
                      )}
                    </div>
                  </div>

                  {selectedConversation.last_message && (
                    <div>
                      <h4 className="font-medium mb-2">Last Message</h4>
                      <div className="bg-gray-50 p-3 rounded">
                        <p className="text-sm">
                          {selectedConversation.last_message.content ||
                            `${selectedConversation.last_message.message_type} message`}
                        </p>
                        <p className="text-xs text-muted-foreground mt-1">
                          {new Date(
                            selectedConversation.last_message_at!
                          ).toLocaleString()}
                        </p>
                      </div>
                    </div>
                  )}

                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <label className="font-medium">Type</label>
                      <p className="capitalize">{selectedConversation.type}</p>
                    </div>
                    <div>
                      <label className="font-medium">Unread Messages</label>
                      <p>{selectedConversation.unread_count || 0}</p>
                    </div>
                    <div>
                      <label className="font-medium">Created</label>
                      <p>
                        {new Date(
                          selectedConversation.created_at
                        ).toLocaleDateString()}
                      </p>
                    </div>
                    <div>
                      <label className="font-medium">Last Activity</label>
                      <p>
                        {selectedConversation.last_message_at
                          ? new Date(
                              selectedConversation.last_message_at
                            ).toLocaleDateString()
                          : "No activity"}
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowConversationDetails(false)}
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
            <DialogTitle>Delete Message</DialogTitle>
            <DialogDescription>
              This action cannot be undone. The message will be permanently
              deleted.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">Reason</label>
              <Textarea
                placeholder="Enter reason for deleting this message..."
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
              onClick={executeMessageAction}
              variant="destructive"
              disabled={!actionReason.trim()}
            >
              Delete Message
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
              You are about to {bulkAction} {selectedIds.length} message(s).
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
              variant="destructive"
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

export default withAuth(MessagesPage);
