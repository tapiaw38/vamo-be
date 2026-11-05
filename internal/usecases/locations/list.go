package locations

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/location"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
)

type ListUsecase interface {
	Execute(context.Context) ([]ProvinceData, *apperrors.ApplicationError)
}
type listUsecase struct{ repository location.Repository }

func NewListUsecase(repository location.Repository) ListUsecase {
	return &listUsecase{repository: repository}
}
func (u *listUsecase) Execute(ctx context.Context) ([]ProvinceData, *apperrors.ApplicationError) {
	items, err := u.repository.ListProvinces(ctx)
	if err != nil {
		return nil, apperrors.New("locations:list-failed", 500, "could not list provinces")
	}
	out := make([]ProvinceData, 0, len(items))
	for _, item := range items {
		out = append(out, provinceData(item))
	}
	return out, nil
}
