package domain

import "time"

type TripRating struct {
	ID, TripID, BookingID, PassengerID, DriverProfileID, Comment string
	Score                                                        int
	CreatedAt                                                    time.Time
}
