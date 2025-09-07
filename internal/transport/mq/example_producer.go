package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"example-api-template/internal/domain"
	"example-api-template/internal/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// EventType represents different types of events
type EventType string

const (
	EventTypeExampleCreated EventType = "example.created"
	EventTypeExampleUpdated EventType = "example.updated"
	EventTypeExampleDeleted EventType = "example.deleted"
)

// ExampleEvent represents an event related to an example
type ExampleEvent struct {
	ID        string                       `json:"id"`
	Type      EventType                    `json:"type"`
	Timestamp time.Time                    `json:"timestamp"`
	Data      *usecase.ExampleWithMetadata `json:"data,omitempty"`
	Metadata  map[string]interface{}       `json:"metadata,omitempty"`
}

// ExampleDeletedEventData represents data for deletion events
type ExampleDeletedEventData struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ExampleProducer defines the interface for publishing example events
type ExampleProducer interface {
	PublishExampleCreated(ctx context.Context, example *usecase.ExampleWithMetadata) error
	PublishExampleUpdated(ctx context.Context, example *usecase.ExampleWithMetadata) error
	PublishExampleDeleted(ctx context.Context, exampleID, email, name string) error
	Close() error
}

// RabbitMQProducer implements ExampleProducer using RabbitMQ
type RabbitMQProducer struct {
	connection    *amqp.Connection
	channel       *amqp.Channel
	exchangeName  string
	routingPrefix string
	logger        *zap.Logger
}

// RabbitMQProducerConfig holds configuration for RabbitMQ producer
type RabbitMQProducerConfig struct {
	URL           string
	ExchangeName  string
	RoutingPrefix string
	Durable       bool
	AutoDelete    bool
}

// NewRabbitMQProducer creates a new RabbitMQ producer
func NewRabbitMQProducer(config *RabbitMQProducerConfig, logger *zap.Logger) (*RabbitMQProducer, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		config.ExchangeName, // name
		"topic",             // type
		config.Durable,      // durable
		config.AutoDelete,   // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	producer := &RabbitMQProducer{
		connection:    conn,
		channel:       ch,
		exchangeName:  config.ExchangeName,
		routingPrefix: config.RoutingPrefix,
		logger:        logger,
	}

	// Set up connection close handler
	go producer.handleConnectionClose()

	logger.Info("RabbitMQ producer initialized",
		zap.String("exchange", config.ExchangeName),
		zap.String("routing_prefix", config.RoutingPrefix),
	)

	return producer, nil
}

// PublishExampleCreated publishes an example created event
func (p *RabbitMQProducer) PublishExampleCreated(ctx context.Context, example *usecase.ExampleWithMetadata) error {
	event := &ExampleEvent{
		ID:        generateEventID(),
		Type:      EventTypeExampleCreated,
		Timestamp: time.Now(),
		Data:      example,
		Metadata: map[string]interface{}{
			"source":   "example-api",
			"version":  "1.0",
			"user_id":  extractUserID(ctx),
			"trace_id": extractTraceID(ctx),
		},
	}

	routingKey := fmt.Sprintf("%s.%s", p.routingPrefix, EventTypeExampleCreated)
	return p.publishEvent(ctx, event, routingKey)
}

// PublishExampleUpdated publishes an example updated event
func (p *RabbitMQProducer) PublishExampleUpdated(ctx context.Context, example *usecase.ExampleWithMetadata) error {
	event := &ExampleEvent{
		ID:        generateEventID(),
		Type:      EventTypeExampleUpdated,
		Timestamp: time.Now(),
		Data:      example,
		Metadata: map[string]interface{}{
			"source":   "example-api",
			"version":  "1.0",
			"user_id":  extractUserID(ctx),
			"trace_id": extractTraceID(ctx),
		},
	}

	routingKey := fmt.Sprintf("%s.%s", p.routingPrefix, EventTypeExampleUpdated)
	return p.publishEvent(ctx, event, routingKey)
}

// PublishExampleDeleted publishes an example deleted event
func (p *RabbitMQProducer) PublishExampleDeleted(ctx context.Context, exampleID, email, name string) error {
	event := &ExampleEvent{
		ID:        generateEventID(),
		Type:      EventTypeExampleDeleted,
		Timestamp: time.Now(),
		Data: &usecase.ExampleWithMetadata{
			Example: &domain.Example{
				ID:    exampleID,
				Name:  name,
				Email: email,
			},
		},
		Metadata: map[string]interface{}{
			"source":   "example-api",
			"version":  "1.0",
			"user_id":  extractUserID(ctx),
			"trace_id": extractTraceID(ctx),
		},
	}

	routingKey := fmt.Sprintf("%s.%s", p.routingPrefix, EventTypeExampleDeleted)
	return p.publishEvent(ctx, event, routingKey)
}

// publishEvent publishes an event to the message queue
func (p *RabbitMQProducer) publishEvent(ctx context.Context, event *ExampleEvent, routingKey string) error {
	body, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("Failed to marshal event", zap.Error(err), zap.String("event_id", event.ID))
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Set publishing options
	publishing := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // Make message persistent
		MessageId:    event.ID,
		Timestamp:    event.Timestamp,
		Type:         string(event.Type),
		Headers: amqp.Table{
			"source":   "example-api",
			"version":  "1.0",
			"user_id":  extractUserID(ctx),
			"trace_id": extractTraceID(ctx),
		},
		Body: body,
	}

	// Set timeout for publishing
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = p.channel.PublishWithContext(
		publishCtx,
		p.exchangeName, // exchange
		routingKey,     // routing key
		false,          // mandatory
		false,          // immediate
		publishing,
	)

	if err != nil {
		p.logger.Error("Failed to publish event",
			zap.Error(err),
			zap.String("event_id", event.ID),
			zap.String("event_type", string(event.Type)),
			zap.String("routing_key", routingKey),
		)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Info("Event published successfully",
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.Type)),
		zap.String("routing_key", routingKey),
	)

	return nil
}

// Close closes the producer connection
func (p *RabbitMQProducer) Close() error {
	var errs []error

	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if p.connection != nil {
		if err := p.connection.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	p.logger.Info("RabbitMQ producer closed successfully")
	return nil
}

// handleConnectionClose handles connection close events
func (p *RabbitMQProducer) handleConnectionClose() {
	closeError := <-p.connection.NotifyClose(make(chan *amqp.Error))
	if closeError != nil {
		p.logger.Error("RabbitMQ connection closed unexpectedly",
			zap.Error(closeError),
		)
	}
}

// MockProducer is a mock implementation for testing
type MockProducer struct {
	events []ExampleEvent
	logger *zap.Logger
}

// NewMockProducer creates a new mock producer
func NewMockProducer(logger *zap.Logger) *MockProducer {
	return &MockProducer{
		events: make([]ExampleEvent, 0),
		logger: logger,
	}
}

// PublishExampleCreated mock implementation
func (m *MockProducer) PublishExampleCreated(ctx context.Context, example *usecase.ExampleWithMetadata) error {
	event := ExampleEvent{
		ID:        generateEventID(),
		Type:      EventTypeExampleCreated,
		Timestamp: time.Now(),
		Data:      example,
	}
	m.events = append(m.events, event)
	m.logger.Info("Mock: Example created event published", zap.String("example_id", example.ID))
	return nil
}

// PublishExampleUpdated mock implementation
func (m *MockProducer) PublishExampleUpdated(ctx context.Context, example *usecase.ExampleWithMetadata) error {
	event := ExampleEvent{
		ID:        generateEventID(),
		Type:      EventTypeExampleUpdated,
		Timestamp: time.Now(),
		Data:      example,
	}
	m.events = append(m.events, event)
	m.logger.Info("Mock: Example updated event published", zap.String("example_id", example.ID))
	return nil
}

// PublishExampleDeleted mock implementation
func (m *MockProducer) PublishExampleDeleted(ctx context.Context, exampleID, email, name string) error {
	event := ExampleEvent{
		ID:        generateEventID(),
		Type:      EventTypeExampleDeleted,
		Timestamp: time.Now(),
		Data: &usecase.ExampleWithMetadata{
			Example: &domain.Example{
				ID:    exampleID,
				Name:  name,
				Email: email,
			},
		},
	}
	m.events = append(m.events, event)
	m.logger.Info("Mock: Example deleted event published", zap.String("example_id", exampleID))
	return nil
}

// Close mock implementation
func (m *MockProducer) Close() error {
	m.logger.Info("Mock producer closed")
	return nil
}

// GetEvents returns all published events (for testing)
func (m *MockProducer) GetEvents() []ExampleEvent {
	return m.events
}

// ClearEvents clears all published events (for testing)
func (m *MockProducer) ClearEvents() {
	m.events = m.events[:0]
}

// Helper functions

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// extractUserID extracts user ID from context
func extractUserID(ctx context.Context) string {
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return "system"
}

// extractTraceID extracts trace ID from context
func extractTraceID(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}
