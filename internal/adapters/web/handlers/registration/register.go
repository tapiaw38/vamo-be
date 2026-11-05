package registration

import (
	"net/http"

	"github.com/gin-gonic/gin"
	usecase "github.com/tapiaw38/vamo/backend/internal/usecases/registration"
)

func NewRegisterHandler(uc usecase.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input usecase.Input
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "common:invalid-body", "message": "invalid request body"})
			return
		}
		body, appErr := uc.Execute(c.Request.Context(), input)
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", body)
	}
}
