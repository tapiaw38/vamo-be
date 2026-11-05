package drivers

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"testing"
)

type fakeRepository struct{ item *domain.DriverProfile }

func (f *fakeRepository) GetByUserID(context.Context, string) (*domain.DriverProfile, error) {
	return f.item, nil
}
func (f *fakeRepository) Save(_ context.Context, item domain.DriverProfile) (*domain.DriverProfile, error) {
	item.ID = "driver-1"
	f.item = &item
	return &item, nil
}
func (f *fakeRepository) SetAvailability(context.Context, string, bool) (*domain.DriverProfile, error) {
	return f.item, nil
}

func TestSaveTransitionsDraftToPendingReview(t *testing.T) {
	repo := &fakeRepository{}
	uc := NewSaveUsecase(repo)
	output, appErr := uc.Execute(context.Background(), "user-1", SaveInput{SubmitForReview: true})
	if appErr != nil || output.Status != "pending_review" {
		t.Fatalf("expected pending_review, output=%+v error=%v", output, appErr)
	}
}

func TestSaveRejectsInvalidTransition(t *testing.T) {
	repo := &fakeRepository{item: &domain.DriverProfile{ID: "driver-1", UserID: "user-1", Status: "verified"}}
	_, appErr := NewSaveUsecase(repo).Execute(context.Background(), "user-1", SaveInput{SubmitForReview: true})
	if appErr == nil || appErr.Code != "drivers:invalid-transition" {
		t.Fatalf("expected invalid transition, got %v", appErr)
	}
}
