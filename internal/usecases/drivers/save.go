package drivers

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
	"strings"
)

type SaveInput struct {
	LicenseNumber     string         `json:"license_number"`
	SensitiveMetadata map[string]any `json:"sensitive_metadata"`
	VerificationNotes string         `json:"verification_notes"`
	SubmitForReview   bool           `json:"submit_for_review"`
}
type SaveUsecase interface {
	Execute(context.Context, string, SaveInput) (DriverProfileData, *apperrors.ApplicationError)
}
type saveUsecase struct{ repository driver.Repository }

func NewSaveUsecase(repository driver.Repository) SaveUsecase {
	return &saveUsecase{repository: repository}
}
func (u *saveUsecase) Execute(ctx context.Context, userID string, input SaveInput) (DriverProfileData, *apperrors.ApplicationError) {
	if strings.TrimSpace(userID) == "" {
		return DriverProfileData{}, apperrors.New("auth:identity-required", http.StatusUnauthorized, "identity required")
	}
	current, err := u.repository.GetByUserID(ctx, userID)
	if err != nil {
		return DriverProfileData{}, apperrors.New("drivers:get-failed", 500, "could not get driver profile")
	}
	status := "draft"
	if current != nil {
		status = current.Status
	}
	if input.SubmitForReview {
		if status != "draft" {
			return DriverProfileData{}, apperrors.New("drivers:invalid-transition", http.StatusConflict, "only draft profiles can be submitted")
		}
		status = "pending_review"
	} else if current != nil && status != "draft" && status != "pending_review" {
		return DriverProfileData{}, apperrors.New("drivers:invalid-transition", http.StatusConflict, "profile status cannot be edited")
	}
	metadata := input.SensitiveMetadata
	if metadata == nil && current != nil {
		metadata = current.SensitiveMetadata
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	saved, err := u.repository.Save(ctx, domain.DriverProfile{UserID: userID, Status: status, LicenseNumber: strings.TrimSpace(input.LicenseNumber), SensitiveMetadata: metadata, VerificationNotes: input.VerificationNotes})
	if err != nil {
		return DriverProfileData{}, apperrors.New("drivers:save-failed", 500, "could not save driver profile")
	}
	return profileData(saved), nil
}
