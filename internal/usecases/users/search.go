package users

import (
	"context"
	userrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/user"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"strings"
)

type SearchUsecase interface {
	Execute(context.Context, string) ([]SearchData, *apperrors.ApplicationError)
}
type searchUsecase struct{ repository userrepo.Repository }

func NewSearchUsecase(repository userrepo.Repository) SearchUsecase {
	return &searchUsecase{repository: repository}
}
func (u *searchUsecase) Execute(ctx context.Context, query string) ([]SearchData, *apperrors.ApplicationError) {
	if len(strings.TrimSpace(query)) < 2 {
		return []SearchData{}, nil
	}
	items, err := u.repository.Search(ctx, strings.TrimSpace(query))
	if err != nil {
		return nil, apperrors.New("users:search-failed", 500, "could not search users")
	}
	out := make([]SearchData, 0, len(items))
	for _, item := range items {
		out = append(out, SearchData{ID: item.ID, FirstName: item.FirstName, LastName: item.LastName, Email: item.Email})
	}
	return out, nil
}
