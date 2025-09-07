package mq

import (
	"time"

	"example-api-template/internal/domain"
	"example-api-template/internal/repository"
	"example-api-template/internal/usecase"
)

// Test fixtures shared between consumer and producer tests

// createTestExample creates a test example entity
func createTestExample() *domain.Example {
	example, _ := domain.NewExample("test-id", "John Doe", "john@example.com", 30)
	return example
}

// createTestExampleWithMetadata creates a test example with external metadata
func createTestExampleWithMetadata() *usecase.ExampleWithMetadata {
	return &usecase.ExampleWithMetadata{
		Example: createTestExample(),
		ExternalData: &repository.ExternalExampleData{
			ExternalID: "ext_test_123",
			Metadata: map[string]string{
				"source":  "test",
				"version": "1.0",
			},
			Score:        0.85,
			LastModified: time.Now(),
		},
	}
}

// createTestEvent creates a test event of the specified type
func createTestEvent(eventType EventType) *ExampleEvent {
	return &ExampleEvent{
		ID:        "evt_test_123",
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      createTestExampleWithMetadata(),
		Metadata: map[string]interface{}{
			"source":   "test",
			"version":  "1.0",
			"trace_id": "test_trace_123",
		},
	}
}
