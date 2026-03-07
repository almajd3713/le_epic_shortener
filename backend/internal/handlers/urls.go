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
		}
	}
	c.JSON(http.StatusOK, items)
}