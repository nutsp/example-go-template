package mq

import (
	"context"
	"encoding/json"
	"errors"
	"example-api-template/internal/usecase"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// ExampleEventHandler defines the interface for handling example events
type ExampleEventHandler interface {
	HandleExampleCreated(ctx context.Context, event *ExampleEvent) error
	HandleExampleUpdated(ctx context.Context, event *ExampleEvent) error
	HandleExampleDeleted(ctx context.Context, event *ExampleEvent) error
}

// ExampleConsumer defines the interface for consuming example events
type ExampleConsumer interface {
	Start(ctx context.Context) error
	Stop() error
}

// RabbitMQConsumer implements ExampleConsumer using RabbitMQ
type RabbitMQConsumer struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
	queueName    string
	routingKeys  []string
	handler      ExampleEventHandler
	logger       *zap.Logger
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex
	isRunning    bool
}

// RabbitMQConsumerConfig holds configuration for RabbitMQ consumer
type RabbitMQConsumerConfig struct {
	URL           string
	ExchangeName  string
	QueueName     string
	RoutingKeys   []string
	Durable       bool
	AutoDelete    bool
	Exclusive     bool
	NoWait        bool
	PrefetchCount int
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer
func NewRabbitMQConsumer(
	config *RabbitMQConsumerConfig,
	handler ExampleEventHandler,
	logger *zap.Logger,
) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS
	err = ch.Qos(config.PrefetchCount, 0, false)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
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

	// Declare queue
	queue, err := ch.QueueDeclare(
		config.QueueName,  // name
		config.Durable,    // durable
		config.AutoDelete, // delete when unused
		config.Exclusive,  // exclusive
		config.NoWait,     // no-wait
		nil,               // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange with routing keys
	for _, routingKey := range config.RoutingKeys {
		err = ch.QueueBind(
			queue.Name,          // queue name
			routingKey,          // routing key
			config.ExchangeName, // exchange
			false,               // no-wait
			nil,                 // arguments
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to bind queue to routing key %s: %w", routingKey, err)
		}
	}

	consumer := &RabbitMQConsumer{
		connection:   conn,
		channel:      ch,
		exchangeName: config.ExchangeName,
		queueName:    queue.Name,
		routingKeys:  config.RoutingKeys,
		handler:      handler,
		logger:       logger,
		stopChan:     make(chan struct{}),
	}

	logger.Info("RabbitMQ consumer initialized",
		zap.String("exchange", config.ExchangeName),
		zap.String("queue", queue.Name),
		zap.Strings("routing_keys", config.RoutingKeys),
	)

	return consumer, nil
}

// Start starts consuming messages
func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return errors.New("consumer is already running")
	}

	// Register consumer
	msgs, err := c.channel.Consume(
		c.queueName, // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	c.isRunning = true
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()
		c.logger.Info("Starting message consumption")

		for {
			select {
			case <-c.stopChan:
				c.logger.Info("Stopping message consumption")
				return
			case <-ctx.Done():
				c.logger.Info("Context cancelled, stopping message consumption")
				return
			case delivery, ok := <-msgs:
				if !ok {
					c.logger.Warn("Message channel closed")
					return
				}
				c.handleMessage(ctx, delivery)
			}
		}
	}()

	// Set up connection close handler
	go c.handleConnectionClose()

	c.logger.Info("Consumer started successfully")
	return nil
}

// Stop stops the consumer
func (c *RabbitMQConsumer) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning {
		return nil
	}

	c.logger.Info("Stopping consumer...")

	close(c.stopChan)
	c.wg.Wait()

	var errs []error

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if c.connection != nil {
		if err := c.connection.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	c.isRunning = false

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	c.logger.Info("Consumer stopped successfully")
	return nil
}

// handleMessage handles incoming messages
func (c *RabbitMQConsumer) handleMessage(ctx context.Context, delivery amqp.Delivery) {
	logger := c.logger.With(
		zap.String("message_id", delivery.MessageId),
		zap.String("routing_key", delivery.RoutingKey),
		zap.String("exchange", delivery.Exchange),
	)

	logger.Debug("Processing message")

	// Parse event
	var event ExampleEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		logger.Error("Failed to unmarshal event", zap.Error(err))
		c.rejectMessage(delivery, false)
		return
	}

	// Add message metadata to context
	msgCtx := context.WithValue(ctx, "message_id", delivery.MessageId)
	msgCtx = context.WithValue(msgCtx, "routing_key", delivery.RoutingKey)
	msgCtx = context.WithValue(msgCtx, "delivery_tag", delivery.DeliveryTag)

	// Handle event based on type
	var err error
	switch event.Type {
	case EventTypeExampleCreated:
		err = c.handler.HandleExampleCreated(msgCtx, &event)
	case EventTypeExampleUpdated:
		err = c.handler.HandleExampleUpdated(msgCtx, &event)
	case EventTypeExampleDeleted:
		err = c.handler.HandleExampleDeleted(msgCtx, &event)
	default:
		logger.Warn("Unknown event type", zap.String("event_type", string(event.Type)))
		c.ackMessage(delivery)
		return
	}

	if err != nil {
		logger.Error("Failed to handle event",
			zap.Error(err),
			zap.String("event_type", string(event.Type)),
			zap.String("event_id", event.ID),
		)

		// Check if this is a retryable error
		if c.isRetryableError(err) {
			c.rejectMessage(delivery, true) // Requeue for retry
		} else {
			c.rejectMessage(delivery, false) // Don't requeue
		}
		return
	}

	// Acknowledge successful processing
	c.ackMessage(delivery)
	logger.Info("Event processed successfully",
		zap.String("event_type", string(event.Type)),
		zap.String("event_id", event.ID),
	)
}

// ackMessage acknowledges a message
func (c *RabbitMQConsumer) ackMessage(delivery amqp.Delivery) {
	if err := delivery.Ack(false); err != nil {
		c.logger.Error("Failed to ack message",
			zap.Error(err),
			zap.String("message_id", delivery.MessageId),
		)
	}
}

// rejectMessage rejects a message
func (c *RabbitMQConsumer) rejectMessage(delivery amqp.Delivery, requeue bool) {
	if err := delivery.Reject(requeue); err != nil {
		c.logger.Error("Failed to reject message",
			zap.Error(err),
			zap.String("message_id", delivery.MessageId),
			zap.Bool("requeue", requeue),
		)
	}
}

// isRetryableError determines if an error is retryable
func (c *RabbitMQConsumer) isRetryableError(err error) bool {
	// Add logic to determine if error is retryable
	// For example, network errors might be retryable, but validation errors are not
	retryableErrors := []string{
		"connection",
		"timeout",
		"temporary",
		"unavailable",
	}

	errStr := err.Error()
	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}

	return false
}

// handleConnectionClose handles connection close events
func (c *RabbitMQConsumer) handleConnectionClose() {
	closeError := <-c.connection.NotifyClose(make(chan *amqp.Error))
	if closeError != nil {
		c.logger.Error("RabbitMQ connection closed unexpectedly",
			zap.Error(closeError),
		)
		// In production, you might want to implement reconnection logic here
	}
}

// DefaultExampleEventHandler provides a default implementation of ExampleEventHandler
type DefaultExampleEventHandler struct {
	useCase usecase.ExampleUseCase
	logger  *zap.Logger
}

// NewDefaultExampleEventHandler creates a new default event handler
func NewDefaultExampleEventHandler(useCase usecase.ExampleUseCase, logger *zap.Logger) *DefaultExampleEventHandler {
	return &DefaultExampleEventHandler{
		useCase: useCase,
		logger:  logger,
	}
}

// HandleExampleCreated handles example created events
func (h *DefaultExampleEventHandler) HandleExampleCreated(ctx context.Context, event *ExampleEvent) error {
	h.logger.Info("Handling example created event",
		zap.String("event_id", event.ID),
		zap.String("example_id", event.Data.ID),
	)

	// Example: Send welcome email, update analytics, etc.
	// This is where you'd integrate with other services

	// Simulate some processing
	time.Sleep(100 * time.Millisecond)

	h.logger.Info("Example created event processed successfully",
		zap.String("event_id", event.ID),
		zap.String("example_id", event.Data.ID),
	)

	return nil
}

// HandleExampleUpdated handles example updated events
func (h *DefaultExampleEventHandler) HandleExampleUpdated(ctx context.Context, event *ExampleEvent) error {
	h.logger.Info("Handling example updated event",
		zap.String("event_id", event.ID),
		zap.String("example_id", event.Data.ID),
	)

	// Example: Update search index, sync with external systems, etc.

	// Simulate some processing
	time.Sleep(100 * time.Millisecond)

	h.logger.Info("Example updated event processed successfully",
		zap.String("event_id", event.ID),
		zap.String("example_id", event.Data.ID),
	)

	return nil
}

// HandleExampleDeleted handles example deleted events
func (h *DefaultExampleEventHandler) HandleExampleDeleted(ctx context.Context, event *ExampleEvent) error {
	h.logger.Info("Handling example deleted event",
		zap.String("event_id", event.ID),
		zap.String("example_id", event.Data.ID),
	)

	// Example: Clean up related data, remove from search index, etc.

	// Simulate some processing
	time.Sleep(100 * time.Millisecond)

	h.logger.Info("Example deleted event processed successfully",
		zap.String("event_id", event.ID),
		zap.String("example_id", event.Data.ID),
	)

	return nil
}

// MockConsumer is a mock implementation for testing
type MockConsumer struct {
	handler   ExampleEventHandler
	logger    *zap.Logger
	isRunning bool
	events    []ExampleEvent
}

// NewMockConsumer creates a new mock consumer
func NewMockConsumer(handler ExampleEventHandler, logger *zap.Logger) *MockConsumer {
	return &MockConsumer{
		handler: handler,
		logger:  logger,
		events:  make([]ExampleEvent, 0),
	}
}

// Start mock implementation
func (m *MockConsumer) Start(ctx context.Context) error {
	m.isRunning = true
	m.logger.Info("Mock consumer started")
	return nil
}

// Stop mock implementation
func (m *MockConsumer) Stop() error {
	m.isRunning = false
	m.logger.Info("Mock consumer stopped")
	return nil
}

// SimulateEvent simulates receiving an event (for testing)
func (m *MockConsumer) SimulateEvent(ctx context.Context, event *ExampleEvent) error {
	m.events = append(m.events, *event)

	switch event.Type {
	case EventTypeExampleCreated:
		return m.handler.HandleExampleCreated(ctx, event)
	case EventTypeExampleUpdated:
		return m.handler.HandleExampleUpdated(ctx, event)
	case EventTypeExampleDeleted:
		return m.handler.HandleExampleDeleted(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

// GetProcessedEvents returns all processed events (for testing)
func (m *MockConsumer) GetProcessedEvents() []ExampleEvent {
	return m.events
}

// ClearProcessedEvents clears all processed events (for testing)
func (m *MockConsumer) ClearProcessedEvents() {
	m.events = m.events[:0]
}

// IsRunning returns whether the consumer is running
func (m *MockConsumer) IsRunning() bool {
	return m.isRunning
}

// Helper functions

// contains checks if a string contains a substring (case-insensitive)
func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			(len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					containsInMiddle(str, substr))))
}

// containsInMiddle checks if substr is contained in the middle of str (not at start or end)
func containsInMiddle(str, substr string) bool {
	if len(substr) == 0 || len(str) <= len(substr) {
		return false
	}

	// Check if substring is at the beginning or end
	if len(str) >= len(substr) {
		if str[:len(substr)] == substr || str[len(str)-len(substr):] == substr {
			return false // Found at start or end, not in middle
		}
	}

	// Look for substring in the middle
	for i := 1; i <= len(str)-len(substr)-1; i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
