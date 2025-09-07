package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example-api-template/internal/config"
	"example-api-template/internal/repository"
	"example-api-template/internal/service"
	httpTransport "example-api-template/internal/transport/http"
	"example-api-template/internal/transport/mq"
	"example-api-template/internal/usecase"
	"example-api-template/pkg/logger"
	"example-api-template/pkg/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	appLogger.Info("Starting application",
		zap.String("name", cfg.App.Name),
		zap.String("version", cfg.App.Version),
		zap.String("environment", cfg.App.Environment),
	)

	// Initialize dependencies
	deps, err := initializeDependencies(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize dependencies", zap.Error(err))
	}

	// Initialize Echo server
	e := setupEcho(cfg, appLogger)

	// Register routes
	deps.Handler.RegisterRoutes(e)

	// Start server
	startServer(e, cfg, appLogger, deps)
}

// Dependencies holds all application dependencies
type Dependencies struct {
	Repository  repository.ExampleRepository
	ExternalAPI repository.ExternalExampleAPI
	Service     service.ExampleService
	UseCase     usecase.ExampleUseCase
	Validator   validator.Validator
	Handler     *httpTransport.ExampleHandler
	Producer    mq.ExampleProducer
}

// initializeDependencies initializes all application dependencies
func initializeDependencies(cfg *config.Config, logger *logger.Logger) (*Dependencies, error) {
	// Initialize validator
	validator := validator.New()

	// Initialize repository
	var repo repository.ExampleRepository
	switch cfg.Database.Type {
	case "memory":
		repo = repository.NewInMemoryExampleRepository()
		logger.Info("Using in-memory repository")
	default:
		// In a real application, you would initialize database connections here
		// For now, fall back to in-memory
		repo = repository.NewInMemoryExampleRepository()
		logger.Warn("Unsupported database type, falling back to in-memory repository",
			zap.String("type", cfg.Database.Type))
	}

	// Initialize external API
	var externalAPI repository.ExternalExampleAPI
	if cfg.ExternalAPI.EnableMock {
		externalAPI = repository.NewMockExternalExampleAPI(
			cfg.ExternalAPI.MockShouldFail,
			cfg.ExternalAPI.MockDelay,
		)
		logger.Info("Using mock external API")
	} else {
		// In a real application, you would initialize the actual external API client here
		externalAPI = repository.NewMockExternalExampleAPI(false, 100*time.Millisecond)
		logger.Warn("Real external API not implemented, using mock")
	}

	// Initialize service
	svc := service.NewExampleService(repo, logger.Logger)

	// Initialize use case
	uc := usecase.NewExampleUseCase(svc, externalAPI, logger.Logger)

	// Initialize HTTP handler
	handler := httpTransport.NewExampleHandler(uc, validator, logger.Logger)

	// Initialize message queue producer only (consumer runs separately)
	var producer mq.ExampleProducer

	if cfg.MessageQueue.EnableMock {
		// Use mock implementation
		producer = mq.NewMockProducer(logger.Logger)
		logger.Info("Using mock message queue producer")
	} else {
		// Use real RabbitMQ implementation
		if cfg.MessageQueue.EnableProducer {
			producerConfig := &mq.RabbitMQProducerConfig{
				URL:           cfg.MessageQueue.URL,
				ExchangeName:  cfg.MessageQueue.ExchangeName,
				RoutingPrefix: cfg.MessageQueue.RoutingPrefix,
				Durable:       cfg.MessageQueue.Durable,
				AutoDelete:    cfg.MessageQueue.AutoDelete,
			}

			var err error
			producer, err = mq.NewRabbitMQProducer(producerConfig, logger.Logger)
			if err != nil {
				logger.Warn("Failed to initialize RabbitMQ producer, using mock", zap.Error(err))
				producer = mq.NewMockProducer(logger.Logger)
			} else {
				logger.Info("Using RabbitMQ producer")
			}
		} else {
			producer = mq.NewMockProducer(logger.Logger)
			logger.Info("Producer disabled, using mock")
		}
	}

	return &Dependencies{
		Repository:  repo,
		ExternalAPI: externalAPI,
		Service:     svc,
		UseCase:     uc,
		Validator:   validator,
		Handler:     handler,
		Producer:    producer,
	}, nil
}

// setupEcho configures the Echo web framework
func setupEcho(cfg *config.Config, logger *logger.Logger) *echo.Echo {
	e := echo.New()

	// Hide Echo banner
	e.HideBanner = true
	e.HidePort = true

	// Configure Echo
	e.Debug = cfg.App.Debug

	// Middleware
	e.Use(middleware.RequestID())
	e.Use(createLoggingMiddleware(logger))
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: cfg.Server.ReadTimeout,
	}))

	if cfg.Server.EnableCORS {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodDelete,
				http.MethodOptions,
			},
			AllowHeaders: []string{
				echo.HeaderOrigin,
				echo.HeaderContentType,
				echo.HeaderAccept,
				echo.HeaderAuthorization,
				"X-Request-ID",
			},
		}))
	}

	// Security headers
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	// Rate limiting (basic)
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	// Compression
	e.Use(middleware.Gzip())

	return e
}

// createLoggingMiddleware creates a custom logging middleware
func createLoggingMiddleware(logger *logger.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogMethod:    true,
		LogLatency:   true,
		LogError:     true,
		LogRequestID: true,
		LogUserAgent: true,
		LogRemoteIP:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fields := []zap.Field{
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.Int("status", v.Status),
				zap.Duration("latency", v.Latency),
				zap.String("remote_ip", v.RemoteIP),
				zap.String("user_agent", v.UserAgent),
				zap.String("request_id", v.RequestID),
			}

			if v.Error != nil {
				fields = append(fields, zap.Error(v.Error))
				logger.Error("Request failed", fields...)
			} else {
				logger.Info("Request completed", fields...)
			}

			return nil
		},
	})
}

// startServer starts the HTTP server with graceful shutdown
func startServer(e *echo.Echo, cfg *config.Config, logger *logger.Logger, deps *Dependencies) {
	// Server configuration
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.ReadTimeout * 2,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server",
			zap.String("address", server.Addr),
			zap.Duration("read_timeout", server.ReadTimeout),
			zap.Duration("write_timeout", server.WriteTimeout),
		)

		if err := e.StartServer(server); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Close message queue producer
	if err := deps.Producer.Close(); err != nil {
		logger.Error("Failed to close message queue producer", zap.Error(err))
	} else {
		logger.Info("Message queue producer closed")
	}

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	// Shutdown server
	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("Server exited gracefully")
	}
}

// Health check for the application
func init() {
	// Ensure the application can start properly
	if os.Getenv("HEALTH_CHECK") == "true" {
		fmt.Println("OK")
		os.Exit(0)
	}
}
