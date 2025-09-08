package database

import (
	"fmt"
	"time"

	"example-api-template/internal/config"
	"example-api-template/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// PostgreSQLConnection holds the database connection and configuration
type PostgreSQLConnection struct {
	DB     *gorm.DB
	Config *config.DatabaseConfig
	Logger *logger.Logger
}

// NewPostgreSQLConnection creates a new PostgreSQL database connection
func NewPostgreSQLConnection(cfg *config.DatabaseConfig, logger *logger.Logger) (*PostgreSQLConnection, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database configuration is required")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// Build DSN (Data Source Name)
	dsn := buildPostgresDSN(cfg)

	// Configure GORM logger
	gormLogLevel := gormlogger.Silent
	// Enable GORM logging for debug builds or when explicitly requested
	// We'll keep it simple and use Silent by default for production
	gormLogLevel = gormlogger.Warn

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true, // Enable prepared statement cache
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	logger.Info("Successfully connected to PostgreSQL database",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Name),
		zap.String("ssl_mode", cfg.SSLMode),
		zap.Int("max_connections", cfg.MaxConnections),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
		zap.Duration("conn_max_lifetime", cfg.ConnMaxLifetime),
	)

	return &PostgreSQLConnection{
		DB:     db,
		Config: cfg,
		Logger: logger,
	}, nil
}

// Close closes the database connection
func (c *PostgreSQLConnection) Close() error {
	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}

		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}

		c.Logger.Info("Database connection closed")
	}
	return nil
}

// Ping tests the database connection
func (c *PostgreSQLConnection) Ping() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// Stats returns database connection statistics
func (c *PostgreSQLConnection) Stats() map[string]interface{} {
	sqlDB, err := c.DB.DB()
	if err != nil {
		c.Logger.Error("Failed to get underlying sql.DB for stats", zap.Error(err))
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// HealthCheck performs a comprehensive health check
func (c *PostgreSQLConnection) HealthCheck() error {
	// Test basic connectivity
	if err := c.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Test query execution
	var version string
	if err := c.DB.Raw("SELECT version()").Scan(&version).Error; err != nil {
		return fmt.Errorf("version query failed: %w", err)
	}

	// Test transaction
	tx := c.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("transaction begin failed: %w", tx.Error)
	}

	if err := tx.Rollback().Error; err != nil {
		return fmt.Errorf("transaction rollback failed: %w", err)
	}

	c.Logger.Debug("Database health check passed",
		zap.String("version", version[:50]+"..."), // Truncate long version string
	)

	return nil
}

// Migrate runs database migrations
func (c *PostgreSQLConnection) Migrate(models ...interface{}) error {
	if len(models) == 0 {
		return fmt.Errorf("no models provided for migration")
	}

	c.Logger.Info("Starting database migration", zap.Int("models_count", len(models)))

	if err := c.DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	c.Logger.Info("Database migration completed successfully")
	return nil
}

// buildPostgresDSN builds a PostgreSQL Data Source Name from configuration
func buildPostgresDSN(cfg *config.DatabaseConfig) string {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)
	return dsn
}

// TestConnection tests the database connection with retry logic
func TestConnection(cfg *config.DatabaseConfig, logger *logger.Logger, maxRetries int, retryDelay time.Duration) (*PostgreSQLConnection, error) {
	var conn *PostgreSQLConnection
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = NewPostgreSQLConnection(cfg, logger)
		if err == nil {
			if err := conn.HealthCheck(); err == nil {
				return conn, nil
			}
			conn.Close()
		}

		if i < maxRetries-1 {
			logger.Warn("Database connection attempt failed, retrying",
				zap.Int("attempt", i+1),
				zap.Int("max_retries", maxRetries),
				zap.Duration("retry_delay", retryDelay),
				zap.Error(err),
			)
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}
