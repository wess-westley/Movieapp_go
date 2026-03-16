package middleware

import (
	"Magic/utilis"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authmiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := utilis.GetAccessToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no token provided"})
			c.Abort()
			return
		}
		claims, err := utilis.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err})
			c.Abort()
			return
		}
		c.Set("userId", claims.UserId)
		c.Set("role", claims.Role)
		c.Next()

	}
}
