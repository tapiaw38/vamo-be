package users

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/user"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
	"strings"
)

type UpdateInput struct {
	FirstName           string  `json:"first_name"`
	LastName            string  `json:"last_name"`
	Phone               string  `json:"phone"`
	ResidenceLocalityID *string `json:"residence_locality_id"`
}
type UpdateUsecase interface {
	Execute(context.Context, string, UpdateInput) (ProfileData, *apperrors.ApplicationError)
}
type updateUsecase struct{ repository user.Repository }

func NewUpdateUsecase(repository user.Repository) UpdateUsecase {
	return &updateUsecase{repository: repository}
}
func (u *updateUsecase) Execute(ctx context.Context, userID string, input UpdateInput) (ProfileData, *apperrors.ApplicationError) {
	if strings.TrimSpace(userID) == "" {
		return ProfileData{}, apperrors.New("users:identity-required", http.StatusUnauthorized, "identity required")
	}
	saved, err := u.repository.Save(ctx, domain.User{ID: userID, FirstName: input.FirstName, LastName: input.LastName, Phone: input.Phone, ResidenceLocalityID: input.ResidenceLocalityID})
	if err != nil {
		return ProfileData{}, apperrors.New("users:update-failed", 500, "could not update profile")
	}
	return profileData(saved), nil
}
