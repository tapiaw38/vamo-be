package booking

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

var (
	ErrNotFound        = errors.New("booking not found")
	ErrForbidden       = errors.New("booking access forbidden")
	ErrConflict        = errors.New("booking state conflict")
	ErrTripUnavailable = errors.New("trip unavailable")
	ErrSelfBooking     = errors.New("publisher cannot book own trip")
)

type Repository interface {
	Create(context.Context, string, string, int) (*domain.Booking, error)
	ListMine(context.Context, string, string) ([]domain.Booking, error)
	ListReceived(context.Context, string, string) ([]domain.Booking, error)
	GetVisible(context.Context, string, string) (*domain.Booking, error)
	Cancel(context.Context, string, string) (*domain.Booking, error)
	Confirm(context.Context, string, string) (*domain.Booking, error)
	Reject(context.Context, string, string, string) (*domain.Booking, error)
}

type repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) Repository { return &repository{db: db} }

func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }

func (r *repository) completeDeparted(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `UPDATE bookings SET status='completed', updated_at=NOW() WHERE status='confirmed' AND trip_id IN (SELECT id FROM trips WHERE departure_at <= NOW())`)
	return err
}
