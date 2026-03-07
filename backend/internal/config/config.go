package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port        string
	BaseURL     string
	DatabaseURL string
	LogLevel    string
	Environment string

	AllowedOrigins []string
	TrustedProxies []string

	Cache CacheConfig
}

type CacheConfig struct {
	URL             string
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
}

func Load() *Config {
	return &Config{
		Port:           ":" + getEnv("PORT", "8080"),
		BaseURL:        getEnv("BASE_URL", "http://localhost:8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgresql://user:password@localhost/db"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		Environment:    getEnv("ENV", "development"),
		AllowedOrigins: getEnvURLs("ALLOWED_ORIGINS", []string{"http://localhost:8080"}),
		TrustedProxies: getEnvURLs("TRUSTED_PROXIES", []string{"http://localhost:8080"}),

		Cache: CacheConfig{
			URL:             getEnv("REDIS_URL", "redis://localhost:6379"),
			MaxRetries:      getEnvInt("REDIS_MAX_RETRIES", 3),
			MinRetryBackoff: getEnvDuration("REDIS_MIN_RETRY_BACKOFF", time.Second),
			MaxRetryBackoff: getEnvDuration("REDIS_MAX_RETRY_BACKOFF", 10*time.Second),
		},
	}
}

func getEnv(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultValue
}

func getEnvURLs(key string, defaultValue []string) []string {
	if val := os.Getenv(key); val != "" {
		origins := strings.Split(val, ",")
		result := make([]string, 0, len(origins))
		for _, origin := range origins {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}
