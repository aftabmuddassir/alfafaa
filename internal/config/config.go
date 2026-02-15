package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	Upload      UploadConfig
	CORS        CORSConfig
	RateLimit   RateLimitConfig
	Security    SecurityConfig
	GoogleOAuth GoogleOAuthConfig
}

// GoogleOAuthConfig holds Google OAuth configuration
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURLs []string
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
	Host string
	Mode string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret            string
	Expiration        time.Duration
	RefreshExpiration time.Duration
}

// UploadConfig holds file upload configuration
type UploadConfig struct {
	MaxSize int64
	Path    string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int
	Duration time.Duration
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EnableRateLimit    bool
	EnableSanitization bool
	TrustedProxies     []string
	LogPath            string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "alfafaa_blog"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:            getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			Expiration:        parseDuration(getEnv("JWT_EXPIRATION", "24h")),
			RefreshExpiration: parseDuration(getEnv("JWT_REFRESH_EXPIRATION", "168h")),
		},
		Upload: UploadConfig{
			MaxSize: parseInt64(getEnv("UPLOAD_MAX_SIZE", "10485760")),
			Path:    getEnv("UPLOAD_PATH", "./uploads"),
		},
		CORS: CORSConfig{
			AllowedOrigins: parseSlice(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,https://blog-alfafaa.netlify.app")),
		},
		RateLimit: RateLimitConfig{
			Requests: parseInt(getEnv("RATE_LIMIT_REQUESTS", "100")),
			Duration: parseDuration(getEnv("RATE_LIMIT_DURATION", "1m")),
		},
		Security: SecurityConfig{
			EnableRateLimit:    parseBool(getEnv("ENABLE_RATE_LIMIT", "true")),
			EnableSanitization: parseBool(getEnv("ENABLE_SANITIZATION", "true")),
			TrustedProxies:     parseSlice(getEnv("TRUSTED_PROXIES", "127.0.0.1")),
			LogPath:            getEnv("LOG_PATH", ""),
		},
		GoogleOAuth: GoogleOAuthConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURLs: parseSlice(getEnv("GOOGLE_REDIRECT_URLS", "http://localhost:5173,http://localhost:3000")),
		},
	}, nil
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// parseDuration parses a duration string with a fallback
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

// parseInt parses an integer string with a fallback
func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 100
	}
	return i
}

// parseInt64 parses an int64 string with a fallback
func parseInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 10485760
	}
	return i
}

// parseBool parses a boolean string with a fallback
func parseBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return true
	}
	return b
}

// parseSlice parses a comma-separated string into a slice
func parseSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return "host=" + c.Host +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.DBName +
		" port=" + c.Port +
		" sslmode=" + c.SSLMode +
		" TimeZone=UTC"
}
