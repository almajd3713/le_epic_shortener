package handlers

import (
	"log/slog"
	"net/http"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	newUrl, err := h.service.ShortenURL(req.LongURL)
	if err != nil {
		reqLogger.Error("failed to shorten URL", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to shorten URL"})
		return
	}
	var res models.URLResponse
	res.ShortCode = newUrl.ShortCode
	res.ShortenedURL = c.Request.Host + "/" + newUrl.ShortCode

	reqLogger.Info("URL shortened successfully", "short_code", newUrl.ShortCode, "long_url", req.LongURL)
	c.JSON(http.StatusOK, res)
}
