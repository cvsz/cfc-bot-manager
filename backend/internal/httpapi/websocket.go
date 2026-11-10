package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/camfrog"
	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
	"github.com/gorilla/websocket"
)

type wsHub struct {
	clients    map[string]map[*wsClient]bool
	mu         sync.RWMutex
	repo       repository.RepositoryProvider
	upgrader   websocket.Upgrader
}

type wsClient struct {
	conn     *websocket.Conn
	botID    domain.BotID
	orgID    domain.OrganizationID
	send     chan []byte
	closed   bool
	mu       sync.Mutex
}

type wsMessage struct {
	Type      string          `json:"type"`
	BotID     string          `json:"bot_id,omitempty"`
	RoomID    string          `json:"room_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Error     string          `json:"error,omitempty"`
	Timestamp string          `json:"timestamp"`
}

func NewWSHub(repo repository.RepositoryProvider) *wsHub {
	return &wsHub{
		clients: make(map[string]map[*wsClient]bool),
		repo:    repo,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins in development; restrict in production
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (h *wsHub) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Extract bot_id and org_id from query params
	botIDStr := r.URL.Query().Get("bot_id")
	orgIDStr := r.URL.Query().Get("org_id")

	if botIDStr == "" || orgIDStr == "" {
		http.Error(w, "bot_id and org_id query parameters required", http.StatusBadRequest)
		return
	}

	botID, err := domain.ParseBotID(botIDStr)
	if err != nil {
		http.Error(w, "invalid bot_id", http.StatusBadRequest)
		return
	}

	orgID, err := domain.ParseOrganizationID(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org_id", http.StatusBadRequest)
		return
	}

	// Verify bot belongs to organization
	bot, err := h.repo.Bots().GetByID(r.Context(), botID)
	if err != nil || bot.OrganizationID != orgID {
		http.Error(w, "bot not found or access denied", http.StatusForbidden)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &wsClient{
		conn:  conn,
		botID: botID,
		orgID: orgID,
		send:  make(chan []byte, 256),
	}

	// Register client
	h.mu.Lock()
	if h.clients[botIDStr] == nil {
		h.clients[botIDStr] = make(map[*wsClient]bool)
	}
	h.clients[botIDStr][client] = true
	h.mu.Unlock()

	// Start goroutines
	go client.writePump()
	go client.readPump(h)

	// Send initial status
	h.sendBotStatus(client)
}

func (h *wsHub) sendBotStatus(client *wsClient) {
	// This would need access to the adapter - for now send a placeholder
	msg := wsMessage{
		Type:      "status",
		BotID:     client.botID.String(),
		Timestamp: time.Now().Format(time.RFC3339),
		Payload:   json.RawMessage(`{"status":"connecting"}`),
	}

	data, _ := json.Marshal(msg)
	select {
	case client.send <- data:
	default:
		// Channel full, client will be cleaned up
	}
}

func (h *wsHub) BroadcastBotStatus(botID domain.BotID, status *camfrog.StatusResult) {
	botIDStr := botID.String()

	msg := wsMessage{
		Type:      "status",
		BotID:     botIDStr,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	if status != nil {
		payload, _ := json.Marshal(status)
		msg.Payload = payload
	}

	data, _ := json.Marshal(msg)

	h.mu.RLock()
	clients := h.clients[botIDStr]
	h.mu.RUnlock()

	for client := range clients {
		client.mu.Lock()
		if !client.closed {
			select {
			case client.send <- data:
			default:
				// Client buffer full, close it
				client.close()
			}
		}
		client.mu.Unlock()
	}
}

func (h *wsHub) BroadcastBotEvent(botID domain.BotID, eventType string, payload interface{}) {
	botIDStr := botID.String()

	payloadBytes, _ := json.Marshal(payload)
	msg := wsMessage{
		Type:      eventType,
		BotID:     botIDStr,
		Timestamp: time.Now().Format(time.RFC3339),
		Payload:   payloadBytes,
	}

	data, _ := json.Marshal(msg)

	h.mu.RLock()
	clients := h.clients[botIDStr]
	h.mu.RUnlock()

	for client := range clients {
		client.mu.Lock()
		if !client.closed {
			select {
			case client.send <- data:
			default:
				client.close()
			}
		}
		client.mu.Unlock()
	}
}

func (c *wsClient) readPump(h *wsHub) {
	defer func() {
		c.close()
		h.unregister(c)
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Handle incoming messages (ping, subscribe, etc.)
		var msg wsMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "ping":
			c.sendPong()
		case "subscribe":
			// Handle subscription to specific events
		}
	}
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.mu.Lock()
			if !ok || c.closed {
				c.mu.Unlock()
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			c.mu.Unlock()
			if err != nil {
				return
			}
		case <-ticker.C:
			c.mu.Lock()
			if c.closed {
				c.mu.Unlock()
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			c.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

func (c *wsClient) sendPong() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_ = c.conn.WriteMessage(websocket.PongMessage, nil)
}

func (c *wsClient) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	close(c.send)
}

func (h *wsHub) unregister(c *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	botIDStr := c.botID.String()
	if clients, ok := h.clients[botIDStr]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.clients, botIDStr)
		}
	}
}

// WebSocketHandler returns an http.Handler for WebSocket connections
func WebSocketHandler(repo repository.RepositoryProvider) http.Handler {
	hub := NewWSHub(repo)
	return http.HandlerFunc(hub.ServeWS)
}