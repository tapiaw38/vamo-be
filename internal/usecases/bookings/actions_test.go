package bookings

import (
	"context"
	bookingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/booking"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"testing"
)

type fakeBookingRepo struct{ err error }

func (f fakeBookingRepo) Create(context.Context, string, string, int) (*domain.Booking, error) {
	return nil, f.err
}
func (f fakeBookingRepo) ListMine(context.Context, string, string) ([]domain.Booking, error) {
	return nil, f.err
}
func (f fakeBookingRepo) ListReceived(context.Context, string, string) ([]domain.Booking, error) {
	return nil, f.err
}
func (f fakeBookingRepo) GetVisible(context.Context, string, string) (*domain.Booking, error) {
	return nil, f.err
}
func (f fakeBookingRepo) Cancel(context.Context, string, string) (*domain.Booking, error) {
	return nil, f.err
}
func (f fakeBookingRepo) Confirm(context.Context, string, string) (*domain.Booking, error) {
	return nil, f.err
}
func (f fakeBookingRepo) Reject(context.Context, string, string, string) (*domain.Booking, error) {
	return nil, f.err
}

func TestInvalidTransitionIsRejected(t *testing.T) {
	_, appErr := NewConfirmUsecase(fakeBookingRepo{err: bookingrepo.ErrConflict}).Execute(context.Background(), "driver", "booking")
	if appErr == nil || appErr.Code != "bookings:invalid-transition" {
		t.Fatalf("expected invalid transition, got %v", appErr)
	}
}
func TestForeignBookingAccessIsRejected(t *testing.T) {
	_, appErr := NewCancelUsecase(fakeBookingRepo{err: bookingrepo.ErrForbidden}).Execute(context.Background(), "other-passenger", "booking")
	if appErr == nil || appErr.Code != "bookings:forbidden" {
		t.Fatalf("expected forbidden, got %v", appErr)
	}
}
