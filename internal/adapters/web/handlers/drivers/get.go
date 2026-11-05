package drivers

import (
	"github.com/gin-gonic/gin"
	usecase "github.com/tapiaw38/vamo/backend/internal/usecases/drivers"
	"net/http"
)

func NewGetHandler(uc usecase.GetUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetString("vamo.user_id")
		output, appErr := uc.Execute(c.Request.Context(), id)
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": output})
	}
}
