package vehicles

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"testing"
)

type fakeDriverRepository struct{ item *domain.DriverProfile }

func (f *fakeDriverRepository) GetByUserID(context.Context, string) (*domain.DriverProfile, error) {
	return f.item, nil
}
func (f *fakeDriverRepository) Save(context.Context, domain.DriverProfile) (*domain.DriverProfile, error) {
	return f.item, nil
}
func (f *fakeDriverRepository) SetAvailability(context.Context, string, bool) (*domain.DriverProfile, error) {
	return f.item, nil
}

type fakeVehicleRepository struct{ item *domain.Vehicle }

func (f *fakeVehicleRepository) Create(context.Context, domain.Vehicle) (*domain.Vehicle, error) {
	return f.item, nil
}
func (f *fakeVehicleRepository) ListByDriver(context.Context, string) ([]domain.Vehicle, error) {
	return nil, nil
}
func (f *fakeVehicleRepository) GetByIDAndDriver(_ context.Context, _, driverID string) (*domain.Vehicle, error) {
	if f.item == nil || f.item.DriverProfileID != driverID {
		return nil, pgx.ErrNoRows
	}
	return f.item, nil
}
func (f *fakeVehicleRepository) Update(context.Context, domain.Vehicle) (*domain.Vehicle, error) {
	return f.item, nil
}
func (f *fakeVehicleRepository) CreateForCompany(context.Context, domain.Vehicle) (*domain.Vehicle, error) {
	return f.item, nil
}
func (f *fakeVehicleRepository) ListByCompany(context.Context, string) ([]domain.Vehicle, error) {
	return nil, nil
}
func (f *fakeVehicleRepository) GetByIDAndCompany(context.Context, string, string) (*domain.Vehicle, error) {
	return f.item, nil
}

func TestGetVehicleDoesNotCrossDriverOwnership(t *testing.T) {
	uc := NewGetUsecase(&fakeDriverRepository{item: &domain.DriverProfile{ID: "driver-a"}}, &fakeVehicleRepository{item: &domain.Vehicle{ID: "vehicle-1", DriverProfileID: "driver-b"}})
	_, appErr := uc.Execute(context.Background(), "user-a", "vehicle-1")
	if appErr == nil || appErr.Code != "vehicles:not-found" {
		t.Fatalf("expected not found for foreign vehicle, got %v", appErr)
	}
}
