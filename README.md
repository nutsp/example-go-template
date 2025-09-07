# Example API Template

A complete Go REST API service built with Clean Architecture, featuring comprehensive business logic, external API integration, structured logging, and extensive testing.

## 🏗️ Architecture

This project follows Clean Architecture principles with clear separation of concerns:

```
example-api-template/
├── cmd/server/               # Entry point
│   └── main.go              # Application bootstrap with DI
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
│   ├── transport/http/      # HTTP presentation layer
│   │   ├── example_handler.go        # Echo handlers
│   │   └── dto.go                    # Request/Response DTOs
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
- **External API Integration**: Validation, enrichment, and notification services
- **Pagination**: Efficient list operations with limit/offset pagination

### Technical Features
- **Clean Architecture**: Domain-driven design with dependency injection
- **Structured Logging**: Comprehensive logging with Zap (JSON/Console formats)
- **Input Validation**: Request validation with custom rules
- **Error Handling**: Comprehensive error types with proper HTTP status codes
- **Configuration**: Environment-based configuration with validation
- **Testing**: Extensive unit tests with mocks and fixtures
- **Graceful Shutdown**: Proper server lifecycle management
- **Middleware**: CORS, rate limiting, security headers, request logging

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

## 🛠️ Getting Started

### Prerequisites
- Go 1.21 or higher
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

3. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:8080`

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
- [ ] Configure monitoring and alerting
- [ ] Set up health checks
- [ ] Configure load balancer
- [ ] Set up log aggregation

### Docker Support
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### Health Check
```bash
# Simple health check
curl http://localhost:8080/api/v1/health

# Environment variable health check
HEALTH_CHECK=true go run cmd/server/main.go
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
