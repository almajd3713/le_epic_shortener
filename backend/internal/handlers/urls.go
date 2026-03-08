package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"shortener.reeler.com/backend/internal/models"
	"shortener.reeler.com/backend/internal/services"
)

type URLHandler struct {
	service *services.URLService
	logger  *slog.Logger
}

func NewURLHandler(urlService *services.URLService) *URLHandler {
	return &URLHandler{service: urlService}
}

func (h *URLHandler) GET(c *gin.Context) {
	logger, ok := c.Get("logger")
	if !ok {
		logger = slog.New(slog.NewTextHandler(nil, nil))
	}
	reqLogger := logger.(*slog.Logger).With("handler", "GET")

	shortCode := c.Param("code")
	longURL, err := h.service.GetOriginalURL(c, shortCode)
	if err != nil {
		reqLogger.Error("failed to get long URL", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"long_url": longURL})
}

func (h *URLHandler) GET_ALL(c *gin.Context) {
	logger, ok := c.Get("logger")
	if !ok {
		logger = h.logger
	}
	reqLogger := logger.(*slog.Logger).With("handler", "URLHandler")
	urls, err := h.service.GetAllURLs()
	if err != nil {
		reqLogger.Error("failed to get URLs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve URLs"})
		return
	}

	baseURL, _ := c.Get("baseURL")
	items := make([]models.URLListItem, len(urls))
	for i, u := range urls {
		var expiresAt *string
		if u.ExpiresAt != nil {
			s := u.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
			expiresAt = &s
		}
		items[i] = models.URLListItem{
			ShortCode: u.ShortCode,
			LongURL:   u.LongURL,
			ShortURL:  baseURL.(string) + "/" + u.ShortCode,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			ExpiresAt: expiresAt,
			IsActive:  u.IsActive,
		}
	}
	c.JSON(http.StatusOK, items)
}

// PATCH is responsible for activating/deactivating a shortened URL. The request body should contain an "action" field with value "activate" or "deactivate".
func (h *URLHandler) PATCH(c *gin.Context) {
	logger, ok := c.Get("logger")
	if !ok {
		logger = h.logger
	}
	reqLogger := logger.(*slog.Logger).With("handler", "URLHandler")

	shortCode := c.Param("code")

	var req models.URLUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reqLogger.Error("invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	switch req.Action {
	case "activate":
		err := h.service.ActivateURL(c, shortCode)
		if err != nil {
			reqLogger.Error("failed to activate URL", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate URL"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "URL activated successfully"})
	case "deactivate":
		err := h.service.DeactivateURL(c, shortCode)
		if err != nil {
			reqLogger.Error("failed to deactivate URL", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate URL"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "URL deactivated successfully"})
	default:
		reqLogger.Error("invalid action", "action", req.Action)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
	}
}

func (h *URLHandler) DELETE(c *gin.Context) {
	logger, ok := c.Get("logger")
	if !ok {
		logger = h.logger
	}
	reqLogger := logger.(*slog.Logger).With("handler", "URLHandler")

	shortCode := c.Param("code")
	err := h.service.DeleteURL(c, shortCode)
	if err != nil {
		reqLogger.Error("failed to delete URL", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete URL"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "URL deleted successfully"})
}
