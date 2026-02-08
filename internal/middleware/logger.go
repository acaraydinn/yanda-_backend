package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs all incoming requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Get status code
		statusCode := c.Writer.Status()

		// Log request
		log.Printf("[%s] %d | %v | %s | %s %s",
			c.ClientIP(),
			statusCode,
			latency,
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.URL.RawQuery,
		)
	}
}
