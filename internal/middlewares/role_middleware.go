package middlewares

import (
	"net/http"

	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
)

// RequireRole creates a middleware that requires one of the specified roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if userRole == "" {
			utils.ErrorResponseJSON(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
			c.Abort()
			return
		}

		allowed := false
		for _, role := range roles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			utils.ErrorResponseJSON(c, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireMinRole creates a middleware that requires at least the specified role level
func RequireMinRole(minRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoleStr := GetUserRole(c)
		if userRoleStr == "" {
			utils.ErrorResponseJSON(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
			c.Abort()
			return
		}

		userRole := models.UserRole(userRoleStr)
		if !userRole.HasPermission(minRole) {
			utils.ErrorResponseJSON(c, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAuthor requires at least author role
func RequireAuthor() gin.HandlerFunc {
	return RequireMinRole(models.RoleAuthor)
}

// RequireEditor requires at least editor role
func RequireEditor() gin.HandlerFunc {
	return RequireMinRole(models.RoleEditor)
}

// RequireAdmin requires admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(string(models.RoleAdmin))
}

// IsEditor checks if the user has at least editor role
func IsEditor(c *gin.Context) bool {
	userRoleStr := GetUserRole(c)
	if userRoleStr == "" {
		return false
	}
	userRole := models.UserRole(userRoleStr)
	return userRole.HasPermission(models.RoleEditor)
}

// IsAdmin checks if the user is an admin
func IsAdmin(c *gin.Context) bool {
	return GetUserRole(c) == string(models.RoleAdmin)
}
