package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasRole(c.Request.Context(), "superadmin") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "admin:forbidden", "message": "admin role required"})
			return
		}
		c.Next()
	}
}
