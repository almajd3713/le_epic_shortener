package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"

	"shortener.reeler.com/backend/internal/cache"
	"shortener.reeler.com/backend/internal/config"
	"shortener.reeler.com/backend/internal/db"
	"shortener.reeler.com/backend/internal/handlers"
	"shortener.reeler.com/backend/internal/middleware"
	"shortener.reeler.com/backend/internal/repository"
	"shortener.reeler.com/backend/internal/server"
	"shortener.reeler.com/backend/internal/services"
)

func main() {
	godotenv.Load()

	startServer()
}

func startServer() {
	// Load configuration
	cfg := config.Load()

	// Setup Logger
	var logLevel slog.Level
	switch strings.ToUpper(cfg.LogLevel) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	environment := cfg.Environment
	var logHandler slog.Handler
	if environment == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, opts)
	}
	logger := slog.New(logHandler)

	// Database
	ctx := context.Background()
	connString := cfg.DatabaseURL
	pool, err := db.NewPool(ctx, connString)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer pool.Close()
	// Run database migrations
	if err := db.RunMigrations(ctx, pool); err != nil {
		panic("Failed to run migrations: " + err.Error())
	}

	// Cache
	cacheClient, err := cache.NewRedisClient(ctx, cfg.Cache)
	if err != nil {
		panic("Failed to connect to cache: " + err.Error())
	}
	defer cacheClient.Close()
	cache := cache.NewCache(cacheClient)

	// Repositories
	urlRepo := repository.NewURLRepository(pool)
	clickRepo := repository.NewClickRepository(pool)

	// Services
	cacheService := services.NewCacheService(cache, logger)
	urlService := services.NewURLService(urlRepo, cacheService, logger)
	clickService := services.NewClickService(clickRepo, logger)
	shortenerService := services.NewShortenerService(urlRepo, cacheService, logger)
	redirectService := services.NewRedirectorService(urlService, clickService, cacheService, logger)

	// Handlers
	urlHandler := handlers.NewURLHandler(urlService)
	shortenerHandler := handlers.NewShortenerHandler(shortenerService)
	redirectHandler := handlers.NewRedirectHandler(redirectService)
	clickHandler := handlers.NewClickHandler(clickService)

	// Metrics
	metricsRegistry := prometheus.NewRegistry()
	middleware.InitializeMetrics(metricsRegistry)

	// Routes
	r := gin.New()
	r.SetTrustedProxies(cfg.TrustedProxies)

	r.Use(gin.Recovery())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.BaseURL(cfg.BaseURL))
	r.Use(middleware.Metrics())

	// CORs
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.AllowedOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	corsConfig.AllowCredentials = false

	r.Use(middleware.CORS(corsConfig))

	server.SetupRoutes(r,
		*urlHandler,
		*shortenerHandler,
		*redirectHandler,
		*clickHandler,
		metricsRegistry,
	)

	PORT := cfg.Port
	logger.Info("Server starting on port " + PORT)
	r.Run(
		PORT,
	)
}
