package users

import (
	"github.com/gin-gonic/gin"
	usecase "github.com/tapiaw38/vamo/backend/internal/usecases/users"
	"net/http"
)

func NewGetHandler(uc usecase.GetUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("vamo.user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": "auth:identity-required", "message": "identity required"})
			return
		}
		output, appErr := uc.Execute(c.Request.Context(), userID.(string), c.GetHeader("Authorization"))
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": output})
	}
}
