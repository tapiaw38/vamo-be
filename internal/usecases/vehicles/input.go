package vehicles

type Input struct {
	Make            string `json:"make"`
	Model           string `json:"model"`
	Year            int    `json:"year"`
	LicensePlate    string `json:"license_plate"`
	AuthorizedSeats int    `json:"authorized_seats"`
	VehicleType     string `json:"vehicle_type"`
	PhotoURL        string `json:"photo_url"`
}
