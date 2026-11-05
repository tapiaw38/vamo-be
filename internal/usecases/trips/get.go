package trips

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/trip"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"github.com/tapiaw38/vamo/backend/internal/platform/storage"
	"net/http"
)

type GetUsecase interface {
	Execute(context.Context, string) (TripData, *apperrors.ApplicationError)
}
type getUsecase struct {
	trips  trip.Repository
	images storage.ImageStorage
}

func NewGetUsecase(trips trip.Repository, images ...storage.ImageStorage) GetUsecase {
	var imageStorage storage.ImageStorage
	if len(images) > 0 {
		imageStorage = images[0]
	}
	return &getUsecase{trips: trips, images: imageStorage}
}
func (u *getUsecase) Execute(ctx context.Context, id string) (TripData, *apperrors.ApplicationError) {
	item, err := u.trips.GetPublic(ctx, id)
	if err != nil || item == nil {
		return TripData{}, apperrors.New("trips:not-found", http.StatusNotFound, "trip not found")
	}
	return withPresignedPhoto(publicTripData(item), u.images), nil
}
