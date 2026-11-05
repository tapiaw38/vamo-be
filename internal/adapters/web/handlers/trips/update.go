package trips

import (
	"github.com/gin-gonic/gin"
	usecase "github.com/tapiaw38/vamo/backend/internal/usecases/trips"
	"net/http"
)

func NewUpdateHandler(uc usecase.UpdateUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input usecase.Input
		if !bind(c, &input) {
			return
		}
		output, appErr := uc.Execute(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"), input)
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": output})
	}
}
