package utils

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global structured logger instance
var Logger *zap.Logger

// parseLogLevel converts a LOG_LEVEL env var string to a zapcore.Level.
// Supported values: debug, info, warn, error. Defaults to info for production, debug for development.
func parseLogLevel(levelStr string, mode string) zapcore.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		// Default based on mode
		if mode == "release" || mode == "production" {
			return zapcore.InfoLevel
		}
		return zapcore.DebugLevel
	}
}

// InitLogger initializes the global logger based on the application mode.
// Log level can be overridden via the LOG_LEVEL environment variable (debug, info, warn, error).
func InitLogger(mode string) error {
	var err error

	logLevel := parseLogLevel(os.Getenv("LOG_LEVEL"), mode)

	if mode == "release" || mode == "production" {
		// Production logger - JSON format
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(logLevel)
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		Logger, err = config.Build()
	} else {
		// Development logger - console format
		config := zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(logLevel)
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		Logger, err = config.Build()
	}

	if err != nil {
		return err
	}

	return nil
}

// InitLoggerWithFile initializes logger that writes to both console and file
func InitLoggerWithFile(mode string, logPath string) error {
	// Create log file
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	if mode == "release" || mode == "production" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create cores for console and file
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// Set log level
	level := zapcore.InfoLevel
	if mode == "debug" {
		level = zapcore.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(file), level),
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

// CloseLogger flushes any buffered log entries
func CloseLogger() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	}
}
