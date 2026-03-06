package main

import (
	"context"
	"os"
	"strings"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"shortener.reeler.com/backend/internal/db"
	"shortener.reeler.com/backend/internal/server"
	"shortener.reeler.com/backend/internal/repository"
	"shortener.reeler.com/backend/internal/services"
	"shortener.reeler.com/backend/internal/handlers"
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
	case "DEBUG": logLevel = slog.LevelDebug
	case "WARN": logLevel = slog.LevelWarn
	case "ERROR": logLevel = slog.LevelError
	default: logLevel = slog.LevelInfo
	}
	
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logHandler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

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
	shortenerService := services.NewShortenerService(*urlRepo)
	redirectService := services.NewRedirectorService(*urlRepo)

	// Handlers
	shortenerHandler := handlers.NewShortenerHandler(*shortenerService)
	redirectHandler := handlers.NewRedirectHandler(*redirectService)

	// Routes
	r := gin.Default()
	server.SetupRoutes(r,
		*shortenerHandler,
		*redirectHandler,
	)

	PORT := os.Getenv("PORT")
	r.Run(
		PORT,
	)
}
