package trip

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"strconv"
	"time"
)

const tripColumns = `t.id, t.publisher_type, t.driver_profile_id, t.vehicle_id, t.origin_locality_id, t.destination_locality_id, COALESCE(t.company_id::text, ''), t.departure_at, t.arrival_at, t.timezone, t.price_per_seat::text, t.currency, t.total_seats, t.available_seats, t.booking_mode, t.status, t.notes, t.created_at, t.updated_at, CONCAT_WS(' ', u.first_name, u.last_name), COALESCE(ratings.average_score, NULL), COALESCE(ratings.total_ratings, 0), v.make, v.model, v.vehicle_type, COALESCE(v.license_plate, ''), COALESCE(v.photo_url, ''), ol.name, dl.name`
const tripJoins = ` FROM trips t JOIN driver_profiles dp ON dp.id = t.driver_profile_id JOIN users u ON u.id = dp.user_id JOIN vehicles v ON v.id = t.vehicle_id JOIN localities ol ON ol.id = t.origin_locality_id JOIN localities dl ON dl.id = t.destination_locality_id LEFT JOIN (SELECT driver_profile_id, ROUND(AVG(score)::numeric, 1) AS average_score, COUNT(*) AS total_ratings FROM trip_ratings GROUP BY driver_profile_id) ratings ON ratings.driver_profile_id = t.driver_profile_id `

func scan(row pgx.Row) (*domain.Trip, error) {
	var item domain.Trip
	err := row.Scan(&item.ID, &item.PublisherType, &item.DriverProfileID, &item.VehicleID, &item.OriginLocalityID, &item.DestinationLocalityID, &item.CompanyID, &item.DepartureAt, &item.ArrivalAt, &item.Timezone, &item.PricePerSeat, &item.Currency, &item.TotalSeats, &item.AvailableSeats, &item.BookingMode, &item.Status, &item.Notes, &item.CreatedAt, &item.UpdatedAt, &item.DriverName, &item.AverageScore, &item.TotalRatings, &item.VehicleMake, &item.VehicleModel, &item.VehicleType, &item.VehicleLicensePlate, &item.VehiclePhotoURL, &item.OriginName, &item.DestinationName)
	return &item, err
}

func (r *repository) Create(ctx context.Context, item domain.Trip) (*domain.Trip, error) {
	var companyID any = nil
	if item.CompanyID != "" {
		companyID = item.CompanyID
	}
	var id string
	err := r.db.QueryRow(ctx, `INSERT INTO trips (publisher_type, driver_profile_id, company_id, vehicle_id, origin_locality_id, destination_locality_id, departure_at, arrival_at, timezone, price_per_seat, currency, total_seats, available_seats, booking_mode, status, notes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12, $13, 'published', $14) RETURNING id`, item.PublisherType, item.DriverProfileID, companyID, item.VehicleID, item.OriginLocalityID, item.DestinationLocalityID, item.DepartureAt, nullableTime(item.ArrivalAt), item.Timezone, item.PricePerSeat, item.Currency, item.TotalSeats, item.BookingMode, item.Notes).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.get(ctx, `WHERE t.id = $1`, id)
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value
}

func (r *repository) Search(ctx context.Context, filter domain.TripSearch) ([]domain.Trip, error) {
	query := `SELECT ` + tripColumns + tripJoins + `WHERE t.status IN ('published', 'full')`
	args := make([]any, 0, 5)
	n := 1
	if filter.OriginLocalityID != "" {
		query += ` AND t.origin_locality_id = $` + strconv.Itoa(n)
		args = append(args, filter.OriginLocalityID)
		n++
	}
	if filter.DestinationLocalityID != "" {
		query += ` AND t.destination_locality_id = $` + strconv.Itoa(n)
		args = append(args, filter.DestinationLocalityID)
		n++
	}
	if filter.Date != "" {
		query += ` AND t.departure_at::date = $` + strconv.Itoa(n)
		args = append(args, filter.Date)
		n++
	}
	if filter.Passengers > 0 {
		query += ` AND t.available_seats >= $` + strconv.Itoa(n)
		args = append(args, filter.Passengers)
		n++
	}
	if filter.VehicleType != "" {
		query += ` AND v.vehicle_type = $` + strconv.Itoa(n)
		args = append(args, filter.VehicleType)
		n++
	}
	query += ` ORDER BY t.departure_at ASC LIMIT $` + strconv.Itoa(n) + ` OFFSET $` + strconv.Itoa(n+1)
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.Trip, 0)
	for rows.Next() {
		item, scanErr := scan(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *repository) GetPublic(ctx context.Context, id string) (*domain.Trip, error) {
	return r.get(ctx, `WHERE t.id = $1 AND t.status <> 'cancelled'`, id)
}
func (r *repository) GetOwned(ctx context.Context, id, driverID string) (*domain.Trip, error) {
	return r.get(ctx, `WHERE t.id = $1 AND t.driver_profile_id = $2`, id, driverID)
}
func (r *repository) get(ctx context.Context, where string, args ...any) (*domain.Trip, error) {
	item, err := scan(r.db.QueryRow(ctx, `SELECT `+tripColumns+tripJoins+where, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}
	return item, err
}

func (r *repository) Update(ctx context.Context, item domain.Trip) (*domain.Trip, error) {
	var id string
	err := r.db.QueryRow(ctx, `UPDATE trips SET origin_locality_id=$2, destination_locality_id=$3, departure_at=$4, arrival_at=$5, timezone=$6, price_per_seat=$7, total_seats=$8, available_seats=$9, booking_mode=$10, notes=$11, updated_at=NOW() WHERE id=$1 AND driver_profile_id=$12 AND status NOT IN ('cancelled', 'departed', 'completed') RETURNING id`, item.ID, item.OriginLocalityID, item.DestinationLocalityID, item.DepartureAt, nullableTime(item.ArrivalAt), item.Timezone, item.PricePerSeat, item.TotalSeats, item.AvailableSeats, item.BookingMode, item.Notes, item.DriverProfileID).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.get(ctx, `WHERE t.id = $1 AND t.driver_profile_id = $2`, id, item.DriverProfileID)
}
func (r *repository) Cancel(ctx context.Context, id, driverID string) (*domain.Trip, error) {
	var returnedID string
	err := r.db.QueryRow(ctx, `UPDATE trips SET status='cancelled', updated_at=NOW() WHERE id=$1 AND driver_profile_id=$2 AND departure_at > NOW() AND status NOT IN ('cancelled','departed','completed') RETURNING id`, id, driverID).Scan(&returnedID)
	if err != nil {
		return nil, err
	}
	return r.get(ctx, `WHERE t.id = $1`, returnedID)
}
func (r *repository) ListMine(ctx context.Context, driverID, status string) ([]domain.Trip, error) {
	query := `SELECT ` + tripColumns + tripJoins + `WHERE t.driver_profile_id=$1`
	args := []any{driverID}
	if status != "" {
		query += ` AND t.status=$2`
		args = append(args, status)
	}
	query += ` ORDER BY t.departure_at ASC`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.Trip, 0)
	for rows.Next() {
		item, scanErr := scan(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}
func (r *repository) ListMineForUser(ctx context.Context, userID, status string) ([]domain.Trip, error) {
	query := `SELECT ` + tripColumns + tripJoins + `WHERE (dp.user_id=$1 OR EXISTS(SELECT 1 FROM company_members cm WHERE cm.company_id=t.company_id AND cm.user_id=$1 AND cm.status='active' AND cm.role IN ('owner','admin','dispatcher'))) `
	args := []any{userID}
	if status != "" {
		query += `AND t.status=$2 `
		args = append(args, status)
	}
	query += `ORDER BY t.departure_at ASC`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Trip{}
	for rows.Next() {
		item, e := scan(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}
