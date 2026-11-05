package drivers

import (
	"github.com/gin-gonic/gin"
	usecase "github.com/tapiaw38/vamo/backend/internal/usecases/drivers"
	"net/http"
)

func NewSaveHandler(uc usecase.SaveUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input usecase.SaveInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "common:invalid-body", "message": "invalid request body"})
			return
		}
		output, appErr := uc.Execute(c.Request.Context(), c.GetString("vamo.user_id"), input)
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": output})
	}
}
