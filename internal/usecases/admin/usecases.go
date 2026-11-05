package admin

import (
	"context"
	adminrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/admin"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
)

type Usecases struct{ Repo *adminrepo.Repository }

func mapErr(err error) *apperrors.ApplicationError {
	if err == nil {
		return nil
	}
	if err.Error() == "no rows in result set" {
		return apperrors.New("admin:not-found", 404, "entity not found")
	}
	return apperrors.New("admin:operation-failed", http.StatusInternalServerError, "admin operation failed")
}
func (u Usecases) Dashboard(ctx context.Context) (map[string]any, *apperrors.ApplicationError) {
	o, e := u.Repo.Dashboard(ctx)
	return o, mapErr(e)
}
func (u Usecases) List(ctx context.Context, kind string) ([]map[string]any, *apperrors.ApplicationError) {
	o, e := u.Repo.List(ctx, kind)
	return o, mapErr(e)
}
func (u Usecases) Verify(ctx context.Context, actor, kind, id, status string) *apperrors.ApplicationError {
	return mapErr(u.Repo.Verify(ctx, actor, kind, id, status))
}
