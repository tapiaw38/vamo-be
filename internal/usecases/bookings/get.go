package bookings

import (
	"context"
	bookingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/booking"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
)

type GetUsecase interface {
	Execute(context.Context, string, string) (BookingData, *apperrors.ApplicationError)
}
type getUsecase struct{ repository bookingrepo.Repository }

func NewGetUsecase(repository bookingrepo.Repository) GetUsecase {
	return &getUsecase{repository: repository}
}
func (u *getUsecase) Execute(ctx context.Context, actorID, id string) (BookingData, *apperrors.ApplicationError) {
	item, err := u.repository.GetVisible(ctx, id, actorID)
	if err != nil {
		return BookingData{}, mapError(err)
	}
	return bookingData(item), nil
}
