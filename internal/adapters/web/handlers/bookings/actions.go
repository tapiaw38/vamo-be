package bookings

import (
	"github.com/gin-gonic/gin"
	usecase "github.com/tapiaw38/vamo/backend/internal/usecases/bookings"
	"net/http"
)

func NewCancelHandler(uc usecase.CancelUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		output, appErr := uc.Execute(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"))
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": output})
	}
}
func NewConfirmHandler(uc usecase.ConfirmUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		output, appErr := uc.Execute(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"))
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": output})
	}
}
func NewRejectHandler(uc usecase.RejectUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Reason string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&input); err != nil && c.Request.ContentLength > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": "common:invalid-body", "message": "invalid request body"})
			return
		}
		output, appErr := uc.Execute(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"), input.Reason)
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": output})
	}
}
