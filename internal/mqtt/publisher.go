package mqtt

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
)

// Publisher publishes production line events to MQTT
type Publisher struct {
	client *Client
	logger *zap.Logger
}

// NewPublisher creates a new MQTT publisher
func NewPublisher(client *Client, logger *zap.Logger) *Publisher {
	return &Publisher{
		client: client,
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

	return p.publishEvent(TopicEventCreated, event)
}

// PublishUpdated publishes a production line updated event
func (p *Publisher) PublishUpdated(line *domain.ProductionLine) error {
	event := domain.LineEvent{
		Type:      domain.EventUpdated,
		Timestamp: time.Now(),
		Data:      line,
	}

	return p.publishEvent(TopicEventUpdated, event)
}

// PublishDeleted publishes a production line deleted event
func (p *Publisher) PublishDeleted(id uuid.UUID, code string) error {
	event := domain.LineDeletedEvent{
		Type:      domain.EventDeleted,
		Timestamp: time.Now(),
		ID:        id,
		Code:      code,
	}

	return p.publishEvent(TopicEventDeleted, event)
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

	return p.publishEvent(TopicEventStatus, event)
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
