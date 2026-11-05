package admin

import (
	"github.com/gin-gonic/gin"
	uc "github.com/tapiaw38/vamo/backend/internal/usecases/admin"
	"net/http"
)

func NewDashboard(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, e := u.Dashboard(c.Request.Context())
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(200, gin.H{"data": o})
	}
}
func NewList(u uc.Usecases, kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, e := u.List(c.Request.Context(), kind)
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(200, gin.H{"data": o})
	}
}
func NewVerify(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in struct {
			Status string `json:"status"`
		}
		if c.ShouldBindJSON(&in) != nil {
			c.JSON(400, gin.H{"code": "common:invalid-body"})
			return
		}
		if e := u.Verify(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("kind"), c.Param("id"), in.Status); e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.Status(http.StatusNoContent)
	}
}
