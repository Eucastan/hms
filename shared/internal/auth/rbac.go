package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequiredRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.MustGet("role").(string)
		if userRole != role {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			c.Abort()
			return
		}

		c.Next()
	}
}
