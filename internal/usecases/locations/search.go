package locations

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/location"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
	"strings"
)

type SearchUsecase interface {
	Execute(context.Context, string) ([]LocalityData, *apperrors.ApplicationError)
}
type searchUsecase struct{ repository location.Repository }

func NewSearchUsecase(repository location.Repository) SearchUsecase {
	return &searchUsecase{repository: repository}
}
func (u *searchUsecase) Execute(ctx context.Context, query string) ([]LocalityData, *apperrors.ApplicationError) {
	if strings.TrimSpace(query) == "" {
		return nil, apperrors.New("locations:query-required", http.StatusBadRequest, "query is required")
	}
	items, err := u.repository.SearchLocalities(ctx, normalizeQuery(query))
	if err != nil {
		return nil, apperrors.New("locations:search-failed", 500, "could not search localities")
	}
	out := make([]LocalityData, 0, len(items))
	for _, item := range items {
		out = append(out, localityData(item))
	}
	return out, nil
}
func normalizeQuery(value string) string {
	return strings.ToLower(strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n").Replace(strings.TrimSpace(value)))
}
