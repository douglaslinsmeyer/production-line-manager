package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new configured logger based on the log level
func New(level string) (*zap.Logger, error) {
	// Parse log level
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level %s: %w", level, err)
	}

	// Create logger configuration
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	// Use console encoding for better readability in development
	// In production, JSON encoding is better for log aggregation
	if level == "debug" {
		config.Encoding = "console"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config.Encoding = "json"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Build logger
	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return logger, nil
}
