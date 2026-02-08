package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yandas/backend/internal/config"
)

// RateLimiter middleware limits request rate per IP
func RateLimiter(cfg *config.Config, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redisClient == nil {
			// Skip rate limiting if Redis is not available
			c.Next()
			return
		}

		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)

		ctx := context.Background()

		// Get current count
		count, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		if count >= cfg.RateLimitRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		// Increment counter
		pipe := redisClient.Pipeline()
		pipe.Incr(ctx, key)
		if count == 0 {
			pipe.Expire(ctx, key, time.Duration(cfg.RateLimitWindow)*time.Second)
		}
		pipe.Exec(ctx)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RateLimitRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", cfg.RateLimitRequests-count-1))

		c.Next()
	}
}
