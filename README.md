# Example API Template

A complete Go REST API service built with Clean Architecture, featuring comprehensive business logic, external API integration, structured logging, and extensive testing.

## 🏗️ Architecture

This project follows Clean Architecture principles with clear separation of concerns:

```
example-api-template/
├── cmd/
│   ├── server/              # HTTP API server
│   │   └── main.go          # Server application with HTTP handlers
│   └── consumer/            # Message queue consumer
│       └── main.go          # Consumer application for processing events
├── internal/
│   ├── domain/              # Business entities
│   │   └── example.go       # Core domain model
│   ├── repository/          # Data access layer
│   │   ├── example_repository.go     # Repository interface & in-memory impl
│   │   └── external_example_api.go   # External API interface & mock
│   ├── service/             # Business logic layer
│   │   └── example_service.go        # Business rules & validation
│   ├── usecase/             # Application orchestration
│   │   └── example_usecase.go        # Use cases with external integration
│   ├── transport/
│   │   ├── http/            # HTTP presentation layer
│   │   │   ├── example_handler.go    # Echo handlers
│   │   │   └── dto.go                # Request/Response DTOs
│   │   └── mq/              # Message Queue transport layer
│   │       ├── example_producer.go   # Event publishing
│   │       └── example_consumer.go   # Event consumption
│   └── config/              # Configuration management
│       └── config.go        # Environment-based config
├── pkg/
│   ├── logger/              # Structured logging with Zap
│   └── validator/           # Request validation
├── tests/
│   ├── mocks/               # Test mocks
│   └── fixtures/            # Test data fixtures
├── go.mod
└── README.md
```

## 🚀 Features

### Core Functionality
- **CRUD Operations**: Complete Create, Read, Update, Delete operations for Example entities
- **Business Logic**: Name validation, email uniqueness, age restrictions, corporate/VIP domain rules
- **Database Support**: In-memory and PostgreSQL repositories with GORM ORM
- **Internationalization (i18n)**: Multi-language support with localized error messages and responses
- **External API Integration**: Validation, enrichment, and notification services
- **Message Queue Integration**: Asynchronous event publishing and consumption with RabbitMQ
- **Pagination**: Efficient list operations with limit/offset pagination

### Technical Features
- **Clean Architecture**: Domain-driven design with dependency injection
- **Structured Logging**: Comprehensive logging with Zap (JSON/Console formats)
- **Input Validation**: Request validation with custom rules
- **Error Handling**: Comprehensive error types with proper HTTP status codes
- **Configuration**: Environment-based configuration with validation
- **Testing**: Extensive unit tests with mocks and fixtures
- **Graceful Shutdown**: Proper server and message queue lifecycle management
- **Middleware**: CORS, rate limiting, security headers, request logging
- **Event-Driven Architecture**: Asynchronous event processing with reliable message delivery

## 📋 API Endpoints

### Examples
- `POST /api/v1/examples` - Create a new example
- `GET /api/v1/examples` - List examples (paginated)
- `GET /api/v1/examples/{id}` - Get example by ID
- `GET /api/v1/examples/email/{email}` - Get example by email
- `PUT /api/v1/examples/{id}` - Update example
- `DELETE /api/v1/examples/{id}` - Delete example
- `POST /api/v1/examples/validate` - Create with external validation

### Health & Monitoring
- `GET /api/v1/health` - Health check endpoint

## 📨 Message Queue Events

The service publishes events to RabbitMQ for asynchronous processing:

### Event Types
- **example.created** - Published when a new example is created
- **example.updated** - Published when an example is updated
- **example.deleted** - Published when an example is deleted

### Event Structure
```json
{
  "id": "evt_1640995200000000000",
  "type": "example.created",
  "timestamp": "2023-12-01T10:00:00Z",
  "data": {
    "id": "ex_joh_8",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "age": 30,
    "created_at": "2023-12-01T10:00:00Z",
    "updated_at": "2023-12-01T10:00:00Z",
    "external_data": {
      "external_id": "ext_ex_joh_8",
      "metadata": {
        "source": "mock_api",
        "version": "1.0"
      },
      "score": 0.85
    }
  },
  "metadata": {
    "source": "example-api",
    "version": "1.0",
    "user_id": "system",
    "trace_id": "abc123"
  }
}
```

### Consumer Implementation
The service includes both embedded and standalone consumer options:

#### Embedded Consumer (Default)
- Runs alongside the HTTP server
- Uses mock implementation by default
- Good for development and simple deployments

#### Standalone Consumer (`cmd/consumer/main.go`)
- Runs as a separate application
- Dedicated for processing message queue events
- Supports independent scaling and deployment
- Includes the same event handler capabilities:
  - Logs all events for audit purposes
  - Can be extended to integrate with external systems
  - Supports error handling and retry logic
  - Processes events asynchronously

## 🛠️ Getting Started

### Prerequisites
- Go 1.21 or higher
- PostgreSQL (optional, will use in-memory by default)
- RabbitMQ (optional, will use mock by default)
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd example-api-template
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Optional: Start PostgreSQL** (skip if using in-memory)
   ```bash
   # Using Docker
   docker run -d --name postgres \
     -e POSTGRES_DB=example_db \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=password \
     -p 5432:5432 postgres:15
   
   # Or using Homebrew on macOS
   brew install postgresql@15
   brew services start postgresql@15
   createdb example_db
   ```

4. **Optional: Start RabbitMQ** (skip if using mock)
   ```bash
   # Using Docker
   docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
   
   # Or using Homebrew on macOS
   brew install rabbitmq
   brew services start rabbitmq
   ```

5. **Run the applications**
   
   **Option A: Run server only (with mock consumer)**
   ```bash
   go run cmd/server/main.go
   ```
   
   **Option B: Run server and consumer separately**
   ```bash
   # Terminal 1: Start the HTTP API server
   go run cmd/server/main.go
   
   # Terminal 2: Start the message queue consumer
   go run cmd/consumer/main.go
   ```

   **Option C: Run with PostgreSQL**
   ```bash
   # Set database configuration
   export DB_TYPE=postgres
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_NAME=example_db
   export DB_USERNAME=postgres
   export DB_PASSWORD=password
   
   # Start applications
   go run cmd/server/main.go     # Terminal 1
   go run cmd/consumer/main.go   # Terminal 2
   ```

The server will start on `http://localhost:8080` and the consumer will process events from the message queue

### Internationalization (i18n) Usage

The API supports multiple languages through HTTP headers:

#### Language Detection (Priority Order)
1. **Query Parameter**: `?lang=es`
2. **Custom Header**: `X-Language: es`
3. **Accept-Language Header**: `Accept-Language: es-ES,es;q=0.9,en;q=0.8`
4. **Cookie**: `language=es`
5. **Default**: Falls back to configured default language

#### Example Requests
```bash
# Using query parameter
curl -X GET "http://localhost:8080/examples?lang=es"

# Using custom header
curl -X GET "http://localhost:8080/examples" \
  -H "X-Language: fr"

# Using Accept-Language header
curl -X GET "http://localhost:8080/examples" \
  -H "Accept-Language: es-ES,es;q=0.9,en;q=0.8"

# Response includes Content-Language header
HTTP/1.1 200 OK
Content-Language: es
Content-Type: application/json
{
  "message": "Ejemplos obtenidos exitosamente",
  "data": [...]
}
```

#### Supported Languages
- **English (en)**: Default language
- **Spanish (es)**: Full translation support
- **French (fr)**: Full translation support
- **Thai (th)**: Full translation support

#### Adding New Languages
1. Create translation file: `translations/{language}.json`
2. Add language to `I18N_LANGUAGES` environment variable
3. Restart the application

#### Testing Thai Language Support
```bash
# Run the comprehensive Thai language test
./examples/test_thai_api.sh

# Or test manually with curl
curl -H "Accept-Language: th" http://localhost:8080/api/v1/examples
curl -H "X-Language: th" http://localhost:8080/api/v1/examples
curl "http://localhost:8080/api/v1/examples?lang=th"
```

#### Example Thai API Responses
```json
// Thai success response
{
  "message": "ตัวอย่างสร้างสำเร็จ",
  "id": "123",
  "name": "สมชาย ใจดี",
  "email": "somchai@example.com",
  "age": 28
}

// Thai error response
{
  "error": "ไม่พบตัวอย่างที่มี ID '123'",
  "message": "ไม่พบตัวอย่าง"
}
```

### Configuration

Configure the application using environment variables:

#### Server Configuration
```bash
SERVER_HOST=localhost          # Server host (default: localhost)
SERVER_PORT=8080              # Server port (default: 8080)
SERVER_READ_TIMEOUT=10s       # Read timeout (default: 10s)
SERVER_WRITE_TIMEOUT=10s      # Write timeout (default: 10s)
SERVER_SHUTDOWN_TIMEOUT=30s   # Graceful shutdown timeout (default: 30s)
SERVER_ENABLE_CORS=true       # Enable CORS (default: true)
```

#### Database Configuration
```bash
DB_TYPE=memory                # Database type: memory, postgres, mysql (default: memory)
DB_HOST=localhost             # Database host
DB_PORT=5432                  # Database port
DB_NAME=example_db            # Database name
DB_USERNAME=user              # Database username
DB_PASSWORD=pass              # Database password
```

#### External API Configuration
```bash
EXTERNAL_API_ENABLE_MOCK=true        # Use mock external API (default: true)
EXTERNAL_API_MOCK_DELAY=100ms        # Mock API delay (default: 100ms)
EXTERNAL_API_MOCK_SHOULD_FAIL=false  # Make mock API fail (default: false)
EXTERNAL_API_TIMEOUT=30s             # External API timeout (default: 30s)
```

#### Database Configuration
```bash
DB_TYPE=memory                    # Database type: memory, postgres (default: memory)
DB_HOST=localhost                 # PostgreSQL host (default: localhost)
DB_PORT=5432                      # PostgreSQL port (default: 5432)
DB_NAME=example_db                # Database name (default: example_db)
DB_USERNAME=postgres              # Database username (default: empty)
DB_PASSWORD=password              # Database password (default: empty)
DB_SSL_MODE=disable               # SSL mode: disable, require, verify-ca, verify-full (default: disable)
DB_MAX_CONNECTIONS=25             # Maximum open connections (default: 25)
DB_MAX_IDLE_CONNS=5               # Maximum idle connections (default: 5)
DB_CONN_MAX_LIFETIME=5m           # Connection max lifetime (default: 5m)
```

#### Internationalization Configuration
```bash
I18N_DEFAULT_LANGUAGE=en          # Default language (default: en)
I18N_LANGUAGES=en,es,fr,th        # Supported languages (default: en,es,fr,th)
I18N_TRANSLATION_DIR=translations # Translation files directory (default: translations)
```

#### Message Queue Configuration
```bash
MQ_URL=amqp://guest:guest@localhost:5672/  # RabbitMQ connection URL (default: localhost)
MQ_EXCHANGE_NAME=examples                   # Exchange name (default: examples)
MQ_QUEUE_NAME=example-events               # Queue name (default: example-events)
MQ_ROUTING_PREFIX=example                   # Routing key prefix (default: example)
MQ_ROUTING_KEYS=example.created,example.updated,example.deleted  # Routing keys
MQ_ENABLE_PRODUCER=true                     # Enable message producer (default: true)
MQ_ENABLE_CONSUMER=true                     # Enable message consumer (default: true)
MQ_ENABLE_MOCK=true                         # Use mock MQ (default: true)
MQ_PREFETCH_COUNT=10                        # Consumer prefetch count (default: 10)
MQ_DURABLE=true                             # Make queues durable (default: true)
```

#### Logging Configuration
```bash
LOG_LEVEL=info                # Log level: debug, info, warn, error (default: info)
LOG_FORMAT=json               # Log format: json, console (default: json)
LOG_DEVELOPMENT=false         # Development mode (default: false)
```

#### Application Configuration
```bash
APP_NAME=example-api          # Application name (default: example-api)
APP_VERSION=1.0.0             # Application version (default: 1.0.0)
APP_ENVIRONMENT=development   # Environment: development, staging, production (default: development)
APP_DEBUG=false               # Debug mode (default: false)
```

## 📝 Usage Examples

### Create an Example
```bash
curl -X POST http://localhost:8080/api/v1/examples \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john.doe@example.com",
    "age": 30
  }'
```

### Get an Example
```bash
curl http://localhost:8080/api/v1/examples/ex_joh_8
```

### List Examples
```bash
curl "http://localhost:8080/api/v1/examples?limit=10&offset=0"
```

### Update an Example
```bash
curl -X PUT http://localhost:8080/api/v1/examples/ex_joh_8 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane.doe@example.com",
    "age": 28
  }'
```

### Delete an Example
```bash
curl -X DELETE http://localhost:8080/api/v1/examples/ex_joh_8
```

### Create with External Validation
```bash
curl -X POST http://localhost:8080/api/v1/examples/validate \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Smith",
    "email": "alice@example.com",
    "age": 25
  }'
```

## 🧪 Testing

### Run All Tests
```bash
go test ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Run Specific Test
```bash
go test -run TestExampleService_CreateExample ./internal/service
```

### Run Benchmarks
```bash
go test -bench=. ./internal/domain
```

### Test Structure
- **Unit Tests**: Comprehensive tests for all layers using mocks
- **Test Fixtures**: Reusable test data and scenarios
- **Mocks**: Generated mocks for all interfaces using testify/mock
- **Benchmarks**: Performance tests for critical operations

## 🏢 Business Rules

The application implements several business rules:

### Validation Rules
- **Name**: 1-100 characters, letters/spaces/hyphens/apostrophes only
- **Email**: Valid email format, unique across all examples
- **Age**: 0-150 years

### Business Logic
- **Profanity Filter**: Names cannot contain inappropriate content
- **Corporate Domains**: Users with corporate emails (@corp.com, @enterprise.com) must be 18+
- **VIP Domains**: Users with VIP emails (@vip.com, @premium.com) must be 21+

### External API Integration
- **Validation**: External validation for data quality
- **Enrichment**: Additional metadata from external sources
- **Notifications**: Async notifications for new user creation

## 🔧 Development

### Project Structure Principles
- **Domain Layer**: Pure business logic, no external dependencies
- **Repository Layer**: Data access abstraction
- **Service Layer**: Business rules and validation
- **Use Case Layer**: Application orchestration and external integration
- **Transport Layer**: HTTP handlers and DTOs

### Code Quality
- **SOLID Principles**: Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, Dependency Inversion
- **DRY**: Don't Repeat Yourself - common logic is abstracted
- **Clean Code**: Readable, maintainable, and well-documented code
- **Error Handling**: Comprehensive error types and proper error propagation

### Adding New Features

1. **Add Domain Logic**: Start with domain entities and business rules
2. **Create Repository Interface**: Define data access needs
3. **Implement Service**: Add business logic and validation
4. **Create Use Case**: Orchestrate service calls and external integration
5. **Add HTTP Handlers**: Create DTOs and endpoints
6. **Write Tests**: Add unit tests with mocks
7. **Update Documentation**: Keep README and code comments current

## 🚀 Deployment

### Production Checklist
- [ ] Set `APP_ENVIRONMENT=production`
- [ ] Set `LOG_FORMAT=json`
- [ ] Configure proper database connection
- [ ] Set up external API credentials
- [ ] Configure RabbitMQ cluster
- [ ] Set `MQ_ENABLE_MOCK=false`
- [ ] Deploy consumer applications separately
- [ ] Configure monitoring and alerting
- [ ] Set up health checks
- [ ] Configure load balancer
- [ ] Set up log aggregation

### Docker Support

**Server Dockerfile:**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server"]
```

**Consumer Dockerfile:**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o consumer cmd/consumer/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/consumer .
CMD ["./consumer"]
```

**Docker Compose:**
```yaml
version: '3.8'
services:
  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest

  server:
    build:
      context: .
      dockerfile: Dockerfile.server
    ports:
      - "8080:8080"
    environment:
      - MQ_ENABLE_MOCK=false
      - MQ_URL=amqp://guest:guest@rabbitmq:5672/
    depends_on:
      - rabbitmq

  consumer:
    build:
      context: .
      dockerfile: Dockerfile.consumer
    environment:
      - MQ_ENABLE_MOCK=false
      - MQ_URL=amqp://guest:guest@rabbitmq:5672/
    depends_on:
      - rabbitmq
    deploy:
      replicas: 2
```

### Health Check
```bash
# Simple health check
curl http://localhost:8080/api/v1/health

# Environment variable health check
HEALTH_CHECK=true go run cmd/server/main.go
HEALTH_CHECK=true go run cmd/consumer/main.go
```

## 📊 Monitoring

### Metrics
The application provides built-in monitoring capabilities:
- Request/response logging
- Performance metrics
- Error tracking
- Health status

### Logging
Structured logging with configurable levels:
- **Debug**: Detailed debugging information
- **Info**: General application flow
- **Warn**: Warning conditions
- **Error**: Error conditions
- **Fatal**: Application termination

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards
- Follow Go conventions and best practices
- Write comprehensive tests for new features
- Update documentation for API changes
- Use meaningful commit messages
- Ensure all tests pass before submitting PR

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Clean Architecture by Robert C. Martin
- Echo web framework
- Zap structured logging
- Testify testing toolkit
- Go community for excellent tooling and libraries

- Go community for excellent tooling and libraries
