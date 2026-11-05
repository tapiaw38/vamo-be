package drivers

import (
	"context"
	"net/http"

	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
)

type AvailabilityInput struct {
	IsAvailable bool `json:"is_available"`
}

type AvailabilityUsecase interface {
	Execute(context.Context, string, AvailabilityInput) (DriverProfileData, *apperrors.ApplicationError)
}

type availabilityRepository interface {
	SetAvailability(context.Context, string, bool) (*domain.DriverProfile, error)
}
type availabilityUsecase struct{ repository availabilityRepository }

func NewAvailabilityUsecase(repository availabilityRepository) AvailabilityUsecase {
	return &availabilityUsecase{repository: repository}
}

func (u *availabilityUsecase) Execute(ctx context.Context, userID string, input AvailabilityInput) (DriverProfileData, *apperrors.ApplicationError) {
	item, err := u.repository.SetAvailability(ctx, userID, input.IsAvailable)
	if err != nil {
		return DriverProfileData{}, apperrors.New("drivers:availability-failed", http.StatusConflict, "driver profile required")
	}
	return profileData(item), nil
}
