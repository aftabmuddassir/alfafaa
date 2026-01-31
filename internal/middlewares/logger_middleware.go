package middlewares

import (
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerMiddleware logs all HTTP requests with structured logging
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for health checks and swagger
		path := c.Request.URL.Path
		if path == "/health" || path == "/api/v1/health" {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get response status
		status := c.Writer.Status()

		// Get request ID and user ID from context
		requestID := GetRequestID(c)
		userID := GetUserID(c)

		// Build log fields
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("response_size", c.Writer.Size()),
		}

		// Add optional fields if present
		if query != "" {
			fields = append(fields, zap.String("query", query))
		}
		if requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}
		if userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}

		// Log at appropriate level based on status code
		if utils.Logger != nil {
			switch {
			case status >= 500:
				utils.Logger.Error("HTTP Request", fields...)
			case status >= 400:
				utils.Logger.Warn("HTTP Request", fields...)
			default:
				utils.Logger.Info("HTTP Request", fields...)
			}
		}

		// Log errors if any
		if len(c.Errors) > 0 && utils.Logger != nil {
			for _, err := range c.Errors {
				utils.Logger.Error("Request error",
					zap.String("request_id", requestID),
					zap.Error(err.Err),
					zap.String("meta", err.Meta.(string)),
				)
			}
		}
	}
}

// LoggerMiddlewareWithSkipPaths returns a logger middleware that skips certain paths
func LoggerMiddlewareWithSkipPaths(skipPaths []string) gin.HandlerFunc {
	skipMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipMap[path] = true
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip logging for specified paths
		if skipMap[path] {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get response status
		status := c.Writer.Status()

		// Get request ID and user ID from context
		requestID := GetRequestID(c)
		userID := GetUserID(c)

		// Build log fields
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.Int("response_size", c.Writer.Size()),
		}

		if query != "" {
			fields = append(fields, zap.String("query", query))
		}
		if requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}
		if userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}

		if utils.Logger != nil {
			switch {
			case status >= 500:
				utils.Logger.Error("HTTP Request", fields...)
			case status >= 400:
				utils.Logger.Warn("HTTP Request", fields...)
			default:
				utils.Logger.Info("HTTP Request", fields...)
			}
		}
	}
}
