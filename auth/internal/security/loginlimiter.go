package security

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type LoginLimiter struct {
	ips    map[string]*rate.Limiter
	emails map[string]*rate.Limiter
	mu     sync.Mutex
}

var loginLimiter = &LoginLimiter{
	ips:    make(map[string]*rate.Limiter),
	emails: make(map[string]*rate.Limiter),
}

func LoginRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path != "/login" || c.Request.Method != "POST" {
			c.Next()
			return
		}

		ip := c.ClientIP()
		var email string

		// Try to read email from body (only for login)
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // rewind
		var loginReq struct{ Email string }
		json.Unmarshal(body, &loginReq)
		email = loginReq.Email

		loginLimiter.mu.Lock()
		defer loginLimiter.mu.Unlock()

		// Per-IP limit (5/min)
		ipLim, ok := loginLimiter.ips[ip]
		if !ok {
			ipLim = rate.NewLimiter(rate.Every(time.Minute), 5)
			loginLimiter.ips[ip] = ipLim
		}
		if !ipLim.Allow() {
			c.AbortWithStatusJSON(429, gin.H{"error": "too many login attempts from this IP"})
			return
		}

		// Per-email limit (3/min) — prevents targeting specific accounts
		if email != "" {
			emailLim, ok := loginLimiter.emails[email]
			if !ok {
				emailLim = rate.NewLimiter(rate.Every(time.Minute), 3)
				loginLimiter.emails[email] = emailLim
			}
			if !emailLim.Allow() {
				c.AbortWithStatusJSON(429, gin.H{"error": "too many login attempts for this email"})
				return
			}
		}

		c.Next()
	}
}
