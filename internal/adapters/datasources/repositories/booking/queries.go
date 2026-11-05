package booking

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"time"
)

const bookingColumns = `b.id, b.trip_id, b.passenger_id, b.seats, b.unit_price_snapshot::text, b.total_price::text, b.currency, b.status, b.requested_at, b.confirmed_at, b.cancelled_at, b.expires_at, b.created_at, b.updated_at`
const bookingReturnColumns = `id, trip_id, passenger_id, seats, unit_price_snapshot::text, total_price::text, currency, status, requested_at, confirmed_at, cancelled_at, expires_at, created_at, updated_at`

func scanBooking(row pgx.Row) (*domain.Booking, error) {
	var item domain.Booking
	err := row.Scan(&item.ID, &item.TripID, &item.PassengerID, &item.Seats, &item.UnitPriceSnapshot, &item.TotalPrice, &item.Currency, &item.Status, &item.RequestedAt, &item.ConfirmedAt, &item.CancelledAt, &item.ExpiresAt, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}

func scanBookingWithRating(row pgx.Row) (*domain.Booking, error) {
	var item domain.Booking
	err := row.Scan(&item.ID, &item.TripID, &item.PassengerID, &item.Seats, &item.UnitPriceSnapshot, &item.TotalPrice, &item.Currency, &item.Status, &item.RequestedAt, &item.ConfirmedAt, &item.CancelledAt, &item.ExpiresAt, &item.CreatedAt, &item.UpdatedAt, &item.Rated)
	return &item, err
}

func (r *repository) Create(ctx context.Context, passengerID, tripID string, seats int) (*domain.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var departure time.Time
	var driverUserID, unitPrice, currency, mode, status string
	var available int
	err = tx.QueryRow(ctx, `SELECT t.departure_at, dp.user_id, t.price_per_seat::text, t.currency, t.booking_mode, t.status, t.available_seats FROM trips t JOIN driver_profiles dp ON dp.id=t.driver_profile_id WHERE t.id=$1 FOR UPDATE`, tripID).Scan(&departure, &driverUserID, &unitPrice, &currency, &mode, &status, &available)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if driverUserID == passengerID {
		return nil, ErrSelfBooking
	}
	if departure.Before(time.Now()) || (status != "published" && status != "full") || seats <= 0 || available < seats {
		return nil, ErrTripUnavailable
	}
	finalStatus, expiresAt, confirmedAt := "confirmed", any(nil), any(nil)
	if mode == "manual" {
		finalStatus = "pending"
		expiresAt = time.Now().Add(24 * time.Hour)
	} else {
		confirmedAt = time.Now()
	}
	var item domain.Booking
	err = tx.QueryRow(ctx, `INSERT INTO bookings(trip_id, passenger_id, seats, unit_price_snapshot, total_price, currency, status, confirmed_at, expires_at) VALUES($1, $2, $3::int, $4::numeric, ($4::numeric * $3::numeric), $5, $6, $7, $8) RETURNING `+bookingReturnColumns, tripID, passengerID, seats, unitPrice, currency, finalStatus, confirmedAt, expiresAt).Scan(&item.ID, &item.TripID, &item.PassengerID, &item.Seats, &item.UnitPriceSnapshot, &item.TotalPrice, &item.Currency, &item.Status, &item.RequestedAt, &item.ConfirmedAt, &item.CancelledAt, &item.ExpiresAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, err
	}
	if _, err = tx.Exec(ctx, `UPDATE trips SET available_seats=available_seats-$2, status=CASE WHEN available_seats-$2=0 THEN 'full' ELSE 'published' END, updated_at=NOW() WHERE id=$1`, tripID, seats); err != nil {
		return nil, err
	}
	if err = insertEvent(ctx, tx, item.ID, nil, finalStatus, passengerID, "booking requested"); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &item, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func insertEvent(ctx context.Context, tx pgx.Tx, bookingID string, fromStatus *string, toStatus, actor, reason string) error {
	_, err := tx.Exec(ctx, `INSERT INTO booking_events(booking_id, from_status, to_status, actor_user_id, reason) VALUES($1,$2,$3,$4,$5)`, bookingID, fromStatus, toStatus, actor, reason)
	return err
}
