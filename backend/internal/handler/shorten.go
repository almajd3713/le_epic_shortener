package handler

import (
	"github.com/gin-gonic/gin"
)

var json struct {
	URL string `json:"url" binding:"required"`
}

func (h *Handler) Shorten(c *gin.Context) {
	req := c.Request
	if req.Method != "POST" {
		c.JSON(405, gin.H{"error": "Method not allowed"})
		return
	}

	// Validate JSON body
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	
}