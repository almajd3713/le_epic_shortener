package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"shortener.reeler.com/backend/internal/services"
)
type ClickHandler struct {
	service *services.ClickService
	logger  *slog.Logger
}

func NewClickHandler(clickService *services.ClickService) *ClickHandler {
	return &ClickHandler{service: clickService}
}

func (h *ClickHandler) GET(c *gin.Context) {
	logger, ok := c.Get("logger")
	if !ok {
		logger = h.logger
	}
	reqLogger := logger.(*slog.Logger).With("handler", "ClickHandler")

	code := c.Param("code")
	clicks, err := h.service.GetClicksByCode(c, code)
	if err != nil {
		reqLogger.Error("failed to get clicks by code", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clicks"})
		return
	}

	c.JSON(http.StatusOK, clicks)
}

func (h *ClickHandler) COUNT(c *gin.Context) {
	logger, ok := c.Get("logger")
	if !ok {
		logger = h.logger
	}
	reqLogger := logger.(*slog.Logger).With("handler", "ClickHandler")

	code := c.Param("code")
	count, err := h.service.GetClickCountByCode(c, code)
	if err != nil {
		reqLogger.Error("failed to get click count by code", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch click count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}
