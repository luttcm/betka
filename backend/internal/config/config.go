package config

import (
	"os"
	"time"
)

type Config struct {
	Port                   string
	DatabaseURL            string
	RequirePostgres        bool
	RedisURL               string
	AuthJWTSecret          string
	AuthTokenTTL           time.Duration
	EmailFrom              string
	EmailVerifyBaseURL     string
	SMTPHost               string
	SMTPPort               string
	SMTPUsername           string
	SMTPPassword           string
	BootstrapAdminEmail    string
	BootstrapAdminPassword string
}

func Load() Config {
	port := envOrDefault("PORT", "3000")
	databaseURL := envOrDefault("DATABASE_URL", "postgresql://bet_user:bet_password@localhost:5432/bet_mvp?sslmode=disable")
	requirePostgres := envBoolOrDefault("REQUIRE_POSTGRES", true)
	redisURL := envOrDefault("REDIS_URL", "redis://localhost:6379")
	authJWTSecret := envOrDefault("AUTH_JWT_SECRET", "dev-secret")
	authTokenTTL := envDurationOrDefault("AUTH_TOKEN_TTL", 24*time.Hour)
	emailFrom := envOrDefault("EMAIL_FROM", "noreply@bet.local")
	emailVerifyBaseURL := envOrDefault("EMAIL_VERIFY_BASE_URL", "http://localhost:3000/v1/auth/verify-email")
	smtpHost := envOrDefault("SMTP_HOST", "")
	smtpPort := envOrDefault("SMTP_PORT", "")
	smtpUsername := envOrDefault("SMTP_USERNAME", "")
	smtpPassword := envOrDefault("SMTP_PASSWORD", "")
	bootstrapAdminEmail := envOrDefault("BOOTSTRAP_ADMIN_EMAIL", "")
	bootstrapAdminPassword := envOrDefault("BOOTSTRAP_ADMIN_PASSWORD", "")

	return Config{
		Port:                   port,
		DatabaseURL:            databaseURL,
		RequirePostgres:        requirePostgres,
		RedisURL:               redisURL,
		AuthJWTSecret:          authJWTSecret,
		AuthTokenTTL:           authTokenTTL,
		EmailFrom:              emailFrom,
		EmailVerifyBaseURL:     emailVerifyBaseURL,
		SMTPHost:               smtpHost,
		SMTPPort:               smtpPort,
		SMTPUsername:           smtpUsername,
		SMTPPassword:           smtpPassword,
		BootstrapAdminEmail:    bootstrapAdminEmail,
		BootstrapAdminPassword: bootstrapAdminPassword,
	}
}

func envBoolOrDefault(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "TRUE", "True", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "False", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
