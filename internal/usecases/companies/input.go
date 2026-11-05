package companies

type CompanyInput struct {
	LegalName    string `json:"legal_name"`
	DisplayName  string `json:"display_name"`
	TaxID        string `json:"tax_id"`
	ProvinceID   string `json:"province_id"`
	LocalityID   string `json:"locality_id"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
}
type MemberInput struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}
type VehicleInput struct {
	Make            string `json:"make"`
	Model           string `json:"model"`
	Year            int    `json:"year"`
	LicensePlate    string `json:"license_plate"`
	AuthorizedSeats int    `json:"authorized_seats"`
	VehicleType     string `json:"vehicle_type"`
	PhotoURL        string `json:"photo_url"`
}
