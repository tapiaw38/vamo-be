package domain

import "time"

type Booking struct {
	ID, TripID, PassengerID, UnitPriceSnapshot, TotalPrice, Currency, Status, FromStatus, Reason string
	Seats                                                                                        int
	RequestedAt, ConfirmedAt, CancelledAt, ExpiresAt, CreatedAt, UpdatedAt                       *time.Time
	Trip                                                                                         *Trip
	Rated                                                                                        bool
}
