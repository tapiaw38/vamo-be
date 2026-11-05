package riderequests

import (
	"github.com/gin-gonic/gin"
	uc "github.com/tapiaw38/vamo/backend/internal/usecases/riderequests"
	"net/http"
)

func NewCreate(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in uc.Input
		if c.ShouldBindJSON(&in) != nil {
			c.JSON(400, gin.H{"code": "common:invalid-body"})
			return
		}
		o, e := u.Create(c.Request.Context(), c.GetString("vamo.user_id"), in, c.GetHeader("Authorization"))
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": o})
	}
}
func NewList(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, e := u.List(c.Request.Context(), c.GetString("vamo.user_id"), uc.SearchInput{OriginLocalityID: c.Query("origin_locality_id"), DestinationLocalityID: c.Query("destination_locality_id")})
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(200, gin.H{"data": o})
	}
}
func NewActive(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, e := u.Active(c.Request.Context(), c.GetString("vamo.user_id"))
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": o})
	}
}
func NewRecent(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, e := u.Recent(c.Request.Context(), c.GetString("vamo.user_id"))
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": o})
	}
}
func NewCancel(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		o, e := u.Cancel(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"))
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(200, gin.H{"data": o})
	}
}
func NewAccept(u uc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input uc.AcceptInput
		if c.Request.ContentLength != 0 && c.ShouldBindJSON(&input) != nil {
			c.JSON(400, gin.H{"code": "common:invalid-body"})
			return
		}
		o, e := u.Accept(c.Request.Context(), c.GetString("vamo.user_id"), c.Param("id"), input.VehicleID)
		if e != nil {
			c.JSON(e.Status, e)
			return
		}
		c.JSON(200, gin.H{"data": o})
	}
}
