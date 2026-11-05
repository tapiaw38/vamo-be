package drivers

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
)

type GetUsecase interface {
	Execute(context.Context, string) (DriverProfileData, *apperrors.ApplicationError)
}
type getUsecase struct{ repository driver.Repository }

func NewGetUsecase(repository driver.Repository) GetUsecase {
	return &getUsecase{repository: repository}
}
func (u *getUsecase) Execute(ctx context.Context, userID string) (DriverProfileData, *apperrors.ApplicationError) {
	item, err := u.repository.GetByUserID(ctx, userID)
	if err != nil {
		return DriverProfileData{}, apperrors.New("drivers:get-failed", 500, "could not get driver profile")
	}
	if item == nil {
		return DriverProfileData{}, apperrors.New("drivers:not-found", http.StatusNotFound, "driver profile not found")
	}
	return profileData(item), nil
}
