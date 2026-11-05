package user

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) Save(ctx context.Context, item domain.User) (*domain.User, error) {
	var saved domain.User
	err := r.db.QueryRow(ctx, `
		WITH saved AS (
			INSERT INTO users (id, first_name, last_name, email, phone, residence_locality_id, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
			ON CONFLICT (id) DO UPDATE SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, email = CASE WHEN EXCLUDED.email <> '' THEN EXCLUDED.email ELSE users.email END, phone = EXCLUDED.phone, residence_locality_id = EXCLUDED.residence_locality_id, updated_at = NOW()
			RETURNING id, first_name, last_name, email, phone, residence_locality_id
		)
		SELECT s.id, s.first_name, s.last_name, s.email, s.phone, s.residence_locality_id,
		       l.name, p.id, p.name
		FROM saved s
		LEFT JOIN localities l ON l.id = s.residence_locality_id
		LEFT JOIN provinces p ON p.id = l.province_id`, item.ID, item.FirstName, item.LastName, item.Email, item.Phone, item.ResidenceLocalityID).Scan(&saved.ID, &saved.FirstName, &saved.LastName, &saved.Email, &saved.Phone, &saved.ResidenceLocalityID, &saved.ResidenceLocalityName, &saved.ResidenceProvinceID, &saved.ResidenceProvinceName)
	if err != nil {
		return nil, err
	}
	return &saved, nil
}
