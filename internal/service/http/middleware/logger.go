package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/draw-hub/internal/modules/logs"
)

func isDrawApi(path string) bool {

	return strings.Contains(path, "/task/slow") || strings.Contains(path, "/task/fast")
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		c.Next()

		statusCode := c.Writer.Status()
		duration := time.Since(start)

		// 根据状态码和响应时间使用不同的日志级别
		logEvent := logs.Logger.With().
			Str("method", method).
			Str("path", path).
			Str("client_ip", clientIP).
			Int("status", statusCode).
			Dur("duration", duration).
			Logger()

		switch {
		case statusCode >= 500:
			// 服务器错误使用error级别
			logEvent.Error().Msg("request failed with server error")
		case statusCode >= 400:
			// 客户端错误使用warn级别
			logEvent.Warn().Msg("request failed with client error")
		case duration > 3*time.Second && !isDrawApi(path):
			// 慢请求使用warn级别
			logEvent.Warn().Msg("slow request detected")
		default:
			// 其他情况使用debug级别（生产环境不会显示）
			logEvent.Debug().Msg("request processed")
		}
	}
}
