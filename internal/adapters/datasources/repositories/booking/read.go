package booking

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"time"
)

func (r *repository) ListMine(ctx context.Context, passengerID, status string) ([]domain.Booking, error) {
	if err := r.completeDeparted(ctx); err != nil {
		return nil, err
	}
	var ids []string
	rows, err := r.db.Query(ctx, `SELECT id FROM bookings WHERE passenger_id=$1 AND status='pending' AND expires_at IS NOT NULL AND expires_at <= NOW()`, passengerID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	for _, id := range ids {
		_, _ = r.expireIfNeeded(ctx, id, passengerID)
	}
	query := `SELECT ` + bookingColumns + `, EXISTS(SELECT 1 FROM trip_ratings tr WHERE tr.booking_id=b.id) FROM bookings b WHERE b.passenger_id=$1`
	args := []any{passengerID}
	if status != "" {
		query += ` AND b.status=$2`
		args = append(args, status)
	}
	query += ` ORDER BY b.requested_at DESC`
	rows, err = r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.Booking, 0)
	for rows.Next() {
		item, scanErr := scanBookingWithRating(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		if err := r.attachTrip(ctx, item); err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *repository) GetVisible(ctx context.Context, id, actorID string) (*domain.Booking, error) {
	if err := r.completeDeparted(ctx); err != nil {
		return nil, err
	}
	var passengerID, driverUserID string
	err := r.db.QueryRow(ctx, `SELECT b.passenger_id, dp.user_id FROM bookings b JOIN trips t ON t.id=b.trip_id JOIN driver_profiles dp ON dp.id=t.driver_profile_id WHERE b.id=$1`, id).Scan(&passengerID, &driverUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if actorID != passengerID && actorID != driverUserID {
		return nil, ErrForbidden
	}
	_, _ = r.expireIfNeeded(ctx, id, actorID)
	item, err := scanBookingWithRating(r.db.QueryRow(ctx, `SELECT `+bookingColumns+`, EXISTS(SELECT 1 FROM trip_ratings tr WHERE tr.booking_id=b.id) FROM bookings b WHERE b.id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := r.attachTrip(ctx, item); err != nil {
		return nil, err
	}
	return item, err
}

func (r *repository) attachTrip(ctx context.Context, item *domain.Booking) error {
	var trip domain.Trip
	err := r.db.QueryRow(ctx, `SELECT t.id, t.publisher_type, t.driver_profile_id, t.vehicle_id, t.origin_locality_id, t.destination_locality_id, COALESCE(t.company_id::text,''), t.departure_at, t.arrival_at, t.timezone, t.price_per_seat::text, t.currency, t.total_seats, t.available_seats, t.booking_mode, t.status, t.notes, t.created_at, t.updated_at, CONCAT_WS(' ',u.first_name,u.last_name), v.make, v.model, v.vehicle_type, ol.name, dl.name FROM trips t JOIN driver_profiles dp ON dp.id=t.driver_profile_id JOIN users u ON u.id=dp.user_id JOIN vehicles v ON v.id=t.vehicle_id JOIN localities ol ON ol.id=t.origin_locality_id JOIN localities dl ON dl.id=t.destination_locality_id WHERE t.id=$1`, item.TripID).Scan(&trip.ID, &trip.PublisherType, &trip.DriverProfileID, &trip.VehicleID, &trip.OriginLocalityID, &trip.DestinationLocalityID, &trip.CompanyID, &trip.DepartureAt, &trip.ArrivalAt, &trip.Timezone, &trip.PricePerSeat, &trip.Currency, &trip.TotalSeats, &trip.AvailableSeats, &trip.BookingMode, &trip.Status, &trip.Notes, &trip.CreatedAt, &trip.UpdatedAt, &trip.DriverName, &trip.VehicleMake, &trip.VehicleModel, &trip.VehicleType, &trip.OriginName, &trip.DestinationName)
	if err != nil {
		return err
	}
	item.Trip = &trip
	return nil
}

func (r *repository) expireIfNeeded(ctx context.Context, id, actorID string) (*domain.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var tripID string
	err = tx.QueryRow(ctx, `SELECT trip_id FROM bookings WHERE id=$1`, id).Scan(&tripID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var lockedTripID string
	if err = tx.QueryRow(ctx, `SELECT id FROM trips WHERE id=$1 FOR UPDATE`, tripID).Scan(&lockedTripID); err != nil {
		return nil, err
	}
	var item domain.Booking
	err = tx.QueryRow(ctx, `SELECT `+bookingReturnColumns+` FROM bookings WHERE id=$1 FOR UPDATE`, id).Scan(&item.ID, &item.TripID, &item.PassengerID, &item.Seats, &item.UnitPriceSnapshot, &item.TotalPrice, &item.Currency, &item.Status, &item.RequestedAt, &item.ConfirmedAt, &item.CancelledAt, &item.ExpiresAt, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if item.Status != "pending" || item.ExpiresAt == nil || item.ExpiresAt.After(now()) {
		return &item, nil
	}
	if _, err = tx.Exec(ctx, `UPDATE bookings SET status='expired', updated_at=NOW() WHERE id=$1`, id); err != nil {
		return nil, err
	}
	if err = releaseSeats(ctx, tx, item.TripID, item.Seats); err != nil {
		return nil, err
	}
	if err = insertEvent(ctx, tx, id, &item.Status, "expired", actorID, "pending booking expired"); err != nil {
		return nil, err
	}
	_ = lockedTripID
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	item.Status = "expired"
	return &item, nil
}

func now() time.Time { return time.Now() }
