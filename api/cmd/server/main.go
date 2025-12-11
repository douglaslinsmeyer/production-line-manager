// Package main is the entry point for the Production Line API server
//
//	@title						Production Line API
//	@version					1.0
//	@description				API backend for managing production line operations in a manufacturing facility
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@example.com
//	@license.name				Proprietary
//	@license.url				http://www.example.com/license
//	@host						localhost:8080
//	@BasePath					/api/v1
//	@schemes					http https
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	_ "ping/production-line-api/docs"
	"ping/production-line-api/internal/config"
	"ping/production-line-api/internal/database"
	"ping/production-line-api/internal/external/openholidays"
	"ping/production-line-api/internal/handler"
	"ping/production-line-api/internal/logger"
	"ping/production-line-api/internal/mqtt"
	"ping/production-line-api/internal/repository"
	"ping/production-line-api/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Production Line API",
		zap.String("port", cfg.Port),
		zap.String("log_level", cfg.LogLevel),
		zap.String("mqtt_broker", cfg.MQTTBrokerURL),
	)

	// Connect to database
	ctx := context.Background()
	log.Info("Connecting to database...")
	dbConfig := database.DefaultConfig(cfg.DatabaseURL)
	db, err := database.Connect(ctx, dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close(db)
	log.Info("Database connection established")

	// Create repositories
	lineRepo := repository.NewLineRepository(db)
	statusLogRepo := repository.NewStatusLogRepository(db)
	labelRepo := repository.NewLabelRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	complianceRepo := repository.NewComplianceRepository(db, scheduleRepo)
	deviceRepo := repository.NewDeviceRepository(db)

	// Setup MQTT client
	log.Info("Connecting to MQTT broker...", zap.String("broker", cfg.MQTTBrokerURL))
	mqttConfig := &mqtt.Config{
		BrokerURL: cfg.MQTTBrokerURL,
		ClientID:  cfg.MQTTClientID,
		QOS:       cfg.MQTTQOS,
	}
	mqttClient, err := mqtt.NewClient(mqttConfig, log)
	if err != nil {
		log.Fatal("Failed to create MQTT client", zap.Error(err))
	}
	defer mqttClient.Disconnect()
	log.Info("MQTT client connected")

	// Setup MQTT publisher and device discovery handler
	mqttPublisher := mqtt.NewPublisher(mqttClient, log)
	deviceDiscoveryHandler := mqtt.NewDeviceDiscoveryHandler(deviceRepo, mqttPublisher, log)

	// Setup external clients
	var holidaysClient *openholidays.Client
	if cfg.HolidaysAPIEnabled {
		holidaysClient = openholidays.NewClient(log)
		log.Info("Holidays API client enabled", zap.String("country", cfg.HolidaysCountryCode))
	}

	// Create services
	lineService := service.NewLineService(lineRepo, statusLogRepo, deviceRepo, mqttPublisher, log)
	labelService := service.NewLabelService(labelRepo, lineRepo, mqttPublisher, log)
	analyticsService := service.NewAnalyticsService(analyticsRepo, lineRepo, log)
	scheduleService := service.NewScheduleService(scheduleRepo, lineRepo, holidaysClient, cfg.HolidaysCountryCode, log)
	complianceService := service.NewComplianceService(complianceRepo, lineRepo, log)

	// Setup MQTT subscriber (after services are created)
	mqttSubscriber := mqtt.NewSubscriber(mqttClient, lineService, deviceDiscoveryHandler, log)

	// Start MQTT subscriber
	if err := mqttSubscriber.Start(); err != nil {
		log.Fatal("Failed to start MQTT subscriber", zap.Error(err))
	}
	log.Info("MQTT subscriber started")

	// Create device handler (uses repository directly, no service layer)
	deviceHandler := handler.NewDeviceHandler(deviceRepo, mqttPublisher, log)

	// Setup router
	router := handler.NewRouter(
		lineService,
		labelService,
		analyticsService,
		scheduleService,
		complianceService,
		deviceHandler,
		cfg.CORSAllowedOrigins,
		log,
	)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Info("HTTP server starting", zap.String("addr", srv.Addr))
		serverErrors <- srv.ListenAndServe()
	}()

	// Wait for interrupt signal or server error
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatal("Server error", zap.Error(err))

	case sig := <-shutdown:
		log.Info("Shutdown signal received", zap.String("signal", sig.String()))

		// Create context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		// Gracefully shutdown the server
		if err := srv.Shutdown(ctx); err != nil {
			log.Error("Graceful shutdown failed", zap.Error(err))
			if err := srv.Close(); err != nil {
				log.Error("Failed to close server", zap.Error(err))
			}
		}

		log.Info("Server stopped gracefully")
	}
}
