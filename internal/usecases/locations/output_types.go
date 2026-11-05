package locations

import "github.com/tapiaw38/vamo/backend/internal/domain"

type ProvinceData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}
type LocalityData struct {
	ID           string   `json:"id"`
	ProvinceID   string   `json:"province_id"`
	Name         string   `json:"name"`
	OfficialCode string   `json:"official_code"`
	Lat          *float64 `json:"lat,omitempty"`
	Lon          *float64 `json:"lon,omitempty"`
}

func provinceData(item domain.Province) ProvinceData {
	return ProvinceData{ID: item.ID, Name: item.Name, Code: item.Code}
}
func localityData(item domain.Locality) LocalityData {
	return LocalityData{ID: item.ID, ProvinceID: item.ProvinceID, Name: item.Name, OfficialCode: item.OfficialCode, Lat: item.Lat, Lon: item.Lon}
}
