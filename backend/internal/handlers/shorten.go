package handlers

import (
	"github.com/gin-gonic/gin"

	"shortener.reeler.com/backend/internal/models"
	"shortener.reeler.com/backend/internal/services"
)

type ShortenerHandler struct {
	shortenerService services.ShortenerService
}

func NewShortenerHandler(shortenerService services.ShortenerService) *ShortenerHandler {
	return &ShortenerHandler{shortenerService: shortenerService}
}

func (h *ShortenerHandler) POST(c *gin.Context) {
	var req models.URLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	code, err := h.shortenerService.ShortenURL(req.LongURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to shorten URL"})
		return
	}

	c.JSON(200, gin.H{"shortened_url": code})
}