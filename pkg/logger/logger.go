package logger

import (
	"fmt"
	"os"

	"example-api-template/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger with additional functionality
type Logger struct {
	*zap.Logger
}

// New creates a new logger instance based on configuration
func New(cfg *config.LoggerConfig) (*Logger, error) {
	// Create encoder config
	encoderConfig := createEncoderConfig(cfg)

	// Create encoder
	var encoder zapcore.Encoder
	switch cfg.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return nil, fmt.Errorf("unsupported log format: %s", cfg.Format)
	}

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %s", cfg.Level)
	}

	// Create output paths
	outputPaths := cfg.OutputPaths
	if len(outputPaths) == 0 {
		outputPaths = []string{"stdout"}
	}

	// Create writer syncer
	writeSyncers := make([]zapcore.WriteSyncer, 0, len(outputPaths))
	for _, path := range outputPaths {
		var ws zapcore.WriteSyncer
		if path == "stdout" {
			ws = zapcore.AddSync(os.Stdout)
		} else if path == "stderr" {
			ws = zapcore.AddSync(os.Stderr)
		} else {
			// For file paths, open the file
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return nil, fmt.Errorf("failed to open log file %s: %w", path, err)
			}
			ws = zapcore.AddSync(file)
		}
		writeSyncers = append(writeSyncers, ws)
	}

	// Combine all write syncers
	writeSyncer := zapcore.NewMultiWriteSyncer(writeSyncers...)

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Create logger options
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	if cfg.Development {
		options = append(options, zap.Development())
	}

	// Create logger
	logger := zap.New(core, options...)

	return &Logger{Logger: logger}, nil
}

// NewDevelopment creates a development logger with sensible defaults
func NewDevelopment() (*Logger, error) {
	cfg := &config.LoggerConfig{
		Level:       "debug",
		Format:      "console",
		Development: true,
		EnableColor: true,
		OutputPaths: []string{"stdout"},
	}
	return New(cfg)
}

// NewProduction creates a production logger with sensible defaults
func NewProduction() (*Logger, error) {
	cfg := &config.LoggerConfig{
		Level:       "info",
		Format:      "json",
		Development: false,
		EnableColor: false,
		OutputPaths: []string{"stdout"},
	}
	return New(cfg)
}

// createEncoderConfig creates encoder configuration based on logger config
func createEncoderConfig(cfg *config.LoggerConfig) zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:     "timestamp",
		LevelKey:    "level",
		NameKey:     "logger",
		CallerKey:   "caller",
		FunctionKey: zapcore.OmitKey,
		MessageKey:  "message",
		// StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.Development {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		if cfg.EnableColor {
			encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
	}

	if cfg.Format == "console" {
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	}

	return encoderConfig
}

// WithFields adds fields to the logger context
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.Logger.With(fields...)}
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.Logger.With(zap.Error(err))}
}

// WithRequestID adds a request ID field to the logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("request_id", requestID))}
}

// WithUserID adds a user ID field to the logger
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("user_id", userID))}
}

// WithComponent adds a component field to the logger
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("component", component))}
}

// WithOperation adds an operation field to the logger
func (l *Logger) WithOperation(operation string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("operation", operation))}
}

// LogHTTPRequest logs HTTP request details
func (l *Logger) LogHTTPRequest(method, path, userAgent, clientIP string, statusCode int, duration int64) {
	l.Logger.Info("HTTP request",
		zap.String("method", method),
		zap.String("path", path),
		zap.String("user_agent", userAgent),
		zap.String("client_ip", clientIP),
		zap.Int("status_code", statusCode),
		zap.Int64("duration_ms", duration),
	)
}

// LogDatabaseQuery logs database query details
func (l *Logger) LogDatabaseQuery(query string, args []interface{}, duration int64, err error) {
	fields := []zap.Field{
		zap.String("query", query),
		zap.Any("args", args),
		zap.Int64("duration_ms", duration),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Logger.Error("Database query failed", fields...)
	} else {
		l.Logger.Debug("Database query executed", fields...)
	}
}

// LogExternalAPICall logs external API call details
func (l *Logger) LogExternalAPICall(url, method string, statusCode int, duration int64, err error) {
	fields := []zap.Field{
		zap.String("url", url),
		zap.String("method", method),
		zap.Int("status_code", statusCode),
		zap.Int64("duration_ms", duration),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Logger.Error("External API call failed", fields...)
	} else {
		l.Logger.Info("External API call completed", fields...)
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Close closes the logger and flushes any buffered entries
func (l *Logger) Close() error {
	return l.Sync()
}

// Global logger instance for convenience
var globalLogger *Logger

// SetGlobal sets the global logger instance
func SetGlobal(logger *Logger) {
	globalLogger = logger
}

// GetGlobal returns the global logger instance
func GetGlobal() *Logger {
	if globalLogger == nil {
		// Create a default logger if none is set
		logger, _ := NewDevelopment()
		globalLogger = logger
	}
	return globalLogger
}

// Convenience functions that use the global logger

// Debug logs a debug message using the global logger
func Debug(msg string, fields ...zap.Field) {
	GetGlobal().Debug(msg, fields...)
}

// Info logs an info message using the global logger
func Info(msg string, fields ...zap.Field) {
	GetGlobal().Info(msg, fields...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, fields ...zap.Field) {
	GetGlobal().Warn(msg, fields...)
}

// Error logs an error message using the global logger
func Error(msg string, fields ...zap.Field) {
	GetGlobal().Error(msg, fields...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(msg string, fields ...zap.Field) {
	GetGlobal().Fatal(msg, fields...)
}

// Panic logs a panic message using the global logger and panics
func Panic(msg string, fields ...zap.Field) {
	GetGlobal().Panic(msg, fields...)
}
