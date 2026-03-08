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

	newUrl, err := h.service.ShortenURL(c, req.LongURL, req.ExpiresAt)
	if err != nil {
		reqLogger.Error("failed to shorten URL", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to shorten URL"})
		return
	}
	baseURL, _ := c.Get("baseURL")
	res := models.URLResponse{
		ShortCode: newUrl.ShortCode,
		ShortURL:  baseURL.(string) + "/" + newUrl.ShortCode,
		CreatedAt: newUrl.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	reqLogger.Info("URL shortened successfully", "short_code", newUrl.ShortCode, "long_url", req.LongURL)
	c.JSON(http.StatusOK, res)
}
