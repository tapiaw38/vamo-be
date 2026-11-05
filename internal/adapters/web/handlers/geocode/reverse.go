package geocode

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	geocodeuc "github.com/tapiaw38/vamo/backend/internal/usecases/geocode"
)

func NewReverse(service *geocodeuc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		lat, latErr := strconv.ParseFloat(c.Query("lat"), 64)
		lon, lonErr := strconv.ParseFloat(c.Query("lon"), 64)
		if latErr != nil || lonErr != nil || lat < -90 || lat > 90 || lon < -180 || lon > 180 {
			c.JSON(http.StatusBadRequest, gin.H{"code": "geocode:coordinates-invalid", "message": "valid lat and lon are required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": service.Reverse(c.Request.Context(), lat, lon)})
	}
}
