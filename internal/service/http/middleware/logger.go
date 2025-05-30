package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/draw-hub/internal/modules/logs"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		c.Next()

		statusCode := c.Writer.Status()
		duration := time.Since(start)

		logs.Logger.Info().Str("method", method).
			Str("path", path).
			Str("client_ip", clientIP).
			Int("status", statusCode).
			Dur("duration", duration).
			Msg("request log")
	}
}
