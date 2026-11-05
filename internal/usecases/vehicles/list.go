package vehicles

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/vehicle"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"github.com/tapiaw38/vamo/backend/internal/platform/storage"
)

type ListUsecase interface {
	Execute(context.Context, string) ([]VehicleData, *apperrors.ApplicationError)
}
type listUsecase struct {
	drivers  driver.Repository
	vehicles vehicle.Repository
	images   storage.ImageStorage
}

func NewListUsecase(drivers driver.Repository, vehicles vehicle.Repository, images ...storage.ImageStorage) ListUsecase {
	var imageStorage storage.ImageStorage
	if len(images) > 0 {
		imageStorage = images[0]
	}
	return &listUsecase{drivers: drivers, vehicles: vehicles, images: imageStorage}
}
func (u *listUsecase) Execute(ctx context.Context, userID string) ([]VehicleData, *apperrors.ApplicationError) {
	profile, err := u.drivers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, apperrors.New("drivers:get-failed", 500, "could not get driver profile")
	}
	if profile == nil {
		return []VehicleData{}, nil
	}
	items, err := u.vehicles.ListByDriver(ctx, profile.ID)
	if err != nil {
		return nil, apperrors.New("vehicles:list-failed", 500, "could not list vehicles")
	}
	output := make([]VehicleData, 0, len(items))
	for i := range items {
		output = append(output, vehicleData(&items[i], u.images))
	}
	return output, nil
}
