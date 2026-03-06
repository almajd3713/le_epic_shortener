package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"shortener.reeler.com/backend/internal/services"
)

type RedirectHandler struct {
	service *services.RedirectorService
	logger  *slog.Logger
}

func NewRedirectHandler(redirectService *services.RedirectorService) *RedirectHandler {
	return &RedirectHandler{service: redirectService}
}

func (r *RedirectHandler) GET(c *gin.Context) {
	logger, ok := c.Get("logger")
	if !ok {
		logger = r.logger
	}
	reqLogger := logger.(*slog.Logger).With("handler", "RedirectHandler")

	shortCode := c.Param("code")
	originalURL, err := r.service.Redirect(shortCode)
	if err != nil {
		reqLogger.Error("failed to redirect URL", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}
