package locations

import (
	"context"
	"net/http"

	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/location"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
)

type NearestUsecase interface {
	Execute(context.Context, float64, float64) (LocalityData, *apperrors.ApplicationError)
}
type nearestUsecase struct{ repository location.Repository }

func NewNearestUsecase(repository location.Repository) NearestUsecase {
	return &nearestUsecase{repository: repository}
}
func (u *nearestUsecase) Execute(ctx context.Context, lat, lon float64) (LocalityData, *apperrors.ApplicationError) {
	item, err := u.repository.NearestLocality(ctx, lat, lon)
	if err != nil {
		return LocalityData{}, apperrors.New("locations:nearest-failed", 500, "could not resolve nearest locality")
	}
	if item == nil {
		return LocalityData{}, apperrors.New("locations:nearest-not-found", http.StatusNotFound, "no locality has coordinates")
	}
	return localityData(*item), nil
}
