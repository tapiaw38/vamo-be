package vehicle

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) CreateForCompany(ctx context.Context, item domain.Vehicle) (*domain.Vehicle, error) {
	return scanVehicle(r.db.QueryRow(ctx, `INSERT INTO vehicles(owner_type,driver_profile_id,company_id,make,model,year,license_plate,authorized_seats,vehicle_type,photo_url,verification_status,active) VALUES('company',NULL,$1,$2,$3,$4,$5,$6,$7,$8,'pending',TRUE) RETURNING `+vehicleColumns, item.CompanyID, item.Make, item.Model, item.Year, item.LicensePlate, item.AuthorizedSeats, item.VehicleType, item.PhotoURL))
}
func (r *repository) ListByCompany(ctx context.Context, companyID string) ([]domain.Vehicle, error) {
	rows, err := r.db.Query(ctx, `SELECT `+vehicleColumns+` FROM vehicles WHERE company_id=$1 ORDER BY created_at DESC`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Vehicle{}
	for rows.Next() {
		item, e := scanVehicle(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}
