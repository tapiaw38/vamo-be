package drivers

import (
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"time"
)

type DriverProfileData struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	Status            string    `json:"status"`
	LicenseNumber     string    `json:"license_number"`
	VerificationNotes string    `json:"verification_notes"`
	IsAvailable       bool      `json:"is_available"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func profileData(item *domain.DriverProfile) DriverProfileData {
	return DriverProfileData{ID: item.ID, UserID: item.UserID, Status: item.Status, LicenseNumber: item.LicenseNumber, VerificationNotes: item.VerificationNotes, IsAvailable: item.IsAvailable, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
