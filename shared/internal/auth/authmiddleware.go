package auth

import (
	"net/http"
	"strings"

	"github.com/Eucastan/hms/auth/internal/configs"
	"github.com/Eucastan/shared/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *configs.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		tokenStr := parts[1]

		claims, err := utils.ValidateToken(tokenStr, cfg)
		if err != nil || !claims.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("token", tokenStr)

		c.Set("user_id", claims.UserId)
		c.Set("role", claims.Role)

		c.Next()
	}
}
