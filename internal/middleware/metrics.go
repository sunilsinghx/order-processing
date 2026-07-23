package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunilsinghx/order-processing/internal/metrics"
)

func TrackMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		metrics.HttpRequestsTotal.WithLabelValues(path, method, status).Inc()
		metrics.HttpRequestDuration.WithLabelValues(path, method).Observe(duration)
	}
}
