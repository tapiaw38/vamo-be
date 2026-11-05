package domain

import "time"

type Company struct {
	ID, LegalName, DisplayName, TaxID, ProvinceID, LocalityID, Status, VerificationStatus, ContactEmail, ContactPhone string
	CreatedAt, UpdatedAt                                                                                              time.Time
}
type CompanyMember struct {
	CompanyID, UserID, Role, Status string
	DriverProfileID                 string
	CreatedAt, UpdatedAt            time.Time
}
type AuditLog struct {
	ID, ActorUserID, Action, EntityType, EntityID string
	Metadata                                      map[string]any
	CreatedAt                                     time.Time
}
