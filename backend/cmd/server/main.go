package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

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
	// Setup Logger
	level := os.Getenv("LOG_LEVEL")
	var logLevel slog.Level
	switch strings.ToUpper(level) {
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
	environment := os.Getenv("ENV")
	var logHandler slog.Handler
	if environment == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, opts)
	}
	logger := slog.New(logHandler)

	// Initialize Database
	ctx := context.Background()
	connString := os.Getenv("DATABASE_URL")
	pool, err := db.NewPool(ctx, connString)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer pool.Close()

	// Run database migrations
	if err := db.RunMigrations(ctx, pool); err != nil {
		panic("Failed to run migrations: " + err.Error())
	}

	// Repositories
	urlRepo := repository.NewURLRepository(pool)

	// Services
	shortenerService := services.NewShortenerService(*urlRepo, logger)
	redirectService := services.NewRedirectorService(*urlRepo, logger)

	// Handlers
	shortenerHandler := handlers.NewShortenerHandler(shortenerService)
	redirectHandler := handlers.NewRedirectHandler(redirectService)

	// Routes
	r := gin.Default()

	r.Use(gin.Recovery())
	r.Use(middleware.Logger(logger))

	server.SetupRoutes(r,
		*shortenerHandler,
		*redirectHandler,
	)

	PORT := ":" + os.Getenv("PORT")
	logger.Info("Server starting on port " + PORT)
	r.Run(
		PORT,
	)
}
