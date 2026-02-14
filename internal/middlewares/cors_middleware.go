package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Request-ID"},
		ExposedHeaders:   []string{"Content-Length", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// ProductionCORSConfig returns a CORS configuration for production
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	return CORSConfig{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Request-ID"},
		ExposedHeaders:   []string{"Content-Length", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           43200, // 12 hours
	}
}

// CORSMiddleware creates a CORS middleware with the given configuration
func CORSMiddleware(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		utils.Debug("CORS: request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("origin", origin),
		)

		// Check if the origin is allowed
		allowedOrigin := ""
		for _, o := range config.AllowedOrigins {
			if o == "*" {
				allowedOrigin = "*"
				break
			}
			if o == origin {
				allowedOrigin = origin
				break
			}
		}

		if allowedOrigin == "" {
			utils.Warn("CORS: origin not allowed",
				zap.String("origin", origin),
				zap.Strings("allowed", config.AllowedOrigins),
			)
		}

		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))

			if config.AllowCredentials && allowedOrigin != "*" {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			if config.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
			}
		}

		// Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			utils.Debug("CORS: preflight response", zap.String("allowed_origin", allowedOrigin))
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CORSWithOrigins creates a CORS middleware with specific allowed origins
func CORSWithOrigins(origins []string) gin.HandlerFunc {
	config := DefaultCORSConfig()
	config.AllowedOrigins = origins
	return CORSMiddleware(config)
}

// DevelopmentCORS creates a permissive CORS middleware for development
// WARNING: Do not use in production!
func DevelopmentCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
