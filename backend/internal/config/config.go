package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
}

func Load() Config {
	port := envOrDefault("PORT", "3000")
	databaseURL := envOrDefault("DATABASE_URL", "postgresql://bet_user:bet_password@localhost:5432/bet_mvp?sslmode=disable")
	redisURL := envOrDefault("REDIS_URL", "redis://localhost:6379")

	return Config{
		Port:        port,
		DatabaseURL: databaseURL,
		RedisURL:    redisURL,
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
