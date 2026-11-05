package booking

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) ListReceived(ctx context.Context, driverUserID, status string) ([]domain.Booking, error) {
	if err := r.completeDeparted(ctx); err != nil {
		return nil, err
	}
	expiredRows, err := r.db.Query(ctx, `SELECT b.id FROM bookings b JOIN trips t ON t.id=b.trip_id JOIN driver_profiles dp ON dp.id=t.driver_profile_id WHERE dp.user_id=$1 AND b.status='pending' AND b.expires_at IS NOT NULL AND b.expires_at <= NOW()`, driverUserID)
	if err != nil {
		return nil, err
	}
	var expiredIDs []string
	for expiredRows.Next() {
		var id string
		if scanErr := expiredRows.Scan(&id); scanErr != nil {
			expiredRows.Close()
			return nil, scanErr
		}
		expiredIDs = append(expiredIDs, id)
	}
	expiredRows.Close()
	for _, id := range expiredIDs {
		_, _ = r.expireIfNeeded(ctx, id, driverUserID)
	}
	query := `SELECT ` + bookingColumns + ` FROM bookings b JOIN trips t ON t.id=b.trip_id JOIN driver_profiles dp ON dp.id=t.driver_profile_id WHERE dp.user_id=$1`
	args := []any{driverUserID}
	if status != "" {
		query += ` AND b.status=$2`
		args = append(args, status)
	}
	query += ` ORDER BY b.requested_at DESC`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.Booking, 0)
	for rows.Next() {
		item, scanErr := scanBooking(rows)
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
