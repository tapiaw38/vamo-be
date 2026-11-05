package trips

import (
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"github.com/tapiaw38/vamo/backend/internal/platform/storage"
	"strings"
	"time"
)

type TripData struct {
	ID                    string     `json:"id"`
	PublisherType         string     `json:"publisher_type"`
	DriverProfileID       string     `json:"driver_profile_id"`
	VehicleID             string     `json:"vehicle_id"`
	OriginLocalityID      string     `json:"origin_locality_id"`
	DestinationLocalityID string     `json:"destination_locality_id"`
	OriginName            string     `json:"origin_name"`
	DestinationName       string     `json:"destination_name"`
	DriverName            string     `json:"driver_name"`
	AverageScore          *float64   `json:"average_score,omitempty"`
	TotalRatings          int        `json:"total_ratings"`
	VehicleMake           string     `json:"vehicle_make"`
	VehicleModel          string     `json:"vehicle_model"`
	VehicleType           string     `json:"vehicle_type"`
	VehicleLicensePlate   string     `json:"vehicle_license_plate,omitempty"`
	VehiclePhotoURL       string     `json:"vehicle_photo_url,omitempty"`
	DepartureAt           time.Time  `json:"departure_at"`
	ArrivalAt             *time.Time `json:"arrival_at,omitempty"`
	Timezone              string     `json:"timezone"`
	PricePerSeat          string     `json:"price_per_seat"`
	Currency              string     `json:"currency"`
	BookingMode           string     `json:"booking_mode"`
	Status                string     `json:"status"`
	Notes                 string     `json:"notes"`
	TotalSeats            int        `json:"total_seats"`
	AvailableSeats        int        `json:"available_seats"`
}

func tripData(item *domain.Trip) TripData {
	return TripData{ID: item.ID, PublisherType: item.PublisherType, DriverProfileID: item.DriverProfileID, VehicleID: item.VehicleID, OriginLocalityID: item.OriginLocalityID, DestinationLocalityID: item.DestinationLocalityID, OriginName: item.OriginName, DestinationName: item.DestinationName, DriverName: item.DriverName, AverageScore: item.AverageScore, TotalRatings: item.TotalRatings, VehicleMake: item.VehicleMake, VehicleModel: item.VehicleModel, VehicleType: item.VehicleType, VehicleLicensePlate: item.VehicleLicensePlate, VehiclePhotoURL: item.VehiclePhotoURL, DepartureAt: item.DepartureAt, ArrivalAt: item.ArrivalAt, Timezone: item.Timezone, PricePerSeat: item.PricePerSeat, Currency: item.Currency, BookingMode: item.BookingMode, Status: item.Status, Notes: item.Notes, TotalSeats: item.TotalSeats, AvailableSeats: item.AvailableSeats}
}

func publicTripData(item *domain.Trip) TripData {
	out := tripData(item)
	out.VehicleLicensePlate = maskLicensePlate(item.VehicleLicensePlate)
	out.VehiclePhotoURL = item.VehiclePhotoURL
	return out
}

func withPresignedPhoto(out TripData, images storage.ImageStorage) TripData {
	if images == nil || out.VehiclePhotoURL == "" {
		return out
	}
	if signed, ok := images.PresignGetURL(out.VehiclePhotoURL, 15*time.Minute); ok {
		out.VehiclePhotoURL = signed
	}
	return out
}

func maskLicensePlate(plate string) string {
	if len(plate) <= 3 {
		return "***"
	}
	return plate[:3] + strings.Repeat("*", len(plate)-3)
}
