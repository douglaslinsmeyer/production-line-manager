package sse

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Event represents an SSE event to be broadcast to clients
type Event struct {
	Type string          `json:"type"` // "line.status", "line.updated", etc
	Data json.RawMessage `json:"data"` // JSON-encoded event data
}

// Client represents a connected SSE client
type Client struct {
	id        uuid.UUID
	eventChan chan Event
	ctx       context.Context
	cancel    context.CancelFunc
}

// Hub manages all SSE client connections and broadcasts events
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Event
	register   chan *Client
	unregister chan *Client
	logger     *zap.Logger
	mu         sync.RWMutex
	shutdown   chan struct{}
	done       chan struct{}
}

// NewHub creates a new SSE hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Event, 256), // Buffer for burst events
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
		shutdown:   make(chan struct{}),
		done:       make(chan struct{}),
	}
}

// Run starts the hub's main event loop (should be called as goroutine)
func (h *Hub) Run() {
	h.logger.Info("SSE hub started")
	defer close(h.done)

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Info("SSE client registered",
				zap.String("client_id", client.id.String()),
				zap.Int("total_clients", len(h.clients)))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.eventChan)
				h.logger.Info("SSE client unregistered",
					zap.String("client_id", client.id.String()),
					zap.Int("total_clients", len(h.clients)))
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			h.mu.RLock()
			clientCount := len(h.clients)
			h.mu.RUnlock()

			if clientCount == 0 {
				h.logger.Debug("No clients connected - skipping broadcast",
					zap.String("event_type", event.Type))
				continue
			}

			h.logger.Debug("Broadcasting event to clients",
				zap.String("event_type", event.Type),
				zap.Int("client_count", clientCount))

			// Send event to all connected clients
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.eventChan <- event:
					// Event sent successfully
				case <-client.ctx.Done():
					// Client disconnected, will be unregistered
				default:
					// Client channel full - skip to prevent blocking
					h.logger.Warn("Client event channel full - skipping event",
						zap.String("client_id", client.id.String()),
						zap.String("event_type", event.Type))
				}
			}
			h.mu.RUnlock()

		case <-h.shutdown:
			h.logger.Info("SSE hub shutting down")
			h.mu.Lock()
			for client := range h.clients {
				close(client.eventChan)
				client.cancel()
			}
			h.clients = make(map[*Client]bool)
			h.mu.Unlock()
			return
		}
	}
}

// Broadcast sends an event to all connected clients
func (h *Hub) Broadcast(event Event) {
	select {
	case h.broadcast <- event:
		// Event queued for broadcast
	default:
		h.logger.Warn("Broadcast channel full - dropping event",
			zap.String("event_type", event.Type))
	}
}

// RegisterClient adds a new client to the hub
func (h *Hub) RegisterClient(ctx context.Context) *Client {
	clientCtx, cancel := context.WithCancel(ctx)
	client := &Client{
		id:        uuid.New(),
		eventChan: make(chan Event, 16), // Buffer for client-specific events
		ctx:       clientCtx,
		cancel:    cancel,
	}

	h.register <- client
	return client
}

// UnregisterClient removes a client from the hub
func (h *Hub) UnregisterClient(client *Client) {
	client.cancel()
	h.unregister <- client
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown() {
	close(h.shutdown)
	<-h.done
	h.logger.Info("SSE hub shut down")
}

// ClientCount returns the current number of connected clients
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
