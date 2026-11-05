package pricing

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	pricinguc "github.com/tapiaw38/vamo/backend/internal/usecases/pricing"
)

func NewQuote(u pricinguc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		coordinates := make([]float64, 0, 4)
		provided := 0
		for _, name := range []string{"origin_lat", "origin_lon", "destination_lat", "destination_lon"} {
			value := c.Query(name)
			if value == "" {
				continue
			}
			parsed, err := strconv.ParseFloat(value, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": "pricing:invalid-coordinates", "message": "pin coordinates must be numeric"})
				return
			}
			coordinates = append(coordinates, parsed)
			provided++
		}
		if provided != 0 && provided != 4 {
			c.JSON(http.StatusBadRequest, gin.H{"code": "pricing:invalid-coordinates", "message": "all pin coordinates are required"})
			return
		}
		quote, appErr := u.Quote(c.Request.Context(), c.Query("origin_locality_id"), c.Query("destination_locality_id"), coordinates...)
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": pricinguc.QuoteOutput(quote)})
	}
}

func NewGet(u pricinguc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		settings, appErr := u.Get(c.Request.Context())
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": pricinguc.SettingsOutput(settings)})
	}
}

func NewUpdate(u pricinguc.Usecases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			BasePrice       string `json:"base_price"`
			PricePerKm      string `json:"price_per_km"`
			UrbanBasePrice  string `json:"urban_base_price"`
			UrbanPricePerKm string `json:"urban_price_per_km"`
			Currency        string `json:"currency"`
		}
		if c.ShouldBindJSON(&input) != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "common:invalid-body", "message": "invalid pricing settings"})
			return
		}
		settings, appErr := u.Update(c.Request.Context(), input.BasePrice, input.PricePerKm, input.UrbanBasePrice, input.UrbanPricePerKm, input.Currency, c.GetString("vamo.user_id"))
		if appErr != nil {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": pricinguc.SettingsOutput(settings)})
	}
}
