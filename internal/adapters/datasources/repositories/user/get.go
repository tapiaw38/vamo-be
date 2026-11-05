package user

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) Get(ctx context.Context, id string) (*domain.User, error) {
	var item domain.User
	err := r.db.QueryRow(ctx, `
		SELECT u.id, u.first_name, u.last_name, u.email, u.phone, u.residence_locality_id,
		       l.name, p.id, p.name
		FROM users u
		LEFT JOIN localities l ON l.id = u.residence_locality_id
		LEFT JOIN provinces p ON p.id = l.province_id
		WHERE u.id = $1`, id).Scan(&item.ID, &item.FirstName, &item.LastName, &item.Email, &item.Phone, &item.ResidenceLocalityID, &item.ResidenceLocalityName, &item.ResidenceProvinceID, &item.ResidenceProvinceName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
