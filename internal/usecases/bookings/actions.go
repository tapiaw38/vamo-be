package bookings

import (
	"context"
	bookingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/booking"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
)

type CancelUsecase interface {
	Execute(context.Context, string, string) (BookingData, *apperrors.ApplicationError)
}
type ConfirmUsecase interface {
	Execute(context.Context, string, string) (BookingData, *apperrors.ApplicationError)
}
type RejectUsecase interface {
	Execute(context.Context, string, string, string) (BookingData, *apperrors.ApplicationError)
}
type actionUsecase struct{ repository bookingrepo.Repository }

func NewCancelUsecase(repository bookingrepo.Repository) CancelUsecase {
	return &cancelUsecase{repository: repository}
}

type cancelUsecase struct{ repository bookingrepo.Repository }

func (u *cancelUsecase) Execute(ctx context.Context, actorID, id string) (BookingData, *apperrors.ApplicationError) {
	item, err := u.repository.Cancel(ctx, id, actorID)
	if err != nil {
		return BookingData{}, mapError(err)
	}
	return bookingData(item), nil
}
func NewConfirmUsecase(repository bookingrepo.Repository) ConfirmUsecase {
	return &confirmUsecase{repository: repository}
}

type confirmUsecase struct{ repository bookingrepo.Repository }

func (u *confirmUsecase) Execute(ctx context.Context, actorID, id string) (BookingData, *apperrors.ApplicationError) {
	item, err := u.repository.Confirm(ctx, id, actorID)
	if err != nil {
		return BookingData{}, mapError(err)
	}
	return bookingData(item), nil
}
func NewRejectUsecase(repository bookingrepo.Repository) RejectUsecase {
	return &rejectUsecase{repository: repository}
}

type rejectUsecase struct{ repository bookingrepo.Repository }

func (u *rejectUsecase) Execute(ctx context.Context, actorID, id, reason string) (BookingData, *apperrors.ApplicationError) {
	item, err := u.repository.Reject(ctx, id, actorID, reason)
	if err != nil {
		return BookingData{}, mapError(err)
	}
	return bookingData(item), nil
}
