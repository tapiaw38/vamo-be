package trips

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/trip"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
	"time"
)

type CancelUsecase interface {
	Execute(context.Context, string, string) (TripData, *apperrors.ApplicationError)
}
type cancelUsecase struct {
	drivers driver.Repository
	trips   trip.Repository
}

func NewCancelUsecase(drivers driver.Repository, trips trip.Repository) CancelUsecase {
	return &cancelUsecase{drivers: drivers, trips: trips}
}
func (u *cancelUsecase) Execute(ctx context.Context, userID, id string) (TripData, *apperrors.ApplicationError) {
	profile, err := u.drivers.GetByUserID(ctx, userID)
	if err != nil || profile == nil {
		return TripData{}, apperrors.New("trips:not-found", http.StatusNotFound, "trip not found")
	}
	item, err := u.trips.Cancel(ctx, id, profile.ID)
	if err != nil || item == nil || !item.DepartureAt.After(time.Now()) {
		return TripData{}, apperrors.New("trips:cannot-cancel", http.StatusConflict, "trip already departed or is not owned by user")
	}
	return tripData(item), nil
}
