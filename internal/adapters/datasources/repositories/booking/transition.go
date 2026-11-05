package booking

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"time"
)

func (r *repository) Cancel(ctx context.Context, id, actorID string) (*domain.Booking, error) {
	return r.transition(ctx, id, actorID, "cancelled", false, "passenger cancelled booking")
}
func (r *repository) Confirm(ctx context.Context, id, actorID string) (*domain.Booking, error) {
	return r.transition(ctx, id, actorID, "confirmed", true, "publisher confirmed booking")
}
func (r *repository) Reject(ctx context.Context, id, actorID, reason string) (*domain.Booking, error) {
	return r.transition(ctx, id, actorID, "rejected", true, reason)
}

func (r *repository) transition(ctx context.Context, id, actorID, target string, publisherAction bool, reason string) (*domain.Booking, error) {
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
	var driverUserID string
	var tripStatus string
	var departure time.Time
	var mode string
	if err = tx.QueryRow(ctx, `SELECT dp.user_id, t.status, t.departure_at, t.booking_mode FROM trips t JOIN driver_profiles dp ON dp.id=t.driver_profile_id WHERE t.id=$1 FOR UPDATE`, tripID).Scan(&driverUserID, &tripStatus, &departure, &mode); err != nil {
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
	if publisherAction && driverUserID != actorID {
		return nil, ErrForbidden
	}
	if !publisherAction && item.PassengerID != actorID {
		return nil, ErrForbidden
	}
	if item.Status == "pending" && item.ExpiresAt != nil && item.ExpiresAt.Before(now()) {
		if _, err = tx.Exec(ctx, `UPDATE bookings SET status='expired', updated_at=NOW() WHERE id=$1`, id); err != nil {
			return nil, err
		}
		if err = releaseSeats(ctx, tx, item.TripID, item.Seats); err != nil {
			return nil, err
		}
		if err = insertEvent(ctx, tx, id, &item.Status, "expired", actorID, "pending booking expired"); err != nil {
			return nil, err
		}
		if err = tx.Commit(ctx); err != nil {
			return nil, err
		}
		return nil, ErrConflict
	}
	if target == "confirmed" {
		if item.Status != "pending" || mode != "manual" || departure.Before(time.Now()) || tripStatus == "cancelled" {
			return nil, ErrConflict
		}
	} else if target == "rejected" {
		if item.Status != "pending" || mode != "manual" {
			return nil, ErrConflict
		}
	} else if target == "cancelled" {
		if item.Status != "pending" && item.Status != "confirmed" {
			return nil, ErrConflict
		}
		if departure.Before(time.Now()) {
			return nil, ErrConflict
		}
	}
	from := item.Status
	var confirmedAt, cancelledAt any
	if target == "confirmed" {
		confirmedAt = time.Now()
	}
	if target == "cancelled" {
		cancelledAt = time.Now()
	}
	if _, err = tx.Exec(ctx, `UPDATE bookings SET status=$2, confirmed_at=COALESCE($3, confirmed_at), cancelled_at=COALESCE($4, cancelled_at), updated_at=NOW() WHERE id=$1`, id, target, confirmedAt, cancelledAt); err != nil {
		return nil, err
	}
	if target == "rejected" || target == "cancelled" {
		if err = releaseSeats(ctx, tx, item.TripID, item.Seats); err != nil {
			return nil, err
		}
	}
	if err = insertEvent(ctx, tx, id, &from, target, actorID, reason); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	item.FromStatus, item.Status, item.ConfirmedAt, item.CancelledAt = from, target, nil, nil
	return &item, nil
}

func releaseSeats(ctx context.Context, tx pgx.Tx, tripID string, seats int) error {
	_, err := tx.Exec(ctx, `UPDATE trips SET available_seats=LEAST(total_seats, available_seats+$2), status=CASE WHEN status='full' THEN 'published' ELSE status END, updated_at=NOW() WHERE id=$1`, tripID, seats)
	return err
}
