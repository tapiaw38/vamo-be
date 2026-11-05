package domain

type User struct {
	ID, FirstName, LastName, Email, Phone string
	ResidenceLocalityID                   *string
	ResidenceLocalityName                 *string
	ResidenceProvinceID                   *string
	ResidenceProvinceName                 *string
}
