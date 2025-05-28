// hub.go
package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Client lookup by user ID (supports multiple sessions per user)
	userClients map[primitive.ObjectID]map[*Client]bool

	// Channel subscriptions
	channels map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to all clients
	broadcast chan BroadcastMessage

	// Channel subscription management
	subscribe chan *Subscription

	// Message handlers
	messageHandler      MessageHandlerInterface
	notificationHandler NotificationHandlerInterface

	// Statistics
	stats *HubStats

	// Mutex for thread safety
	mutex sync.RWMutex

	// Configuration
	config *HubConfig

	// Cleanup ticker
	cleanupTicker *time.Ticker

	// Shutdown channel
	shutdown chan struct{}
}

// BroadcastMessage represents a message to be broadcast
type BroadcastMessage struct {
	Type      string               `json:"type"`
	Target    BroadcastTarget      `json:"target"`
	Channel   string               `json:"channel,omitempty"`
	UserID    primitive.ObjectID   `json:"user_id,omitempty"`
	UserIDs   []primitive.ObjectID `json:"user_ids,omitempty"`
	Message   WebSocketMessage     `json:"message"`
	ExcludeID primitive.ObjectID   `json:"exclude_id,omitempty"` // Exclude specific user
}

// BroadcastTarget defines the target for broadcasting
type BroadcastTarget string

const (
	BroadcastAll     BroadcastTarget = "all"
	BroadcastUser    BroadcastTarget = "user"
	BroadcastUsers   BroadcastTarget = "users"
	BroadcastChannel BroadcastTarget = "channel"
)

// Subscription represents a channel subscription request
type Subscription struct {
	Client  *Client `json:"client"`
	Channel string  `json:"channel"`
	Action  string  `json:"action"` // "subscribe" or "unsubscribe"
}

// HubStats contains hub statistics
type HubStats struct {
	ConnectedClients    int            `json:"connected_clients"`
	TotalConnections    int64          `json:"total_connections"`
	TotalDisconnections int64          `json:"total_disconnections"`
	TotalMessages       int64          `json:"total_messages"`
	ChannelCounts       map[string]int `json:"channel_counts"`
	UserCounts          map[string]int `json:"user_counts"`
	StartTime           time.Time      `json:"start_time"`
	LastActivity        time.Time      `json:"last_activity"`
	mutex               sync.RWMutex
}

// HubConfig contains hub configuration
type HubConfig struct {
	MaxClientsPerUser int           `json:"max_clients_per_user"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	StatsInterval     time.Duration `json:"stats_interval"`
	EnableMetrics     bool          `json:"enable_metrics"`
	MaxChannels       int           `json:"max_channels"`
}

// DefaultHubConfig returns default hub configuration
func DefaultHubConfig() *HubConfig {
	return &HubConfig{
		MaxClientsPerUser: 5,
		CleanupInterval:   5 * time.Minute,
		StatsInterval:     1 * time.Minute,
		EnableMetrics:     true,
		MaxChannels:       1000,
	}
}

// NewHub creates a new WebSocket hub
func NewHub(config *HubConfig) *Hub {
	if config == nil {
		config = DefaultHubConfig()
	}

	hub := &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[primitive.ObjectID]map[*Client]bool),
		channels:    make(map[string]map[*Client]bool),
		register:    make(chan *Client, 100),
		unregister:  make(chan *Client, 100),
		broadcast:   make(chan BroadcastMessage, 1000),
		subscribe:   make(chan *Subscription, 100),
		config:      config,
		shutdown:    make(chan struct{}),
		stats: &HubStats{
			ChannelCounts: make(map[string]int),
			UserCounts:    make(map[string]int),
			StartTime:     time.Now(),
			LastActivity:  time.Now(),
		},
	}

	return hub
}

// SetMessageHandler sets the message handler
func (h *Hub) SetMessageHandler(handler MessageHandlerInterface) {
	h.messageHandler = handler
}

// SetNotificationHandler sets the notification handler
func (h *Hub) SetNotificationHandler(handler NotificationHandlerInterface) {
	h.notificationHandler = handler
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	log.Println("WebSocket Hub started")

	// Start cleanup ticker
	h.cleanupTicker = time.NewTicker(h.config.CleanupInterval)
	defer h.cleanupTicker.Stop()

	// Start stats ticker if enabled
	var statsTicker *time.Ticker
	if h.config.EnableMetrics {
		statsTicker = time.NewTicker(h.config.StatsInterval)
		defer statsTicker.Stop()
	}

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.handleBroadcast(message)

		case subscription := <-h.subscribe:
			h.handleSubscription(subscription)

		case <-h.cleanupTicker.C:
			h.cleanup()

		case <-statsTicker.C:
			if h.config.EnableMetrics {
				h.updateStats()
			}

		case <-h.shutdown:
			log.Println("WebSocket Hub shutting down")
			return
		}
	}
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown() {
	log.Println("WebSocket Hub shutdown requested")

	// Close all client connections
	h.mutex.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mutex.RUnlock()

	for _, client := range clients {
		client.Close()
	}

	close(h.shutdown)
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Check max clients per user
	userClientMap := h.userClients[client.UserID]
	if userClientMap != nil && len(userClientMap) >= h.config.MaxClientsPerUser {
		// Close oldest connection
		var oldestClient *Client
		var oldestTime time.Time

		for existingClient := range userClientMap {
			if oldestClient == nil || existingClient.ConnectedAt.Before(oldestTime) {
				oldestClient = existingClient
				oldestTime = existingClient.ConnectedAt
			}
		}

		if oldestClient != nil {
			log.Printf("Closing oldest connection for user %s due to max clients limit", client.Username)
			oldestClient.Close()
		}
	}

	// Register client
	h.clients[client] = true

	// Add to user clients map
	if h.userClients[client.UserID] == nil {
		h.userClients[client.UserID] = make(map[*Client]bool)
	}
	h.userClients[client.UserID][client] = true

	// Update stats
	h.stats.mutex.Lock()
	h.stats.ConnectedClients++
	h.stats.TotalConnections++
	h.stats.UserCounts[client.UserID.Hex()]++
	h.stats.LastActivity = time.Now()
	h.stats.mutex.Unlock()

	// Auto-subscribe to user's personal channel
	h.subscribeClientToChannel(client, "user:"+client.UserID.Hex())

	// Send welcome message
	welcomeMessage := WebSocketMessage{
		Type: "connected",
		Data: map[string]interface{}{
			"session_id":   client.SessionID,
			"user_id":      client.UserID.Hex(),
			"connected_at": client.ConnectedAt,
			"server_time":  time.Now(),
		},
	}
	client.SendMessage(welcomeMessage)

	log.Printf("Client registered: %s (total: %d)", client.Username, len(h.clients))
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client]; ok {
		// Remove from clients map
		delete(h.clients, client)

		// Remove from user clients map
		if userClientMap := h.userClients[client.UserID]; userClientMap != nil {
			delete(userClientMap, client)
			if len(userClientMap) == 0 {
				delete(h.userClients, client.UserID)
			}
		}

		// Remove from all channel subscriptions
		for channel, clientMap := range h.channels {
			if clientMap[client] {
				delete(clientMap, client)
				if len(clientMap) == 0 {
					delete(h.channels, channel)
				}
			}
		}

		// Update stats
		h.stats.mutex.Lock()
		h.stats.ConnectedClients--
		h.stats.TotalDisconnections++
		h.stats.LastActivity = time.Now()
		h.stats.mutex.Unlock()

		log.Printf("Client unregistered: %s (total: %d)", client.Username, len(h.clients))
	}
}

// handleBroadcast handles broadcast messages
func (h *Hub) handleBroadcast(broadcastMsg BroadcastMessage) {
	h.stats.mutex.Lock()
	h.stats.TotalMessages++
	h.stats.LastActivity = time.Now()
	h.stats.mutex.Unlock()

	switch broadcastMsg.Target {
	case BroadcastAll:
		h.broadcastToAll(broadcastMsg.Message, broadcastMsg.ExcludeID)
	case BroadcastUser:
		h.broadcastToUser(broadcastMsg.UserID, broadcastMsg.Message)
	case BroadcastUsers:
		h.broadcastToUsers(broadcastMsg.UserIDs, broadcastMsg.Message, broadcastMsg.ExcludeID)
	case BroadcastChannel:
		h.broadcastToChannel(broadcastMsg.Channel, broadcastMsg.Message, broadcastMsg.ExcludeID)
	}
}

// handleSubscription handles channel subscription requests
func (h *Hub) handleSubscription(subscription *Subscription) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	switch subscription.Action {
	case "subscribe":
		h.subscribeClientToChannel(subscription.Client, subscription.Channel)
	case "unsubscribe":
		h.unsubscribeClientFromChannel(subscription.Client, subscription.Channel)
	}
}

// subscribeClientToChannel subscribes a client to a channel
func (h *Hub) subscribeClientToChannel(client *Client, channel string) {
	if h.channels[channel] == nil {
		h.channels[channel] = make(map[*Client]bool)
	}
	h.channels[channel][client] = true

	// Update stats
	h.stats.mutex.Lock()
	h.stats.ChannelCounts[channel]++
	h.stats.mutex.Unlock()
}

// unsubscribeClientFromChannel unsubscribes a client from a channel
func (h *Hub) unsubscribeClientFromChannel(client *Client, channel string) {
	if clientMap := h.channels[channel]; clientMap != nil {
		delete(clientMap, client)
		if len(clientMap) == 0 {
			delete(h.channels, channel)
		}

		// Update stats
		h.stats.mutex.Lock()
		if h.stats.ChannelCounts[channel] > 0 {
			h.stats.ChannelCounts[channel]--
		}
		h.stats.mutex.Unlock()
	}
}

// Broadcasting methods

// BroadcastToAll broadcasts a message to all connected clients
func (h *Hub) BroadcastToAll(message WebSocketMessage) {
	broadcastMsg := BroadcastMessage{
		Type:    "broadcast_all",
		Target:  BroadcastAll,
		Message: message,
	}
	h.broadcast <- broadcastMsg
}

// BroadcastToUser broadcasts a message to a specific user (all their sessions)
func (h *Hub) BroadcastToUser(userID string, message WebSocketMessage) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Printf("Invalid user ID for broadcast: %s", userID)
		return
	}

	broadcastMsg := BroadcastMessage{
		Type:    "broadcast_user",
		Target:  BroadcastUser,
		UserID:  userObjectID,
		Message: message,
	}
	h.broadcast <- broadcastMsg
}

// BroadcastToUsers broadcasts a message to multiple users
func (h *Hub) BroadcastToUsers(userIDs []string, message WebSocketMessage, excludeUserID string) {
	userObjectIDs := make([]primitive.ObjectID, 0, len(userIDs))
	for _, userID := range userIDs {
		if objectID, err := primitive.ObjectIDFromHex(userID); err == nil {
			userObjectIDs = append(userObjectIDs, objectID)
		}
	}

	var excludeID primitive.ObjectID
	if excludeUserID != "" {
		excludeID, _ = primitive.ObjectIDFromHex(excludeUserID)
	}

	broadcastMsg := BroadcastMessage{
		Type:      "broadcast_users",
		Target:    BroadcastUsers,
		UserIDs:   userObjectIDs,
		Message:   message,
		ExcludeID: excludeID,
	}
	h.broadcast <- broadcastMsg
}

// BroadcastToChannel broadcasts a message to all clients subscribed to a channel
func (h *Hub) BroadcastToChannel(channel string, message WebSocketMessage, excludeUserID primitive.ObjectID) {
	broadcastMsg := BroadcastMessage{
		Type:      "broadcast_channel",
		Target:    BroadcastChannel,
		Channel:   channel,
		Message:   message,
		ExcludeID: excludeUserID,
	}
	h.broadcast <- broadcastMsg
}

// Internal broadcasting methods

// broadcastToAll broadcasts to all clients
func (h *Hub) broadcastToAll(message WebSocketMessage, excludeUserID primitive.ObjectID) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		if client.UserID != excludeUserID {
			select {
			case <-time.After(100 * time.Millisecond):
				// Skip slow clients
				log.Printf("Skipping slow client: %s", client.Username)
			default:
				client.SendMessage(message)
			}
		}
	}
}

// broadcastToUser broadcasts to a specific user
func (h *Hub) broadcastToUser(userID primitive.ObjectID, message WebSocketMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clientMap := h.userClients[userID]; clientMap != nil {
		for client := range clientMap {
			client.SendMessage(message)
		}
	}
}

// broadcastToUsers broadcasts to multiple users
func (h *Hub) broadcastToUsers(userIDs []primitive.ObjectID, message WebSocketMessage, excludeUserID primitive.ObjectID) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, userID := range userIDs {
		if userID != excludeUserID {
			if clientMap := h.userClients[userID]; clientMap != nil {
				for client := range clientMap {
					client.SendMessage(message)
				}
			}
		}
	}
}

// broadcastToChannel broadcasts to channel subscribers
func (h *Hub) broadcastToChannel(channel string, message WebSocketMessage, excludeUserID primitive.ObjectID) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clientMap := h.channels[channel]; clientMap != nil {
		for client := range clientMap {
			if client.UserID != excludeUserID {
				client.SendMessage(message)
			}
		}
	}
}

// Utility methods

// GetConnectedUsers returns a list of connected user IDs
func (h *Hub) GetConnectedUsers() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	userIDs := make([]string, 0, len(h.userClients))
	for userID := range h.userClients {
		userIDs = append(userIDs, userID.Hex())
	}
	return userIDs
}

// IsUserOnline checks if a user is currently online
func (h *Hub) IsUserOnline(userID string) bool {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	_, exists := h.userClients[userObjectID]
	return exists
}

// GetUserSessionCount returns the number of active sessions for a user
func (h *Hub) GetUserSessionCount(userID string) int {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clientMap := h.userClients[userObjectID]; clientMap != nil {
		return len(clientMap)
	}
	return 0
}

// GetChannelSubscriberCount returns the number of subscribers to a channel
func (h *Hub) GetChannelSubscriberCount(channel string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clientMap := h.channels[channel]; clientMap != nil {
		return len(clientMap)
	}
	return 0
}

// GetStats returns hub statistics
func (h *Hub) GetStats() HubStats {
	h.stats.mutex.RLock()
	defer h.stats.mutex.RUnlock()

	// Create a copy to avoid race conditions
	statsCopy := *h.stats
	statsCopy.ChannelCounts = make(map[string]int)
	statsCopy.UserCounts = make(map[string]int)

	for k, v := range h.stats.ChannelCounts {
		statsCopy.ChannelCounts[k] = v
	}
	for k, v := range h.stats.UserCounts {
		statsCopy.UserCounts[k] = v
	}

	return statsCopy
}

// GetConnectedClients returns information about all connected clients
func (h *Hub) GetConnectedClients() []ClientInfo {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients := make([]ClientInfo, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client.GetInfo())
	}
	return clients
}

// cleanup performs periodic cleanup tasks
func (h *Hub) cleanup() {
	log.Println("Running WebSocket hub cleanup")

	// Clean up empty channels
	h.mutex.Lock()
	for channel, clientMap := range h.channels {
		if len(clientMap) == 0 {
			delete(h.channels, channel)
		}
	}
	h.mutex.Unlock()

	// Clean up stats
	h.stats.mutex.Lock()
	for channel := range h.stats.ChannelCounts {
		if h.stats.ChannelCounts[channel] <= 0 {
			delete(h.stats.ChannelCounts, channel)
		}
	}
	h.stats.mutex.Unlock()
}

// updateStats updates hub statistics
func (h *Hub) updateStats() {
	h.mutex.RLock()
	h.stats.mutex.Lock()

	h.stats.ConnectedClients = len(h.clients)

	// Update channel counts
	for channel, clientMap := range h.channels {
		h.stats.ChannelCounts[channel] = len(clientMap)
	}

	h.stats.mutex.Unlock()
	h.mutex.RUnlock()
}

// SendSystemMessage sends a system-wide message to all clients
func (h *Hub) SendSystemMessage(messageType, content string) {
	systemMessage := WebSocketMessage{
		Type: "system",
		Data: map[string]interface{}{
			"message_type": messageType,
			"content":      content,
			"timestamp":    time.Now(),
		},
	}
	h.BroadcastToAll(systemMessage)
}

// SendMaintenanceNotice sends a maintenance notice to all clients
func (h *Hub) SendMaintenanceNotice(message string, scheduledTime time.Time) {
	maintenanceMessage := WebSocketMessage{
		Type: "maintenance",
		Data: map[string]interface{}{
			"message":        message,
			"scheduled_time": scheduledTime,
			"server_time":    time.Now(),
		},
	}
	h.BroadcastToAll(maintenanceMessage)
}

// Debug and monitoring methods

// GetHubInfo returns detailed hub information for debugging
func (h *Hub) GetHubInfo() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return map[string]interface{}{
		"connected_clients": len(h.clients),
		"unique_users":      len(h.userClients),
		"active_channels":   len(h.channels),
		"config":            h.config,
		"stats":             h.GetStats(),
		"uptime":            time.Since(h.stats.StartTime),
	}
}

// ExportMetrics exports metrics in JSON format
func (h *Hub) ExportMetrics() ([]byte, error) {
	metrics := map[string]interface{}{
		"hub_info": h.GetHubInfo(),
		"clients":  h.GetConnectedClients(),
		"stats":    h.GetStats(),
	}

	return json.Marshal(metrics)
}
