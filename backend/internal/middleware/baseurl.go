package middleware

import "github.com/gin-gonic/gin"

func BaseURL(baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("baseURL", baseURL)
		c.Next()
	}
}
