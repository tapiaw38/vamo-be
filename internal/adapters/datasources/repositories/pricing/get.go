package pricing

import (
	"context"

	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) Get(ctx context.Context) (domain.PricingSettings, error) {
	var item domain.PricingSettings
	err := r.db.QueryRow(ctx, `SELECT base_price::text, price_per_km::text, urban_base_price::text, urban_price_per_km::text, currency, updated_at, COALESCE(updated_by, '') FROM pricing_settings WHERE id=1`).Scan(&item.BasePrice, &item.PricePerKm, &item.UrbanBasePrice, &item.UrbanPricePerKm, &item.Currency, &item.UpdatedAt, &item.UpdatedBy)
	return item, err
}

func (r *repository) GetRouteCoordinates(ctx context.Context, originID, destinationID string) (float64, float64, float64, float64, error) {
	var originLat, originLon, destinationLat, destinationLon *float64
	err := r.db.QueryRow(ctx, `SELECT ol.lat, ol.lon, dl.lat, dl.lon FROM localities ol JOIN localities dl ON dl.id=$2 WHERE ol.id=$1`, originID, destinationID).Scan(&originLat, &originLon, &destinationLat, &destinationLon)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	if originLat == nil || originLon == nil || destinationLat == nil || destinationLon == nil {
		return 0, 0, 0, 0, ErrMissingCoordinates
	}
	return *originLat, *originLon, *destinationLat, *destinationLon, nil
}
