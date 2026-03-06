package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"shortener.reeler.com/backend/internal/models"
	"shortener.reeler.com/backend/internal/services"
)

type ShortenerHandler struct {
	service *services.ShortenerService
	logger  *slog.Logger
}

func NewShortenerHandler(shortenerService *services.ShortenerService) *ShortenerHandler {
	return &ShortenerHandler{service: shortenerService}
}

func (h *ShortenerHandler) POST(c *gin.Context) {
	logger, ok := c.Get("logger")
	if !ok {
		logger = h.logger
	}
	reqLogger := logger.(*slog.Logger).With("handler", "ShortenerHandler")

	var req models.URLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reqLogger.Warn("invalid JSON body", "error", err)
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	code, err := h.service.ShortenURL(req.LongURL)
	if err != nil {
		reqLogger.Error("failed to shorten URL", "error", err)
		c.JSON(500, gin.H{"error": "Failed to shorten URL"})
		return
	}

	c.JSON(200, gin.H{"shortened_url": code})
}
