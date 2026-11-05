package domain

type Province struct{ ID, Name, Code string }
type Locality struct {
	ID, ProvinceID, Name, NormalizedName, OfficialCode string
	Lat, Lon                                           *float64
}
