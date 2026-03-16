package auth

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  sync.Mutex
}

var (
	ipLimiter = &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
	}
	// Global fallback (very high limit)
	//globalLimiter = rate.NewLimiter(rate.Every(time.Minute), 100)
)

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		// 10 requests per minute per IP
		limiter = rate.NewLimiter(rate.Every(time.Minute), 10)
		i.ips[ip] = limiter
	}

	return limiter
}

func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for health checks / internal paths
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Get client IP
		clientIP := c.ClientIP()

		limiter := ipLimiter.GetLimiter(clientIP)

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": "60 seconds",
			})
			return
		}

		c.Next()
	}
}
