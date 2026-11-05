package trips

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/trip"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"github.com/tapiaw38/vamo/backend/internal/platform/storage"
)

type SearchInput struct {
	OriginLocalityID, DestinationLocalityID, Date string
	VehicleType                                   string
	Passengers, Page, PageSize                    int
}
type SearchUsecase interface {
	Execute(context.Context, SearchInput) ([]TripData, *apperrors.ApplicationError)
}
type searchUsecase struct {
	trips  trip.Repository
	images storage.ImageStorage
}

func NewSearchUsecase(trips trip.Repository, images ...storage.ImageStorage) SearchUsecase {
	var imageStorage storage.ImageStorage
	if len(images) > 0 {
		imageStorage = images[0]
	}
	return &searchUsecase{trips: trips, images: imageStorage}
}
func (u *searchUsecase) Execute(ctx context.Context, input SearchInput) ([]TripData, *apperrors.ApplicationError) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}
	items, err := u.trips.Search(ctx, domain.TripSearch{OriginLocalityID: input.OriginLocalityID, DestinationLocalityID: input.DestinationLocalityID, Date: input.Date, VehicleType: input.VehicleType, Passengers: input.Passengers, Page: input.Page, PageSize: input.PageSize})
	if err != nil {
		return nil, apperrors.New("trips:search-failed", 500, "could not search trips")
	}
	output := make([]TripData, 0, len(items))
	for i := range items {
		output = append(output, withPresignedPhoto(publicTripData(&items[i]), u.images))
	}
	return output, nil
}
