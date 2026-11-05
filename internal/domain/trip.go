package domain

import "time"

type Trip struct {
	ID, PublisherType, DriverProfileID, VehicleID, OriginLocalityID, DestinationLocalityID                                string
	CompanyID, Timezone, PricePerSeat, Currency, BookingMode, Status, Notes                                               string
	ArrivalAt                                                                                                             *time.Time
	DepartureAt                                                                                                           time.Time
	TotalSeats, AvailableSeats                                                                                            int
	DriverName, VehicleMake, VehicleModel, VehicleType, VehicleLicensePlate, VehiclePhotoURL, OriginName, DestinationName string
	AverageScore                                                                                                          *float64
	TotalRatings                                                                                                          int
	CreatedAt, UpdatedAt                                                                                                  time.Time
}

type TripSearch struct {
	OriginLocalityID, DestinationLocalityID, Date, VehicleType string
	Passengers, Page, PageSize                                 int
}
