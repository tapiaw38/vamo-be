package pricing

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"

	pricingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/pricing"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
)

type Calculator interface {
	Calculate(context.Context, string, string, ...float64) (domain.PriceQuote, error)
}

type Usecases struct{ Repo pricingrepo.Repository }

func NewUsecases(repo pricingrepo.Repository) Usecases { return Usecases{Repo: repo} }

func (u Usecases) Get(ctx context.Context) (domain.PricingSettings, *apperrors.ApplicationError) {
	item, err := u.Repo.Get(ctx)
	if err != nil {
		return domain.PricingSettings{}, apperrors.New("pricing:get-failed", http.StatusInternalServerError, "could not load pricing settings")
	}
	return item, nil
}

func (u Usecases) Update(ctx context.Context, basePrice, pricePerKm, urbanBasePrice, urbanPricePerKm, currency, actor string) (domain.PricingSettings, *apperrors.ApplicationError) {
	if !nonNegative(basePrice) || !nonNegative(pricePerKm) || !nonNegative(urbanBasePrice) || !nonNegative(urbanPricePerKm) || currency == "" {
		return domain.PricingSettings{}, apperrors.New("pricing:invalid-settings", http.StatusBadRequest, "prices must be non-negative and currency is required")
	}
	item, err := u.Repo.Update(ctx, basePrice, pricePerKm, urbanBasePrice, urbanPricePerKm, currency, actor)
	if err != nil {
		return domain.PricingSettings{}, apperrors.New("pricing:update-failed", http.StatusInternalServerError, "could not update pricing settings")
	}
	return item, nil
}

func (u Usecases) Calculate(ctx context.Context, originID, destinationID string, pinCoordinates ...float64) (domain.PriceQuote, error) {
	settings, err := u.Repo.Get(ctx)
	if err != nil {
		return domain.PriceQuote{}, err
	}
	coordinates := pinCoordinates
	if len(coordinates) != 4 {
		originLat, originLon, destinationLat, destinationLon, coordinateErr := u.Repo.GetRouteCoordinates(ctx, originID, destinationID)
		if coordinateErr != nil {
			return domain.PriceQuote{}, coordinateErr
		}
		coordinates = []float64{originLat, originLon, destinationLat, destinationLon}
	}
	distance := haversine(coordinates[0], coordinates[1], coordinates[2], coordinates[3])
	baseText, rateText := settings.BasePrice, settings.PricePerKm
	fareKind := "intercity"
	if originID == destinationID {
		baseText, rateText, fareKind = settings.UrbanBasePrice, settings.UrbanPricePerKm, "urban"
	}
	base, err := strconv.ParseFloat(baseText, 64)
	if err != nil {
		return domain.PriceQuote{}, err
	}
	rate, err := strconv.ParseFloat(rateText, 64)
	if err != nil {
		return domain.PriceQuote{}, err
	}
	return domain.PriceQuote{
		DistanceKM:   distance,
		BasePrice:    baseText,
		PricePerKm:   rateText,
		PricePerSeat: formatMoney(base + rate*distance),
		Currency:     settings.Currency,
		FareKind:     fareKind,
	}, nil
}

func (u Usecases) Quote(ctx context.Context, originID, destinationID string, pinCoordinates ...float64) (domain.PriceQuote, *apperrors.ApplicationError) {
	if originID == "" || destinationID == "" {
		return domain.PriceQuote{}, apperrors.New("pricing:invalid-route", http.StatusBadRequest, "origin and destination are required")
	}
	quote, err := u.Calculate(ctx, originID, destinationID, pinCoordinates...)
	if errors.Is(err, pricingrepo.ErrMissingCoordinates) {
		return domain.PriceQuote{}, apperrors.New("pricing:coordinates-missing", http.StatusUnprocessableEntity, "one locality has no coordinates")
	}
	if err != nil {
		return domain.PriceQuote{}, apperrors.New("pricing:quote-failed", http.StatusInternalServerError, "could not calculate quote")
	}
	return quote, nil
}

func nonNegative(value string) bool {
	n, err := strconv.ParseFloat(value, 64)
	return err == nil && n >= 0
}

func formatMoney(value float64) string {
	return strconv.FormatFloat(math.Round(value*100)/100, 'f', 2, 64)
}

// haversine calculates great-circle distance between two latitude/longitude points.
// Earth radius is expressed in kilometres; this precision is used for monetary pricing.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKM = 6371.0088
	lat1Rad, lat2Rad := lat1*math.Pi/180, lat2*math.Pi/180
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return earthRadiusKM * 2 * math.Asin(math.Sqrt(a))
}
