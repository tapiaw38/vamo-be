package riderequest

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

var (
	ErrNotFound    = errors.New("ride request not found")
	ErrForbidden   = errors.New("ride request access forbidden")
	ErrUnavailable = errors.New("ride request unavailable")
	ErrConflict    = errors.New("ride request conflict")
)

const expirationGrace = 15 * time.Minute

type Repository interface {
	Create(context.Context, domain.RideRequest) (*domain.RideRequest, error)
	Get(context.Context, string) (*domain.RideRequest, error)
	GetActive(context.Context, string) (*domain.RideRequest, error)
	RecentDestinations(context.Context, string) ([]domain.RecentRideRequestDestination, error)
	ListOpen(context.Context, string, string) ([]domain.RideRequest, error)
	Cancel(context.Context, string, string) (*domain.RideRequest, error)
	Accept(context.Context, string, string, string, string) (domain.RideRequest, string, string, error)
}

type repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) Repository { return &repository{db: db} }

const columns = `id, passenger_id, origin_locality_id, destination_locality_id, requested_at, max_price_per_seat::text, seats, status, COALESCE(accepted_by_driver_profile_id::text,''), COALESCE(accepted_trip_id::text,''), origin_lat, origin_lon, destination_lat, destination_lon, origin_address, destination_address, notes, created_at, updated_at`
const selectColumns = `rr.id, rr.passenger_id, rr.origin_locality_id, rr.destination_locality_id, rr.requested_at, rr.max_price_per_seat::text, rr.seats, rr.status, COALESCE(rr.accepted_by_driver_profile_id::text,''), COALESCE(rr.accepted_trip_id::text,''), rr.origin_lat, rr.origin_lon, rr.destination_lat, rr.destination_lon, rr.origin_address, rr.destination_address, rr.notes, rr.created_at, rr.updated_at, ol.name, dl.name`
const activeColumns = selectColumns + `, COALESCE(CONCAT_WS(' ', u.first_name, u.last_name), ''), ratings.average_score, COALESCE(ratings.total_ratings, 0), COALESCE(v.make, ''), COALESCE(v.model, ''), COALESCE(v.vehicle_type, ''), COALESCE(v.license_plate, '')`

func scan(row pgx.Row) (*domain.RideRequest, error) {
	return scanBase(row)
}

func scanJoined(row pgx.Row) (*domain.RideRequest, error) {
	var item domain.RideRequest
	err := row.Scan(&item.ID, &item.PassengerID, &item.OriginLocalityID, &item.DestinationLocalityID, &item.RequestedAt, &item.MaxPricePerSeat, &item.Seats, &item.Status, &item.AcceptedByDriverProfileID, &item.AcceptedTripID, &item.OriginLat, &item.OriginLon, &item.DestinationLat, &item.DestinationLon, &item.OriginAddress, &item.DestinationAddress, &item.Notes, &item.CreatedAt, &item.UpdatedAt, &item.OriginName, &item.DestinationName)
	return &item, err
}

func scanBase(row pgx.Row) (*domain.RideRequest, error) {
	var item domain.RideRequest
	err := row.Scan(&item.ID, &item.PassengerID, &item.OriginLocalityID, &item.DestinationLocalityID, &item.RequestedAt, &item.MaxPricePerSeat, &item.Seats, &item.Status, &item.AcceptedByDriverProfileID, &item.AcceptedTripID, &item.OriginLat, &item.OriginLon, &item.DestinationLat, &item.DestinationLon, &item.OriginAddress, &item.DestinationAddress, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}

func (r *repository) Create(ctx context.Context, item domain.RideRequest) (*domain.RideRequest, error) {
	return scan(r.db.QueryRow(ctx, `INSERT INTO ride_requests(passenger_id,origin_locality_id,destination_locality_id,requested_at,max_price_per_seat,seats,origin_lat,origin_lon,destination_lat,destination_lon,origin_address,destination_address,notes) VALUES($1,$2,$3,$4,$5::numeric,$6,$7::numeric,$8::numeric,$9::numeric,$10::numeric,$11,$12,$13) RETURNING `+columns, item.PassengerID, item.OriginLocalityID, item.DestinationLocalityID, item.RequestedAt, item.MaxPricePerSeat, item.Seats, item.OriginLat, item.OriginLon, item.DestinationLat, item.DestinationLon, item.OriginAddress, item.DestinationAddress, item.Notes))
}

func (r *repository) Get(ctx context.Context, id string) (*domain.RideRequest, error) {
	item, err := scan(r.db.QueryRow(ctx, `SELECT `+columns+` FROM ride_requests WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return item, err
}

func (r *repository) refreshStatuses(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `UPDATE ride_requests SET status='expired', updated_at=NOW() WHERE status='open' AND requested_at <= NOW() - ($1 * INTERVAL '1 minute')`, int(expirationGrace/time.Minute))
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `UPDATE ride_requests SET status='completed', updated_at=NOW() WHERE status='accepted' AND accepted_trip_id IN (SELECT id FROM trips WHERE departure_at <= NOW())`)
	return err
}

func (r *repository) GetActive(ctx context.Context, passenger string) (*domain.RideRequest, error) {
	if err := r.refreshStatuses(ctx); err != nil {
		return nil, err
	}
	var item domain.RideRequest
	err := r.db.QueryRow(ctx, `SELECT `+activeColumns+` FROM ride_requests rr
		JOIN localities ol ON ol.id=rr.origin_locality_id
		JOIN localities dl ON dl.id=rr.destination_locality_id
		LEFT JOIN trips t ON t.id=rr.accepted_trip_id
		LEFT JOIN driver_profiles dp ON dp.id=t.driver_profile_id
		LEFT JOIN users u ON u.id=dp.user_id
		LEFT JOIN vehicles v ON v.id=t.vehicle_id
		LEFT JOIN (SELECT driver_profile_id, ROUND(AVG(score)::numeric, 1) AS average_score, COUNT(*) AS total_ratings FROM trip_ratings GROUP BY driver_profile_id) ratings ON ratings.driver_profile_id=dp.id
		WHERE rr.passenger_id=$1 AND rr.status IN ('open','accepted') ORDER BY rr.created_at DESC LIMIT 1`, passenger).Scan(
		&item.ID, &item.PassengerID, &item.OriginLocalityID, &item.DestinationLocalityID, &item.RequestedAt, &item.MaxPricePerSeat, &item.Seats, &item.Status, &item.AcceptedByDriverProfileID, &item.AcceptedTripID, &item.OriginLat, &item.OriginLon, &item.DestinationLat, &item.DestinationLon, &item.OriginAddress, &item.DestinationAddress, &item.Notes, &item.CreatedAt, &item.UpdatedAt, &item.OriginName, &item.DestinationName, &item.AcceptedDriverName, &item.AcceptedDriverAverageScore, &item.AcceptedDriverTotalRatings, &item.AcceptedVehicleMake, &item.AcceptedVehicleModel, &item.AcceptedVehicleType, &item.AcceptedVehicleLicensePlate)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &item, err
}

func (r *repository) RecentDestinations(ctx context.Context, passenger string) ([]domain.RecentRideRequestDestination, error) {
	rows, err := r.db.Query(ctx, `SELECT destination_locality_id, destination_name, destination_address, destination_lat, destination_lon FROM (SELECT DISTINCT ON (rr.destination_address) rr.destination_locality_id, l.name AS destination_name, rr.destination_address, rr.destination_lat, rr.destination_lon, rr.created_at FROM ride_requests rr JOIN localities l ON l.id=rr.destination_locality_id WHERE rr.passenger_id=$1 ORDER BY rr.destination_address, rr.created_at DESC) recent ORDER BY created_at DESC LIMIT 5`, passenger)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.RecentRideRequestDestination, 0, 5)
	for rows.Next() {
		var item domain.RecentRideRequestDestination
		if err := rows.Scan(&item.DestinationLocalityID, &item.DestinationName, &item.DestinationAddress, &item.DestinationLat, &item.DestinationLon); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *repository) ListOpen(ctx context.Context, origin, destination string) ([]domain.RideRequest, error) {
	if err := r.refreshStatuses(ctx); err != nil {
		return nil, err
	}
	query := `SELECT ` + selectColumns + ` FROM ride_requests rr JOIN localities ol ON ol.id=rr.origin_locality_id JOIN localities dl ON dl.id=rr.destination_locality_id WHERE rr.status='open' AND rr.origin_locality_id=$1`
	args := []any{origin}
	if destination != "" {
		query += ` AND rr.destination_locality_id=$2`
		args = append(args, destination)
	}
	query += ` ORDER BY rr.requested_at ASC`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.RideRequest{}
	for rows.Next() {
		item, e := scanJoined(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func (r *repository) Cancel(ctx context.Context, id, passenger string) (*domain.RideRequest, error) {
	var item domain.RideRequest
	if err := r.refreshStatuses(ctx); err != nil {
		return nil, err
	}
	err := r.db.QueryRow(ctx, `UPDATE ride_requests SET status='cancelled',updated_at=NOW() WHERE id=$1 AND passenger_id=$2 AND status='open' RETURNING `+columns, id, passenger).Scan(&item.ID, &item.PassengerID, &item.OriginLocalityID, &item.DestinationLocalityID, &item.RequestedAt, &item.MaxPricePerSeat, &item.Seats, &item.Status, &item.AcceptedByDriverProfileID, &item.AcceptedTripID, &item.OriginLat, &item.OriginLon, &item.DestinationLat, &item.DestinationLon, &item.OriginAddress, &item.DestinationAddress, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrForbidden
	}
	return &item, err
}

func (r *repository) Accept(ctx context.Context, id, driverProfileID, requestedVehicleID, pricePerSeat string) (domain.RideRequest, string, string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.RideRequest{}, "", "", err
	}
	defer tx.Rollback(ctx)
	var item domain.RideRequest
	err = tx.QueryRow(ctx, `SELECT `+columns+` FROM ride_requests WHERE id=$1 FOR UPDATE`, id).Scan(&item.ID, &item.PassengerID, &item.OriginLocalityID, &item.DestinationLocalityID, &item.RequestedAt, &item.MaxPricePerSeat, &item.Seats, &item.Status, &item.AcceptedByDriverProfileID, &item.AcceptedTripID, &item.OriginLat, &item.OriginLon, &item.DestinationLat, &item.DestinationLon, &item.OriginAddress, &item.DestinationAddress, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, "", "", ErrNotFound
	}
	if err != nil {
		return item, "", "", err
	}
	if item.Status != "open" {
		return item, "", "", ErrUnavailable
	}
	var driverUser string
	if err = tx.QueryRow(ctx, `SELECT user_id FROM driver_profiles WHERE id=$1 AND is_available=TRUE`, driverProfileID).Scan(&driverUser); err != nil {
		return item, "", "", ErrForbidden
	}
	if driverUser == item.PassengerID {
		return item, "", "", ErrConflict
	}
	vehicleID := requestedVehicleID
	var vehicleCapacity int
	if vehicleID != "" {
		err = tx.QueryRow(ctx, `SELECT id, authorized_seats FROM vehicles WHERE id=$1 AND driver_profile_id=$2 AND active=TRUE`, vehicleID, driverProfileID).Scan(&vehicleID, &vehicleCapacity)
	} else {
		err = tx.QueryRow(ctx, `SELECT id, authorized_seats FROM vehicles WHERE driver_profile_id=$1 AND active=TRUE ORDER BY created_at ASC LIMIT 1`, driverProfileID).Scan(&vehicleID, &vehicleCapacity)
	}
	if err != nil {
		return item, "", "", ErrUnavailable
	}
	if item.Seats > vehicleCapacity {
		return item, "", "", ErrUnavailable
	}
	var tripID string
	err = tx.QueryRow(ctx, `INSERT INTO trips(publisher_type,driver_profile_id,vehicle_id,origin_locality_id,destination_locality_id,departure_at,price_per_seat,currency,total_seats,available_seats,booking_mode,status,notes) VALUES('driver',$1::uuid,$2::uuid,$3::text,$4::text,$5::timestamptz,$6::numeric,'ARS',$7::int,($7-$8)::int,'automatic',CASE WHEN $7=$8 THEN 'full' ELSE 'published' END,$9::text) RETURNING id`, driverProfileID, vehicleID, item.OriginLocalityID, item.DestinationLocalityID, item.RequestedAt, pricePerSeat, vehicleCapacity, item.Seats, item.Notes).Scan(&tripID)
	if err != nil {
		return item, "", "", err
	}
	var bookingID string
	err = tx.QueryRow(ctx, `INSERT INTO bookings(trip_id,passenger_id,seats,unit_price_snapshot,total_price,currency,status,confirmed_at) VALUES($1::uuid,$2::text,$3::int,$4::numeric,($4::numeric*$3::numeric),'ARS','confirmed',NOW()) RETURNING id`, tripID, item.PassengerID, item.Seats, pricePerSeat).Scan(&bookingID)
	if err != nil {
		return item, "", "", err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO booking_events(booking_id,from_status,to_status,actor_user_id,reason) VALUES($1,NULL,'confirmed',$2,'ride request accepted')`, bookingID, driverUser); err != nil {
		return item, "", "", err
	}
	err = tx.QueryRow(ctx, `UPDATE ride_requests SET status='accepted',accepted_by_driver_profile_id=$2,accepted_trip_id=$3,updated_at=NOW() WHERE id=$1 RETURNING `+columns, id, driverProfileID, tripID).Scan(&item.ID, &item.PassengerID, &item.OriginLocalityID, &item.DestinationLocalityID, &item.RequestedAt, &item.MaxPricePerSeat, &item.Seats, &item.Status, &item.AcceptedByDriverProfileID, &item.AcceptedTripID, &item.OriginLat, &item.OriginLon, &item.DestinationLat, &item.DestinationLon, &item.OriginAddress, &item.DestinationAddress, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return item, "", "", err
	}
	if err = tx.Commit(ctx); err != nil {
		return item, "", "", err
	}
	return item, tripID, bookingID, nil
}

func now() time.Time { return time.Now() }
