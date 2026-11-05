package registration

import (
	"context"
	"errors"
	"testing"

	userrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/user"
	"github.com/tapiaw38/vamo/backend/internal/adapters/web/integrations/authapi"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

type registerAuthStub struct{}

func (registerAuthStub) Register(context.Context, authapi.RegisterInput) (authapi.RegisterResult, []byte, error) {
	var result authapi.RegisterResult
	result.Data.Username = "new-user"
	return result, []byte(`{"data":{"username":"new-user"}}`), nil
}
func (registerAuthStub) GetByEmail(context.Context, string, string) (*authapi.UserInfo, error) {
	return nil, nil
}
func (registerAuthStub) GetBatch(context.Context, string, []string) ([]authapi.UserInfo, error) {
	return nil, nil
}
func (registerAuthStub) GetTokenVersion(context.Context, string, string) (uint, error) { return 0, nil }
func (registerAuthStub) GetMe(context.Context, string) (*authapi.UserInfo, error)      { return nil, nil }
func (registerAuthStub) Login(context.Context, authapi.LoginInput) (authapi.LoginResult, error) {
	return authapi.LoginResult{}, nil
}
func (registerAuthStub) AssignRoles(context.Context, string, string, []string) error { return nil }

type failingProfileRepository struct{}

func (failingProfileRepository) Get(context.Context, string) (*domain.User, error) { return nil, nil }
func (failingProfileRepository) Save(context.Context, domain.User) (*domain.User, error) {
	return nil, errors.New("database unavailable")
}
func (failingProfileRepository) Search(context.Context, string) ([]domain.User, error) {
	return nil, nil
}

var _ userrepo.Repository = failingProfileRepository{}

func TestRegisterDoesNotReturnSuccessWhenLocalProfileBootstrapFails(t *testing.T) {
	uc := Usecase{Auth: registerAuthStub{}, Profile: failingProfileRepository{}}
	_, appErr := uc.Execute(context.Background(), Input{Email: "new@example.com", Password: "secret"})
	if appErr == nil || appErr.Code != "auth:profile-bootstrap-failed" || appErr.Status != 502 {
		t.Fatalf("expected explicit local bootstrap failure, got %#v", appErr)
	}
}
