package trips

import (
	"github.com/gin-gonic/gin"
	usecase "github.com/tapiaw38/vamo/backend/internal/usecases/trips"
	"net/http"
	"strconv"
)

func NewSearchHandler(uc usecase.SearchUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		size, _ := strconv.Atoi(c.Query("page_size"))
		passengers, _ := strconv.Atoi(c.Query("passengers"))
		output, appErr := uc.Execute(c.Request.Context(), usecase.SearchInput{OriginLocalityID: c.Query("origin_locality_id"), DestinationLocalityID: c.Query("destination_locality_id"), Date: c.Query("date"), VehicleType: c.Query("vehicle_type"), Passengers: passengers, Page: page, PageSize: size})
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": output, "meta": gin.H{"page": page, "page_size": size}})
	}
}
