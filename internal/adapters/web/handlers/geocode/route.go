package geocode

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	routinguc "github.com/tapiaw38/vamo/backend/internal/usecases/routing"
)

func NewRoute(service *routinguc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		originLat, originLatErr := strconv.ParseFloat(c.Query("origin_lat"), 64)
		originLon, originLonErr := strconv.ParseFloat(c.Query("origin_lon"), 64)
		destinationLat, destinationLatErr := strconv.ParseFloat(c.Query("destination_lat"), 64)
		destinationLon, destinationLonErr := strconv.ParseFloat(c.Query("destination_lon"), 64)
		if originLatErr != nil || originLonErr != nil || destinationLatErr != nil || destinationLonErr != nil || originLat < -90 || originLat > 90 || destinationLat < -90 || destinationLat > 90 || originLon < -180 || originLon > 180 || destinationLon < -180 || destinationLon > 180 {
			c.JSON(http.StatusBadRequest, gin.H{"code": "routing:coordinates-invalid", "message": "valid origin and destination coordinates are required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": service.Route(c.Request.Context(), originLat, originLon, destinationLat, destinationLon)})
	}
}
