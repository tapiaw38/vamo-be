package geocode

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	geocodeuc "github.com/tapiaw38/vamo/backend/internal/usecases/geocode"
)

func NewSearch(service *geocodeuc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
		lon, _ := strconv.ParseFloat(c.Query("lon"), 64)
		c.JSON(http.StatusOK, gin.H{"data": service.Forward(c.Request.Context(), c.Query("q"), lat, lon)})
	}
}
