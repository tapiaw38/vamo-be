package vehicle

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

const vehicleColumns = `id, owner_type, COALESCE(driver_profile_id::text, ''), COALESCE(company_id::text, ''), make, model, year, license_plate, authorized_seats, vehicle_type, COALESCE(photo_url, ''), verification_status, active, created_at, updated_at`

func scanVehicle(row pgx.Row) (*domain.Vehicle, error) {
	var item domain.Vehicle
	err := row.Scan(&item.ID, &item.OwnerType, &item.DriverProfileID, &item.CompanyID, &item.Make, &item.Model, &item.Year, &item.LicensePlate, &item.AuthorizedSeats, &item.VehicleType, &item.PhotoURL, &item.VerificationStatus, &item.Active, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}

func (r *repository) Create(ctx context.Context, item domain.Vehicle) (*domain.Vehicle, error) {
	return scanVehicle(r.db.QueryRow(ctx, `INSERT INTO vehicles (owner_type, driver_profile_id, company_id, make, model, year, license_plate, authorized_seats, vehicle_type, photo_url, verification_status, active) VALUES ('driver', $1, NULL, $2, $3, $4, $5, $6, $7, $8, 'pending', TRUE) RETURNING `+vehicleColumns, item.DriverProfileID, item.Make, item.Model, item.Year, item.LicensePlate, item.AuthorizedSeats, item.VehicleType, item.PhotoURL))
}

func (r *repository) ListByDriver(ctx context.Context, driverID string) ([]domain.Vehicle, error) {
	rows, err := r.db.Query(ctx, `SELECT `+vehicleColumns+` FROM vehicles WHERE driver_profile_id = $1 ORDER BY created_at DESC`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.Vehicle, 0)
	for rows.Next() {
		item, scanErr := scanVehicle(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *repository) GetByIDAndDriver(ctx context.Context, id, driverID string) (*domain.Vehicle, error) {
	item, err := scanVehicle(r.db.QueryRow(ctx, `SELECT `+vehicleColumns+` FROM vehicles WHERE id = $1 AND driver_profile_id = $2`, id, driverID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}
	return item, err
}

func (r *repository) Update(ctx context.Context, item domain.Vehicle) (*domain.Vehicle, error) {
	return scanVehicle(r.db.QueryRow(ctx, `UPDATE vehicles SET make = $2, model = $3, year = $4, license_plate = $5, authorized_seats = $6, vehicle_type = $7, photo_url = $8, updated_at = NOW() WHERE id = $1 AND driver_profile_id = $9 RETURNING `+vehicleColumns, item.ID, item.Make, item.Model, item.Year, item.LicensePlate, item.AuthorizedSeats, item.VehicleType, item.PhotoURL, item.DriverProfileID))
}
