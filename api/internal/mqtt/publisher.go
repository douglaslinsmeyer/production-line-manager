package mqtt

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/sse"
)

// Publisher publishes production line events to MQTT and SSE
type Publisher struct {
	client *Client
	sseHub *sse.Hub
	logger *zap.Logger
}

// NewPublisher creates a new MQTT publisher
func NewPublisher(client *Client, sseHub *sse.Hub, logger *zap.Logger) *Publisher {
	return &Publisher{
		client: client,
		sseHub: sseHub,
		logger: logger,
	}
}

// PublishCreated publishes a production line created event
func (p *Publisher) PublishCreated(line *domain.ProductionLine) error {
	event := domain.LineEvent{
		Type:      domain.EventCreated,
		Timestamp: time.Now(),
		Data:      line,
	}

	if err := p.publishEvent(TopicEventCreated, event); err != nil {
		return err
	}

	// Also broadcast to SSE clients
	p.broadcastToSSE("line.created", event)
	return nil
}

// PublishUpdated publishes a production line updated event
func (p *Publisher) PublishUpdated(line *domain.ProductionLine) error {
	event := domain.LineEvent{
		Type:      domain.EventUpdated,
		Timestamp: time.Now(),
		Data:      line,
	}

	if err := p.publishEvent(TopicEventUpdated, event); err != nil {
		return err
	}

	// Also broadcast to SSE clients
	p.broadcastToSSE("line.updated", event)
	return nil
}

// PublishDeleted publishes a production line deleted event
func (p *Publisher) PublishDeleted(id uuid.UUID, code string) error {
	event := domain.LineDeletedEvent{
		Type:      domain.EventDeleted,
		Timestamp: time.Now(),
		ID:        id,
		Code:      code,
	}

	if err := p.publishEvent(TopicEventDeleted, event); err != nil {
		return err
	}

	// Also broadcast to SSE clients
	p.broadcastToSSE("line.deleted", event)
	return nil
}

// PublishStatus publishes a production line status change event
func (p *Publisher) PublishStatus(line *domain.ProductionLine) error {
	event := domain.StatusEvent{
		Type:      domain.EventStatus,
		Timestamp: time.Now(),
		ID:        line.ID,
		Code:      line.Code,
		Status:    line.Status,
	}

	if err := p.publishEvent(TopicEventStatus, event); err != nil {
		return err
	}

	// Also broadcast to SSE clients
	p.broadcastToSSE("line.status", event)
	return nil
}

// PublishRaw publishes raw payload to a topic
func (p *Publisher) PublishRaw(topic string, payload []byte) error {
	if err := p.client.Publish(topic, payload); err != nil {
		return err
	}

	p.logger.Info("message published",
		zap.String("topic", topic),
		zap.Int("payload_size", len(payload)))

	return nil
}

// publishEvent is a helper function to marshal and publish events
func (p *Publisher) publishEvent(topic string, event interface{}) error {
	payload, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("failed to marshal event",
			zap.String("topic", topic),
			zap.Error(err))
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := p.client.Publish(topic, payload); err != nil {
		return err
	}

	p.logger.Info("event published",
		zap.String("topic", topic),
		zap.Int("payload_size", len(payload)))

	return nil
}

// broadcastToSSE broadcasts an event to SSE clients
func (p *Publisher) broadcastToSSE(eventType string, event interface{}) {
	// Skip if no SSE hub configured
	if p.sseHub == nil {
		return
	}

	// Marshal event data
	data, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("failed to marshal event for SSE",
			zap.String("event_type", eventType),
			zap.Error(err))
		return
	}

	// Broadcast to SSE clients
	p.sseHub.Broadcast(sse.Event{
		Type: eventType,
		Data: data,
	})

	p.logger.Debug("event broadcast to SSE clients",
		zap.String("event_type", eventType))
}
