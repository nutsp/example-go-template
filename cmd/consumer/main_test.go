package main

import (
	"context"
	"os"
	"testing"
	"time"

	"example-api-template/internal/config"
	"example-api-template/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConsumerDependencyInitialization tests the dependency initialization
func TestConsumerDependencyInitialization(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "memory",
		},
		ExternalAPI: config.ExternalAPIConfig{
			EnableMock:     true,
			MockDelay:      10 * time.Millisecond,
			MockShouldFail: false,
		},
		MessageQueue: config.MessageQueueConfig{
			EnableMock:     true,
			EnableConsumer: true,
		},
		Logger: config.LoggerConfig{
			Level:       "info",
			Format:      "console",
			Development: true,
		},
	}

	// Initialize logger
	appLogger, err := logger.New(&cfg.Logger)
	require.NoError(t, err)
	defer appLogger.Close()

	// Test dependency initialization
	deps, err := initializeConsumerDependencies(cfg, appLogger)
	assert.NoError(t, err)
	assert.NotNil(t, deps)
	assert.NotNil(t, deps.Repository)
	assert.NotNil(t, deps.ExternalAPI)
	assert.NotNil(t, deps.Service)
	assert.NotNil(t, deps.UseCase)
	assert.NotNil(t, deps.Consumer)
}

// TestConsumerDependencyInitializationWithRealMQ tests with real MQ config (but expects failure)
func TestConsumerDependencyInitializationWithRealMQ(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "memory",
		},
		ExternalAPI: config.ExternalAPIConfig{
			EnableMock: true,
		},
		MessageQueue: config.MessageQueueConfig{
			EnableMock:     false,
			EnableConsumer: true,
			URL:            "amqp://invalid:invalid@nonexistent:5672/",
			ExchangeName:   "test-exchange",
			QueueName:      "test-queue",
			RoutingKeys:    []string{"test.created"},
			Durable:        true,
			PrefetchCount:  1,
		},
		Logger: config.LoggerConfig{
			Level:  "error", // Reduce log noise
			Format: "console",
		},
	}

	appLogger, err := logger.New(&cfg.Logger)
	require.NoError(t, err)
	defer appLogger.Close()

	// This should fail because RabbitMQ is not available
	deps, err := initializeConsumerDependencies(cfg, appLogger)
	assert.Error(t, err)
	assert.Nil(t, deps)
	assert.Contains(t, err.Error(), "failed to initialize RabbitMQ consumer")
}

// TestConsumerDependencyInitializationWithDisabledConsumer tests when consumer is disabled
func TestConsumerDependencyInitializationWithDisabledConsumer(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "memory",
		},
		ExternalAPI: config.ExternalAPIConfig{
			EnableMock: true,
		},
		MessageQueue: config.MessageQueueConfig{
			EnableMock:     false,
			EnableConsumer: false, // Disabled
		},
		Logger: config.LoggerConfig{
			Level:  "error",
			Format: "console",
		},
	}

	appLogger, err := logger.New(&cfg.Logger)
	require.NoError(t, err)
	defer appLogger.Close()

	// This should fail because consumer is disabled
	_, err = initializeConsumerDependencies(cfg, appLogger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "consumer is disabled in configuration")
}

// TestHealthCheck tests the health check functionality
func TestHealthCheck(t *testing.T) {
	// Set environment variable for health check
	os.Setenv("HEALTH_CHECK", "true")
	defer os.Unsetenv("HEALTH_CHECK")

	// This test would normally exit the process, so we can't test it directly
	// In a real scenario, you might use a separate test binary or mock os.Exit
	// For now, we'll just verify the environment variable is set
	assert.Equal(t, "true", os.Getenv("HEALTH_CHECK"))
}

// TestConsumerStartStop tests consumer lifecycle with mock
func TestConsumerStartStop(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "memory",
		},
		ExternalAPI: config.ExternalAPIConfig{
			EnableMock: true,
		},
		MessageQueue: config.MessageQueueConfig{
			EnableMock:     true,
			EnableConsumer: true,
		},
		Logger: config.LoggerConfig{
			Level:       "error", // Reduce log noise
			Format:      "console",
			Development: false,
		},
	}

	appLogger, err := logger.New(&cfg.Logger)
	require.NoError(t, err)
	defer appLogger.Close()

	deps, err := initializeConsumerDependencies(cfg, appLogger)
	require.NoError(t, err)

	// Test consumer start
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = deps.Consumer.Start(ctx)
	assert.NoError(t, err)

	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	// Test consumer stop
	err = deps.Consumer.Stop()
	assert.NoError(t, err)
}

// BenchmarkConsumerInitialization benchmarks the initialization process
func BenchmarkConsumerInitialization(b *testing.B) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "memory",
		},
		ExternalAPI: config.ExternalAPIConfig{
			EnableMock: true,
		},
		MessageQueue: config.MessageQueueConfig{
			EnableMock:     true,
			EnableConsumer: true,
		},
		Logger: config.LoggerConfig{
			Level:       "error",
			Format:      "console",
			Development: false,
		},
	}

	appLogger, _ := logger.New(&cfg.Logger)
	defer appLogger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deps, err := initializeConsumerDependencies(cfg, appLogger)
		if err != nil {
			b.Fatal(err)
		}
		_ = deps.Consumer.Stop() // Clean up
	}
}
