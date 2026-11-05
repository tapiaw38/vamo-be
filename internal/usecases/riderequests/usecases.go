package riderequests

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	driverrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	pricingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/pricing"
	repo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/riderequest"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	pricinguc "github.com/tapiaw38/vamo/backend/internal/usecases/pricing"
	"github.com/tapiaw38/vamo/backend/internal/usecases/profilebootstrap"
)

type Input struct {
	OriginLocalityID      string   `json:"origin_locality_id"`
	DestinationLocalityID string   `json:"destination_locality_id"`
	OriginLat             *float64 `json:"origin_lat"`
	OriginLon             *float64 `json:"origin_lon"`
	DestinationLat        *float64 `json:"destination_lat"`
	DestinationLon        *float64 `json:"destination_lon"`
	OriginAddress         string   `json:"origin_address"`
	DestinationAddress    string   `json:"destination_address"`
	RequestedAt           string   `json:"requested_at"`
	MaxPricePerSeat       string   `json:"max_price_per_seat"`
	Seats                 int      `json:"seats"`
	Notes                 string   `json:"notes"`
}
type SearchInput struct {
	OriginLocalityID      string
	DestinationLocalityID string
}
type RecentData struct {
	DestinationLocalityID string   `json:"destination_locality_id"`
	DestinationName       string   `json:"destination_name"`
	DestinationAddress    string   `json:"destination_address"`
	DestinationLat        *float64 `json:"destination_lat,omitempty"`
	DestinationLon        *float64 `json:"destination_lon,omitempty"`
}
type Data struct {
	ID                          string    `json:"id"`
	PassengerID                 string    `json:"passenger_id"`
	OriginLocalityID            string    `json:"origin_locality_id"`
	DestinationLocalityID       string    `json:"destination_locality_id"`
	OriginName                  string    `json:"origin_name"`
	DestinationName             string    `json:"destination_name"`
	OriginLat                   *float64  `json:"origin_lat,omitempty"`
	OriginLon                   *float64  `json:"origin_lon,omitempty"`
	DestinationLat              *float64  `json:"destination_lat,omitempty"`
	DestinationLon              *float64  `json:"destination_lon,omitempty"`
	OriginAddress               string    `json:"origin_address"`
	DestinationAddress          string    `json:"destination_address"`
	RequestedAt                 time.Time `json:"requested_at"`
	MaxPricePerSeat             string    `json:"max_price_per_seat"`
	Seats                       int       `json:"seats"`
	Status                      string    `json:"status"`
	AcceptedByDriverProfileID   string    `json:"accepted_by_driver_profile_id,omitempty"`
	AcceptedTripID              string    `json:"accepted_trip_id,omitempty"`
	AcceptedDriverName          string    `json:"accepted_driver_name,omitempty"`
	AcceptedDriverAverageScore  *float64  `json:"accepted_driver_average_score,omitempty"`
	AcceptedDriverTotalRatings  int       `json:"accepted_driver_total_ratings,omitempty"`
	AcceptedVehicleMake         string    `json:"accepted_vehicle_make,omitempty"`
	AcceptedVehicleModel        string    `json:"accepted_vehicle_model,omitempty"`
	AcceptedVehicleType         string    `json:"accepted_vehicle_type,omitempty"`
	AcceptedVehicleLicensePlate string    `json:"accepted_vehicle_license_plate,omitempty"`
	Notes                       string    `json:"notes"`
	CreatedAt                   time.Time `json:"created_at"`
}
type AcceptData struct {
	Request   Data   `json:"request"`
	TripID    string `json:"trip_id"`
	BookingID string `json:"booking_id"`
}
type AcceptInput struct {
	VehicleID string `json:"vehicle_id"`
}

type Usecases struct {
	Repo    repo.Repository
	Drivers driverrepo.Repository
	Pricing pricinguc.Calculator
	Profile profilebootstrap.Service
	Logger  *slog.Logger
}

func data(item *domain.RideRequest) Data {
	return Data{ID: item.ID, PassengerID: item.PassengerID, OriginLocalityID: item.OriginLocalityID, DestinationLocalityID: item.DestinationLocalityID, OriginName: item.OriginName, DestinationName: item.DestinationName, OriginLat: item.OriginLat, OriginLon: item.OriginLon, DestinationLat: item.DestinationLat, DestinationLon: item.DestinationLon, OriginAddress: item.OriginAddress, DestinationAddress: item.DestinationAddress, RequestedAt: item.RequestedAt, MaxPricePerSeat: item.MaxPricePerSeat, Seats: item.Seats, Status: item.Status, AcceptedByDriverProfileID: item.AcceptedByDriverProfileID, AcceptedTripID: item.AcceptedTripID, AcceptedDriverName: item.AcceptedDriverName, AcceptedDriverAverageScore: item.AcceptedDriverAverageScore, AcceptedDriverTotalRatings: item.AcceptedDriverTotalRatings, AcceptedVehicleMake: item.AcceptedVehicleMake, AcceptedVehicleModel: item.AcceptedVehicleModel, AcceptedVehicleType: item.AcceptedVehicleType, AcceptedVehicleLicensePlate: maskPlate(item.AcceptedVehicleLicensePlate), Notes: item.Notes, CreatedAt: item.CreatedAt}
}

func maskPlate(plate string) string {
	if plate == "" {
		return ""
	}
	if len(plate) <= 3 {
		return "***"
	}
	return plate[:3] + "***"
}
func (u Usecases) Create(ctx context.Context, passenger string, in Input, bearerTokens ...string) (Data, *apperrors.ApplicationError) {
	now := time.Now()
	requested := now
	if in.RequestedAt != "" {
		parsed, err := time.Parse(time.RFC3339, in.RequestedAt)
		if err != nil || parsed.Before(now.Add(-5*time.Minute)) {
			return Data{}, apperrors.New("ride-requests:invalid-requested-at", 400, "la fecha de salida es inválida o quedó demasiado en el pasado")
		}
		requested = parsed
	}
	if in.OriginLocalityID == "" {
		return Data{}, apperrors.New("ride-requests:origin-required", 400, "falta el origen")
	}
	if in.DestinationLocalityID == "" {
		return Data{}, apperrors.New("ride-requests:destination-required", 400, "falta el destino")
	}
	if u.Pricing == nil {
		return Data{}, apperrors.New("ride-requests:pricing-unavailable", 503, "el servicio de precios no está disponible")
	}
	active, activeErr := u.Repo.GetActive(ctx, passenger)
	if activeErr != nil {
		return Data{}, apperrors.New("ride-requests:active-check-failed", http.StatusInternalServerError, "no pudimos verificar tu pedido activo")
	}
	if active != nil {
		err := apperrors.New("ride-requests:active-exists", http.StatusConflict, "ya tenés un pedido activo; cancelalo o esperá que termine antes de pedir otro")
		err.Data = map[string]string{"active_request_id": active.ID}
		return Data{}, err
	}
	if in.Seats <= 0 {
		in.Seats = 1
	}
	bearerToken := ""
	if len(bearerTokens) > 0 {
		bearerToken = bearerTokens[0]
	}
	if err := u.Profile.Ensure(ctx, passenger, bearerToken); err != nil {
		if u.Logger != nil {
			u.Logger.Error("ride request profile bootstrap failed", "user_id", passenger, "error", err)
		}
		return Data{}, apperrors.New("ride-requests:profile-unavailable", http.StatusInternalServerError, "tu perfil todavía no está disponible, volvé a intentar")
	}
	coordinates := make([]float64, 0, 4)
	if in.OriginLat != nil && in.OriginLon != nil && in.DestinationLat != nil && in.DestinationLon != nil {
		coordinates = append(coordinates, *in.OriginLat, *in.OriginLon, *in.DestinationLat, *in.DestinationLon)
	}
	quote, quoteErr := u.Pricing.Calculate(ctx, in.OriginLocalityID, in.DestinationLocalityID, coordinates...)
	if errors.Is(quoteErr, pricingrepo.ErrMissingCoordinates) {
		return Data{}, apperrors.New("pricing:coordinates-missing", http.StatusUnprocessableEntity, "one locality has no coordinates")
	}
	if quoteErr != nil {
		return Data{}, apperrors.New("ride-requests:quote-failed", 500, "could not calculate ride request price")
	}
	item, e := u.Repo.Create(ctx, domain.RideRequest{PassengerID: passenger, OriginLocalityID: in.OriginLocalityID, DestinationLocalityID: in.DestinationLocalityID, OriginLat: in.OriginLat, OriginLon: in.OriginLon, DestinationLat: in.DestinationLat, DestinationLon: in.DestinationLon, OriginAddress: in.OriginAddress, DestinationAddress: in.DestinationAddress, RequestedAt: requested, MaxPricePerSeat: quote.PricePerSeat, Seats: in.Seats, Notes: in.Notes})
	if e != nil {
		if u.Logger != nil {
			attrs := []any{"user_id", passenger, "error", e}
			var pgErr *pgconn.PgError
			if errors.As(e, &pgErr) {
				attrs = append(attrs, "pg_code", pgErr.Code, "constraint", pgErr.ConstraintName)
				if pgErr.Code == "23505" && pgErr.ConstraintName == "ride_requests_one_active_per_passenger" {
					active, activeErr := u.Repo.GetActive(ctx, passenger)
					if activeErr == nil && active != nil {
						err := apperrors.New("ride-requests:active-exists", http.StatusConflict, "ya tenés un pedido activo; cancelalo o esperá que termine antes de pedir otro")
						err.Data = map[string]string{"active_request_id": active.ID}
						return Data{}, err
					}
				}
				if pgErr.ConstraintName == "ride_requests_passenger_id_fkey" {
					return Data{}, apperrors.New("ride-requests:profile-unavailable", http.StatusInternalServerError, "tu perfil todavía no está disponible, volvé a intentar")
				}
			}
			u.Logger.Error("ride request creation failed", attrs...)
		}
		return Data{}, apperrors.New("ride-requests:create-failed", 500, "could not create ride request")
	}
	return data(item), nil
}
func (u Usecases) List(ctx context.Context, user string, in SearchInput) ([]Data, *apperrors.ApplicationError) {
	profile, e := u.Drivers.GetByUserID(ctx, user)
	if e != nil || profile == nil || !profile.IsAvailable {
		return nil, apperrors.New("ride-requests:driver-unavailable", http.StatusForbidden, "driver must be available")
	}
	items, e := u.Repo.ListOpen(ctx, in.OriginLocalityID, in.DestinationLocalityID)
	if e != nil {
		return nil, apperrors.New("ride-requests:list-failed", 500, "could not list ride requests")
	}
	out := make([]Data, 0, len(items))
	for i := range items {
		out = append(out, data(&items[i]))
	}
	return out, nil
}
func (u Usecases) Active(ctx context.Context, passenger string) (*Data, *apperrors.ApplicationError) {
	item, err := u.Repo.GetActive(ctx, passenger)
	if err != nil {
		return nil, apperrors.New("ride-requests:active-failed", http.StatusInternalServerError, "could not get active ride request")
	}
	if item == nil {
		return nil, nil
	}
	output := data(item)
	return &output, nil
}
func (u Usecases) Recent(ctx context.Context, passenger string) ([]RecentData, *apperrors.ApplicationError) {
	items, err := u.Repo.RecentDestinations(ctx, passenger)
	if err != nil {
		return nil, apperrors.New("ride-requests:recent-failed", http.StatusInternalServerError, "could not get recent destinations")
	}
	out := make([]RecentData, 0, len(items))
	for _, item := range items {
		out = append(out, RecentData{DestinationLocalityID: item.DestinationLocalityID, DestinationName: item.DestinationName, DestinationAddress: item.DestinationAddress, DestinationLat: item.DestinationLat, DestinationLon: item.DestinationLon})
	}
	return out, nil
}
func (u Usecases) Cancel(ctx context.Context, passenger, id string) (Data, *apperrors.ApplicationError) {
	item, e := u.Repo.Cancel(ctx, id, passenger)
	if e != nil {
		return Data{}, apperrors.New("ride-requests:forbidden", 403, "ride request cannot be cancelled")
	}
	return data(item), nil
}
func (u Usecases) Accept(ctx context.Context, user, id string, vehicleIDs ...string) (AcceptData, *apperrors.ApplicationError) {
	vehicleID := ""
	if len(vehicleIDs) > 0 {
		vehicleID = vehicleIDs[0]
	}
	profile, e := u.Drivers.GetByUserID(ctx, user)
	if e != nil || profile == nil || !profile.IsAvailable {
		return AcceptData{}, apperrors.New("ride-requests:driver-unavailable", 403, "driver must be available")
	}
	request, e := u.Repo.Get(ctx, id)
	if e == repo.ErrNotFound {
		return AcceptData{}, apperrors.New("ride-requests:not-found", 404, "ride request not found")
	}
	if e != nil || u.Pricing == nil {
		return AcceptData{}, apperrors.New("ride-requests:unavailable", 409, "ride request is no longer available")
	}
	coordinates := make([]float64, 0, 4)
	if request.OriginLat != nil && request.OriginLon != nil && request.DestinationLat != nil && request.DestinationLon != nil {
		coordinates = append(coordinates, *request.OriginLat, *request.OriginLon, *request.DestinationLat, *request.DestinationLon)
	}
	quote, e := u.Pricing.Calculate(ctx, request.OriginLocalityID, request.DestinationLocalityID, coordinates...)
	if e != nil {
		return AcceptData{}, apperrors.New("ride-requests:quote-failed", 500, "could not calculate ride request price")
	}
	item, trip, booking, e := u.Repo.Accept(ctx, id, profile.ID, vehicleID, quote.PricePerSeat)
	if e == repo.ErrNotFound {
		return AcceptData{}, apperrors.New("ride-requests:not-found", 404, "ride request not found")
	}
	if e != nil {
		return AcceptData{}, apperrors.New("ride-requests:unavailable", 409, "ride request is no longer available")
	}
	return AcceptData{Request: data(&item), TripID: trip, BookingID: booking}, nil
}
