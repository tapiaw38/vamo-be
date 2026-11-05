package trips

import (
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"math/big"
	"net/http"
	"strings"
	"time"
)

func parseTripInput(input Input, authorizedSeats int, requireFuture bool) (time.Time, *time.Time, *apperrors.ApplicationError) {
	departure, err := time.Parse(time.RFC3339, input.DepartureAt)
	if err != nil || (requireFuture && !departure.After(time.Now())) {
		return time.Time{}, nil, apperrors.New("trips:invalid-departure", http.StatusBadRequest, "departure must be a valid future RFC3339 datetime")
	}
	if input.OriginLocalityID == "" || input.DestinationLocalityID == "" || input.OriginLocalityID == input.DestinationLocalityID {
		return time.Time{}, nil, apperrors.New("trips:invalid-route", http.StatusBadRequest, "origin and destination must differ")
	}
	price, ok := new(big.Rat).SetString(strings.TrimSpace(input.PricePerSeat))
	if !ok || price.Sign() <= 0 {
		return time.Time{}, nil, apperrors.New("trips:invalid-price", http.StatusBadRequest, "price must be positive")
	}
	if input.TotalSeats <= 0 || input.TotalSeats > authorizedSeats {
		return time.Time{}, nil, apperrors.New("trips:invalid-seats", http.StatusBadRequest, "seats exceed vehicle authorized capacity")
	}
	if input.BookingMode != "automatic" && input.BookingMode != "manual" {
		return time.Time{}, nil, apperrors.New("trips:invalid-booking-mode", http.StatusBadRequest, "booking mode is invalid")
	}
	var arrival *time.Time
	if input.ArrivalAt != nil && strings.TrimSpace(*input.ArrivalAt) != "" {
		parsed, parseErr := time.Parse(time.RFC3339, *input.ArrivalAt)
		if parseErr != nil || !parsed.After(departure) {
			return time.Time{}, nil, apperrors.New("trips:invalid-arrival", http.StatusBadRequest, "arrival must be after departure")
		}
		arrival = &parsed
	}
	return departure, arrival, nil
}
