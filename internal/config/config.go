package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server       ServerConfig       `json:"server"`
	Database     DatabaseConfig     `json:"database"`
	ExternalAPI  ExternalAPIConfig  `json:"external_api"`
	MessageQueue MessageQueueConfig `json:"message_queue"`
	Logger       LoggerConfig       `json:"logger"`
	App          AppConfig          `json:"app"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	EnableCORS      bool          `json:"enable_cors"`
	EnableMetrics   bool          `json:"enable_metrics"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type            string        `json:"type"`
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Name            string        `json:"name"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	SSLMode         string        `json:"ssl_mode"`
	MaxConnections  int           `json:"max_connections"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}

// ExternalAPIConfig holds external API configuration
type ExternalAPIConfig struct {
	BaseURL        string            `json:"base_url"`
	APIKey         string            `json:"api_key"`
	Timeout        time.Duration     `json:"timeout"`
	RetryAttempts  int               `json:"retry_attempts"`
	RetryDelay     time.Duration     `json:"retry_delay"`
	EnableMock     bool              `json:"enable_mock"`
	MockDelay      time.Duration     `json:"mock_delay"`
	MockShouldFail bool              `json:"mock_should_fail"`
	Headers        map[string]string `json:"headers"`
}

// MessageQueueConfig holds message queue configuration
type MessageQueueConfig struct {
	URL               string        `json:"url"`
	ExchangeName      string        `json:"exchange_name"`
	QueueName         string        `json:"queue_name"`
	RoutingPrefix     string        `json:"routing_prefix"`
	RoutingKeys       []string      `json:"routing_keys"`
	Durable           bool          `json:"durable"`
	AutoDelete        bool          `json:"auto_delete"`
	Exclusive         bool          `json:"exclusive"`
	NoWait            bool          `json:"no_wait"`
	PrefetchCount     int           `json:"prefetch_count"`
	EnableProducer    bool          `json:"enable_producer"`
	EnableConsumer    bool          `json:"enable_consumer"`
	EnableMock        bool          `json:"enable_mock"`
	ReconnectInterval time.Duration `json:"reconnect_interval"`
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level       string   `json:"level"`
	Format      string   `json:"format"` // json, console
	Development bool     `json:"development"`
	EnableColor bool     `json:"enable_color"`
	OutputPaths []string `json:"output_paths"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Debug       bool   `json:"debug"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "localhost"),
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			EnableCORS:      getEnvAsBool("SERVER_ENABLE_CORS", true),
			EnableMetrics:   getEnvAsBool("SERVER_ENABLE_METRICS", true),
		},
		Database: DatabaseConfig{
			Type:            getEnv("DB_TYPE", "memory"), // memory, postgres, mysql
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			Name:            getEnv("DB_NAME", "example_db"),
			Username:        getEnv("DB_USERNAME", ""),
			Password:        getEnv("DB_PASSWORD", ""),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxConnections:  getEnvAsInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		ExternalAPI: ExternalAPIConfig{
			BaseURL:        getEnv("EXTERNAL_API_BASE_URL", "https://api.example.com"),
			APIKey:         getEnv("EXTERNAL_API_KEY", ""),
			Timeout:        getEnvAsDuration("EXTERNAL_API_TIMEOUT", 30*time.Second),
			RetryAttempts:  getEnvAsInt("EXTERNAL_API_RETRY_ATTEMPTS", 3),
			RetryDelay:     getEnvAsDuration("EXTERNAL_API_RETRY_DELAY", 1*time.Second),
			EnableMock:     getEnvAsBool("EXTERNAL_API_ENABLE_MOCK", true),
			MockDelay:      getEnvAsDuration("EXTERNAL_API_MOCK_DELAY", 100*time.Millisecond),
			MockShouldFail: getEnvAsBool("EXTERNAL_API_MOCK_SHOULD_FAIL", false),
			Headers:        getEnvAsMap("EXTERNAL_API_HEADERS", map[string]string{}),
		},
		MessageQueue: MessageQueueConfig{
			URL:               getEnv("MQ_URL", "amqp://guest:guest@localhost:5672/"),
			ExchangeName:      getEnv("MQ_EXCHANGE_NAME", "examples"),
			QueueName:         getEnv("MQ_QUEUE_NAME", "example-events"),
			RoutingPrefix:     getEnv("MQ_ROUTING_PREFIX", "example"),
			RoutingKeys:       getEnvAsSlice("MQ_ROUTING_KEYS", []string{"example.created", "example.updated", "example.deleted"}),
			Durable:           getEnvAsBool("MQ_DURABLE", true),
			AutoDelete:        getEnvAsBool("MQ_AUTO_DELETE", false),
			Exclusive:         getEnvAsBool("MQ_EXCLUSIVE", false),
			NoWait:            getEnvAsBool("MQ_NO_WAIT", false),
			PrefetchCount:     getEnvAsInt("MQ_PREFETCH_COUNT", 10),
			EnableProducer:    getEnvAsBool("MQ_ENABLE_PRODUCER", true),
			EnableConsumer:    getEnvAsBool("MQ_ENABLE_CONSUMER", true),
			EnableMock:        getEnvAsBool("MQ_ENABLE_MOCK", true),
			ReconnectInterval: getEnvAsDuration("MQ_RECONNECT_INTERVAL", 5*time.Second),
		},
		Logger: LoggerConfig{
			Level:       getEnv("LOG_LEVEL", "info"),
			Format:      getEnv("LOG_FORMAT", "json"),
			Development: getEnvAsBool("LOG_DEVELOPMENT", false),
			EnableColor: getEnvAsBool("LOG_ENABLE_COLOR", false),
			OutputPaths: getEnvAsSlice("LOG_OUTPUT_PATHS", []string{"stdout"}),
		},
		App: AppConfig{
			Name:        getEnv("APP_NAME", "example-api"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
			Environment: getEnv("APP_ENVIRONMENT", "development"),
			Debug:       getEnvAsBool("APP_DEBUG", false),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	var errs []string

	// Validate server config
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, "server port must be between 1 and 65535")
	}
	if c.Server.ReadTimeout <= 0 {
		errs = append(errs, "server read timeout must be positive")
	}
	if c.Server.WriteTimeout <= 0 {
		errs = append(errs, "server write timeout must be positive")
	}

	// Validate database config
	if c.Database.Type != "memory" && c.Database.Type != "postgres" && c.Database.Type != "mysql" {
		errs = append(errs, "database type must be one of: memory, postgres, mysql")
	}
	if c.Database.Type != "memory" {
		if c.Database.Host == "" {
			errs = append(errs, "database host is required for non-memory databases")
		}
		if c.Database.Port < 1 || c.Database.Port > 65535 {
			errs = append(errs, "database port must be between 1 and 65535")
		}
		if c.Database.Name == "" {
			errs = append(errs, "database name is required for non-memory databases")
		}
	}

	// Validate external API config
	if c.ExternalAPI.Timeout <= 0 {
		errs = append(errs, "external API timeout must be positive")
	}
	if c.ExternalAPI.RetryAttempts < 0 {
		errs = append(errs, "external API retry attempts must be non-negative")
	}

	// Validate logger config
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLogLevels, c.Logger.Level) {
		errs = append(errs, "logger level must be one of: debug, info, warn, error, fatal, panic")
	}
	if c.Logger.Format != "json" && c.Logger.Format != "console" {
		errs = append(errs, "logger format must be either 'json' or 'console'")
	}

	// Validate app config
	if c.App.Name == "" {
		errs = append(errs, "app name is required")
	}
	if c.App.Version == "" {
		errs = append(errs, "app version is required")
	}
	validEnvironments := []string{"development", "staging", "production"}
	if !contains(validEnvironments, c.App.Environment) {
		errs = append(errs, "app environment must be one of: development, staging, production")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetDatabaseDSN returns the database DSN (Data Source Name)
func (c *Config) GetDatabaseDSN() string {
	if c.Database.Type == "memory" {
		return "memory"
	}

	switch c.Database.Type {
	case "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			c.Database.Username,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.Name,
			c.Database.SSLMode,
		)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			c.Database.Username,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.Name,
		)
	default:
		return ""
	}
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getEnvAsMap(key string, defaultValue map[string]string) map[string]string {
	if value := os.Getenv(key); value != "" {
		result := make(map[string]string)
		pairs := strings.Split(value, ",")
		for _, pair := range pairs {
			if kv := strings.Split(pair, "="); len(kv) == 2 {
				result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
		return result
	}
	return defaultValue
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
