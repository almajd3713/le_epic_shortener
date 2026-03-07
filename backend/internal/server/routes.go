package server

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"shortener.reeler.com/backend/internal/handlers"
)

func SetupRoutes(r *gin.Engine, 
	urlHandler handlers.URLHandler,
	shortenerHandler handlers.ShortenerHandler,
	redirectHandler handlers.RedirectHandler,
) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/:code", redirectHandler.GET) 
	r.GET("/geturl/:code", urlHandler.GET)
	r.POST("/api/shorten", shortenerHandler.POST)
	r.GET("/api/urls", urlHandler.GET_ALL)

	r.PATCH("/api/toggle/:code", urlHandler.PATCH)
	r.DELETE("/api/delete/:code", urlHandler.DELETE)
}