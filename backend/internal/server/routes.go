package server

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"shortener.reeler.com/backend/internal/handlers"
)

func SetupRoutes(r *gin.Engine, 
	shortenerHandler handlers.ShortenerHandler,
	redirectHandler handlers.RedirectHandler,
) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/:code", redirectHandler.GET) 
	r.POST("/api/shorten", shortenerHandler.POST)
}