# Testing Guide for Message Queue Consumer

This guide explains how to test the message queue consumer components in the example-api-template project.

## Testing Strategy

### 1. Unit Tests
- **Mock Consumer**: Test event handling logic without external dependencies
- **Event Handlers**: Test business logic for processing events
- **Configuration**: Test configuration parsing and validation
- **Helper Functions**: Test utility functions and event serialization

### 2. Integration Tests
- **Consumer Application**: Test the complete consumer application initialization
- **RabbitMQ Integration**: Test with real RabbitMQ (optional, requires setup)
- **End-to-End**: Test producer → RabbitMQ → consumer flow

### 3. Performance Tests
- **Benchmarks**: Measure event processing performance
- **Load Testing**: Test consumer under high event volume
- **Memory Usage**: Monitor resource consumption

## Running Tests

### Run All Consumer Tests
```bash
# Run all MQ-related tests
go test ./internal/transport/mq/... -v

# Run consumer application tests
go test ./cmd/consumer/... -v

# Run with coverage
go test ./internal/transport/mq/... -cover -v
```

### Run Specific Test Categories

**Unit Tests Only:**
```bash
go test ./internal/transport/mq/ -run "Test.*Consumer" -v
go test ./internal/transport/mq/ -run "Test.*Producer" -v
```

**Integration Tests (requires RabbitMQ):**
```bash
# Skip integration tests
go test ./internal/transport/mq/... -short -v

# Run integration tests (requires RabbitMQ)
go test ./internal/transport/mq/... -v
```

**Benchmarks:**
```bash
go test ./internal/transport/mq/ -bench=. -v
go test ./cmd/consumer/ -bench=. -v
```

## Test Examples

### Testing Mock Consumer

```go
func TestMockConsumer(t *testing.T) {
    mockHandler := &MockEventHandler{}
    logger := zap.NewNop()
    consumer := NewMockConsumer(mockHandler, logger)

    // Test lifecycle
    ctx := context.Background()
    err := consumer.Start(ctx)
    assert.NoError(t, err)
    assert.True(t, consumer.IsRunning())

    // Test event processing
    event := createTestEvent(EventTypeExampleCreated)
    mockHandler.On("HandleExampleCreated", mock.Anything, event).Return(nil)
    
    err = consumer.SimulateEvent(ctx, event)
    assert.NoError(t, err)

    // Verify
    events := consumer.GetProcessedEvents()
    assert.Len(t, events, 1)
    mockHandler.AssertExpectations(t)
}
```

### Testing Event Handlers

```go
func TestEventHandler(t *testing.T) {
    logger := zap.NewNop()
    handler := NewDefaultExampleEventHandler(nil, logger)

    event := &ExampleEvent{
        ID:   "test-event",
        Type: EventTypeExampleCreated,
        Data: createTestExampleWithMetadata(),
    }

    err := handler.HandleExampleCreated(context.Background(), event)
    assert.NoError(t, err)
}
```

### Testing Consumer Application

```go
func TestConsumerApplication(t *testing.T) {
    cfg := &config.Config{
        MessageQueue: config.MessageQueueConfig{
            EnableMock:     true,
            EnableConsumer: true,
        },
        // ... other config
    }

    deps, err := initializeConsumerDependencies(cfg, logger)
    assert.NoError(t, err)
    assert.NotNil(t, deps.Consumer)
}
```

## Manual Testing

### Testing with Mock Consumer

1. **Start the consumer application:**
   ```bash
   go run cmd/consumer/main.go
   ```

2. **In another terminal, start the server:**
   ```bash
   go run cmd/server/main.go
   ```

3. **Create an example via API:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/examples \
     -H "Content-Type: application/json" \
     -d '{"name":"Test User","email":"test@example.com","age":25}'
   ```

4. **Check consumer logs** for event processing messages.

### Testing with Real RabbitMQ

1. **Start RabbitMQ:**
   ```bash
   docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
   ```

2. **Configure for real RabbitMQ:**
   ```bash
   export MQ_ENABLE_MOCK=false
   export MQ_URL=amqp://guest:guest@localhost:5672/
   ```

3. **Start consumer:**
   ```bash
   go run cmd/consumer/main.go
   ```

4. **Start server:**
   ```bash
   go run cmd/server/main.go
   ```

5. **Test API operations** and observe events in RabbitMQ Management UI (http://localhost:15672).

## Test Data and Fixtures

### Creating Test Events

```go
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
```

### Mock Event Handler

```go
type MockEventHandler struct {
    mock.Mock
}

func (m *MockEventHandler) HandleExampleCreated(ctx context.Context, event *ExampleEvent) error {
    args := m.Called(ctx, event)
    return args.Error(0)
}
```

## Debugging Tests

### Enable Debug Logging

```go
// In test setup
logger := zap.NewDevelopment()
consumer := NewMockConsumer(handler, logger)
```

### Verbose Test Output

```bash
go test ./internal/transport/mq/ -v -args -test.v
```

### Test with Race Detection

```bash
go test ./internal/transport/mq/ -race -v
```

## Performance Testing

### Benchmark Event Processing

```bash
go test ./internal/transport/mq/ -bench=BenchmarkEventHandling -v
```

### Load Testing with Real RabbitMQ

```go
func TestConsumerLoad(t *testing.T) {
    // Setup consumer
    consumer := setupRealConsumer(t)
    
    // Send many events
    for i := 0; i < 1000; i++ {
        publishEvent(fmt.Sprintf("event-%d", i))
    }
    
    // Verify all processed
    eventually(t, func() bool {
        return consumer.GetProcessedCount() == 1000
    }, 30*time.Second)
}
```

## Troubleshooting

### Common Issues

1. **RabbitMQ Connection Failed**
   ```
   Error: failed to connect to RabbitMQ: dial tcp :5672: connect: connection refused
   ```
   **Solution**: Ensure RabbitMQ is running or use mock mode.

2. **Consumer Not Processing Events**
   ```
   Consumer started but no events processed
   ```
   **Solution**: Check routing keys and queue bindings.

3. **Test Timeouts**
   ```
   Test timed out waiting for event processing
   ```
   **Solution**: Increase timeout or check for deadlocks.

### Debug Commands

```bash
# Check RabbitMQ status
docker exec rabbitmq rabbitmqctl status

# List queues
docker exec rabbitmq rabbitmqctl list_queues

# Monitor logs
docker logs -f rabbitmq
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Test Consumer
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      rabbitmq:
        image: rabbitmq:3-management
        ports:
          - 5672:5672
        options: >-
          --health-cmd "rabbitmq-diagnostics -q ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Run unit tests
      run: go test ./internal/transport/mq/... -short -v
      
    - name: Run integration tests
      run: |
        export MQ_ENABLE_MOCK=false
        export MQ_URL=amqp://guest:guest@localhost:5672/
        go test ./internal/transport/mq/... -v
      env:
        MQ_ENABLE_MOCK: false
        MQ_URL: amqp://guest:guest@localhost:5672/
```

This comprehensive testing approach ensures your message queue consumer is reliable, performant, and maintainable.
