package middlewares

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next() // 處理請求

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 在生產環境中，這裡應該接入 zap 或 logrus 等日誌庫
		gin.DefaultWriter.Write([]byte(
			"[GIN] " + fmt.Sprint(statusCode) + " | " + 
			latency.String() + " | " + 
			c.ClientIP() + " | " + 
			c.Request.Method + " " + path + "?" + query + "\n",
		))
	}
}