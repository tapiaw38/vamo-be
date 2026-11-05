package user

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) Search(ctx context.Context, query string) ([]domain.User, error) {
	rows, err := r.db.Query(ctx, `SELECT id, first_name, last_name, email, phone, residence_locality_id FROM users WHERE LOWER(email) LIKE '%' || LOWER($1) || '%' OR LOWER(first_name || ' ' || last_name) LIKE '%' || LOWER($1) || '%' ORDER BY first_name, last_name LIMIT 20`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.User, 0)
	for rows.Next() {
		var item domain.User
		if err := rows.Scan(&item.ID, &item.FirstName, &item.LastName, &item.Email, &item.Phone, &item.ResidenceLocalityID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
