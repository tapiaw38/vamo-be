package trips

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"testing"
	"time"
)

type fakeDriverRepo struct{ profile *domain.DriverProfile }

func (f *fakeDriverRepo) GetByUserID(context.Context, string) (*domain.DriverProfile, error) {
	return f.profile, nil
}
func (f *fakeDriverRepo) Save(context.Context, domain.DriverProfile) (*domain.DriverProfile, error) {
	return f.profile, nil
}
func (f *fakeDriverRepo) SetAvailability(context.Context, string, bool) (*domain.DriverProfile, error) {
	return f.profile, nil
}

type fakeVehicleRepo struct{ vehicle *domain.Vehicle }

func (f *fakeVehicleRepo) Create(context.Context, domain.Vehicle) (*domain.Vehicle, error) {
	return f.vehicle, nil
}
func (f *fakeVehicleRepo) ListByDriver(context.Context, string) ([]domain.Vehicle, error) {
	return nil, nil
}
func (f *fakeVehicleRepo) GetByIDAndDriver(_ context.Context, _, driverID string) (*domain.Vehicle, error) {
	if f.vehicle == nil || f.vehicle.DriverProfileID != driverID {
		return nil, pgx.ErrNoRows
	}
	return f.vehicle, nil
}
func (f *fakeVehicleRepo) Update(context.Context, domain.Vehicle) (*domain.Vehicle, error) {
	return f.vehicle, nil
}
func (f *fakeVehicleRepo) CreateForCompany(context.Context, domain.Vehicle) (*domain.Vehicle, error) {
	return f.vehicle, nil
}
func (f *fakeVehicleRepo) ListByCompany(context.Context, string) ([]domain.Vehicle, error) {
	return nil, nil
}
func (f *fakeVehicleRepo) GetByIDAndCompany(context.Context, string, string) (*domain.Vehicle, error) {
	return f.vehicle, nil
}

type fakeTripRepo struct {
	item           *domain.Trip
	receivedDriver string
}

func (f *fakeTripRepo) Create(_ context.Context, item domain.Trip) (*domain.Trip, error) {
	f.receivedDriver = item.DriverProfileID
	item.ID = "trip-1"
	item.Status = "published"
	f.item = &item
	return &item, nil
}
func (f *fakeTripRepo) Search(context.Context, domain.TripSearch) ([]domain.Trip, error) {
	return nil, nil
}
func (f *fakeTripRepo) GetPublic(context.Context, string) (*domain.Trip, error) { return f.item, nil }
func (f *fakeTripRepo) GetOwned(context.Context, string, string) (*domain.Trip, error) {
	return f.item, nil
}
func (f *fakeTripRepo) Update(context.Context, domain.Trip) (*domain.Trip, error) { return f.item, nil }
func (f *fakeTripRepo) Cancel(context.Context, string, string) (*domain.Trip, error) {
	return f.item, nil
}
func (f *fakeTripRepo) ListMine(context.Context, string, string) ([]domain.Trip, error) {
	return nil, nil
}
func (f *fakeTripRepo) ListMineForUser(context.Context, string, string) ([]domain.Trip, error) {
	return nil, nil
}

func validTripInput() Input {
	return Input{VehicleID: "vehicle-1", OriginLocalityID: "origin", DestinationLocalityID: "destination", DepartureAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339), PricePerSeat: "25000.00", TotalSeats: 4, BookingMode: "automatic"}
}
func validCreateUC() (*fakeTripRepo, CreateUsecase) {
	drivers := &fakeDriverRepo{profile: &domain.DriverProfile{ID: "driver-1", UserID: "user-1", Status: "draft"}}
	vehicles := &fakeVehicleRepo{vehicle: &domain.Vehicle{ID: "vehicle-1", DriverProfileID: "driver-1", AuthorizedSeats: 4}}
	trips := &fakeTripRepo{}
	return trips, NewCreateUsecase(drivers, vehicles, trips)
}

func TestCreateRejectsSameLocality(t *testing.T) {
	_, uc := validCreateUC()
	input := validTripInput()
	input.DestinationLocalityID = input.OriginLocalityID
	_, appErr := uc.Execute(context.Background(), "user-1", input)
	if appErr == nil || appErr.Code != "trips:invalid-route" {
		t.Fatalf("expected invalid route, got %v", appErr)
	}
}
func TestCreateRejectsInvalidPriceAndCapacity(t *testing.T) {
	_, uc := validCreateUC()
	input := validTripInput()
	input.PricePerSeat = "0"
	_, appErr := uc.Execute(context.Background(), "user-1", input)
	if appErr == nil || appErr.Code != "trips:invalid-price" {
		t.Fatalf("expected invalid price, got %v", appErr)
	}
	input = validTripInput()
	input.TotalSeats = 5
	_, appErr = uc.Execute(context.Background(), "user-1", input)
	if appErr == nil || appErr.Code != "trips:invalid-seats" {
		t.Fatalf("expected invalid seats, got %v", appErr)
	}
}
