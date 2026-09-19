package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	MongoURI      string
	MongoDatabase string
	RedisURL      string
	JWTSecret     string
	JWTExpires    time.Duration
	FrontendURL   string
	CORSOrigin    string
}

func Load() Config {
	_ = godotenv.Load()
	hours := 24
	if raw := os.Getenv("JWT_EXPIRES_HOURS"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			hours = parsed
		}
	}
	return Config{
		Port:          get("PORT", "8080"),
		MongoURI:      get("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase: get("MONGO_DATABASE", "live_polling"),
		RedisURL:      get("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:     get("JWT_SECRET", "dev-only-change-me"),
		JWTExpires:    time.Duration(hours) * time.Hour,
		FrontendURL:   get("FRONTEND_URL", "http://localhost:4173"),
		CORSOrigin:    get("CORS_ORIGIN", "http://localhost:5173,http://localhost:4173"),
	}
}

func get(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
