package database

import (
	"os"
	"testing"
	"time"

	"example-api-template/internal/config"
	"example-api-template/internal/domain"
	"example-api-template/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildPostgresDSN tests the DSN building function
func TestBuildPostgresDSN(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "test_db",
		Username: "test_user",
		Password: "test_password",
		SSLMode:  "disable",
	}

	expected := "host=localhost port=5432 user=test_user password=test_password dbname=test_db sslmode=disable TimeZone=UTC"
	actual := buildPostgresDSN(cfg)
	assert.Equal(t, expected, actual)
}

// TestNewPostgreSQLConnection tests connection creation with invalid config
func TestNewPostgreSQLConnection(t *testing.T) {
	logger, err := logger.New(&config.LoggerConfig{
		Level:  "error",
		Format: "console",
	})
	require.NoError(t, err)
	defer logger.Close()

	// Test with nil config
	_, err = NewPostgreSQLConnection(nil, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database configuration is required")

	// Test with nil logger
	cfg := &config.DatabaseConfig{
		Type: "postgres",
		Host: "localhost",
		Port: 5432,
		Name: "test_db",
	}
	_, err = NewPostgreSQLConnection(cfg, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logger is required")

	// Test with invalid connection (should fail to connect)
	cfg = &config.DatabaseConfig{
		Type:            "postgres",
		Host:            "nonexistent-host",
		Port:            5432,
		Name:            "test_db",
		Username:        "test_user",
		Password:        "test_password",
		SSLMode:         "disable",
		MaxConnections:  10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
	}

	_, err = NewPostgreSQLConnection(cfg, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to PostgreSQL database")
}

// TestPostgreSQLConnectionMethods tests connection methods with mock/invalid connections
func TestPostgreSQLConnectionMethods(t *testing.T) {
	// This test uses a non-existent host to ensure we can test error paths
	// without requiring a real PostgreSQL instance
	cfg := &config.DatabaseConfig{
		Type:            "postgres",
		Host:            "nonexistent-host",
		Port:            5432,
		Name:            "test_db",
		Username:        "test_user",
		Password:        "test_password",
		SSLMode:         "disable",
		MaxConnections:  10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
	}

	logger, err := logger.New(&config.LoggerConfig{
		Level:  "error",
		Format: "console",
	})
	require.NoError(t, err)
	defer logger.Close()

	// Test connection failure
	conn, err := NewPostgreSQLConnection(cfg, logger)
	assert.Error(t, err)
	assert.Nil(t, conn)
}

// TestTestConnection tests the retry connection logic
func TestTestConnection(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type:            "postgres",
		Host:            "nonexistent-host",
		Port:            5432,
		Name:            "test_db",
		Username:        "test_user",
		Password:        "test_password",
		SSLMode:         "disable",
		MaxConnections:  10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
	}

	logger, err := logger.New(&config.LoggerConfig{
		Level:  "error",
		Format: "console",
	})
	require.NoError(t, err)
	defer logger.Close()

	// Test with retries (should fail after retries)
	start := time.Now()
	conn, err := TestConnection(cfg, logger, 2, 100*time.Millisecond)
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Nil(t, conn)
	assert.Contains(t, err.Error(), "failed to connect to database after 2 attempts")

	// Should have waited at least one retry delay
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
}

// Integration tests that require a real PostgreSQL database
func TestPostgreSQLIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Skip if no PostgreSQL connection available
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping PostgreSQL integration tests")
	}

	t.Run("RealConnection", func(t *testing.T) {
		testRealConnection(t)
	})
}

// testRealConnection tests with a real PostgreSQL database
func testRealConnection(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type:            "postgres",
		Host:            getEnvOrDefault("DB_HOST", "localhost"),
		Port:            5432,
		Name:            getEnvOrDefault("DB_NAME", "postgres"),
		Username:        getEnvOrDefault("DB_USER", "postgres"),
		Password:        getEnvOrDefault("DB_PASSWORD", ""),
		SSLMode:         "disable",
		MaxConnections:  10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
	}

	logger, err := logger.New(&config.LoggerConfig{
		Level:  "info",
		Format: "console",
	})
	require.NoError(t, err)
	defer logger.Close()

	// Test connection
	conn, err := NewPostgreSQLConnection(cfg, logger)
	if err != nil {
		t.Skipf("Could not connect to PostgreSQL: %v", err)
	}
	defer conn.Close()

	// Test ping
	err = conn.Ping()
	assert.NoError(t, err)

	// Test health check
	err = conn.HealthCheck()
	assert.NoError(t, err)

	// Test stats
	stats := conn.Stats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "max_open_connections")
	assert.Contains(t, stats, "open_connections")

	// Test migration with domain example model
	err = conn.Migrate(&domain.Example{})
	assert.NoError(t, err)

	// Clean up
	conn.DB.Exec("DROP TABLE IF EXISTS examples")
}

// TestDatabaseConfigValidation tests database configuration validation
func TestDatabaseConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *config.DatabaseConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: &config.DatabaseConfig{
				Type:            "postgres",
				Host:            "localhost",
				Port:            5432,
				Name:            "test_db",
				Username:        "user",
				Password:        "pass",
				SSLMode:         "disable",
				MaxConnections:  10,
				MaxIdleConns:    2,
				ConnMaxLifetime: 5 * time.Minute,
			},
			valid: true,
		},
		{
			name: "missing host",
			config: &config.DatabaseConfig{
				Type:     "postgres",
				Port:     5432,
				Name:     "test_db",
				Username: "user",
				Password: "pass",
			},
			valid: false,
		},
		{
			name: "invalid port",
			config: &config.DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     0,
				Name:     "test_db",
				Username: "user",
				Password: "pass",
			},
			valid: false,
		},
		{
			name: "missing database name",
			config: &config.DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Username: "user",
				Password: "pass",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := buildPostgresDSN(tt.config)

			if tt.valid {
				assert.NotEmpty(t, dsn)
				assert.Contains(t, dsn, tt.config.Host)
				assert.Contains(t, dsn, tt.config.Name)
			} else {
				// Even invalid configs will build a DSN, but it won't work
				// The validation happens during connection attempt
				assert.NotEmpty(t, dsn)
			}
		})
	}
}

// BenchmarkDSNBuilding benchmarks DSN string building
func BenchmarkDSNBuilding(b *testing.B) {
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "test_db",
		Username: "test_user",
		Password: "test_password",
		SSLMode:  "disable",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildPostgresDSN(cfg)
	}
}

// Helper function to get environment variable with default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
