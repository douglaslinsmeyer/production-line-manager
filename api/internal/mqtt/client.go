package mqtt

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// Client wraps the MQTT client with connection management
type Client struct {
	client   mqtt.Client
	logger   *zap.Logger
	qos      byte
	brokerURL string
	clientID  string
}

// Config holds MQTT client configuration
type Config struct {
	BrokerURL string
	ClientID  string
	QOS       byte
}

// NewClient creates a new MQTT client with auto-reconnect
func NewClient(cfg *Config, logger *zap.Logger) (*Client, error) {
	if cfg.BrokerURL == "" {
		return nil, fmt.Errorf("broker URL is required")
	}

	c := &Client{
		logger:    logger,
		qos:       cfg.QOS,
		brokerURL: cfg.BrokerURL,
		clientID:  cfg.ClientID,
	}

	// Configure MQTT client options
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.BrokerURL).
		SetClientID(cfg.ClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetMaxReconnectInterval(60 * time.Second).
		SetKeepAlive(60 * time.Second).
		SetPingTimeout(10 * time.Second).
		SetWriteTimeout(10 * time.Second).
		SetOnConnectHandler(func(client mqtt.Client) {
			logger.Info("mqtt connected",
				zap.String("broker", cfg.BrokerURL),
				zap.String("client_id", cfg.ClientID))
		}).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			logger.Warn("mqtt connection lost",
				zap.Error(err),
				zap.String("broker", cfg.BrokerURL))
		}).
		SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
			logger.Info("mqtt reconnecting",
				zap.String("broker", cfg.BrokerURL))
		})

	// Create client
	c.client = mqtt.NewClient(opts)

	// Connect
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	return c, nil
}

// Publish publishes a message to a topic
func (c *Client) Publish(topic string, payload []byte) error {
	token := c.client.Publish(topic, c.qos, false, payload)
	token.Wait()

	if err := token.Error(); err != nil {
		c.logger.Error("failed to publish message",
			zap.String("topic", topic),
			zap.Error(err))
		return fmt.Errorf("publish failed: %w", err)
	}

	c.logger.Debug("message published",
		zap.String("topic", topic),
		zap.Int("payload_size", len(payload)))

	return nil
}

// Subscribe subscribes to a topic with a message handler
func (c *Client) Subscribe(topic string, handler mqtt.MessageHandler) error {
	token := c.client.Subscribe(topic, c.qos, handler)
	token.Wait()

	if err := token.Error(); err != nil {
		c.logger.Error("failed to subscribe to topic",
			zap.String("topic", topic),
			zap.Error(err))
		return fmt.Errorf("subscribe failed: %w", err)
	}

	c.logger.Info("subscribed to topic",
		zap.String("topic", topic),
		zap.Uint8("qos", c.qos))

	return nil
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

// Disconnect disconnects the MQTT client
func (c *Client) Disconnect() {
	if c.client.IsConnected() {
		c.client.Disconnect(250) // Wait up to 250ms to send pending messages
		c.logger.Info("mqtt disconnected")
	}
}
