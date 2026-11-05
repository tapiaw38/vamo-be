package trips

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"testing"
	"time"
)

type accessTripRepo struct{}

func (accessTripRepo) Create(context.Context, domain.Trip) (*domain.Trip, error) { return nil, nil }
func (accessTripRepo) Search(context.Context, domain.TripSearch) ([]domain.Trip, error) {
	return nil, nil
}
func (accessTripRepo) GetPublic(context.Context, string) (*domain.Trip, error) { return nil, nil }
func (accessTripRepo) GetOwned(context.Context, string, string) (*domain.Trip, error) {
	return nil, pgx.ErrNoRows
}
func (accessTripRepo) Update(_ context.Context, item domain.Trip) (*domain.Trip, error) {
	if item.DriverProfileID != "driver-owner" {
		return nil, pgx.ErrNoRows
	}
	return &item, nil
}
func (accessTripRepo) Cancel(context.Context, string, string) (*domain.Trip, error) {
	return nil, pgx.ErrNoRows
}
func (accessTripRepo) ListMine(context.Context, string, string) ([]domain.Trip, error) {
	return nil, nil
}
func (accessTripRepo) ListMineForUser(context.Context, string, string) ([]domain.Trip, error) {
	return nil, nil
}

type accessDriverRepo struct{}

func (accessDriverRepo) GetByUserID(context.Context, string) (*domain.DriverProfile, error) {
	return &domain.DriverProfile{ID: "driver-attacker"}, nil
}
func (accessDriverRepo) Save(context.Context, domain.DriverProfile) (*domain.DriverProfile, error) {
	return nil, nil
}
func (accessDriverRepo) SetAvailability(context.Context, string, bool) (*domain.DriverProfile, error) {
	return nil, nil
}

type accessVehicleRepo struct{}

func (accessVehicleRepo) Create(context.Context, domain.Vehicle) (*domain.Vehicle, error) {
	return nil, nil
}
func (accessVehicleRepo) ListByDriver(context.Context, string) ([]domain.Vehicle, error) {
	return nil, nil
}
func (accessVehicleRepo) GetByIDAndDriver(context.Context, string, string) (*domain.Vehicle, error) {
	return &domain.Vehicle{ID: "vehicle-1", DriverProfileID: "driver-attacker", AuthorizedSeats: 4}, nil
}
func (accessVehicleRepo) Update(context.Context, domain.Vehicle) (*domain.Vehicle, error) {
	return nil, nil
}
func (accessVehicleRepo) CreateForCompany(context.Context, domain.Vehicle) (*domain.Vehicle, error) {
	return nil, nil
}
func (accessVehicleRepo) ListByCompany(context.Context, string) ([]domain.Vehicle, error) {
	return nil, nil
}
func (accessVehicleRepo) GetByIDAndCompany(context.Context, string, string) (*domain.Vehicle, error) {
	return nil, nil
}

func TestUpdateCannotEditForeignTrip(t *testing.T) {
	uc := NewUpdateUsecase(accessDriverRepo{}, accessVehicleRepo{}, accessTripRepo{})
	input := validTripInput()
	_, appErr := uc.Execute(context.Background(), "attacker", "trip-owner", input)
	if appErr == nil || appErr.Code != "trips:update-failed" {
		t.Fatalf("expected ownership failure, got %v", appErr)
	}
}
func TestCancelCannotCancelDepartedTrip(t *testing.T) {
	uc := NewCancelUsecase(accessDriverRepo{}, pastCancelTripRepo{})
	_, appErr := uc.Execute(context.Background(), "attacker", "trip-old")
	if appErr == nil || appErr.Code != "trips:cannot-cancel" {
		t.Fatalf("expected cancel rejection, got %v", appErr)
	}
}

type pastCancelTripRepo struct{ accessTripRepo }

func (pastCancelTripRepo) Cancel(context.Context, string, string) (*domain.Trip, error) {
	return &domain.Trip{ID: "trip-old", DepartureAt: time.Now().Add(-time.Hour)}, nil
}
