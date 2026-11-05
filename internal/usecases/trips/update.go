package trips

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/trip"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/vehicle"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
	"strings"
)

type UpdateUsecase interface {
	Execute(context.Context, string, string, Input) (TripData, *apperrors.ApplicationError)
}
type updateUsecase struct {
	drivers  driver.Repository
	vehicles vehicle.Repository
	trips    trip.Repository
}

func NewUpdateUsecase(drivers driver.Repository, vehicles vehicle.Repository, trips trip.Repository) UpdateUsecase {
	return &updateUsecase{drivers: drivers, vehicles: vehicles, trips: trips}
}
func (u *updateUsecase) Execute(ctx context.Context, userID, id string, input Input) (TripData, *apperrors.ApplicationError) {
	profile, err := u.drivers.GetByUserID(ctx, userID)
	if err != nil || profile == nil {
		return TripData{}, apperrors.New("trips:not-found", http.StatusNotFound, "trip not found")
	}
	vehicleItem, err := u.vehicles.GetByIDAndDriver(ctx, input.VehicleID, profile.ID)
	if err != nil || vehicleItem == nil {
		return TripData{}, apperrors.New("vehicles:not-found", http.StatusNotFound, "vehicle not found for driver")
	}
	departure, arrival, appErr := parseTripInput(input, vehicleItem.AuthorizedSeats, true)
	if appErr != nil {
		return TripData{}, appErr
	}
	timezone := input.Timezone
	if strings.TrimSpace(timezone) == "" {
		timezone = "America/Argentina/Buenos_Aires"
	}
	currency := input.Currency
	if currency == "" {
		currency = "ARS"
	}
	item, err := u.trips.Update(ctx, domain.Trip{ID: id, DriverProfileID: profile.ID, VehicleID: vehicleItem.ID, OriginLocalityID: input.OriginLocalityID, DestinationLocalityID: input.DestinationLocalityID, DepartureAt: departure, ArrivalAt: arrival, Timezone: timezone, PricePerSeat: input.PricePerSeat, Currency: currency, TotalSeats: input.TotalSeats, AvailableSeats: input.TotalSeats, BookingMode: input.BookingMode, Notes: input.Notes})
	if err != nil {
		return TripData{}, apperrors.New("trips:update-failed", http.StatusConflict, "trip not found or cannot be edited")
	}
	return tripData(item), nil
}
