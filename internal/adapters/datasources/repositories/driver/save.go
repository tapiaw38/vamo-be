package driver

import (
	"context"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) Save(ctx context.Context, item domain.DriverProfile) (*domain.DriverProfile, error) {
	var saved domain.DriverProfile
	err := r.db.QueryRow(ctx, `INSERT INTO driver_profiles (user_id, status, license_number, sensitive_metadata, verification_notes, updated_at) VALUES ($1, $2, $3, $4, $5, NOW()) ON CONFLICT (user_id) DO UPDATE SET status = EXCLUDED.status, license_number = EXCLUDED.license_number, sensitive_metadata = EXCLUDED.sensitive_metadata, verification_notes = EXCLUDED.verification_notes, updated_at = NOW() RETURNING id, user_id, status, COALESCE(license_number, ''), sensitive_metadata, COALESCE(verification_notes, ''), is_available, created_at, updated_at`, item.UserID, item.Status, item.LicenseNumber, item.SensitiveMetadata, item.VerificationNotes).Scan(&saved.ID, &saved.UserID, &saved.Status, &saved.LicenseNumber, &saved.SensitiveMetadata, &saved.VerificationNotes, &saved.IsAvailable, &saved.CreatedAt, &saved.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

func (r *repository) SetAvailability(ctx context.Context, userID string, available bool) (*domain.DriverProfile, error) {
	var item domain.DriverProfile
	err := r.db.QueryRow(ctx, `UPDATE driver_profiles SET is_available=$2, updated_at=NOW() WHERE user_id=$1 RETURNING id,user_id,status,COALESCE(license_number,''),sensitive_metadata,COALESCE(verification_notes,''),is_available,created_at,updated_at`, userID, available).Scan(&item.ID, &item.UserID, &item.Status, &item.LicenseNumber, &item.SensitiveMetadata, &item.VerificationNotes, &item.IsAvailable, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}
