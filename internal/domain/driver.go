package domain

import "time"

type DriverProfile struct {
	ID, UserID, Status, LicenseNumber, VerificationNotes string
	SensitiveMetadata                                    map[string]any
	IsAvailable                                          bool
	CreatedAt, UpdatedAt                                 time.Time
}

type Vehicle struct {
	ID, OwnerType, DriverProfileID, Make, Model, LicensePlate, VehicleType, VerificationStatus, PhotoURL string
	CompanyID                                                                                            *string
	Year, AuthorizedSeats                                                                                int
	Active                                                                                               bool
	CreatedAt, UpdatedAt                                                                                 time.Time
}

type VehicleDocument struct {
	ID, VehicleID, DocumentType, ExternalURL, VerificationStatus string
	Metadata                                                     map[string]any
}
