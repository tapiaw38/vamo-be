package locations

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/location"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
)

type ListLocalitiesUsecase interface {
	Execute(context.Context, string) ([]LocalityData, *apperrors.ApplicationError)
}
type listLocalitiesUsecase struct{ repository location.Repository }

func NewListLocalitiesUsecase(repository location.Repository) ListLocalitiesUsecase {
	return &listLocalitiesUsecase{repository: repository}
}
func (u *listLocalitiesUsecase) Execute(ctx context.Context, provinceID string) ([]LocalityData, *apperrors.ApplicationError) {
	if provinceID == "" {
		return nil, apperrors.New("locations:province-required", http.StatusBadRequest, "province id is required")
	}
	items, err := u.repository.ListLocalities(ctx, provinceID)
	if err != nil {
		return nil, apperrors.New("locations:list-localities-failed", 500, "could not list localities")
	}
	out := make([]LocalityData, 0, len(items))
	for _, item := range items {
		out = append(out, localityData(item))
	}
	return out, nil
}
