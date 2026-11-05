package companies

import (
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"time"
)

type CompanyData struct {
	ID                 string    `json:"id"`
	LegalName          string    `json:"legal_name"`
	DisplayName        string    `json:"display_name"`
	TaxID              string    `json:"tax_id"`
	ProvinceID         string    `json:"province_id"`
	LocalityID         string    `json:"locality_id"`
	Status             string    `json:"status"`
	VerificationStatus string    `json:"verification_status"`
	ContactEmail       string    `json:"contact_email"`
	ContactPhone       string    `json:"contact_phone"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
type MemberData struct {
	CompanyID       string    `json:"company_id"`
	UserID          string    `json:"user_id"`
	Role            string    `json:"role"`
	Status          string    `json:"status"`
	DriverProfileID string    `json:"driver_profile_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func companyData(c *domain.Company) CompanyData {
	return CompanyData{c.ID, c.LegalName, c.DisplayName, c.TaxID, c.ProvinceID, c.LocalityID, c.Status, c.VerificationStatus, c.ContactEmail, c.ContactPhone, c.CreatedAt, c.UpdatedAt}
}
func memberData(m domain.CompanyMember) MemberData {
	return MemberData{m.CompanyID, m.UserID, m.Role, m.Status, m.DriverProfileID, m.CreatedAt, m.UpdatedAt}
}
