package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/orbita-sh/orbita/internal/response"
)

func RateLimit(rdb *redis.Client, maxAttempts int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s:%s", c.FullPath(), ip)

		ctx := context.Background()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		if count > int64(maxAttempts) {
			response.TooManyRequests(c, "Too many attempts. Please try again later.")
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxAttempts))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(maxAttempts)-count))

		c.Next()
	}
}

func RateLimitAPI(rdb *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := c.ClientIP()
		if apiKey := c.GetString("api_key_id"); apiKey != "" {
			identifier = apiKey
		}

		key := fmt.Sprintf("ratelimit:api:%s", identifier)
		ctx := context.Background()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(maxRequests)-count))

		if count > int64(maxRequests) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "RATE_LIMIT_EXCEEDED", "message": "Rate limit exceeded"},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
