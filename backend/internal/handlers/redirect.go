package handlers

import (
	"github.com/gin-gonic/gin"

	"shortener.reeler.com/backend/internal/services"
)

type RedirectHandler struct {
	redirectService services.RedirectorService
}

func NewRedirectHandler(redirectService services.RedirectorService) *RedirectHandler {
	return &RedirectHandler{redirectService: redirectService}
}

func (r *RedirectHandler) GET(c *gin.Context) {
	shortCode := c.Param("code")
	originalURL, err := r.redirectService.Redirect(shortCode)
	if err != nil {
		c.JSON(404, gin.H{"error": "URL not found"})
		return
	}

	c.Redirect(302, originalURL)
}