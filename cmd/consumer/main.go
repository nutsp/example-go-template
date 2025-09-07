package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"example-api-template/internal/config"
	"example-api-template/internal/repository"
	"example-api-template/internal/service"
	"example-api-template/internal/transport/mq"
	"example-api-template/internal/usecase"
	"example-api-template/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	appLogger, err := logger.New(&cfg.Logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer appLogger.Close()

	// Set global logger
	logger.SetGlobal(appLogger)

	appLogger.Info("Starting message queue consumer",
		zap.String("name", cfg.App.Name+"-consumer"),
		zap.String("version", cfg.App.Version),
		zap.String("environment", cfg.App.Environment),
	)

	// Initialize dependencies for event handling
	deps, err := initializeConsumerDependencies(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize consumer dependencies", zap.Error(err))
	}

	// Start consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := deps.Consumer.Start(ctx); err != nil {
		appLogger.Fatal("Failed to start message queue consumer", zap.Error(err))
	}

	appLogger.Info("Message queue consumer started successfully")

	// Wait for interrupt signal to gracefully shutdown the consumer
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down consumer...")

	// Cancel context to stop consumer
	cancel()

	// Stop consumer gracefully
	if err := deps.Consumer.Stop(); err != nil {
		appLogger.Error("Failed to stop consumer gracefully", zap.Error(err))
	} else {
		appLogger.Info("Consumer stopped gracefully")
	}

	appLogger.Info("Consumer shutdown complete")
}

// ConsumerDependencies holds all dependencies needed for the consumer
type ConsumerDependencies struct {
	Repository  repository.ExampleRepository
	ExternalAPI repository.ExternalExampleAPI
	Service     service.ExampleService
	UseCase     usecase.ExampleUseCase
	Consumer    mq.ExampleConsumer
}

// initializeConsumerDependencies initializes all dependencies needed for the consumer
func initializeConsumerDependencies(cfg *config.Config, logger *logger.Logger) (*ConsumerDependencies, error) {
	// Initialize repository (needed for event handlers that might need to fetch data)
	var repo repository.ExampleRepository
	switch cfg.Database.Type {
	case "memory":
		repo = repository.NewInMemoryExampleRepository()
		logger.Info("Using in-memory repository for consumer")
	default:
		// In a real application, you would initialize database connections here
		// For now, fall back to in-memory
		repo = repository.NewInMemoryExampleRepository()
		logger.Warn("Unsupported database type, falling back to in-memory repository",
			zap.String("type", cfg.Database.Type))
	}

	// Initialize external API (might be needed for event processing)
	var externalAPI repository.ExternalExampleAPI
	if cfg.ExternalAPI.EnableMock {
		externalAPI = repository.NewMockExternalExampleAPI(
			cfg.ExternalAPI.MockShouldFail,
			cfg.ExternalAPI.MockDelay,
		)
		logger.Info("Using mock external API for consumer")
	} else {
		// In a real application, you would initialize the actual external API client here
		externalAPI = repository.NewMockExternalExampleAPI(false, 100)
		logger.Warn("Real external API not implemented, using mock for consumer")
	}

	// Initialize service
	svc := service.NewExampleService(repo, logger.Logger)

	// Initialize use case
	uc := usecase.NewExampleUseCase(svc, externalAPI, logger.Logger)

	// Initialize message queue consumer
	var consumer mq.ExampleConsumer

	if cfg.MessageQueue.EnableMock {
		// Use mock implementation
		eventHandler := mq.NewDefaultExampleEventHandler(uc, logger.Logger)
		consumer = mq.NewMockConsumer(eventHandler, logger.Logger)
		logger.Info("Using mock message queue consumer")
	} else {
		// Use real RabbitMQ implementation
		if !cfg.MessageQueue.EnableConsumer {
			return nil, fmt.Errorf("consumer is disabled in configuration")
		}

		consumerConfig := &mq.RabbitMQConsumerConfig{
			URL:           cfg.MessageQueue.URL,
			ExchangeName:  cfg.MessageQueue.ExchangeName,
			QueueName:     cfg.MessageQueue.QueueName,
			RoutingKeys:   cfg.MessageQueue.RoutingKeys,
			Durable:       cfg.MessageQueue.Durable,
			AutoDelete:    cfg.MessageQueue.AutoDelete,
			Exclusive:     cfg.MessageQueue.Exclusive,
			NoWait:        cfg.MessageQueue.NoWait,
			PrefetchCount: cfg.MessageQueue.PrefetchCount,
		}

		eventHandler := mq.NewDefaultExampleEventHandler(uc, logger.Logger)
		var err error
		consumer, err = mq.NewRabbitMQConsumer(consumerConfig, eventHandler, logger.Logger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize RabbitMQ consumer: %w", err)
		}
		logger.Info("Using RabbitMQ consumer")
	}

	return &ConsumerDependencies{
		Repository:  repo,
		ExternalAPI: externalAPI,
		Service:     svc,
		UseCase:     uc,
		Consumer:    consumer,
	}, nil
}

// Health check for the consumer application
func init() {
	// Ensure the consumer application can start properly
	if os.Getenv("HEALTH_CHECK") == "true" {
		fmt.Println("OK")
		os.Exit(0)
	}
}
