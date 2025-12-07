package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port string `envconfig:"PORT" default:"8080"`

	// Logging configuration
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// Database configuration
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	// MQTT configuration
	MQTTBrokerURL string `envconfig:"MQTT_BROKER_URL" required:"true"`
	MQTTClientID  string `envconfig:"MQTT_CLIENT_ID" default:"production-line-api"`
	MQTTQOS       byte   `envconfig:"MQTT_QOS" default:"1"`

	// Graceful shutdown timeout
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"10s"`

	// CORS configuration
	CORSAllowedOrigins string `envconfig:"CORS_ALLOWED_ORIGINS" default:""`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[cfg.LogLevel] {
		return nil, fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", cfg.LogLevel)
	}

	// Validate MQTT QoS
	if cfg.MQTTQOS > 2 {
		return nil, fmt.Errorf("invalid MQTT QoS: %d (must be 0, 1, or 2)", cfg.MQTTQOS)
	}

	return &cfg, nil
}
