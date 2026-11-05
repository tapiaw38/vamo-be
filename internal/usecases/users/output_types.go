package users

import "github.com/tapiaw38/vamo/backend/internal/domain"

type ProfileData struct {
	ID                  string  `json:"id"`
	FirstName           string  `json:"first_name"`
	LastName            string  `json:"last_name"`
	Phone               string  `json:"phone"`
	ResidenceLocalityID *string `json:"residence_locality_id"`
	LocalityName        *string `json:"locality_name"`
	ProvinceID          *string `json:"province_id"`
	ProvinceName        *string `json:"province_name"`
}

func profileData(item *domain.User) ProfileData {
	return ProfileData{ID: item.ID, FirstName: item.FirstName, LastName: item.LastName, Phone: item.Phone, ResidenceLocalityID: item.ResidenceLocalityID, LocalityName: item.ResidenceLocalityName, ProvinceID: item.ResidenceProvinceID, ProvinceName: item.ResidenceProvinceName}
}

type SearchData struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}
