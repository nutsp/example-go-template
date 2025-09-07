package mq

import (
	"context"
	"testing"
	"time"

	"example-api-template/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestMockProducer tests the mock producer implementation
func TestMockProducer(t *testing.T) {
	logger := zap.NewNop()
	producer := NewMockProducer(logger)

	// Create test data
	example := createTestExampleWithMetadata()
	ctx := context.Background()

	t.Run("publish example created", func(t *testing.T) {
		err := producer.PublishExampleCreated(ctx, example)
		assert.NoError(t, err)

		events := producer.GetEvents()
		assert.Len(t, events, 1)
		assert.Equal(t, EventTypeExampleCreated, events[0].Type)
		assert.Equal(t, example.ID, events[0].Data.ID)
	})

	t.Run("publish example updated", func(t *testing.T) {
		producer.ClearEvents() // Clear previous events

		err := producer.PublishExampleUpdated(ctx, example)
		assert.NoError(t, err)

		events := producer.GetEvents()
		assert.Len(t, events, 1)
		assert.Equal(t, EventTypeExampleUpdated, events[0].Type)
		assert.Equal(t, example.ID, events[0].Data.ID)
	})

	t.Run("publish example deleted", func(t *testing.T) {
		producer.ClearEvents() // Clear previous events

		err := producer.PublishExampleDeleted(ctx, example.ID, example.Email, example.Name)
		assert.NoError(t, err)

		events := producer.GetEvents()
		assert.Len(t, events, 1)
		assert.Equal(t, EventTypeExampleDeleted, events[0].Type)
		assert.Equal(t, example.ID, events[0].Data.ID)
		assert.Equal(t, example.Email, events[0].Data.Email)
		assert.Equal(t, example.Name, events[0].Data.Name)
	})

	t.Run("close producer", func(t *testing.T) {
		err := producer.Close()
		assert.NoError(t, err)
	})

	t.Run("multiple events tracking", func(t *testing.T) {
		producer.ClearEvents()

		// Publish multiple events
		err := producer.PublishExampleCreated(ctx, example)
		assert.NoError(t, err)

		err = producer.PublishExampleUpdated(ctx, example)
		assert.NoError(t, err)

		err = producer.PublishExampleDeleted(ctx, example.ID, example.Email, example.Name)
		assert.NoError(t, err)

		events := producer.GetEvents()
		assert.Len(t, events, 3)
		assert.Equal(t, EventTypeExampleCreated, events[0].Type)
		assert.Equal(t, EventTypeExampleUpdated, events[1].Type)
		assert.Equal(t, EventTypeExampleDeleted, events[2].Type)
	})
}

// TestRabbitMQProducerConfig tests the producer configuration
func TestRabbitMQProducerConfig(t *testing.T) {
	config := &RabbitMQProducerConfig{
		URL:           "amqp://guest:guest@localhost:5672/",
		ExchangeName:  "test-exchange",
		RoutingPrefix: "test",
		Durable:       true,
		AutoDelete:    false,
	}

	assert.Equal(t, "amqp://guest:guest@localhost:5672/", config.URL)
	assert.Equal(t, "test-exchange", config.ExchangeName)
	assert.Equal(t, "test", config.RoutingPrefix)
	assert.True(t, config.Durable)
	assert.False(t, config.AutoDelete)
}

// TestEventGeneration tests event creation and metadata
func TestEventGeneration(t *testing.T) {
	logger := zap.NewNop()
	producer := NewMockProducer(logger)

	example := createTestExampleWithMetadata()
	ctx := context.WithValue(context.Background(), "user_id", "test-user-123")
	ctx = context.WithValue(ctx, "trace_id", "test-trace-456")

	err := producer.PublishExampleCreated(ctx, example)
	assert.NoError(t, err)

	events := producer.GetEvents()
	require.Len(t, events, 1)

	event := events[0]
	assert.NotEmpty(t, event.ID)
	assert.Equal(t, EventTypeExampleCreated, event.Type)
	assert.WithinDuration(t, time.Now(), event.Timestamp, time.Second)
	assert.Equal(t, example, event.Data)
	// Note: Mock producer doesn't set metadata like the real producer does
}

// TestHelperFunctions tests utility functions in producer
func TestProducerHelperFunctions(t *testing.T) {
	t.Run("generateEventID", func(t *testing.T) {
		id1 := generateEventID()
		time.Sleep(1 * time.Nanosecond) // Ensure different timestamps
		id2 := generateEventID()

		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2) // Should be unique
		assert.Contains(t, id1, "evt_")
		assert.Contains(t, id2, "evt_")

		// Test multiple IDs for uniqueness
		ids := make(map[string]bool)
		for i := 0; i < 10; i++ {
			id := generateEventID()
			assert.False(t, ids[id], "Generated duplicate ID: %s", id)
			ids[id] = true
			time.Sleep(1 * time.Nanosecond)
		}
	})

	t.Run("extractUserID", func(t *testing.T) {
		// Test with user ID in context
		ctx := context.WithValue(context.Background(), "user_id", "test-user-123")
		userID := extractUserID(ctx)
		assert.Equal(t, "test-user-123", userID)

		// Test without user ID in context
		ctx = context.Background()
		userID = extractUserID(ctx)
		assert.Equal(t, "system", userID)

		// Test with wrong type in context
		ctx = context.WithValue(context.Background(), "user_id", 123)
		userID = extractUserID(ctx)
		assert.Equal(t, "system", userID)
	})

	t.Run("extractTraceID", func(t *testing.T) {
		// Test with trace ID in context
		ctx := context.WithValue(context.Background(), "trace_id", "test-trace-456")
		traceID := extractTraceID(ctx)
		assert.Equal(t, "test-trace-456", traceID)

		// Test without trace ID in context
		ctx = context.Background()
		traceID = extractTraceID(ctx)
		assert.Equal(t, "", traceID)

		// Test with wrong type in context
		ctx = context.WithValue(context.Background(), "trace_id", 456)
		traceID = extractTraceID(ctx)
		assert.Equal(t, "", traceID)
	})
}

// TestEventTypes tests all event type constants
func TestEventTypes(t *testing.T) {
	assert.Equal(t, EventType("example.created"), EventTypeExampleCreated)
	assert.Equal(t, EventType("example.updated"), EventTypeExampleUpdated)
	assert.Equal(t, EventType("example.deleted"), EventTypeExampleDeleted)
}

// TestExampleEvent tests the event structure
func TestExampleEvent(t *testing.T) {
	event := &ExampleEvent{
		ID:        "test-event-id",
		Type:      EventTypeExampleCreated,
		Timestamp: time.Now(),
		Data:      createTestExampleWithMetadata(),
		Metadata: map[string]interface{}{
			"source":   "test",
			"version":  "1.0",
			"trace_id": "test-trace",
		},
	}

	assert.Equal(t, "test-event-id", event.ID)
	assert.Equal(t, EventTypeExampleCreated, event.Type)
	assert.NotNil(t, event.Data)
	assert.NotNil(t, event.Metadata)
	assert.Equal(t, "test", event.Metadata["source"])
	assert.Equal(t, "1.0", event.Metadata["version"])
	assert.Equal(t, "test-trace", event.Metadata["trace_id"])
}

// TestExternalExampleData tests the external data structure
func TestExternalExampleData(t *testing.T) {
	data := &repository.ExternalExampleData{
		ExternalID: "ext-123",
		Metadata: map[string]string{
			"source":  "external-api",
			"version": "2.0",
		},
		Score:        0.95,
		LastModified: time.Now(),
	}

	assert.Equal(t, "ext-123", data.ExternalID)
	assert.NotNil(t, data.Metadata)
	assert.Equal(t, "external-api", data.Metadata["source"])
	assert.Equal(t, "2.0", data.Metadata["version"])
	assert.Equal(t, 0.95, data.Score)
	assert.WithinDuration(t, time.Now(), data.LastModified, time.Second)
}

// BenchmarkEventGeneration benchmarks event creation
func BenchmarkEventGeneration(b *testing.B) {
	logger := zap.NewNop()
	producer := NewMockProducer(logger)
	example := createTestExampleWithMetadata()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		producer.PublishExampleCreated(ctx, example)
	}
}

// BenchmarkEventIDGeneration benchmarks ID generation
func BenchmarkEventIDGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateEventID()
	}
}

// Integration test for RabbitMQ producer (requires real RabbitMQ)
func TestRabbitMQProducerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require a real RabbitMQ instance
	// For now, we'll just test that the constructor fails gracefully with invalid config

	logger := zap.NewNop()
	config := &RabbitMQProducerConfig{
		URL:           "amqp://invalid:invalid@nonexistent:5672/",
		ExchangeName:  "test-exchange",
		RoutingPrefix: "test",
		Durable:       true,
		AutoDelete:    false,
	}

	producer, err := NewRabbitMQProducer(config, logger)
	assert.Error(t, err)
	assert.Nil(t, producer)
	assert.Contains(t, err.Error(), "failed to connect to RabbitMQ")
}
