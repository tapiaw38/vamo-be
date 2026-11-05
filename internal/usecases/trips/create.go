package trips

import (
	"context"
	companyrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/company"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/trip"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/vehicle"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
	"strings"
)

type CreateUsecase interface {
	Execute(context.Context, string, Input) (TripData, *apperrors.ApplicationError)
}
type createUsecase struct {
	drivers   driver.Repository
	vehicles  vehicle.Repository
	trips     trip.Repository
	companies companyrepo.Repository
}

func NewCreateUsecase(drivers driver.Repository, vehicles vehicle.Repository, trips trip.Repository, companies ...companyrepo.Repository) CreateUsecase {
	var companyRepo companyrepo.Repository
	if len(companies) > 0 {
		companyRepo = companies[0]
	}
	return &createUsecase{drivers: drivers, vehicles: vehicles, trips: trips, companies: companyRepo}
}
func (u *createUsecase) Execute(ctx context.Context, userID string, input Input) (TripData, *apperrors.ApplicationError) {
	publisherType := input.PublisherType
	if publisherType == "" {
		publisherType = "driver"
	}
	if publisherType == "company" {
		if u.companies == nil || input.CompanyID == "" || input.DriverProfileID == "" {
			return TripData{}, apperrors.New("companies:invalid-publisher", http.StatusBadRequest, "company publisher data required")
		}
		ok, e := u.companies.CanDispatch(ctx, input.CompanyID, userID)
		if e != nil {
			return TripData{}, apperrors.New("companies:operation-failed", 500, "could not validate company membership")
		}
		if !ok {
			return TripData{}, apperrors.New("companies:forbidden", http.StatusForbidden, "company dispatch permission required")
		}
		belongs, e := u.companies.DriverBelongs(ctx, input.CompanyID, input.DriverProfileID)
		if e != nil || !belongs {
			return TripData{}, apperrors.New("companies:driver-forbidden", http.StatusBadRequest, "driver does not belong to company")
		}
		driverID := input.DriverProfileID
		vehicleItem, err := u.vehicles.GetByIDAndCompany(ctx, input.VehicleID, input.CompanyID)
		if err != nil || vehicleItem == nil {
			return TripData{}, apperrors.New("vehicles:not-found", http.StatusNotFound, "company vehicle not found")
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
		item, e := u.trips.Create(ctx, domain.Trip{PublisherType: publisherType, CompanyID: input.CompanyID, DriverProfileID: driverID, VehicleID: vehicleItem.ID, OriginLocalityID: input.OriginLocalityID, DestinationLocalityID: input.DestinationLocalityID, DepartureAt: departure, ArrivalAt: arrival, Timezone: timezone, PricePerSeat: input.PricePerSeat, Currency: currency, TotalSeats: input.TotalSeats, BookingMode: input.BookingMode, Notes: input.Notes})
		if e != nil {
			return TripData{}, apperrors.New("trips:create-failed", 500, "could not create trip")
		}
		return tripData(item), nil
	}
	profile, err := u.drivers.GetByUserID(ctx, userID)
	if err != nil {
		return TripData{}, apperrors.New("drivers:get-failed", 500, "could not get driver profile")
	}
	if profile == nil {
		return TripData{}, apperrors.New("drivers:profile-required", http.StatusConflict, "driver profile required")
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
	item, err := u.trips.Create(ctx, domain.Trip{PublisherType: publisherType, DriverProfileID: profile.ID, VehicleID: vehicleItem.ID, OriginLocalityID: input.OriginLocalityID, DestinationLocalityID: input.DestinationLocalityID, DepartureAt: departure, ArrivalAt: arrival, Timezone: timezone, PricePerSeat: input.PricePerSeat, Currency: currency, TotalSeats: input.TotalSeats, BookingMode: input.BookingMode, Notes: input.Notes})
	if err != nil {
		return TripData{}, apperrors.New("trips:create-failed", 500, "could not create trip")
	}
	return tripData(item), nil
}
