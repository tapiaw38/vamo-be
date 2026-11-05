package users

import (
	"testing"

	"github.com/tapiaw38/vamo/backend/internal/domain"
)

func TestProfileDataIncludesResidenceNames(t *testing.T) {
	localityID, localityName := "loc-1", "Tinogasta"
	provinceID, provinceName := "prov-1", "Catamarca"
	profile := profileData(&domain.User{ID: "user-1", ResidenceLocalityID: &localityID, ResidenceLocalityName: &localityName, ResidenceProvinceID: &provinceID, ResidenceProvinceName: &provinceName})
	if profile.LocalityName == nil || *profile.LocalityName != localityName || profile.ProvinceName == nil || *profile.ProvinceName != provinceName || profile.ProvinceID == nil || *profile.ProvinceID != provinceID {
		t.Fatalf("residence names not mapped: %+v", profile)
	}
}
