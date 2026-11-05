package users

import (
	"github.com/gin-gonic/gin"
	usecase "github.com/tapiaw38/vamo/backend/internal/usecases/users"
)

func NewSearchHandler(uc usecase.SearchUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		output, appErr := uc.Execute(c.Request.Context(), c.Query("q"))
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(200, gin.H{"data": output})
	}
}
