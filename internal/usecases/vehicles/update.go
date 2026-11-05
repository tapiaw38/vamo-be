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

type UpdateUsecase interface {
	Execute(context.Context, string, string, Input) (VehicleData, *apperrors.ApplicationError)
}
type updateUsecase struct {
	drivers  driver.Repository
	vehicles vehicle.Repository
	images   storage.ImageStorage
}

func NewUpdateUsecase(drivers driver.Repository, vehicles vehicle.Repository, images storage.ImageStorage) UpdateUsecase {
	return &updateUsecase{drivers: drivers, vehicles: vehicles, images: images}
}
func (u *updateUsecase) Execute(ctx context.Context, userID, id string, input Input) (VehicleData, *apperrors.ApplicationError) {
	profile, err := u.drivers.GetByUserID(ctx, userID)
	if err != nil || profile == nil {
		return VehicleData{}, apperrors.New("vehicles:not-found", http.StatusNotFound, "vehicle not found")
	}
	if input.Year < 1900 || input.Year > time.Now().Year()+1 || input.AuthorizedSeats <= 0 {
		return VehicleData{}, apperrors.New("vehicles:invalid-input", http.StatusBadRequest, "invalid vehicle data")
	}
	photoURL, appErr := storePhoto(ctx, u.images, profile.ID, input.PhotoURL)
	if appErr != nil {
		return VehicleData{}, appErr
	}
	item, err := u.vehicles.Update(ctx, domain.Vehicle{ID: id, DriverProfileID: profile.ID, Make: strings.TrimSpace(input.Make), Model: strings.TrimSpace(input.Model), Year: input.Year, LicensePlate: strings.TrimSpace(input.LicensePlate), AuthorizedSeats: input.AuthorizedSeats, VehicleType: strings.TrimSpace(input.VehicleType), PhotoURL: photoURL})
	if err != nil {
		return VehicleData{}, apperrors.New("vehicles:update-failed", http.StatusConflict, "could not update vehicle; license plate may already be active")
	}
	return vehicleData(item, u.images), nil
}
