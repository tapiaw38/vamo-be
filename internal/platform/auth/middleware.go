package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

const UserIDKey = "vamo.user_id"
const RolesKey = "vamo.roles"

func Middleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "auth:missing-token", "message": "valid bearer token required"})
			return
		}
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "auth:invalid-token", "message": "invalid or expired token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "auth:invalid-claims", "message": "invalid token claims"})
			return
		}
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "auth:missing-user-id", "message": "token user id is required"})
			return
		}
		c.Set(UserIDKey, userID)
		roles := claimRoles(claims["roles"])
		c.Set(RolesKey, roles)
		c.Request = c.Request.WithContext(contextWithRoles(contextWithUserID(c.Request.Context(), userID), roles))
		c.Next()
	}
}

func claimRoles(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	roles := make([]string, 0, len(items))
	for _, item := range items {
		role, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, ok := role["name"].(string)
		if ok && name != "" {
			roles = append(roles, name)
		}
	}
	return roles
}

func RequireConductor() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasRole(c.Request.Context(), "admin") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "driver:role-required", "message": "conductor role required"})
			return
		}
		c.Next()
	}
}
