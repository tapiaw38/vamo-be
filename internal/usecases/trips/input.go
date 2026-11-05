package trips

type Input struct {
	PublisherType         string  `json:"publisher_type"`
	CompanyID             string  `json:"company_id"`
	DriverProfileID       string  `json:"driver_profile_id"`
	VehicleID             string  `json:"vehicle_id"`
	OriginLocalityID      string  `json:"origin_locality_id"`
	DestinationLocalityID string  `json:"destination_locality_id"`
	DepartureAt           string  `json:"departure_at"`
	ArrivalAt             *string `json:"arrival_at"`
	Timezone              string  `json:"timezone"`
	PricePerSeat          string  `json:"price_per_seat"`
	Currency              string  `json:"currency"`
	TotalSeats            int     `json:"total_seats"`
	BookingMode           string  `json:"booking_mode"`
	Notes                 string  `json:"notes"`
}
