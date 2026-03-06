package middleware

import (
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS(config cors.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		corsHandler := cors.New(config)
		corsHandler(ctx)
		if ctx.IsAborted() {
			slog.Debug("CORS preflight request aborted")
			return
		}
	}
}

