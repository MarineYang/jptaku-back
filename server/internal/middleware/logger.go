package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if query != "" {
			path = path + "?" + query
		}

		log.Printf("[%s] %d | %13v | %15s | %s",
			c.Request.Method,
			c.Writer.Status(),
			latency,
			c.ClientIP(),
			path,
		)
	}
}
