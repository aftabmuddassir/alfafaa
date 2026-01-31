package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from panics and logs them with structured logging
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Capture stack trace
				stack := string(debug.Stack())

				// Get request ID for tracing
				requestID := GetRequestID(c)

				// Log the panic with full details
				if utils.Logger != nil {
					utils.Logger.Error("Panic recovered",
						zap.Any("error", err),
						zap.String("stack", stack),
						zap.String("request_id", requestID),
						zap.String("path", c.Request.URL.Path),
						zap.String("method", c.Request.Method),
						zap.String("ip", c.ClientIP()),
						zap.String("user_agent", c.Request.UserAgent()),
					)
				}

				// Return error response to client
				// Don't expose internal details in production
				utils.ErrorResponseJSON(c, http.StatusInternalServerError,
					"INTERNAL_ERROR",
					"An internal server error occurred. Please try again later.",
					nil)

				// Abort further processing
				c.Abort()
			}
		}()

		c.Next()
	}
}

// RecoveryWithWriter recovers from panics and logs them with a custom writer
// Useful for when Logger is not initialized yet
func RecoveryWithWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Capture stack trace
				stack := string(debug.Stack())

				// Log to stderr if Logger is not available
				fmt.Printf("[PANIC RECOVERED] Error: %v\nStack: %s\n", err, stack)

				// Return error response
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "An internal server error occurred. Please try again later.",
					},
				})
			}
		}()

		c.Next()
	}
}
