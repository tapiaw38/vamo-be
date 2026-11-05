package users

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"testing"
)

type fakeRepository struct{ saved domain.User }

func (f *fakeRepository) Get(context.Context, string) (*domain.User, error) { return nil, nil }
func (f *fakeRepository) Save(_ context.Context, item domain.User) (*domain.User, error) {
	f.saved = item
	return &item, nil
}
func (f *fakeRepository) Search(context.Context, string) ([]domain.User, error) { return nil, nil }

func TestUpdateUsesAuthenticatedUserID(t *testing.T) {
	repository := &fakeRepository{}
	usecase := NewUpdateUsecase(repository)
	_, err := usecase.Execute(context.Background(), "authenticated-user", UpdateInput{FirstName: "Updated"})
	if err != nil {
		t.Fatal(err)
	}
	if repository.saved.ID != "authenticated-user" {
		t.Fatalf("profile write targeted %q", repository.saved.ID)
	}
}
