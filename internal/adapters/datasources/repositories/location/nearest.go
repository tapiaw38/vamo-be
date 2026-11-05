package location

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) NearestLocality(ctx context.Context, lat, lon float64) (*domain.Locality, error) {
	var item domain.Locality
	err := r.db.QueryRow(ctx, `SELECT id, province_id, name, normalized_name, COALESCE(official_code, ''), lat, lon FROM localities WHERE lat IS NOT NULL AND lon IS NOT NULL ORDER BY (lat-$1)^2 + (lon-$2)^2 ASC LIMIT 1`, lat, lon).Scan(&item.ID, &item.ProvinceID, &item.Name, &item.NormalizedName, &item.OfficialCode, &item.Lat, &item.Lon)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
