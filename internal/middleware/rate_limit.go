package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sunilsinghx/order-processing/internal/metrics"
)

type RateLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rdb:    rdb,
		limit:  limit,
		window: window,
	}
}

func (rl *RateLimiter) LimitByIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ip := c.ClientIP()
		key := "rate_limit:" + ip

		count, err := rl.rdb.Incr(ctx, key).Result()
		if err != nil {
			fmt.Println("Redis is down")
			c.Next()
			return
		}

		if count == 1 {
			rl.rdb.Expire(ctx, key, rl.window)
		}

		if int(count) > rl.limit {
			metrics.RateLimitTotal.Inc()

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
