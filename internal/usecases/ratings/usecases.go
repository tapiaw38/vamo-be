package ratings

import (
	"context"
	"net/http"

	ratingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/rating"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
)

type Input struct {
	Score   int    `json:"score"`
	Comment string `json:"comment"`
}

type Data struct {
	ID              string `json:"id"`
	TripID          string `json:"trip_id"`
	BookingID       string `json:"booking_id"`
	PassengerID     string `json:"passenger_id"`
	DriverProfileID string `json:"driver_profile_id"`
	Score           int    `json:"score"`
	Comment         string `json:"comment,omitempty"`
}

type StatsData struct {
	AverageScore *float64 `json:"average_score,omitempty"`
	TotalRatings int      `json:"total_ratings"`
}

type Usecases struct{ Repo ratingrepo.Repository }

func (u Usecases) Create(ctx context.Context, passengerID, bookingID string, input Input) (Data, *apperrors.ApplicationError) {
	if input.Score < 1 || input.Score > 5 {
		return Data{}, apperrors.New("ratings:invalid-score", http.StatusBadRequest, "score must be between 1 and 5")
	}
	item, err := u.Repo.Create(ctx, bookingID, passengerID, input.Score, input.Comment)
	if err != nil {
		return Data{}, mapError(err)
	}
	return data(*item), nil
}

func (u Usecases) GetDriver(ctx context.Context, driverID string) (StatsData, *apperrors.ApplicationError) {
	stats, err := u.Repo.GetByDriver(ctx, driverID)
	if err != nil {
		return StatsData{}, apperrors.New("ratings:stats-failed", http.StatusInternalServerError, "could not load driver rating")
	}
	return StatsData{AverageScore: stats.AverageScore, TotalRatings: stats.TotalRatings}, nil
}

func data(item domain.TripRating) Data {
	return Data{ID: item.ID, TripID: item.TripID, BookingID: item.BookingID, PassengerID: item.PassengerID, DriverProfileID: item.DriverProfileID, Score: item.Score, Comment: item.Comment}
}

func mapError(err error) *apperrors.ApplicationError {
	switch {
	case err == ratingrepo.ErrNotFound:
		return apperrors.New("ratings:booking-not-found", http.StatusNotFound, "booking not found")
	case err == ratingrepo.ErrForbidden:
		return apperrors.New("ratings:forbidden", http.StatusForbidden, "only booking passenger can rate")
	case err == ratingrepo.ErrNotEligible:
		return apperrors.New("ratings:not-eligible", http.StatusBadRequest, "booking is not eligible for rating")
	case err == ratingrepo.ErrConflict:
		return apperrors.New("ratings:already-exists", http.StatusConflict, "booking already has a rating")
	default:
		return apperrors.New("ratings:create-failed", http.StatusInternalServerError, "could not create rating")
	}
}
