package config

import (
	"os"
	"strings"
)

type Config struct {
	Port			  	string
	DatabaseURL       	string
	LogLevel          	string
	Environment       	string
	
	AllowedOrigins		[]string
	TrustedProxies		[]string
}

func Load() *Config {
	port := ":" + os.Getenv("PORT")
	databaseURL := os.Getenv("DATABASE_URL")
	logLevel := os.Getenv("LOG_LEVEL")
	environment := os.Getenv("ENV")

	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := getOriginsFromEnv(allowedOriginsStr)

	trustedProxiesStr := os.Getenv("TRUSTED_PROXIES")
	trustedProxies := getOriginsFromEnv(trustedProxiesStr)

	return &Config{
		Port:			port,
		DatabaseURL:	databaseURL,
		LogLevel:		logLevel,
		Environment:   	environment,
		AllowedOrigins:	allowedOrigins,
		TrustedProxies: trustedProxies,
	}
}

func getOriginsFromEnv(str string) []string {
	if str == "" {
		return []string{}
	}
	origins := strings.Split(str, ",")
	result := make([]string, 0, len(origins))
    
    for _, origin := range origins {
        trimmed := strings.TrimSpace(origin)
        if trimmed != "" {
            result = append(result, trimmed)
        }
    }
    return result
}