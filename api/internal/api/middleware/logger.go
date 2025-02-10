package middleware

import (
	"time"

	"github.com/ahmetkoprulu/rtrp/common/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		status := c.Writer.Status()

		if query != "" {
			path = path + "?" + query
		}

		utils.Logger.Info("Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
		)
	}
}
