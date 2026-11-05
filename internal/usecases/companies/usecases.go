package companies

import (
	"context"
	companyrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/company"
	vehiclerepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/vehicle"
	"github.com/tapiaw38/vamo/backend/internal/domain"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
	"strings"
	"time"
)

type Usecases struct {
	Companies   companyrepo.Repository
	VehicleRepo vehiclerepo.Repository
}

func mapErr(err error) *apperrors.ApplicationError {
	if err == companyrepo.ErrForbidden {
		return apperrors.New("companies:forbidden", http.StatusForbidden, "company access forbidden")
	}
	if err == companyrepo.ErrNotFound {
		return apperrors.New("companies:not-found", http.StatusNotFound, "company not found")
	}
	return apperrors.New("companies:operation-failed", 500, "company operation failed")
}
func (u Usecases) Create(ctx context.Context, user string, in CompanyInput) (CompanyData, *apperrors.ApplicationError) {
	if strings.TrimSpace(in.LegalName) == "" || strings.TrimSpace(in.DisplayName) == "" || strings.TrimSpace(in.TaxID) == "" {
		return CompanyData{}, apperrors.New("companies:invalid-input", 400, "company identity is required")
	}
	c, e := u.Companies.Create(ctx, domain.Company{LegalName: in.LegalName, DisplayName: in.DisplayName, TaxID: in.TaxID, ProvinceID: in.ProvinceID, LocalityID: in.LocalityID, ContactEmail: in.ContactEmail, ContactPhone: in.ContactPhone}, user)
	if e != nil {
		return CompanyData{}, mapErr(e)
	}
	return companyData(c), nil
}
func (u Usecases) Get(ctx context.Context, user, id string) (CompanyData, *apperrors.ApplicationError) {
	c, e := u.Companies.Get(ctx, id, user)
	if e != nil {
		return CompanyData{}, mapErr(e)
	}
	return companyData(c), nil
}
func (u Usecases) Update(ctx context.Context, user, id string, in CompanyInput) (CompanyData, *apperrors.ApplicationError) {
	c, e := u.Companies.Update(ctx, domain.Company{ID: id, LegalName: in.LegalName, DisplayName: in.DisplayName, TaxID: in.TaxID, ProvinceID: in.ProvinceID, LocalityID: in.LocalityID, ContactEmail: in.ContactEmail, ContactPhone: in.ContactPhone}, user)
	if e != nil {
		return CompanyData{}, mapErr(e)
	}
	return companyData(c), nil
}
func (u Usecases) List(ctx context.Context, user string) ([]CompanyData, *apperrors.ApplicationError) {
	cs, e := u.Companies.ListMine(ctx, user)
	if e != nil {
		return nil, mapErr(e)
	}
	out := make([]CompanyData, 0, len(cs))
	for i := range cs {
		out = append(out, companyData(&cs[i]))
	}
	return out, nil
}
func (u Usecases) Members(ctx context.Context, user, id string) ([]MemberData, *apperrors.ApplicationError) {
	ms, e := u.Companies.ListMembers(ctx, id, user)
	if e != nil {
		return nil, mapErr(e)
	}
	out := make([]MemberData, 0, len(ms))
	for _, m := range ms {
		out = append(out, memberData(m))
	}
	return out, nil
}
func (u Usecases) AddMember(ctx context.Context, user, id string, in MemberInput) *apperrors.ApplicationError {
	if in.Role != "owner" && in.Role != "admin" && in.Role != "dispatcher" && in.Role != "driver" {
		return apperrors.New("companies:invalid-role", 400, "invalid company role")
	}
	if e := u.Companies.AddMember(ctx, id, domain.CompanyMember{UserID: in.UserID, Role: in.Role}, user); e != nil {
		return mapErr(e)
	}
	return nil
}
func (u Usecases) RemoveMember(ctx context.Context, user, id, member string) *apperrors.ApplicationError {
	if e := u.Companies.RemoveMember(ctx, id, member, user); e != nil {
		return mapErr(e)
	}
	return nil
}
func (u Usecases) Vehicles(ctx context.Context, user, id string) ([]domain.Vehicle, *apperrors.ApplicationError) {
	ok, e := u.Companies.CanManage(ctx, id, user)
	if e != nil {
		return nil, mapErr(e)
	}
	if !ok {
		return nil, mapErr(companyrepo.ErrForbidden)
	}
	vs, e := u.VehicleRepo.ListByCompany(ctx, id)
	if e != nil {
		return nil, apperrors.New("companies:vehicles-failed", 500, "could not list company vehicles")
	}
	return vs, nil
}
func (u Usecases) AddVehicle(ctx context.Context, user, id string, in VehicleInput) (domain.Vehicle, *apperrors.ApplicationError) {
	ok, e := u.Companies.CanManage(ctx, id, user)
	if e != nil {
		return domain.Vehicle{}, mapErr(e)
	}
	if !ok {
		return domain.Vehicle{}, mapErr(companyrepo.ErrForbidden)
	}
	if in.AuthorizedSeats <= 0 || in.Year < 1900 || in.Year > time.Now().Year()+1 {
		return domain.Vehicle{}, apperrors.New("companies:invalid-vehicle", 400, "invalid vehicle")
	}
	v, e := u.VehicleRepo.CreateForCompany(ctx, domain.Vehicle{CompanyID: &id, Make: in.Make, Model: in.Model, Year: in.Year, LicensePlate: in.LicensePlate, AuthorizedSeats: in.AuthorizedSeats, VehicleType: in.VehicleType, PhotoURL: in.PhotoURL})
	if e != nil {
		return domain.Vehicle{}, apperrors.New("companies:vehicle-failed", 409, "could not add company vehicle")
	}
	return *v, nil
}
