package server

import (
	"os"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
)

// Handler struct, to hold paths, 


func main() {
	godotenv.Load()

	// Load environment variables from .env file
	PORT := ":" + os.Getenv("PORT")

	startServer(PORT)
}

func startServer(PORT string) {
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