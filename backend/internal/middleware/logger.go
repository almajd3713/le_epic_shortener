package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		
		// Request ID for tracing. We set it in context too
		requestID := uuid.New().String()
		c.Set("requestID", requestID)

		reqLogger := logger.With(
			slog.String("request_id", requestID),
			slog.String("method", method),
			slog.String("path", path),
		)

		// Store logger in context
		c.Set("logger", reqLogger)

		c.Next()

		// Post-Request Logging
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		reqLogger.Info("Request completed",
			slog.Int("status", statusCode),
			slog.Duration("duration", duration),
		)
	}
}

