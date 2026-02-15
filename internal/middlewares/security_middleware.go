package middlewares

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	// return func(c *gin.Context) {
	// 	// Prevent clickjacking attacks
	// 	// Denies rendering the page in a frame/iframe
	// 	c.Header("X-Frame-Options", "SAMEORIGIN")

	// 	// Prevent MIME type sniffing
	// 	// Forces browser to respect Content-Type header
	// 	c.Header("X-Content-Type-Options", "nosniff")

	// 	// Enable XSS protection in older browsers
	// 	// Modern browsers use CSP instead
	// 	c.Header("X-XSS-Protection", "1; mode=block")

	// 	// Control referrer information
	// 	// Only send origin for cross-origin requests
	// 	//c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

	// 	// Content Security Policy
	// 	// Restricts resource loading to prevent XSS
	// 	// Note: This is a basic API CSP, adjust for frontend needs
	// 	c.Header("Content-Security-Policy",
	// 		"default-src 'self'; "+
	// 			"script-src 'self' 'unsafe-inline'; "+
	// 			"style-src 'self' 'unsafe-inline'; "+
	// 			"img-src 'self' data: https:; "+
	// 			"font-src 'self'; "+
	// 			"connect-src '*'; "+
	// 			"frame-ancestors 'none'")

	// 	// Strict Transport Security (HTTPS only)
	// 	// Force HTTPS for 1 year, including subdomains
	// 	// Only set when using TLS
	// 	if c.Request.TLS != nil {
	// 		c.Header("Strict-Transport-Security",
	// 			"max-age=31536000; includeSubDomains; preload")
	// 	}

	// 	// Permissions Policy (formerly Feature Policy)
	// 	// Disable unused browser features
	// 	c.Header("Permissions-Policy",
	// 		"geolocation=(), "+
	// 			"microphone=(), "+
	// 			"camera=(), "+
	// 			"payment=(), "+
	// 			"usb=(), "+
	// 			"magnetometer=(), "+
	// 			"gyroscope=()")

	// 	// Cache Control for API responses
	// 	// Prevent caching of sensitive data
	// 	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
	// 	c.Header("Pragma", "no-cache")
	// 	c.Header("Expires", "0")

	// 	c.Next()
	// }

	return func(c *gin.Context) {
		// // 1. Strict Transport Security (HTTPS only)
		// // Tells the browser "Always talk to me over HTTPS"
		// if c.Request.TLS != nil {
		// 	c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		// }

		// // 2. Referrer Policy
		// // Controls how much info is sent when linking to other sites
		// c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// // 3. Content Type Options
		// // Prevents the browser from "guessing" the file type (security best practice)
		// c.Header("X-Content-Type-Options", "nosniff")

		// // REMOVED: Content-Security-Policy (Not needed for JSON API)
		// // REMOVED: X-Frame-Options (Not needed for JSON API)
		// // REMOVED: X-XSS-Protection (Not needed for JSON API)

		// c.Next()
	}


}

// SecurityHeadersConfig allows customizing security headers
type SecurityHeadersConfig struct {
	// FrameOptions controls X-Frame-Options (default: DENY)
	FrameOptions string
	// ContentTypeNosniff controls X-Content-Type-Options (default: true)
	ContentTypeNosniff bool
	// XSSProtection controls X-XSS-Protection (default: true)
	XSSProtection bool
	// ReferrerPolicy controls Referrer-Policy (default: strict-origin-when-cross-origin)
	ReferrerPolicy string
	// CSP controls Content-Security-Policy (empty to disable)
	CSP string
	// HSTS controls Strict-Transport-Security (default: enabled for TLS)
	HSTSMaxAge int
	// PermissionsPolicy controls Permissions-Policy (empty to disable)
	PermissionsPolicy string
}

// DefaultSecurityHeadersConfig returns default security headers configuration
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		FrameOptions:       "DENY",
		ContentTypeNosniff: true,
		XSSProtection:      true,
		ReferrerPolicy:     "strict-origin-when-cross-origin",
		CSP: "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'",
		HSTSMaxAge: 31536000, // 1 year
		PermissionsPolicy: "geolocation=(), microphone=(), camera=(), " +
			"payment=(), usb=(), magnetometer=(), gyroscope=()",
	}
}

// SecurityHeadersWithConfig returns a middleware with custom configuration
func SecurityHeadersWithConfig(config SecurityHeadersConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.FrameOptions != "" {
			c.Header("X-Frame-Options", config.FrameOptions)
		}

		if config.ContentTypeNosniff {
			c.Header("X-Content-Type-Options", "nosniff")
		}

		if config.XSSProtection {
			c.Header("X-XSS-Protection", "1; mode=block")
		}

		if config.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", config.ReferrerPolicy)
		}

		if config.CSP != "" {
			c.Header("Content-Security-Policy", config.CSP)
		}

		if c.Request.TLS != nil && config.HSTSMaxAge > 0 {
			c.Header("Strict-Transport-Security",
				"max-age="+string(rune(config.HSTSMaxAge))+"; includeSubDomains; preload")
		}

		if config.PermissionsPolicy != "" {
			c.Header("Permissions-Policy", config.PermissionsPolicy)
		}

		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		c.Next()
	}
}
