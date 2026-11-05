package companies

import (
	"github.com/gin-gonic/gin"
	uc "github.com/tapiaw38/vamo/backend/internal/usecases/companies"
	"net/http"
)

func NewCreate(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in uc.CompanyInput
		if c.ShouldBindJSON(&in) != nil {
			c.JSON(400, gin.H{"code": "common:invalid-body"})
			return
		}
		o, e := u.Create(c.Request.Context(), c.GetString("vamo.user_id"), in)
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(201, gin.H{"data": o})
	}
}
func NewList(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, e := u.List(c.Request.Context(), c.GetString("vamo.user_id"))
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(200, gin.H{"data": o})
	}
}
func NewGet(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, e := u.Get(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"))
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(200, gin.H{"data": o})
	}
}
func NewUpdate(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in uc.CompanyInput
		if c.ShouldBindJSON(&in) != nil {
			c.JSON(400, gin.H{"code": "common:invalid-body"})
			return
		}
		o, e := u.Update(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"), in)
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(200, gin.H{"data": o})
	}
}
func NewMembers(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, e := u.Members(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"))
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(200, gin.H{"data": o})
	}
}
func NewAddMember(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in uc.MemberInput
		if c.ShouldBindJSON(&in) != nil {
			c.JSON(400, gin.H{"code": "common:invalid-body"})
			return
		}
		if e := u.AddMember(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"), in); e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.Status(http.StatusCreated)
	}
}
func NewRemoveMember(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		if e := u.RemoveMember(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"), c.Param("memberID")); e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.Status(http.StatusNoContent)
	}
}
func NewVehicles(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, e := u.Vehicles(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"))
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(200, gin.H{"data": o})
	}
}
func NewAddVehicle(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in uc.VehicleInput
		if c.ShouldBindJSON(&in) != nil {
			c.JSON(400, gin.H{"code": "common:invalid-body"})
			return
		}
		o, e := u.AddVehicle(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"), in)
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(201, gin.H{"data": o})
	}
}
