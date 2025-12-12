package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/service"
)

// Subscriber subscribes to MQTT command topics and processes them
type Subscriber struct {
	client                 *Client
	lineService            *service.LineService
	deviceDiscoveryHandler *DeviceDiscoveryHandler
	logger                 *zap.Logger
	subscriptions          []subscription
	subMu                  sync.RWMutex
}

// subscription represents a single MQTT topic subscription
type subscription struct {
	topic   string
	handler mqtt.MessageHandler
}

// NewSubscriber creates a new MQTT subscriber
func NewSubscriber(
	client *Client,
	lineService *service.LineService,
	deviceDiscoveryHandler *DeviceDiscoveryHandler,
	logger *zap.Logger,
) *Subscriber {
	return &Subscriber{
		client:               client,
		lineService:          lineService,
		deviceDiscoveryHandler: deviceDiscoveryHandler,
		logger:               logger,
	}
}

// subscribeToTopics performs all MQTT subscriptions
func (s *Subscriber) subscribeToTopics() error {
	s.subMu.RLock()
	subs := s.subscriptions
	s.subMu.RUnlock()

	var errors []error
	for _, sub := range subs {
		if err := s.client.Subscribe(sub.topic, sub.handler); err != nil {
			s.logger.Error("failed to subscribe to topic",
				zap.String("topic", sub.topic),
				zap.Error(err))
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to subscribe to %d topic(s)", len(errors))
	}

	s.logger.Info("successfully subscribed to all mqtt topics",
		zap.Int("topic_count", len(subs)))
	return nil
}

// ResubscribeAll re-subscribes to all configured topics (called on reconnect)
func (s *Subscriber) ResubscribeAll() error {
	s.logger.Info("re-subscribing to mqtt topics")
	return s.subscribeToTopics()
}

// Start starts subscribing to command topics
func (s *Subscriber) Start() error {
	// Register all subscriptions
	s.subMu.Lock()
	s.subscriptions = []subscription{
		{
			topic:   TopicCommandStatus,
			handler: s.handleStatusCommand,
		},
		{
			topic:   "devices/announce",
			handler: s.deviceDiscoveryHandler.HandleAnnouncement,
		},
		{
			topic:   "devices/+/status",
			handler: s.deviceDiscoveryHandler.HandleDeviceStatus,
		},
		{
			topic:   "devices/+/input-change",
			handler: s.deviceDiscoveryHandler.HandleInputChange,
		},
	}
	s.subMu.Unlock()

	// Register this subscriber with the client for reconnection handling
	s.client.SetSubscriptionManager(s)

	// Perform initial subscription
	if err := s.subscribeToTopics(); err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	// Start stale device monitor
	s.deviceDiscoveryHandler.StartStaleDeviceMonitor()

	s.logger.Info("mqtt subscriber started",
		zap.Int("topic_count", len(s.subscriptions)))

	return nil
}

// handleStatusCommand handles incoming status change commands from MQTT
func (s *Subscriber) handleStatusCommand(client mqtt.Client, msg mqtt.Message) {
	s.logger.Debug("received status command",
		zap.String("topic", msg.Topic()),
		zap.Int("payload_size", len(msg.Payload())))

	// Parse command
	var cmd domain.StatusCommand
	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		s.logger.Error("failed to parse status command",
			zap.String("topic", msg.Topic()),
			zap.ByteString("payload", msg.Payload()),
			zap.Error(err))
		return
	}

	// Validate command
	if cmd.Code == "" {
		s.logger.Error("status command missing code",
			zap.ByteString("payload", msg.Payload()))
		return
	}

	if !cmd.Status.IsValid() {
		s.logger.Error("invalid status in command",
			zap.String("code", cmd.Code),
			zap.String("status", cmd.Status.String()))
		return
	}

	s.logger.Info("processing status command",
		zap.String("code", cmd.Code),
		zap.String("status", cmd.Status.String()))

	// Get production line by code
	ctx := context.Background()
	line, err := s.lineService.GetByCode(ctx, cmd.Code)
	if err != nil {
		if err == domain.ErrNotFound {
			s.logger.Warn("production line not found for status command",
				zap.String("code", cmd.Code))
		} else {
			s.logger.Error("failed to get production line",
				zap.String("code", cmd.Code),
				zap.Error(err))
		}
		return
	}

	// Set status via service (which will log and publish event)
	sourceDetail := map[string]interface{}{
		"topic":     msg.Topic(),
		"message_id": msg.MessageID(),
	}

	_, err = s.lineService.SetStatus(ctx, line.ID, cmd.Status, "mqtt", sourceDetail)
	if err != nil {
		s.logger.Error("failed to set production line status",
			zap.String("code", cmd.Code),
			zap.String("status", cmd.Status.String()),
			zap.Error(err))
		return
	}

	s.logger.Info("status command processed successfully",
		zap.String("code", cmd.Code),
		zap.String("status", cmd.Status.String()))
}
