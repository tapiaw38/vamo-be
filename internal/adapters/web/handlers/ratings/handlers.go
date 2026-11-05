package ratings

import (
	"net/http"

	"github.com/gin-gonic/gin"
	usecase "github.com/tapiaw38/vamo/backend/internal/usecases/ratings"
)

func NewCreate(u usecase.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input usecase.Input
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "ratings:invalid-body", "message": "invalid rating body"})
			return
		}
		output, appErr := u.Create(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"), input)
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": output})
	}
}

func NewDriverStats(u usecase.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		output, appErr := u.GetDriver(c.Request.Context(), c.Param("id"))
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": output})
	}
}
