package location

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) ListLocalities(ctx context.Context, provinceID string) ([]domain.Locality, error) {
	rows, err := r.db.Query(ctx, `SELECT id, province_id, name, normalized_name, COALESCE(official_code, ''), lat, lon FROM localities WHERE province_id = $1 ORDER BY name`, provinceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.Locality, 0)
	for rows.Next() {
		var item domain.Locality
		if err := rows.Scan(&item.ID, &item.ProvinceID, &item.Name, &item.NormalizedName, &item.OfficialCode, &item.Lat, &item.Lon); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
