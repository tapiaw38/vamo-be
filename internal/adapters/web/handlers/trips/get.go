package trips

import (
	"github.com/gin-gonic/gin"
	usecase "github.com/tapiaw38/vamo/backend/internal/usecases/trips"
	"net/http"
)

func NewGetHandler(uc usecase.GetUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		output, appErr := uc.Execute(c.Request.Context(), c.Param("id"))
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": output})
	}
}
