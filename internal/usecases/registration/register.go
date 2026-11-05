package registration

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	userrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/user"
	"github.com/tapiaw38/vamo/backend/internal/adapters/web/integrations/authapi"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
)

type Input struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	WantsToBeDriver bool   `json:"wants_to_be_driver"`
}

type Usecase struct {
	Auth          authapi.Client
	Profile       userrepo.Repository
	ServiceEmail  string
	ServiceSecret string
	Logger        *slog.Logger
}

func (u Usecase) Execute(ctx context.Context, input Input) ([]byte, *apperrors.ApplicationError) {
	result, body, err := u.Auth.Register(ctx, authapi.RegisterInput{FirstName: input.FirstName, LastName: input.LastName, Email: input.Email, Password: input.Password})
	if err != nil {
		status := http.StatusBadGateway
		message := "auth service registration failed"
		var upstream *authapi.UpstreamError
		if errors.As(err, &upstream) {
			status = upstream.Status
			if parsed := extractUpstreamMessage(body); parsed != "" {
				message = parsed
			}
		}
		return nil, apperrors.New("auth:register-failed", status, message)
	}
	if result.Data.Username != "" && u.Profile != nil {
		profile := domain.User{ID: result.Data.Username, FirstName: input.FirstName, LastName: input.LastName, Email: input.Email}
		var profileErr error
		for attempt := 0; attempt < 2; attempt++ {
			if _, profileErr = u.Profile.Save(ctx, profile); profileErr == nil {
				break
			}
		}
		if profileErr != nil {
			if u.Logger != nil {
				u.Logger.Error("local profile bootstrap failed after retry", "user_id", result.Data.Username, "error", profileErr)
			}
			// The account already exists upstream, so re-registering would only hit a 409.
			// Signing in is the way forward: the profile is created lazily on first use.
			return nil, apperrors.New("auth:profile-bootstrap-failed", http.StatusBadGateway, "tu cuenta se creó, pero no pudimos preparar tu perfil; iniciá sesión para completarlo")
		}
	}
	if input.WantsToBeDriver {
		u.grantConductorRole(ctx, result.Data.ID)
	}
	return body, nil
}

func extractUpstreamMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var parsed struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ""
	}
	return parsed.Message
}

func (u Usecase) grantConductorRole(ctx context.Context, userID string) {
	if userID == "" || u.ServiceEmail == "" || u.ServiceSecret == "" {
		u.Logger.Warn("conductor role grant skipped", "user_id", userID, "reason", "service credentials unavailable")
		return
	}
	serviceSession, err := u.Auth.Login(ctx, authapi.LoginInput{Email: u.ServiceEmail, Password: u.ServiceSecret})
	if err == nil {
		err = u.Auth.AssignRoles(ctx, "Bearer "+serviceSession.Token, userID, []string{"user", "admin"})
	}
	if err != nil {
		u.Logger.Error("conductor role grant failed", "user_id", userID, "error", err)
	}
}
