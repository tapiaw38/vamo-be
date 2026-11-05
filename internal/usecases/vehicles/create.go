package vehicles

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/vehicle"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"github.com/tapiaw38/vamo/backend/internal/platform/storage"
	"net/http"
	"strings"
	"time"
)

type CreateUsecase interface {
	Execute(context.Context, string, Input) (VehicleData, *apperrors.ApplicationError)
}
type createUsecase struct {
	drivers  driver.Repository
	vehicles vehicle.Repository
	images   storage.ImageStorage
}

func NewCreateUsecase(drivers driver.Repository, vehicles vehicle.Repository, images storage.ImageStorage) CreateUsecase {
	return &createUsecase{drivers: drivers, vehicles: vehicles, images: images}
}
func (u *createUsecase) Execute(ctx context.Context, userID string, input Input) (VehicleData, *apperrors.ApplicationError) {
	profile, err := u.drivers.GetByUserID(ctx, userID)
	if err != nil {
		return VehicleData{}, apperrors.New("drivers:get-failed", 500, "could not get driver profile")
	}
	if profile == nil {
		return VehicleData{}, apperrors.New("drivers:profile-required", http.StatusConflict, "driver profile required")
	}
	year := time.Now().Year()
	if input.Year < 1900 || input.Year > year+1 || input.AuthorizedSeats <= 0 || strings.TrimSpace(input.Make) == "" || strings.TrimSpace(input.Model) == "" || strings.TrimSpace(input.LicensePlate) == "" || strings.TrimSpace(input.VehicleType) == "" {
		return VehicleData{}, apperrors.New("vehicles:invalid-input", http.StatusBadRequest, "invalid vehicle data")
	}
	photoURL, appErr := storePhoto(ctx, u.images, profile.ID, input.PhotoURL)
	if appErr != nil {
		return VehicleData{}, appErr
	}
	item, err := u.vehicles.Create(ctx, domain.Vehicle{DriverProfileID: profile.ID, Make: strings.TrimSpace(input.Make), Model: strings.TrimSpace(input.Model), Year: input.Year, LicensePlate: strings.TrimSpace(input.LicensePlate), AuthorizedSeats: input.AuthorizedSeats, VehicleType: strings.TrimSpace(input.VehicleType), PhotoURL: photoURL})
	if err != nil {
		return VehicleData{}, apperrors.New("vehicles:create-failed", http.StatusConflict, "could not create vehicle; license plate may already be active")
	}
	return vehicleData(item, u.images), nil
}
