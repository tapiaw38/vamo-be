package vehicle

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) GetByIDAndCompany(ctx context.Context, id, companyID string) (*domain.Vehicle, error) {
	item, err := scanVehicle(r.db.QueryRow(ctx, `SELECT `+vehicleColumns+` FROM vehicles WHERE id=$1 AND company_id=$2`, id, companyID))
	if err == pgx.ErrNoRows {
		return nil, pgx.ErrNoRows
	}
	return item, err
}
