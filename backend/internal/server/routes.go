package server

import (
	"github.com/gin-gonic/gin"
	"shortener.reeler.com/backend/internal/handler"
)

func SetupRoutes(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/api/shorten", handler.ShortenURL)
}