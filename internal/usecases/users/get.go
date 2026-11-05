package users

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/user"
	"github.com/tapiaw38/vamo/backend/internal/adapters/web/integrations/authapi"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
)

type GetUsecase interface {
	Execute(context.Context, string, string) (ProfileData, *apperrors.ApplicationError)
}
type getUsecase struct {
	repository user.Repository
	auth       authapi.Client
}

func NewGetUsecase(repository user.Repository, auth authapi.Client) GetUsecase {
	return &getUsecase{repository: repository, auth: auth}
}
func (u *getUsecase) Execute(ctx context.Context, userID, bearerToken string) (ProfileData, *apperrors.ApplicationError) {
	item, err := u.repository.Get(ctx, userID)
	if err != nil {
		return ProfileData{}, apperrors.New("users:get-failed", 500, "could not get profile")
	}
	if item != nil {
		return profileData(item), nil
	}
	upstream, err := u.auth.GetMe(ctx, bearerToken)
	if err != nil || upstream == nil {
		return ProfileData{}, apperrors.New("users:auth-profile-unavailable", http.StatusBadGateway, "could not bootstrap profile")
	}
	created, err := u.repository.Save(ctx, domain.User{ID: userID, FirstName: upstream.FirstName, LastName: upstream.LastName, Email: upstream.Email})
	if err != nil {
		return ProfileData{}, apperrors.New("users:create-profile-failed", 500, "could not create profile")
	}
	return profileData(created), nil
}
