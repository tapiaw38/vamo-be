package rating

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

var (
	ErrNotFound    = errors.New("booking not found")
	ErrForbidden   = errors.New("rating forbidden")
	ErrNotEligible = errors.New("booking not eligible for rating")
	ErrConflict    = errors.New("booking already rated")
)

type Stats struct {
	AverageScore *float64
	TotalRatings int
}

type Repository interface {
	Create(context.Context, string, string, int, string) (*domain.TripRating, error)
	GetByDriver(context.Context, string) (Stats, error)
}

type repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, bookingID, passengerID string, score int, comment string) (*domain.TripRating, error) {
	if _, err := r.db.Exec(ctx, `UPDATE bookings SET status='completed', updated_at=NOW() WHERE status='confirmed' AND trip_id IN (SELECT id FROM trips WHERE departure_at <= NOW())`); err != nil {
		return nil, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var storedPassenger, tripID, driverID, status string
	err = tx.QueryRow(ctx, `SELECT b.passenger_id, b.trip_id, t.driver_profile_id, b.status FROM bookings b JOIN trips t ON t.id=b.trip_id WHERE b.id=$1 FOR UPDATE`, bookingID).Scan(&storedPassenger, &tripID, &driverID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if storedPassenger != passengerID {
		return nil, ErrForbidden
	}
	if status != "completed" && status != "confirmed" {
		return nil, ErrNotEligible
	}
	if status != "completed" {
		return nil, ErrNotEligible
	}

	var item domain.TripRating
	err = tx.QueryRow(ctx, `INSERT INTO trip_ratings(trip_id, booking_id, passenger_id, driver_profile_id, score, comment) VALUES($1,$2,$3,$4,$5,$6) RETURNING id, trip_id, booking_id, passenger_id, driver_profile_id, score, COALESCE(comment,''), created_at`, tripID, bookingID, passengerID, driverID, score, nullableComment(comment)).Scan(&item.ID, &item.TripID, &item.BookingID, &item.PassengerID, &item.DriverProfileID, &item.Score, &item.Comment, &item.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &item, nil
}

func nullableComment(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *repository) GetByDriver(ctx context.Context, driverID string) (Stats, error) {
	var stats Stats
	err := r.db.QueryRow(ctx, `SELECT ROUND(AVG(score)::numeric, 1), COUNT(*) FROM trip_ratings WHERE driver_profile_id=$1`, driverID).Scan(&stats.AverageScore, &stats.TotalRatings)
	return stats, err
}
