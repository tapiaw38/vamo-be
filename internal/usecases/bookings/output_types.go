package bookings

import (
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"time"
)

type BookingData struct {
	ID                string          `json:"id"`
	TripID            string          `json:"trip_id"`
	PassengerID       string          `json:"passenger_id"`
	UnitPriceSnapshot string          `json:"unit_price_snapshot"`
	TotalPrice        string          `json:"total_price"`
	Currency          string          `json:"currency"`
	Status            string          `json:"status"`
	Seats             int             `json:"seats"`
	RequestedAt       *time.Time      `json:"requested_at"`
	ConfirmedAt       *time.Time      `json:"confirmed_at,omitempty"`
	CancelledAt       *time.Time      `json:"cancelled_at,omitempty"`
	ExpiresAt         *time.Time      `json:"expires_at,omitempty"`
	Trip              *domainTripData `json:"trip,omitempty"`
	Rated             bool            `json:"rated"`
	CanRate           bool            `json:"can_rate"`
}
type domainTripData struct {
	ID              string    `json:"id"`
	OriginName      string    `json:"origin_name"`
	DestinationName string    `json:"destination_name"`
	DriverName      string    `json:"driver_name"`
	VehicleMake     string    `json:"vehicle_make"`
	VehicleModel    string    `json:"vehicle_model"`
	PricePerSeat    string    `json:"price_per_seat"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	DepartureAt     time.Time `json:"departure_at"`
	AvailableSeats  int       `json:"available_seats"`
	TotalSeats      int       `json:"total_seats"`
}

func bookingData(item *domain.Booking) BookingData {
	var trip *domainTripData
	if item.Trip != nil {
		trip = &domainTripData{ID: item.Trip.ID, OriginName: item.Trip.OriginName, DestinationName: item.Trip.DestinationName, DriverName: item.Trip.DriverName, VehicleMake: item.Trip.VehicleMake, VehicleModel: item.Trip.VehicleModel, PricePerSeat: item.Trip.PricePerSeat, Currency: item.Trip.Currency, Status: item.Trip.Status, DepartureAt: item.Trip.DepartureAt, AvailableSeats: item.Trip.AvailableSeats, TotalSeats: item.Trip.TotalSeats}
	}
	canRate := item.Status == "completed"
	return BookingData{ID: item.ID, TripID: item.TripID, PassengerID: item.PassengerID, UnitPriceSnapshot: item.UnitPriceSnapshot, TotalPrice: item.TotalPrice, Currency: item.Currency, Status: item.Status, Seats: item.Seats, RequestedAt: item.RequestedAt, ConfirmedAt: item.ConfirmedAt, CancelledAt: item.CancelledAt, ExpiresAt: item.ExpiresAt, Trip: trip, Rated: item.Rated, CanRate: canRate && !item.Rated}
}
