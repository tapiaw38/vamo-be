package profilebootstrap

import (
	"context"
	"log/slog"

	userrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/user"
	"github.com/tapiaw38/vamo/backend/internal/adapters/web/integrations/authapi"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

// Service lazily guarantees the local profile required by domain foreign keys.
// Auth data is best-effort; the JWT user id is enough to create a minimal row.
type Service struct {
	Profile userrepo.Repository
	Auth    authapi.Client
	Logger  *slog.Logger
}

func (s Service) Ensure(ctx context.Context, userID, bearerToken string) error {
	if userID == "" || s.Profile == nil {
		return nil
	}

	item, err := s.Profile.Get(ctx, userID)
	if err == nil && item != nil {
		return nil
	}
	if err != nil {
		s.log("local profile lookup failed", userID, err)
	}

	profile := domain.User{ID: userID}
	if s.Auth != nil && bearerToken != "" {
		upstream, authErr := s.Auth.GetMe(ctx, bearerToken)
		if authErr != nil {
			s.log("auth profile lookup failed; using minimal profile", userID, authErr)
		} else if upstream != nil {
			profile.FirstName = upstream.FirstName
			profile.LastName = upstream.LastName
			profile.Email = upstream.Email
		}
	}

	_, err = s.Profile.Save(ctx, profile)
	if err != nil {
		s.log("local profile creation failed", userID, err)
	}
	return err
}

func (s Service) log(message, userID string, err error) {
	if s.Logger != nil {
		s.Logger.Error(message, "user_id", userID, "error", err)
	}
}
