package server

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"shortener.reeler.com/backend/internal/db"
)

// Handler struct, to hold paths,


func main() {
	godotenv.Load()

	// Load environment variables from .env file
	PORT := ":" + os.Getenv("PORT")

	startServer(PORT)
}

func startServer(PORT string) {
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

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Run(
		PORT,
	)
}
