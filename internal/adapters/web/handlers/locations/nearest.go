package locations

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	usecase "github.com/tapiaw38/vamo/backend/internal/usecases/locations"
)

func NewNearestHandler(uc usecase.NearestUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		lat, latErr := strconv.ParseFloat(c.Query("lat"), 64)
		lon, lonErr := strconv.ParseFloat(c.Query("lon"), 64)
		if latErr != nil || lonErr != nil || lat < -90 || lat > 90 || lon < -180 || lon > 180 {
			c.JSON(http.StatusBadRequest, gin.H{"code": "locations:coordinates-invalid", "message": "valid lat and lon are required"})
			return
		}
		output, appErr := uc.Execute(c.Request.Context(), lat, lon)
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": output})
	}
}
