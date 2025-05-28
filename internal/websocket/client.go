// client.go
package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 8192

	// Buffer size for client channels
	channelBufferSize = 256
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking in production
		return true
	},
}

// Client represents a WebSocket client connection
type Client struct {
	// WebSocket connection
	conn *websocket.Conn

	// User information
	UserID   primitive.ObjectID `json:"user_id"`
	Username string             `json:"username"`
	IsActive bool               `json:"is_active"`

	// Connection metadata
	SessionID   string    `json:"session_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastPingAt  time.Time `json:"last_ping_at"`
	UserAgent   string    `json:"user_agent"`
	IPAddress   string    `json:"ip_address"`

	// Channels for sending messages
	send     chan []byte
	register chan *Client

	// Hub reference
	hub *Hub

	// Subscriptions - what channels/rooms this client is subscribed to
	subscriptions map[string]bool
	subMutex      sync.RWMutex

	// Client state
	isClosing bool
	mutex     sync.RWMutex
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Action    string                 `json:"action,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Target    string                 `json:"target,omitempty"`  // Target user/room
	Channel   string                 `json:"channel,omitempty"` // Channel/topic
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"` // For request-response pattern
}

// ClientInfo represents client information for admin/monitoring
type ClientInfo struct {
	UserID        string            `json:"user_id"`
	Username      string            `json:"username"`
	SessionID     string            `json:"session_id"`
	ConnectedAt   time.Time         `json:"connected_at"`
	LastPingAt    time.Time         `json:"last_ping_at"`
	Subscriptions []string          `json:"subscriptions"`
	IsActive      bool              `json:"is_active"`
	UserAgent     string            `json:"user_agent"`
	IPAddress     string            `json:"ip_address"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID primitive.ObjectID, username string, r *http.Request) *Client {
	client := &Client{
		conn:          conn,
		UserID:        userID,
		Username:      username,
		IsActive:      true,
		SessionID:     generateSessionID(),
		ConnectedAt:   time.Now(),
		LastPingAt:    time.Now(),
		UserAgent:     r.Header.Get("User-Agent"),
		IPAddress:     getClientIP(r),
		send:          make(chan []byte, channelBufferSize),
		hub:           hub,
		subscriptions: make(map[string]bool),
		isClosing:     false,
	}

	return client
}

// Start begins the client's read and write loops
func (c *Client) Start() {
	go c.writePump()
	go c.readPump()

	// Register client with hub
	c.hub.register <- c

	log.Printf("WebSocket client connected: user=%s, session=%s, ip=%s",
		c.Username, c.SessionID, c.IPAddress)
}

// Close cleanly closes the client connection
func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isClosing {
		return
	}

	c.isClosing = true
	c.IsActive = false

	// Close channels
	close(c.send)

	// Close WebSocket connection
	c.conn.Close()

	// Unregister from hub
	c.hub.unregister <- c

	log.Printf("WebSocket client disconnected: user=%s, session=%s",
		c.Username, c.SessionID)
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(message WebSocketMessage) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.isClosing {
		return ErrClientClosed
	}

	// Set timestamp
	message.Timestamp = time.Now()

	// Marshal message to JSON
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Send to client (non-blocking)
	select {
	case c.send <- data:
		return nil
	default:
		// Channel is full, client is probably slow/disconnected
		return ErrChannelFull
	}
}

// Subscribe subscribes the client to a channel/room
func (c *Client) Subscribe(channel string) {
	c.subMutex.Lock()
	defer c.subMutex.Unlock()

	c.subscriptions[channel] = true

	// Notify hub about subscription
	c.hub.subscribe <- &Subscription{
		Client:  c,
		Channel: channel,
		Action:  "subscribe",
	}

	log.Printf("Client %s subscribed to channel: %s", c.Username, channel)
}

// Unsubscribe unsubscribes the client from a channel/room
func (c *Client) Unsubscribe(channel string) {
	c.subMutex.Lock()
	defer c.subMutex.Unlock()

	delete(c.subscriptions, channel)

	// Notify hub about unsubscription
	c.hub.subscribe <- &Subscription{
		Client:  c,
		Channel: channel,
		Action:  "unsubscribe",
	}

	log.Printf("Client %s unsubscribed from channel: %s", c.Username, channel)
}

// IsSubscribed checks if client is subscribed to a channel
func (c *Client) IsSubscribed(channel string) bool {
	c.subMutex.RLock()
	defer c.subMutex.RUnlock()

	return c.subscriptions[channel]
}

// GetSubscriptions returns all client subscriptions
func (c *Client) GetSubscriptions() []string {
	c.subMutex.RLock()
	defer c.subMutex.RUnlock()

	channels := make([]string, 0, len(c.subscriptions))
	for channel := range c.subscriptions {
		channels = append(channels, channel)
	}
	return channels
}

// GetInfo returns client information
func (c *Client) GetInfo() ClientInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return ClientInfo{
		UserID:        c.UserID.Hex(),
		Username:      c.Username,
		SessionID:     c.SessionID,
		ConnectedAt:   c.ConnectedAt,
		LastPingAt:    c.LastPingAt,
		Subscriptions: c.GetSubscriptions(),
		IsActive:      c.IsActive,
		UserAgent:     c.UserAgent,
		IPAddress:     c.IPAddress,
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Close()
	}()

	// Configure connection
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.LastPingAt = time.Now()
		return nil
	})

	for {
		_, messageData, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", c.Username, err)
			}
			break
		}

		// Parse message
		var message WebSocketMessage
		if err := json.Unmarshal(messageData, &message); err != nil {
			log.Printf("Invalid WebSocket message from user %s: %v", c.Username, err)
			continue
		}

		// Set message metadata
		message.Timestamp = time.Now()

		// Handle message based on type
		if err := c.handleMessage(message); err != nil {
			log.Printf("Error handling WebSocket message from user %s: %v", c.Username, err)

			// Send error response
			errorMsg := WebSocketMessage{
				Type: "error",
				Data: map[string]interface{}{
					"message":    "Failed to process message",
					"error_code": "PROCESSING_ERROR",
				},
				RequestID: message.RequestID,
			}
			c.SendMessage(errorMsg)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming WebSocket messages
func (c *Client) handleMessage(message WebSocketMessage) error {
	switch message.Type {
	case "ping":
		return c.handlePing(message)
	case "subscribe":
		return c.handleSubscribe(message)
	case "unsubscribe":
		return c.handleUnsubscribe(message)
	case "message":
		return c.hub.messageHandler.HandleMessage(c, message)
	case "notification":
		return c.hub.notificationHandler.HandleNotification(c, message)
	case "typing":
		return c.handleTyping(message)
	case "presence":
		return c.handlePresence(message)
	case "story_view":
		return c.handleStoryView(message)
	default:
		return ErrUnknownMessageType
	}
}

// handlePing handles ping messages
func (c *Client) handlePing(message WebSocketMessage) error {
	pongMessage := WebSocketMessage{
		Type:      "pong",
		Data:      message.Data,
		RequestID: message.RequestID,
	}
	return c.SendMessage(pongMessage)
}

// handleSubscribe handles subscription requests
func (c *Client) handleSubscribe(message WebSocketMessage) error {
	channel, ok := message.Data["channel"].(string)
	if !ok {
		return ErrInvalidChannelName
	}

	// Validate channel access
	if !c.canAccessChannel(channel) {
		return ErrUnauthorizedChannel
	}

	c.Subscribe(channel)

	// Send confirmation
	confirmMessage := WebSocketMessage{
		Type:    "subscribed",
		Channel: channel,
		Data: map[string]interface{}{
			"status": "success",
		},
		RequestID: message.RequestID,
	}
	return c.SendMessage(confirmMessage)
}

// handleUnsubscribe handles unsubscription requests
func (c *Client) handleUnsubscribe(message WebSocketMessage) error {
	channel, ok := message.Data["channel"].(string)
	if !ok {
		return ErrInvalidChannelName
	}

	c.Unsubscribe(channel)

	// Send confirmation
	confirmMessage := WebSocketMessage{
		Type:    "unsubscribed",
		Channel: channel,
		Data: map[string]interface{}{
			"status": "success",
		},
		RequestID: message.RequestID,
	}
	return c.SendMessage(confirmMessage)
}

// handleTyping handles typing indicator messages
func (c *Client) handleTyping(message WebSocketMessage) error {
	conversationID, ok := message.Data["conversation_id"].(string)
	if !ok {
		return ErrInvalidData
	}

	isTyping, ok := message.Data["is_typing"].(bool)
	if !ok {
		return ErrInvalidData
	}

	// Broadcast typing indicator to conversation participants
	typingMessage := WebSocketMessage{
		Type:    "typing",
		Channel: "conversation:" + conversationID,
		Data: map[string]interface{}{
			"user_id":         c.UserID.Hex(),
			"username":        c.Username,
			"conversation_id": conversationID,
			"is_typing":       isTyping,
		},
	}

	c.hub.BroadcastToChannel("conversation:"+conversationID, typingMessage, c.UserID)
	return nil
}

// handlePresence handles presence/online status messages
func (c *Client) handlePresence(message WebSocketMessage) error {
	status, ok := message.Data["status"].(string)
	if !ok {
		return ErrInvalidData
	}

	// Update user's online status in database
	// This would typically be handled by a service
	go func() {
		// TODO: Update user's online status in database
		log.Printf("User %s presence status: %s", c.Username, status)
	}()

	return nil
}

// handleStoryView handles story view notifications
func (c *Client) handleStoryView(message WebSocketMessage) error {
	storyID, ok := message.Data["story_id"].(string)
	if !ok {
		return ErrInvalidData
	}

	// Broadcast story view to story owner
	viewMessage := WebSocketMessage{
		Type: "story_viewed",
		Data: map[string]interface{}{
			"story_id": storyID,
			"viewer": map[string]interface{}{
				"user_id":  c.UserID.Hex(),
				"username": c.Username,
			},
		},
	}

	// This would be sent to the story owner
	// Implementation depends on how you want to handle story view notifications
	c.hub.BroadcastToUser(message.Target, viewMessage)
	return nil
}

// canAccessChannel checks if the client can access a specific channel
func (c *Client) canAccessChannel(channel string) bool {
	// Basic channel access control
	// In production, this should check database permissions

	// User can always access their own channels
	if channel == "user:"+c.UserID.Hex() {
		return true
	}

	// TODO: Implement proper channel access control
	// Check if user has permission to access conversation, group, etc.

	return true
}

// Utility functions

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return primitive.NewObjectID().Hex()
}

// getClientIP extracts client IP from HTTP request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// Custom errors
var (
	ErrClientClosed        = &WebSocketError{Code: "CLIENT_CLOSED", Message: "Client connection is closed"}
	ErrChannelFull         = &WebSocketError{Code: "CHANNEL_FULL", Message: "Client send channel is full"}
	ErrUnknownMessageType  = &WebSocketError{Code: "UNKNOWN_MESSAGE_TYPE", Message: "Unknown message type"}
	ErrInvalidChannelName  = &WebSocketError{Code: "INVALID_CHANNEL", Message: "Invalid channel name"}
	ErrUnauthorizedChannel = &WebSocketError{Code: "UNAUTHORIZED_CHANNEL", Message: "Unauthorized channel access"}
	ErrInvalidData         = &WebSocketError{Code: "INVALID_DATA", Message: "Invalid message data"}
)

// WebSocketError represents WebSocket-specific errors
type WebSocketError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *WebSocketError) Error() string {
	return e.Code + ": " + e.Message
}
