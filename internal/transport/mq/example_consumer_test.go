package mq

import (
	"context"
	"encoding/json"
	"testing"

	"example-api-template/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockEventHandler for testing
type MockEventHandler struct {
	mock.Mock
}

func (m *MockEventHandler) HandleExampleCreated(ctx context.Context, event *ExampleEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventHandler) HandleExampleUpdated(ctx context.Context, event *ExampleEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventHandler) HandleExampleDeleted(ctx context.Context, event *ExampleEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// TestDefaultExampleEventHandler tests the default event handler
func TestDefaultExampleEventHandler(t *testing.T) {
	mockUseCase := &mocks.MockExampleService{} // Using service mock for simplicity
	logger := zap.NewNop()
	handler := NewDefaultExampleEventHandler(nil, logger) // UseCase can be nil for this test

	tests := []struct {
		name      string
		eventType EventType
		testFunc  func(*testing.T, *DefaultExampleEventHandler, *ExampleEvent)
	}{
		{
			name:      "handle example created",
			eventType: EventTypeExampleCreated,
			testFunc: func(t *testing.T, h *DefaultExampleEventHandler, event *ExampleEvent) {
				err := h.HandleExampleCreated(context.Background(), event)
				assert.NoError(t, err)
			},
		},
		{
			name:      "handle example updated",
			eventType: EventTypeExampleUpdated,
			testFunc: func(t *testing.T, h *DefaultExampleEventHandler, event *ExampleEvent) {
				err := h.HandleExampleUpdated(context.Background(), event)
				assert.NoError(t, err)
			},
		},
		{
			name:      "handle example deleted",
			eventType: EventTypeExampleDeleted,
			testFunc: func(t *testing.T, h *DefaultExampleEventHandler, event *ExampleEvent) {
				err := h.HandleExampleDeleted(context.Background(), event)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := createTestEvent(tt.eventType)
			tt.testFunc(t, handler, event)
		})
	}

	mockUseCase.AssertExpectations(t)
}

// TestMockConsumer tests the mock consumer implementation
func TestMockConsumer(t *testing.T) {
	mockHandler := &MockEventHandler{}
	logger := zap.NewNop()
	consumer := NewMockConsumer(mockHandler, logger)

	// Test Start
	t.Run("start consumer", func(t *testing.T) {
		ctx := context.Background()
		err := consumer.Start(ctx)
		assert.NoError(t, err)
		assert.True(t, consumer.IsRunning())
	})

	// Test Stop
	t.Run("stop consumer", func(t *testing.T) {
		err := consumer.Stop()
		assert.NoError(t, err)
		assert.False(t, consumer.IsRunning())
	})

	// Test SimulateEvent
	t.Run("simulate events", func(t *testing.T) {
		// Restart consumer
		ctx := context.Background()
		err := consumer.Start(ctx)
		require.NoError(t, err)

		// Test created event
		createdEvent := createTestEvent(EventTypeExampleCreated)
		mockHandler.On("HandleExampleCreated", mock.Anything, createdEvent).Return(nil)

		err = consumer.SimulateEvent(ctx, createdEvent)
		assert.NoError(t, err)

		// Test updated event
		updatedEvent := createTestEvent(EventTypeExampleUpdated)
		mockHandler.On("HandleExampleUpdated", mock.Anything, updatedEvent).Return(nil)

		err = consumer.SimulateEvent(ctx, updatedEvent)
		assert.NoError(t, err)

		// Test deleted event
		deletedEvent := createTestEvent(EventTypeExampleDeleted)
		mockHandler.On("HandleExampleDeleted", mock.Anything, deletedEvent).Return(nil)

		err = consumer.SimulateEvent(ctx, deletedEvent)
		assert.NoError(t, err)

		// Verify events were processed
		events := consumer.GetProcessedEvents()
		assert.Len(t, events, 3)
		assert.Equal(t, EventTypeExampleCreated, events[0].Type)
		assert.Equal(t, EventTypeExampleUpdated, events[1].Type)
		assert.Equal(t, EventTypeExampleDeleted, events[2].Type)

		// Clear events
		consumer.ClearProcessedEvents()
		events = consumer.GetProcessedEvents()
		assert.Len(t, events, 0)

		mockHandler.AssertExpectations(t)
	})

	// Test unknown event type
	t.Run("unknown event type", func(t *testing.T) {
		unknownEvent := createTestEvent("unknown.event")

		err := consumer.SimulateEvent(context.Background(), unknownEvent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown event type")
	})
}

// TestRabbitMQConsumerConfig tests the consumer configuration
func TestRabbitMQConsumerConfig(t *testing.T) {
	config := &RabbitMQConsumerConfig{
		URL:           "amqp://guest:guest@localhost:5672/",
		ExchangeName:  "test-exchange",
		QueueName:     "test-queue",
		RoutingKeys:   []string{"test.created", "test.updated", "test.deleted"},
		Durable:       true,
		AutoDelete:    false,
		Exclusive:     false,
		NoWait:        false,
		PrefetchCount: 10,
	}

	assert.Equal(t, "amqp://guest:guest@localhost:5672/", config.URL)
	assert.Equal(t, "test-exchange", config.ExchangeName)
	assert.Equal(t, "test-queue", config.QueueName)
	assert.Len(t, config.RoutingKeys, 3)
	assert.True(t, config.Durable)
	assert.False(t, config.AutoDelete)
	assert.Equal(t, 10, config.PrefetchCount)
}

// TestEventHandling tests event handling scenarios
func TestEventHandling(t *testing.T) {
	mockHandler := &MockEventHandler{}
	logger := zap.NewNop()
	consumer := NewMockConsumer(mockHandler, logger)

	ctx := context.Background()
	err := consumer.Start(ctx)
	require.NoError(t, err)

	tests := []struct {
		name          string
		event         *ExampleEvent
		setupMock     func(*MockEventHandler)
		expectError   bool
		errorContains string
	}{
		{
			name:  "successful event handling",
			event: createTestEvent(EventTypeExampleCreated),
			setupMock: func(m *MockEventHandler) {
				m.On("HandleExampleCreated", mock.Anything, mock.AnythingOfType("*mq.ExampleEvent")).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:  "handler returns error",
			event: createTestEvent(EventTypeExampleUpdated),
			setupMock: func(m *MockEventHandler) {
				m.On("HandleExampleUpdated", mock.Anything, mock.AnythingOfType("*mq.ExampleEvent")).
					Return(assert.AnError)
			},
			expectError:   true,
			errorContains: "assert.AnError",
		},
		{
			name:          "unknown event type",
			event:         createTestEvent("invalid.type"),
			setupMock:     func(m *MockEventHandler) {},
			expectError:   true,
			errorContains: "unknown event type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockHandler.ExpectedCalls = nil
			mockHandler.Calls = nil
			tt.setupMock(mockHandler)

			err := consumer.SimulateEvent(ctx, tt.event)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockHandler.AssertExpectations(t)
		})
	}
}

// TestConsumerLifecycle tests the consumer lifecycle
func TestConsumerLifecycle(t *testing.T) {
	mockHandler := &MockEventHandler{}
	logger := zap.NewNop()
	consumer := NewMockConsumer(mockHandler, logger)

	// Test initial state
	assert.False(t, consumer.IsRunning())

	// Test start
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := consumer.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, consumer.IsRunning())

	// Test double start (should not error but should not change state)
	err = consumer.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, consumer.IsRunning())

	// Test stop
	err = consumer.Stop()
	assert.NoError(t, err)
	assert.False(t, consumer.IsRunning())

	// Test double stop (should not error)
	err = consumer.Stop()
	assert.NoError(t, err)
	assert.False(t, consumer.IsRunning())
}

// TestEventSerialization tests event JSON serialization/deserialization
func TestEventSerialization(t *testing.T) {
	originalEvent := createTestEvent(EventTypeExampleCreated)

	// Serialize
	data, err := json.Marshal(originalEvent)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Deserialize
	var deserializedEvent ExampleEvent
	err = json.Unmarshal(data, &deserializedEvent)
	assert.NoError(t, err)

	// Verify
	assert.Equal(t, originalEvent.ID, deserializedEvent.ID)
	assert.Equal(t, originalEvent.Type, deserializedEvent.Type)
	assert.Equal(t, originalEvent.Data.ID, deserializedEvent.Data.ID)
	assert.Equal(t, originalEvent.Data.Name, deserializedEvent.Data.Name)
	assert.Equal(t, originalEvent.Data.Email, deserializedEvent.Data.Email)
}

// TestHelperFunctions tests utility functions
func TestHelperFunctions(t *testing.T) {
	t.Run("contains function", func(t *testing.T) {
		tests := []struct {
			str    string
			substr string
			want   bool
		}{
			{"hello world", "hello", true},
			{"hello world", "world", true},
			{"hello world", "lo wo", true},
			{"hello world", "xyz", false},
			{"", "test", false},
			{"test", "", true}, // Empty substring should be found
		}

		for _, tt := range tests {
			result := contains(tt.str, tt.substr)
			assert.Equal(t, tt.want, result,
				"contains(%q, %q) = %v, want %v", tt.str, tt.substr, result, tt.want)
		}
	})

	t.Run("containsInMiddle function", func(t *testing.T) {
		tests := []struct {
			str    string
			substr string
			want   bool
		}{
			{"hello world", "ello", true},
			{"hello world", "lo w", true},
			{"hello world", "hello", false}, // At start
			{"hello world", "world", false}, // At end
			{"hello world", "xyz", false},   // Not found
		}

		for _, tt := range tests {
			result := containsInMiddle(tt.str, tt.substr)
			assert.Equal(t, tt.want, result,
				"containsInMiddle(%q, %q) = %v, want %v", tt.str, tt.substr, result, tt.want)
		}
	})
}

// BenchmarkEventHandling benchmarks event processing
func BenchmarkEventHandling(b *testing.B) {
	mockHandler := &MockEventHandler{}
	logger := zap.NewNop()
	consumer := NewMockConsumer(mockHandler, logger)

	// Setup mock to always succeed
	mockHandler.On("HandleExampleCreated", mock.Anything, mock.AnythingOfType("*mq.ExampleEvent")).
		Return(nil)

	ctx := context.Background()
	consumer.Start(ctx)

	event := createTestEvent(EventTypeExampleCreated)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		consumer.SimulateEvent(ctx, event)
	}
}

// BenchmarkEventSerialization benchmarks JSON serialization
func BenchmarkEventSerialization(b *testing.B) {
	event := createTestEvent(EventTypeExampleCreated)

	b.Run("Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			json.Marshal(event)
		}
	})

	data, _ := json.Marshal(event)
	b.Run("Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var e ExampleEvent
			json.Unmarshal(data, &e)
		}
	})
}

// Integration test example (would require actual RabbitMQ for full test)
func TestRabbitMQConsumerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require a real RabbitMQ instance
	// For now, we'll just test that the constructor fails gracefully with invalid config

	mockHandler := &MockEventHandler{}
	logger := zap.NewNop()

	config := &RabbitMQConsumerConfig{
		URL:           "amqp://invalid:invalid@nonexistent:5672/",
		ExchangeName:  "test-exchange",
		QueueName:     "test-queue",
		RoutingKeys:   []string{"test.created"},
		Durable:       true,
		PrefetchCount: 1,
	}

	consumer, err := NewRabbitMQConsumer(config, mockHandler, logger)
	assert.Error(t, err)
	assert.Nil(t, consumer)
	assert.Contains(t, err.Error(), "failed to connect to RabbitMQ")
}

// Example test showing how to test with context cancellation
func TestConsumerContextCancellation(t *testing.T) {
	mockHandler := &MockEventHandler{}
	logger := zap.NewNop()
	consumer := NewMockConsumer(mockHandler, logger)

	ctx, cancel := context.WithCancel(context.Background())

	// Start consumer
	err := consumer.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, consumer.IsRunning())

	// Cancel context (simulates shutdown signal)
	cancel()

	// In a real implementation, the consumer would stop automatically
	// For mock, we manually stop it
	err = consumer.Stop()
	assert.NoError(t, err)
	assert.False(t, consumer.IsRunning())
}
