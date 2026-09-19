package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"live-polling-app/backend/utils"
)

const UserIDKey = "userID"

func AuthRequired(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		id, err := jwtManager.Parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Set(UserIDKey, id)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) string {
	value, _ := c.Get(UserIDKey)
	id, _ := value.(string)
	return id
}
