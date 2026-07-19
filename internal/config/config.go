package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Admin    AdminConfig
	Log      LogConfig
	Security SecurityConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
	Host string
	Mode string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Driver       string
	SQLitePath   string
	PostgresHost string
	PostgresPort string
	PostgresUser string
	PostgresPass string
	PostgresDB   string
	PostgresSSL  string
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret      string
	ExpiryHours int
	Issuer      string
}

// AdminConfig holds admin user configuration
type AdminConfig struct {
	Username string
	Password string
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string
	Format string
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EncryptionKey      string
	RateLimitEnabled   bool
	RateLimitPerMinute int
	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	CORSAllowedHeaders []string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "2500"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Driver:       getEnv("DB_DRIVER", "sqlite"),
			SQLitePath:   getEnv("SQLITE_PATH", "./storage/One.db"),
			PostgresHost: getEnv("POSTGRES_HOST", "localhost"),
			PostgresPort: getEnv("POSTGRES_PORT", "5432"),
			PostgresUser: getEnv("POSTGRES_USER", "One"),
			PostgresPass: getEnv("POSTGRES_PASSWORD", "One"),
			PostgresDB:   getEnv("POSTGRES_DB", "One"),
			PostgresSSL:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "One-secret-key"),
			ExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),
			Issuer:      getEnv("JWT_ISSUER", "One-go"),
		},
		Admin: AdminConfig{
			Username: getEnv("ADMIN_USERNAME", "admin"),
			Password: getEnv("ADMIN_PASSWORD", "admin123"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "debug"),
			Format: getEnv("LOG_FORMAT", "console"),
		},
		Security: SecurityConfig{
			EncryptionKey:      getEnv("STORAGE_ENCRYPTION_KEY", ""),
			RateLimitEnabled:   getEnvBool("RATE_LIMIT_ENABLED", true),
			RateLimitPerMinute: getEnvInt("RATE_LIMIT_PER_MINUTE", 60),
			CORSAllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			CORSAllowedMethods: getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			CORSAllowedHeaders: getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization", "X-API-Key"}),
		},
	}

	return cfg, nil
}

// GetJWTExpiryDuration returns JWT expiry as time.Duration
func (c *JWTConfig) GetJWTExpiryDuration() time.Duration {
	return time.Duration(c.ExpiryHours) * time.Hour
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	if c.Driver == "postgres" {
		return "host=" + c.PostgresHost +
			" user=" + c.PostgresUser +
			" password=" + c.PostgresPass +
			" dbname=" + c.PostgresDB +
			" port=" + c.PostgresPort +
			" sslmode=" + c.PostgresSSL
	}
	return c.SQLitePath
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
