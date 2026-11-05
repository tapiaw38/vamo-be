package location

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) ListProvinces(ctx context.Context) ([]domain.Province, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, code FROM provinces ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.Province, 0)
	for rows.Next() {
		var item domain.Province
		if err := rows.Scan(&item.ID, &item.Name, &item.Code); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
