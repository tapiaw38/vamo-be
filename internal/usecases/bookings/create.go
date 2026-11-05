package bookings

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	bookingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/booking"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"github.com/tapiaw38/vamo/backend/internal/usecases/profilebootstrap"
	"net/http"
)

type CreateUsecase interface {
	Execute(context.Context, string, string, CreateInput, ...string) (BookingData, *apperrors.ApplicationError)
}
type createUsecase struct {
	repository bookingrepo.Repository
	profile    profilebootstrap.Service
}

func NewCreateUsecase(repository bookingrepo.Repository, profiles ...profilebootstrap.Service) CreateUsecase {
	var profile profilebootstrap.Service
	if len(profiles) > 0 {
		profile = profiles[0]
	}
	return &createUsecase{repository: repository, profile: profile}
}
func (u *createUsecase) Execute(ctx context.Context, passengerID, tripID string, input CreateInput, bearerTokens ...string) (BookingData, *apperrors.ApplicationError) {
	if input.Seats <= 0 {
		return BookingData{}, apperrors.New("bookings:invalid-seats", http.StatusBadRequest, "seats must be positive")
	}
	bearerToken := ""
	if len(bearerTokens) > 0 {
		bearerToken = bearerTokens[0]
	}
	if err := u.profile.Ensure(ctx, passengerID, bearerToken); err != nil {
		return BookingData{}, apperrors.New("bookings:profile-unavailable", http.StatusInternalServerError, "tu perfil todavía no está disponible, volvé a intentar")
	}
	item, err := u.repository.Create(ctx, passengerID, tripID, input.Seats)
	if err != nil {
		if u.profile.Logger != nil {
			attrs := []any{"passenger_id", passengerID, "trip_id", tripID, "error", err}
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				attrs = append(attrs, "pg_code", pgErr.Code, "constraint", pgErr.ConstraintName)
				if pgErr.ConstraintName == "bookings_passenger_id_fkey" {
					u.profile.Logger.Error("booking creation failed", attrs...)
					return BookingData{}, apperrors.New("bookings:profile-unavailable", http.StatusInternalServerError, "tu perfil todavía no está disponible, volvé a intentar")
				}
			}
			u.profile.Logger.Error("booking creation failed", attrs...)
		}
		return BookingData{}, mapError(err)
	}
	return bookingData(item), nil
}
