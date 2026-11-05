package vehicles

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/vehicle"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"github.com/tapiaw38/vamo/backend/internal/platform/storage"
	"net/http"
)

type GetUsecase interface {
	Execute(context.Context, string, string) (VehicleData, *apperrors.ApplicationError)
}
type getUsecase struct {
	drivers  driver.Repository
	vehicles vehicle.Repository
	images   storage.ImageStorage
}

func NewGetUsecase(drivers driver.Repository, vehicles vehicle.Repository, images ...storage.ImageStorage) GetUsecase {
	var imageStorage storage.ImageStorage
	if len(images) > 0 {
		imageStorage = images[0]
	}
	return &getUsecase{drivers: drivers, vehicles: vehicles, images: imageStorage}
}
func (u *getUsecase) Execute(ctx context.Context, userID, id string) (VehicleData, *apperrors.ApplicationError) {
	profile, err := u.drivers.GetByUserID(ctx, userID)
	if err != nil || profile == nil {
		return VehicleData{}, apperrors.New("vehicles:not-found", http.StatusNotFound, "vehicle not found")
	}
	item, err := u.vehicles.GetByIDAndDriver(ctx, id, profile.ID)
	if err != nil {
		return VehicleData{}, apperrors.New("vehicles:not-found", http.StatusNotFound, "vehicle not found")
	}
	return vehicleData(item, u.images), nil
}
