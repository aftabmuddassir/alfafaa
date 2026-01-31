package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for request ID
	RequestIDKey = "RequestID"
)

// RequestIDMiddleware generates or extracts request ID for tracing
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID exists in header (from upstream proxy or client)
		requestID := c.GetHeader(RequestIDHeader)

		// Generate new ID if not provided
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context for access in handlers/services
		c.Set(RequestIDKey, requestID)

		// Add to response headers for client tracking
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
