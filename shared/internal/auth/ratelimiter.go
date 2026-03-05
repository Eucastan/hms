package auth

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	limiter = rate.NewLimiter(rate.Every(time.Minute), 5)
	mu      sync.Mutex
)

func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		mu.Lock()
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})

			mu.Unlock()
			return
		}
		mu.Unlock()
		c.Next()
	}
}
