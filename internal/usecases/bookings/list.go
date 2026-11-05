package bookings

import (
	"context"
	bookingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/booking"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
)

type ListUsecase interface {
	Execute(context.Context, string, string) ([]BookingData, *apperrors.ApplicationError)
}
type listUsecase struct{ repository bookingrepo.Repository }

func NewListUsecase(repository bookingrepo.Repository) ListUsecase {
	return &listUsecase{repository: repository}
}

type ReceivedUsecase interface {
	Execute(context.Context, string, string) ([]BookingData, *apperrors.ApplicationError)
}
type receivedUsecase struct{ repository bookingrepo.Repository }

func NewReceivedUsecase(repository bookingrepo.Repository) ReceivedUsecase {
	return &receivedUsecase{repository: repository}
}
func (u *receivedUsecase) Execute(ctx context.Context, driverUserID, status string) ([]BookingData, *apperrors.ApplicationError) {
	items, err := u.repository.ListReceived(ctx, driverUserID, status)
	if err != nil {
		return nil, mapError(err)
	}
	output := make([]BookingData, 0, len(items))
	for i := range items {
		output = append(output, bookingData(&items[i]))
	}
	return output, nil
}
func (u *listUsecase) Execute(ctx context.Context, passengerID, status string) ([]BookingData, *apperrors.ApplicationError) {
	items, err := u.repository.ListMine(ctx, passengerID, status)
	if err != nil {
		return nil, mapError(err)
	}
	output := make([]BookingData, 0, len(items))
	for i := range items {
		output = append(output, bookingData(&items[i]))
	}
	return output, nil
}
