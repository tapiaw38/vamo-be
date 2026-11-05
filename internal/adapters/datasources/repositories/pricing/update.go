package pricing

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

var ErrMissingCoordinates = errors.New("pricing:locality-coordinates-missing")

func (r *repository) Update(ctx context.Context, basePrice, pricePerKm, urbanBasePrice, urbanPricePerKm, currency, actor string) (domain.PricingSettings, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.PricingSettings{}, err
	}
	defer tx.Rollback(ctx)

	var oldBase, oldPerKm, oldUrbanBase, oldUrbanPerKm, oldCurrency string
	if err = tx.QueryRow(ctx, `SELECT base_price::text, price_per_km::text, urban_base_price::text, urban_price_per_km::text, currency FROM pricing_settings WHERE id=1 FOR UPDATE`).Scan(&oldBase, &oldPerKm, &oldUrbanBase, &oldUrbanPerKm, &oldCurrency); err != nil {
		return domain.PricingSettings{}, err
	}

	var item domain.PricingSettings
	err = tx.QueryRow(ctx, `UPDATE pricing_settings SET base_price=$1::numeric, price_per_km=$2::numeric, urban_base_price=$3::numeric, urban_price_per_km=$4::numeric, currency=$5, updated_at=NOW(), updated_by=$6 WHERE id=1 RETURNING base_price::text, price_per_km::text, urban_base_price::text, urban_price_per_km::text, currency, updated_at, COALESCE(updated_by, '')`, basePrice, pricePerKm, urbanBasePrice, urbanPricePerKm, currency, actor).Scan(&item.BasePrice, &item.PricePerKm, &item.UrbanBasePrice, &item.UrbanPricePerKm, &item.Currency, &item.UpdatedAt, &item.UpdatedBy)
	if err != nil {
		return domain.PricingSettings{}, err
	}
	metadataBytes, err := json.Marshal(map[string]string{
		"old_base_price": oldBase, "old_price_per_km": oldPerKm, "old_urban_base_price": oldUrbanBase, "old_urban_price_per_km": oldUrbanPerKm, "old_currency": oldCurrency,
		"base_price": item.BasePrice, "price_per_km": item.PricePerKm, "urban_base_price": item.UrbanBasePrice, "urban_price_per_km": item.UrbanPricePerKm, "currency": item.Currency,
	})
	if err != nil {
		return domain.PricingSettings{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(actor_user_id,action,entity_type,entity_id,metadata) VALUES($1,$2,$3,$4,$5::jsonb)`, actor, "admin.pricing.updated", "pricing_settings", "1", metadataBytes); err != nil {
		return domain.PricingSettings{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.PricingSettings{}, err
	}
	return item, nil
}
