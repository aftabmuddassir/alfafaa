package middlewares

import (
	"net/http"
	"strings"

	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and sets user information in context
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromHeader(c)
		if token == "" {
			utils.ErrorResponseJSON(c, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization token is required", nil)
			c.Abort()
			return
		}

		claims, err := utils.ValidateAccessToken(token, jwtSecret)
		if err != nil {
			utils.ErrorResponseJSON(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token", nil)
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// OptionalAuthMiddleware extracts user info if token is present, but doesn't require it
func OptionalAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromHeader(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := utils.ValidateAccessToken(token, jwtSecret)
		if err != nil {
			// Token is invalid, but we don't block the request
			c.Next()
			return
		}

		// Set user information in context
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// extractTokenFromHeader extracts the JWT token from the Authorization header
func extractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// GetUserID returns the user ID from the context
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return ""
	}
	return userID.(string)
}

// GetUserRole returns the user role from the context
func GetUserRole(c *gin.Context) string {
	userRole, exists := c.Get("userRole")
	if !exists {
		return ""
	}
	return userRole.(string)
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated(c *gin.Context) bool {
	return GetUserID(c) != ""
}
