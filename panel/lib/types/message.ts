import { BaseEntity, SoftDeleteEntity, Media, PaginatedResponse } from './common';
import { User } from './auth';

export interface Conversation extends SoftDeleteEntity {
  type: 'direct' | 'group' | 'broadcast' | 'support';
  title?: string;
  description?: string;
  avatar?: string;
  
  // Participants
  participant_count: number;
  max_participants?: number;
  
  // Settings
  is_private: boolean;
  allow_invites: boolean;
  allow_media_sharing: boolean;
  allow_file_sharing: boolean;
  auto_delete_messages?: number; // hours, 0 for never
  
  // Status
  status: 'active' | 'archived' | 'deleted' | 'muted';
  
  // Last message info
  last_message_id?: string;
  last_message_at?: string;
  last_message_preview?: string;
  
  // Creator info
  created_by: string;
  
  // Analytics
  message_count: number;
  active_participants: number;
  
  // Relationships
  participants?: ConversationParticipant[];
  last_message?: Message;
  creator?: User;
}

export interface ConversationParticipant extends BaseEntity {
  conversation_id: string;
  user_id: string;
  role: 'admin' | 'moderator' | 'member';
  status: 'active' | 'left' | 'kicked' | 'banned';
  
  // Permissions
  can_send_messages: boolean;
  can_add_participants: boolean;
  can_remove_participants: boolean;
  can_edit_conversation: boolean;
  can_delete_messages: boolean;
  
  // Settings
  notifications_enabled: boolean;
  is_muted: boolean;
  mute_until?: string;
  nickname?: string;
  
  // Read status
  last_read_message_id?: string;
  last_read_at?: string;
  unread_count: number;
  
  // Join info
  joined_at: string;
  invited_by?: string;
  left_at?: string;
  
  // Relationships
  user?: User;
  conversation?: Conversation;
}

export interface Message extends SoftDeleteEntity {
  conversation_id: string;
  sender_id: string;
  content: string;
  content_type: 'text' | 'image' | 'video' | 'audio' | 'file' | 'link' | 'location' | 'contact' | 'poll' | 'system';
  
  // Reply/Threading
  reply_to_message_id?: string;
  thread_root_id?: string;
  
  // Content details
  media: Media[];
  mentions: string[];
  
  // Message metadata
  priority: 'normal' | 'high' | 'urgent';
  is_edited: boolean;
  edited_at?: string;
  edit_count: number;
  
  // System messages
  system_message_type?: 'user_joined' | 'user_left' | 'user_added' | 'user_removed' | 'conversation_created' | 'title_changed' | 'avatar_changed';
  system_message_data?: Record<string, any>;
  
  // Status
  status: 'sent' | 'delivered' | 'read' | 'failed' | 'deleted';
  
  // Reactions
  reaction_count: number;
  
  // Forward info
  forwarded_from?: string; // conversation_id or user_id
  forward_count: number;
  
  // Ephemeral messages
  expires_at?: string;
  is_ephemeral: boolean;
  
  // Moderation
  is_flagged: boolean;
  flagged_reason?: string;
  
  // Relationships
  sender?: User;
  conversation?: Conversation;
  reply_to?: Message;
  reactions?: MessageReaction[];
  read_receipts?: MessageReadReceipt[];
}

export interface MessageReaction extends BaseEntity {
  message_id: string;
  user_id: string;
  reaction_type: 'like' | 'love' | 'laugh' | 'wow' | 'sad' | 'angry' | 'thumbs_up' | 'thumbs_down';
  is_active: boolean;
}

export interface MessageReadReceipt extends BaseEntity {
  message_id: string;
  user_id: string;
  read_at: string;
  delivered_at?: string;
}

export interface MessageDraft extends BaseEntity {
  conversation_id: string;
  user_id: string;
  content: string;
  content_type: 'text' | 'image' | 'video' | 'audio' | 'file';
  media: Media[];
  reply_to_message_id?: string;
  last_saved_at: string;
}

export interface MessageAttachment extends BaseEntity {
  message_id: string;
  file_id: string;
  filename: string;
  file_size: number;
  mime_type: string;
  download_count: number;
  expires_at?: string;
}

// Specialized message types
export interface PollMessage {
  question: string;
  options: Array<{
    id: string;
    text: string;
    vote_count: number;
  }>;
  expires_at?: string;
  allow_multiple_votes: boolean;
  anonymous_voting: boolean;
  total_votes: number;
}

export interface LocationMessage {
  latitude: number;
  longitude: number;
  address?: string;
  place_name?: string;
  live_location?: boolean;
  live_until?: string;
}

export interface ContactMessage {
  name: string;
  phone?: string;
  email?: string;
  organization?: string;
}

// Search and filter types
export interface MessageSearchParams {
  query?: string;
  conversation_id?: string;
  sender_id?: string;
  content_type?: 'text' | 'image' | 'video' | 'audio' | 'file';
  has_media?: boolean;
  has_reactions?: boolean;
  is_unread?: boolean;
  date_from?: string;
  date_to?: string;
  priority?: 'normal' | 'high' | 'urgent';
  sort_by?: 'created_at' | 'reaction_count';
  sort_order?: 'asc' | 'desc';
  page?: number;
  limit?: number;
}

export interface ConversationSearchParams {
  query?: string;
  type?: 'direct' | 'group' | 'broadcast' | 'support';
  participant_id?: string;
  has_unread?: boolean;
  is_archived?: boolean;
  is_muted?: boolean;
  created_after?: string;
  created_before?: string;
  sort_by?: 'last_message_at' | 'created_at' | 'participant_count';
  sort_order?: 'asc' | 'desc';
  page?: number;
  limit?: number;
}

// Form types
export interface CreateConversationRequest {
  type: 'direct' | 'group' | 'broadcast';
  participant_ids: string[];
  title?: string;
  description?: string;
  is_private?: boolean;
  allow_invites?: boolean;
  allow_media_sharing?: boolean;
  initial_message?: string;
}

export interface UpdateConversationRequest {
  title?: string;
  description?: string;
  avatar?: string;
  allow_invites?: boolean;
  allow_media_sharing?: boolean;
  allow_file_sharing?: boolean;
  auto_delete_messages?: number;
}

export interface SendMessageRequest {
  conversation_id: string;
  content: string;
  content_type?: 'text' | 'image' | 'video' | 'audio' | 'file' | 'location' | 'contact' | 'poll';
  media?: string[]; // media IDs
  reply_to_message_id?: string;
  priority?: 'normal' | 'high' | 'urgent';
  mentions?: string[];
  is_ephemeral?: boolean;
  expires_at?: string;
  poll_data?: {
    question: string;
    options: string[];
    expires_at?: string;
    allow_multiple_votes?: boolean;
    anonymous_voting?: boolean;
  };
  location_data?: {
    latitude: number;
    longitude: number;
    address?: string;
    place_name?: string;
  };
  contact_data?: {
    name: string;
    phone?: string;
    email?: string;
    organization?: string;
  };
}

export interface UpdateMessageRequest {
  content?: string;
  mentions?: string[];
}

export interface AddParticipantsRequest {
  user_ids: string[];
  role?: 'member' | 'moderator';
}

export interface UpdateParticipantRequest {
  user_id: string;
  role?: 'admin' | 'moderator' | 'member';
  can_send_messages?: boolean;
  can_add_participants?: boolean;
  can_remove_participants?: boolean;
  can_edit_conversation?: boolean;
  can_delete_messages?: boolean;
}

export interface MarkAsReadRequest {
  message_ids?: string[];
  last_message_id?: string;
}

// API response types
export type ConversationsResponse = PaginatedResponse<Conversation>;
export type MessagesResponse = PaginatedResponse<Message>;
export type ConversationParticipantsResponse = PaginatedResponse<ConversationParticipant>;
export type MessageReactionsResponse = PaginatedResponse<MessageReaction>;
export type MessageDraftsResponse = PaginatedResponse<MessageDraft>;

// Statistics types
export interface MessageStats {
  total_conversations: number;
  active_conversations: number;
  direct_conversations: number;
  group_conversations: number;
  broadcast_conversations: number;
  total_messages: number;
  messages_today: number;
  messages_week: number;
  messages_month: number;
  avg_messages_per_conversation: number;
  avg_response_time: number; // in minutes
  
  message_type_distribution: Record<string, number>;
  peak_messaging_hours: Array<{
    hour: number;
    message_count: number;
  }>;
  
  most_active_users: Array<{
    user_id: string;
    username: string;
    message_count: number;
    conversation_count: number;
  }>;
}

export interface ConversationAnalytics {
  conversation_id: string;
  participant_count: number;
  message_count: number;
  active_participants: number;
  avg_messages_per_day: number;
  avg_response_time: number;
  peak_activity_hours: number[];
  message_type_breakdown: Record<string, number>;
  engagement_score: number;
  retention_rate: number;
}

// Notification types
export interface MessageNotification {
  conversation_id: string;
  message_id: string;
  type: 'new_message' | 'new_conversation' | 'added_to_group' | 'removed_from_group' | 'conversation_updated';
  title: string;
  body: string;
  data: Record<string, any>;
  recipients: string[];
  send_push: boolean;
  send_email: boolean;
}

// Moderation types
export interface MessageReport {
  id: string;
  message_id: string;
  conversation_id: string;
  reporter_id: string;
  reason: 'spam' | 'harassment' | 'inappropriate_content' | 'scam' | 'other';
  description: string;
  status: 'pending' | 'resolved' | 'rejected';
  created_at: string;
}

export interface MessageModeration {
  message_id: string;
  action: 'flag' | 'delete' | 'edit' | 'warn_user';
  reason: string;
  moderator_id: string;
  created_at: string;
}

// Real-time types
export interface MessageEvent {
  type: 'message_sent' | 'message_read' | 'message_deleted' | 'user_typing' | 'user_online' | 'user_offline';
  conversation_id: string;
  user_id: string;
  data: any;
  timestamp: string;
}

export interface TypingIndicator {
  conversation_id: string;
  user_id: string;
  is_typing: boolean;
  started_at?: string;
}

export interface OnlineStatus {
  user_id: string;
  status: 'online' | 'offline' | 'away' | 'busy';
  last_seen_at?: string;
}