package bookingpayment

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

var (
	ErrNotFound  = errors.New("payment not found")
	ErrForbidden = errors.New("payment forbidden")
	ErrConflict  = errors.New("payment conflict")
)

type Repository interface {
	Prepare(context.Context, string, string) (*domain.BookingPayment, bool, error)
	SetProvider(context.Context, string, string, string, string) error
	UpdateStatus(context.Context, string, string, string) error
	Get(context.Context, string, string) (*domain.BookingPayment, error)
}
type repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) Repository { return &repository{db: db} }

const cols = `id,booking_id,provider,external_reference,COALESCE(provider_payment_id,''),amount::text,currency,status,idempotency_key,created_at,updated_at`

func scan(row pgx.Row) (*domain.BookingPayment, error) {
	var p domain.BookingPayment
	e := row.Scan(&p.ID, &p.BookingID, &p.Provider, &p.ExternalReference, &p.ProviderPaymentID, &p.Amount, &p.Currency, &p.Status, &p.IdempotencyKey, &p.CreatedAt, &p.UpdatedAt)
	return &p, e
}
func (r *repository) Prepare(ctx context.Context, bookingID, userID string) (*domain.BookingPayment, bool, error) {
	var passenger, status, currency, amount string
	var tripID string
	if e := r.db.QueryRow(ctx, `SELECT b.passenger_id,b.status,b.trip_id,b.total_price::text,b.currency FROM bookings b WHERE b.id=$1`, bookingID).Scan(&passenger, &status, &tripID, &amount, &currency); e != nil {
		if e == pgx.ErrNoRows {
			return nil, false, ErrNotFound
		}
		return nil, false, e
	}
	if passenger != userID {
		return nil, false, ErrForbidden
	}
	if status != "confirmed" {
		return nil, false, ErrConflict
	}
	ref := "vamo:booking:" + bookingID
	existing, e := scan(r.db.QueryRow(ctx, `SELECT `+cols+` FROM booking_payments WHERE booking_id=$1`, bookingID))
	if e == nil {
		return existing, false, nil
	}
	if e != pgx.ErrNoRows {
		return nil, false, e
	}
	p, e := scan(r.db.QueryRow(ctx, `INSERT INTO booking_payments(booking_id,provider,external_reference,amount,currency,status,idempotency_key) VALUES($1,'payments-dev',$2,$3::numeric,$4,'initiated',$2) RETURNING `+cols, bookingID, ref, amount, currency))
	if e != nil {
		if isUnique(e) {
			existing, _ = scan(r.db.QueryRow(ctx, `SELECT `+cols+` FROM booking_payments WHERE booking_id=$1`, bookingID))
			if existing != nil {
				return existing, false, nil
			}
		}
		return nil, false, e
	}
	return p, true, nil
}
func isUnique(e error) bool { var p *pgconn.PgError; return errors.As(e, &p) && p.Code == "23505" }
func (r *repository) SetProvider(ctx context.Context, id, providerID, status string, initPoint string) error {
	_, e := r.db.Exec(ctx, `UPDATE booking_payments SET provider_payment_id=COALESCE(NULLIF($2,''),provider_payment_id),status=$3,updated_at=NOW() WHERE id=$1`, id, providerID, status)
	return e
}
func (r *repository) UpdateStatus(ctx context.Context, bookingID, providerID, status string) error {
	tag, e := r.db.Exec(ctx, `UPDATE booking_payments SET provider_payment_id=COALESCE(NULLIF($2,''),provider_payment_id),status=$3,updated_at=NOW() WHERE booking_id=$1`, bookingID, providerID, status)
	if e != nil {
		return e
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}
func (r *repository) Get(ctx context.Context, bookingID, userID string) (*domain.BookingPayment, error) {
	p, e := scan(r.db.QueryRow(ctx, `SELECT `+cols+` FROM booking_payments p JOIN bookings b ON b.id=p.booking_id JOIN trips t ON t.id=b.trip_id JOIN driver_profiles d ON d.id=t.driver_profile_id JOIN users u ON u.id=d.user_id WHERE p.booking_id=$1 AND (b.passenger_id=$2 OR u.id=$2)`, bookingID, userID))
	if e == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return p, e
}
