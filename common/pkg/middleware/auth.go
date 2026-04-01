package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"ticket-tix/common/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func AuthMiddleware(secretKey string, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwt.ParseToken(tokenString, secretKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
			})
			return
		}

		if claims.Type != jwt.AccessType {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
			})
			return
		}

		blacklistKey := fmt.Sprintf("blacklist:access:%s", tokenString)
		exists, err := redisClient.Exists(c.Request.Context(), blacklistKey).Result()
		if err == nil && exists > 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}
