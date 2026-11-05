package driver

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func (r *repository) GetByUserID(ctx context.Context, userID string) (*domain.DriverProfile, error) {
	var item domain.DriverProfile
	err := r.db.QueryRow(ctx, `SELECT id, user_id, status, COALESCE(license_number, ''), sensitive_metadata, COALESCE(verification_notes, ''), is_available, created_at, updated_at FROM driver_profiles WHERE user_id = $1`, userID).Scan(&item.ID, &item.UserID, &item.Status, &item.LicenseNumber, &item.SensitiveMetadata, &item.VerificationNotes, &item.IsAvailable, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}
