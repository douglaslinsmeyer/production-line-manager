package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Handler handles SSE HTTP connections
type Handler struct {
	hub    *Hub
	logger *zap.Logger
}

// NewHandler creates a new SSE handler
func NewHandler(hub *Hub, logger *zap.Logger) *Handler {
	return &Handler{
		hub:    hub,
		logger: logger,
	}
}

// ServeSSE handles SSE connection requests
func (h *Handler) ServeSSE(w http.ResponseWriter, r *http.Request) {
	// Check if response writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.logger.Error("Response writer does not support flushing")
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable buffering in nginx

	// Register client with hub
	client := h.hub.RegisterClient(r.Context())
	defer h.hub.UnregisterClient(client)

	h.logger.Info("SSE client connected",
		zap.String("client_id", client.id.String()),
		zap.String("remote_addr", r.RemoteAddr))

	// Send initial connection success message
	fmt.Fprintf(w, "event: connected\n")
	fmt.Fprintf(w, "data: {\"client_id\":\"%s\"}\n\n", client.id.String())
	flusher.Flush()

	// Setup heartbeat ticker to keep connection alive
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// Stream events to client
	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			h.logger.Info("SSE client disconnected",
				zap.String("client_id", client.id.String()),
				zap.String("remote_addr", r.RemoteAddr))
			return

		case event, ok := <-client.eventChan:
			if !ok {
				// Channel closed, client being unregistered
				return
			}

			// Send event to client in SSE format
			if err := h.sendEvent(w, flusher, event); err != nil {
				h.logger.Error("Failed to send event to client",
					zap.String("client_id", client.id.String()),
					zap.Error(err))
				return
			}

		case <-heartbeat.C:
			// Send heartbeat comment to keep connection alive
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

// sendEvent sends an SSE event to the client
func (h *Handler) sendEvent(w http.ResponseWriter, flusher http.Flusher, event Event) error {
	// Marshal event data to ensure it's valid JSON
	var data interface{}
	if err := json.Unmarshal(event.Data, &data); err != nil {
		h.logger.Error("Failed to unmarshal event data",
			zap.String("event_type", event.Type),
			zap.Error(err))
		return err
	}

	// Re-marshal for consistent formatting
	jsonData, err := json.Marshal(data)
	if err != nil {
		h.logger.Error("Failed to marshal event data",
			zap.String("event_type", event.Type),
			zap.Error(err))
		return err
	}

	// Write SSE formatted event
	// Format: event: <type>\ndata: <json>\n\n
	fmt.Fprintf(w, "event: %s\n", event.Type)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
	flusher.Flush()

	return nil
}
