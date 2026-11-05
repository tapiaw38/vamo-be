package bookings

import (
	"errors"
	bookingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/booking"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
)

func mapError(err error) *apperrors.ApplicationError {
	switch {
	case errors.Is(err, bookingrepo.ErrNotFound):
		return apperrors.New("bookings:not-found", http.StatusNotFound, "booking not found")
	case errors.Is(err, bookingrepo.ErrForbidden):
		return apperrors.New("bookings:forbidden", http.StatusForbidden, "booking access forbidden")
	case errors.Is(err, bookingrepo.ErrSelfBooking):
		return apperrors.New("bookings:self-booking", http.StatusBadRequest, "driver cannot book own trip")
	case errors.Is(err, bookingrepo.ErrTripUnavailable):
		return apperrors.New("bookings:trip-unavailable", http.StatusConflict, "trip is unavailable")
	case errors.Is(err, bookingrepo.ErrConflict):
		return apperrors.New("bookings:invalid-transition", http.StatusConflict, "booking transition is invalid")
	default:
		return apperrors.New("bookings:operation-failed", http.StatusInternalServerError, "booking operation failed")
	}
}
