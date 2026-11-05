package vehicles

import (
	"context"
	"strings"

	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"github.com/tapiaw38/vamo/backend/internal/platform/storage"
)

func storePhoto(ctx context.Context, images storage.ImageStorage, ownerID, value string) (string, *apperrors.ApplicationError) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if !strings.HasPrefix(value, "data:") {
		if images.OwnsFileURL(value, "vehicles", ownerID) {
			return value, nil
		}
		return "", apperrors.New("vehicles:invalid-photo", 400, "la foto debe ser una imagen seleccionada desde tu dispositivo")
	}
	if _, _, err := storage.DecodeImageDataURI(value); err != nil {
		status := 400
		if strings.Contains(err.Error(), "larger") {
			status = 413
		}
		return "", apperrors.New("vehicles:invalid-photo", status, "la foto debe ser JPEG, PNG o WebP y no superar 5 MB")
	}
	url, err := images.UploadDataURI(ctx, "vehicles", ownerID, value)
	if err != nil {
		return "", apperrors.New("vehicles:photo-upload-failed", 502, "no pudimos guardar la foto del vehículo")
	}
	return url, nil
}
