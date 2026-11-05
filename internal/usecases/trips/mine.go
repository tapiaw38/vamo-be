package trips

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/trip"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
)

type MineUsecase interface {
	Execute(context.Context, string, string) ([]TripData, *apperrors.ApplicationError)
}
type mineUsecase struct {
	drivers driver.Repository
	trips   trip.Repository
}

func NewMineUsecase(drivers driver.Repository, trips trip.Repository) MineUsecase {
	return &mineUsecase{drivers: drivers, trips: trips}
}
func (u *mineUsecase) Execute(ctx context.Context, userID, status string) ([]TripData, *apperrors.ApplicationError) {
	items, err := u.trips.ListMineForUser(ctx, userID, status)
	if err != nil {
		return nil, apperrors.New("trips:list-failed", 500, "could not list trips")
	}
	output := make([]TripData, 0, len(items))
	for i := range items {
		output = append(output, tripData(&items[i]))
	}
	return output, nil
}
