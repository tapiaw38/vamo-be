package pricing

import (
	"time"

	"github.com/tapiaw38/vamo/backend/internal/domain"
)

type SettingsData struct {
	BasePrice       string    `json:"base_price"`
	PricePerKm      string    `json:"price_per_km"`
	UrbanBasePrice  string    `json:"urban_base_price"`
	UrbanPricePerKm string    `json:"urban_price_per_km"`
	Currency        string    `json:"currency"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedBy       string    `json:"updated_by,omitempty"`
}

type QuoteData struct {
	DistanceKM   float64 `json:"distance_km"`
	BasePrice    string  `json:"base_price"`
	PricePerKm   string  `json:"price_per_km"`
	PricePerSeat string  `json:"price_per_seat"`
	Currency     string  `json:"currency"`
	FareKind     string  `json:"fare_kind"`
}

func SettingsOutput(item domain.PricingSettings) SettingsData {
	return SettingsData{BasePrice: item.BasePrice, PricePerKm: item.PricePerKm, UrbanBasePrice: item.UrbanBasePrice, UrbanPricePerKm: item.UrbanPricePerKm, Currency: item.Currency, UpdatedAt: item.UpdatedAt, UpdatedBy: item.UpdatedBy}
}

func QuoteOutput(item domain.PriceQuote) QuoteData {
	return QuoteData{DistanceKM: item.DistanceKM, BasePrice: item.BasePrice, PricePerKm: item.PricePerKm, PricePerSeat: item.PricePerSeat, Currency: item.Currency, FareKind: item.FareKind}
}
