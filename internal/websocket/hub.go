package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 30 * time.Second
)

type Client struct {
	ID     string
	UserID string
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	Rooms  map[string]bool
	mu     sync.Mutex
}

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	rooms      map[string]map[*Client]bool
	mu         sync.RWMutex
}

type Message struct {
	Type    string      `json:"type"`
	Room    string      `json:"room,omitempty"`
	Payload interface{} `json:"payload"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
		rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WS] Client registered: UserID=%s (total clients: %d)", client.UserID, len(h.clients))
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				log.Printf("[WS] Client unregistered: UserID=%s", client.UserID)
				delete(h.clients, client)
				close(client.Send)
				for room := range client.Rooms {
					delete(h.rooms[room], client)
					log.Printf("[WS] Client removed from room: %s", room)
				}
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.broadcastMessage(msg)
		}
	}
}

func (h *Hub) broadcastMessage(msg *Message) {
	data, _ := json.Marshal(msg)
	h.mu.RLock()
	defer h.mu.RUnlock()

	if msg.Room != "" {
		if clients, ok := h.rooms[msg.Room]; ok {
			log.Printf("[WS] Broadcasting type=%s to room=%s, %d clients", msg.Type, msg.Room, len(clients))
			for client := range clients {
				log.Printf("[WS]   -> Sending to client UserID=%s", client.UserID)
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		} else {
			log.Printf("[WS] No clients in room=%s for type=%s. Available rooms:", msg.Room, msg.Type)
			for room, clients := range h.rooms {
				log.Printf("[WS]   room=%s, clients=%d", room, len(clients))
			}
		}
	} else {
		log.Printf("[WS] Broadcasting type=%s to ALL %d clients (no room)", msg.Type, len(h.clients))
		for client := range h.clients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
}

func (h *Hub) JoinRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true
	client.Rooms[room] = true
}

func (h *Hub) BroadcastToConversation(convID string, payload interface{}) {
	h.broadcast <- &Message{Type: "message", Room: "conv:" + convID, Payload: payload}
}

func (h *Hub) BroadcastToUser(userID string, msgType string, payload interface{}) {
	log.Printf("[WS] BroadcastToUser called: userID=%s, type=%s", userID, msgType)
	h.broadcast <- &Message{Type: msgType, Room: "user:" + userID, Payload: payload}
}

func HandleConnection(hub *Hub, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	userID, _ := c.Get("user_id")
	log.Printf("[WS] New connection: UserID=%s, RemoteAddr=%s", userID.(string), c.Request.RemoteAddr)

	client := &Client{
		ID:     c.Request.RemoteAddr,
		UserID: userID.(string),
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Rooms:  make(map[string]bool),
	}

	hub.register <- client
	hub.JoinRoom(client, "user:"+client.UserID)
	log.Printf("[WS] Client joined room: user:%s", client.UserID)

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		log.Printf("[WS] readPump ending for UserID=%s, unregistering...", c.UserID)
		c.Hub.unregister <- c
		// Broadcast offline status
		c.Hub.broadcast <- &Message{
			Type:    "user_offline",
			Room:    "",
			Payload: map[string]string{"user_id": c.UserID},
		}
		c.Conn.Close()
	}()

	// Set read deadline and pong handler for keepalive
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Broadcast online status
	c.Hub.broadcast <- &Message{
		Type:    "user_online",
		Room:    "",
		Payload: map[string]string{"user_id": c.UserID},
	}
	log.Printf("[WS] User %s is now online", c.UserID)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] Unexpected close for UserID=%s: %v", c.UserID, err)
			} else {
				log.Printf("[WS] Connection closed for UserID=%s: %v", c.UserID, err)
			}
			break
		}

		// Reset read deadline on any message
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))

		var msg Message
		if json.Unmarshal(message, &msg) == nil {
			switch msg.Type {
			case "ping":
				// Respond to client-level ping with pong
				pong, _ := json.Marshal(Message{Type: "pong"})
				select {
				case c.Send <- pong:
				default:
				}
			case "join":
				if room, ok := msg.Payload.(string); ok {
					c.Hub.JoinRoom(c, room)
					log.Printf("[WS] UserID=%s joined room: %s", c.UserID, room)
				}
			case "typing":
				// Forward typing indicator to conversation room
				if payload, ok := msg.Payload.(map[string]interface{}); ok {
					convID, _ := payload["conversation_id"].(string)
					if convID != "" {
						c.Hub.broadcast <- &Message{
							Type: "typing",
							Room: "conv:" + convID,
							Payload: map[string]interface{}{
								"conversation_id": convID,
								"user_id":         c.UserID,
								"is_typing":       payload["is_typing"],
							},
						}
					}
				}
			case "read":
				// Forward read receipt to conversation room
				if payload, ok := msg.Payload.(map[string]interface{}); ok {
					convID, _ := payload["conversation_id"].(string)
					if convID != "" {
						c.Hub.broadcast <- &Message{
							Type: "read",
							Room: "conv:" + convID,
							Payload: map[string]interface{}{
								"conversation_id": convID,
								"reader_id":       c.UserID,
							},
						}
					}
				}
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("[WS] Write error for UserID=%s: %v", c.UserID, err)
				return
			}
		case <-ticker.C:
			// Send WebSocket-level ping to keep connection alive
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[WS] Ping failed for UserID=%s: %v", c.UserID, err)
				return
			}
		}
	}
}
