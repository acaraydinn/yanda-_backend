package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port    string
	GinMode string

	// Database
	DatabaseURL string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string

	// Redis
	RedisURL string

	// JWT
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	// Storage
	StorageType string
	StoragePath string
	S3Bucket    string
	S3Region    string
	S3AccessKey string
	S3SecretKey string

	// FCM
	FCMServerKey string

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   int

	// Admin
	AdminEmail    string
	AdminPassword string

	// URLs
	AppURL string
	WebURL string
	APIURL string

	// RevenueCat
	RevenueCatAPIKey string

	// Twilio
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioVerifySID  string

	// SMTP Email
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	SMTPFromName string

	// Agora
	AgoraAppID          string
	AgoraAppCertificate string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		// Server
		Port:    getEnv("PORT", "8080"),
		GinMode: getEnv("GIN_MODE", "debug"),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/yandas?sslmode=disable"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "password"),
		DBName:      getEnv("DB_NAME", "yandas"),

		// Redis
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		// JWT
		JWTSecret:        getEnv("JWT_SECRET", "default-secret-change-in-production"),
		JWTAccessExpiry:  parseDuration(getEnv("JWT_ACCESS_EXPIRY", "24h")),
		JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "720h")),

		// Storage
		StorageType: getEnv("STORAGE_TYPE", "local"),
		StoragePath: getEnv("STORAGE_PATH", "./uploads"),
		S3Bucket:    getEnv("S3_BUCKET", ""),
		S3Region:    getEnv("S3_REGION", ""),
		S3AccessKey: getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey: getEnv("S3_SECRET_KEY", ""),

		// FCM
		FCMServerKey: getEnv("FCM_SERVER_KEY", ""),

		// Rate Limiting
		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvInt("RATE_LIMIT_WINDOW", 60),

		// Admin
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@yandas.app"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),

		// URLs
		AppURL: getEnv("APP_URL", "https://yandas.app"),
		WebURL: getEnv("WEB_URL", "https://yandas.app"),
		APIURL: getEnv("API_URL", "https://api.yandas.app"),

		// RevenueCat
		RevenueCatAPIKey: getEnv("REVENUECAT_API_KEY", ""),

		// Twilio
		TwilioAccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioVerifySID:  getEnv("TWILIO_VERIFY_SERVICE_SID", ""),

		// SMTP Email
		SMTPHost:     getEnv("SMTP_HOST", "mail.ubasoft.net"),
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "yandas@ubasoft.net"),
		SMTPFromName: getEnv("SMTP_FROM_NAME", "YANDAŞ"),

		// Agora
		AgoraAppID:          getEnv("AGORA_APP_ID", ""),
		AgoraAppCertificate: getEnv("AGORA_APP_CERTIFICATE", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute // Default
	}
	return d
}
