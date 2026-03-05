package handler

import (
	"shortener.reeler.com/backend/internal/models"
	"shortener.reeler.com/backend/internal/services"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Shorten(c *gin.Context) {
	var req models.URLRequest

	// Validate JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	// Call the service to shorten the URL
	shortenedURL, err := services.ShortenURL(req.URL)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to shorten URL"})
		return
	}

	c.JSON(200, models.URLResponse{ShortenedURL: shortenedURL})
}