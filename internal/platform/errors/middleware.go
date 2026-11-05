package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}
		appErr := InternalServerError
		if typed, ok := c.Errors.Last().Err.(*ApplicationError); ok {
			appErr = typed
		}
		c.JSON(appErr.Status, appErr)
	}
}

func Abort(c *gin.Context, appErr *ApplicationError) {
	c.Error(appErr)
	c.AbortWithStatus(http.StatusInternalServerError)
}
