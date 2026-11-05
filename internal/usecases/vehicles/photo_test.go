package vehicles

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/tapiaw38/vamo/backend/internal/platform/config"
	"github.com/tapiaw38/vamo/backend/internal/platform/storage"
)

func imageData(size int, contentType string) string {
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(make([]byte, size))
}

func TestStorePhotoRejectsOversizeImage(t *testing.T) {
	_, appErr := storePhoto(context.Background(), storage.NoopImageStorage{}, "driver-a", imageData(storage.MaxImageBytes+1, "image/jpeg"))
	if appErr == nil || appErr.Status != 413 {
		t.Fatalf("expected 413, got %#v", appErr)
	}
}

func TestStorePhotoRejectsInvalidContentType(t *testing.T) {
	_, appErr := storePhoto(context.Background(), storage.NoopImageStorage{}, "driver-a", imageData(12, "application/pdf"))
	if appErr == nil || appErr.Status != 400 {
		t.Fatalf("expected 400, got %#v", appErr)
	}
}

func TestStorePhotoKeepsDataURIWithoutS3(t *testing.T) {
	value := imageData(12, "image/jpeg")
	stored, appErr := storePhoto(context.Background(), storage.NoopImageStorage{}, "driver-a", value)
	if appErr != nil || stored != value {
		t.Fatalf("noop storage failed: stored=%q err=%v", stored, appErr)
	}
}

func TestStoredURLMustBelongToOwner(t *testing.T) {
	images := storage.NewS3ImageStorage(config.S3Config{AWSRegion: "region", AWSAccessKeyID: "key", AWSSecretAccessKey: "secret", AWSBucket: "bucket"})
	owned := "https://bucket.s3.region.amazonaws.com/file/vehicles/driver-a/photo.jpg"
	foreign := strings.Replace(owned, "driver-a", "driver-b", 1)
	if !images.OwnsFileURL(owned, "vehicles", "driver-a") || images.OwnsFileURL(foreign, "vehicles", "driver-a") {
		t.Fatal("stored URL ownership check failed")
	}
}
