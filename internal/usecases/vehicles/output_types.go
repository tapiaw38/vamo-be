package vehicles

import (
	"github.com/tapiaw38/vamo/backend/internal/domain"
	"github.com/tapiaw38/vamo/backend/internal/platform/storage"
	"time"
)

type VehicleData struct {
	ID                 string    `json:"id"`
	OwnerType          string    `json:"owner_type"`
	DriverProfileID    string    `json:"driver_profile_id"`
	Make               string    `json:"make"`
	Model              string    `json:"model"`
	LicensePlate       string    `json:"license_plate"`
	VehicleType        string    `json:"vehicle_type"`
	PhotoURL           string    `json:"photo_url,omitempty"`
	VerificationStatus string    `json:"verification_status"`
	Year               int       `json:"year"`
	AuthorizedSeats    int       `json:"authorized_seats"`
	Active             bool      `json:"active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func vehicleData(item *domain.Vehicle, images ...storage.ImageStorage) VehicleData {
	photoURL := item.PhotoURL
	if len(images) > 0 && images[0] != nil {
		if signed, ok := images[0].PresignGetURL(photoURL, 15*time.Minute); ok {
			photoURL = signed
		}
	}
	return VehicleData{ID: item.ID, OwnerType: item.OwnerType, DriverProfileID: item.DriverProfileID, Make: item.Make, Model: item.Model, LicensePlate: item.LicensePlate, VehicleType: item.VehicleType, PhotoURL: photoURL, VerificationStatus: item.VerificationStatus, Year: item.Year, AuthorizedSeats: item.AuthorizedSeats, Active: item.Active, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
