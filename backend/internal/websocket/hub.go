package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client
type Client struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
	Rooms  map[string]bool // conversation_id -> subscribed
	mu     sync.RWMutex
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	Clients    map[uuid.UUID]*Client
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Room    string      `json:"room,omitempty"` // conversation_id or account_id
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID]*Client),
		Broadcast:  make(chan *Message, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Client registered: %s (User: %s)", client.ID, client.UserID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
				log.Printf("Client unregistered: %s", client.ID)
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

// broadcastMessage sends message to all clients in a room
func (h *Hub) broadcastMessage(message *Message) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.Clients {
		// If room is specified, only send to clients in that room
		if message.Room != "" {
			client.mu.RLock()
			inRoom := client.Rooms[message.Room]
			client.mu.RUnlock()

			if !inRoom {
				continue
			}
		}

		select {
		case client.Send <- data:
		default:
			// Client's send channel is full, skip
			log.Printf("Client %s send channel full, skipping message", client.ID)
		}
	}
}

// ReadPump reads messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (subscribe/unsubscribe from rooms)
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		c.handleMessage(&msg)
	}
}

// WritePump writes messages to the WebSocket connection
func (c *Client) WritePump() {
	defer c.Conn.Close()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Error writing message: %v", err)
			return
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case "subscribe":
		if room, ok := msg.Payload.(string); ok {
			c.mu.Lock()
			c.Rooms[room] = true
			c.mu.Unlock()
			log.Printf("Client %s subscribed to room: %s", c.ID, room)
		}

	case "unsubscribe":
		if room, ok := msg.Payload.(string); ok {
			c.mu.Lock()
			delete(c.Rooms, room)
			c.mu.Unlock()
			log.Printf("Client %s unsubscribed from room: %s", c.ID, room)
		}

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// BroadcastToRoom sends a message to all clients in a specific room
func (h *Hub) BroadcastToRoom(room string, messageType string, payload interface{}) {
	h.Broadcast <- &Message{
		Type:    messageType,
		Payload: payload,
		Room:    room,
	}
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID uuid.UUID, messageType string, payload interface{}) {
	message := &Message{
		Type:    messageType,
		Payload: payload,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.Clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
			default:
				log.Printf("Client %s send channel full, skipping message", client.ID)
			}
		}
	}
}
