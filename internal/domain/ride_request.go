package domain

import "time"

type RideRequest struct {
	ID                                string
	PassengerID                       string
	OriginLocalityID                  string
	DestinationLocalityID             string
	OriginLat, OriginLon              *float64
	DestinationLat, DestinationLon    *float64
	OriginAddress, DestinationAddress string
	OriginName                        string
	DestinationName                   string
	RequestedAt                       time.Time
	MaxPricePerSeat                   string
	Seats                             int
	Status                            string
	AcceptedByDriverProfileID         string
	AcceptedTripID                    string
	Notes                             string
	CreatedAt, UpdatedAt              time.Time
	AcceptedDriverName                string
	AcceptedDriverAverageScore        *float64
	AcceptedDriverTotalRatings        int
	AcceptedVehicleMake               string
	AcceptedVehicleModel              string
	AcceptedVehicleType               string
	AcceptedVehicleLicensePlate       string
}

type RecentRideRequestDestination struct {
	DestinationLocalityID string
	DestinationName       string
	DestinationAddress    string
	DestinationLat        *float64
	DestinationLon        *float64
}
