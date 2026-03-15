package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"shortener.reeler.com/backend/internal/handlers"
)

func SetupRoutes(r *gin.Engine,
	urlHandler handlers.URLHandler,
	shortenerHandler handlers.ShortenerHandler,
	redirectHandler handlers.RedirectHandler,
	clickHandler handlers.ClickHandler,
	metricsRegistry *prometheus.Registry,
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

	r.GET("/api/clicks/:code", clickHandler.GET)
	r.GET("/api/clicks/:code/count", clickHandler.COUNT)

	metricsHandler := promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})
	r.GET("/metrics", func(c *gin.Context) {
		metricsHandler.ServeHTTP(c.Writer, c.Request)
	})
}
