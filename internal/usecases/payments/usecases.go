package payments

import (
	"context"
	"fmt"
	"github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/bookingpayment"
	paymentsapi "github.com/tapiaw38/vamo/backend/internal/adapters/web/integrations/payments"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	"net/http"
)

type Data struct {
	ID, BookingID, Provider, ExternalReference, ProviderPaymentID, Amount, Currency, Status string
	InitPoint                                                                               string `json:"init_point,omitempty"`
	SandboxInitPoint                                                                        string `json:"sandbox_init_point,omitempty"`
}
type Usecases struct {
	Repo    bookingpayment.Repository
	Client  paymentsapi.Client
	Enabled bool
}

func mapErr(e error) *apperrors.ApplicationError {
	switch e {
	case bookingpayment.ErrNotFound:
		return apperrors.New("payments:not-found", 404, "booking/payment not found")
	case bookingpayment.ErrForbidden:
		return apperrors.New("payments:forbidden", 403, "payment access forbidden")
	case bookingpayment.ErrConflict:
		return apperrors.New("payments:booking-not-confirmed", 409, "booking is not confirmed")
	}
	return apperrors.New("payments:operation-failed", 500, "payment operation failed")
}
func (u Usecases) Start(ctx context.Context, user, bookingID string) (Data, *apperrors.ApplicationError) {
	if !u.Enabled {
		return Data{}, apperrors.New("payments:disabled", http.StatusServiceUnavailable, "payments are disabled")
	}
	p, created, e := u.Repo.Prepare(ctx, bookingID, user)
	if e != nil {
		return Data{}, mapErr(e)
	}
	out := Data{ID: p.ID, BookingID: p.BookingID, Provider: p.Provider, ExternalReference: p.ExternalReference, ProviderPaymentID: p.ProviderPaymentID, Amount: p.Amount, Currency: p.Currency, Status: p.Status}
	if !created {
		return out, nil
	}
	res, e := u.Client.CreatePreference(ctx, paymentsapi.PreferenceRequest{Items: []paymentsapi.PreferenceItem{{Title: "VAMO! viaje", Quantity: 1, UnitPrice: p.Amount, CurrencyID: p.Currency}}, PayerEmail: user + "@vamo.invalid", ExternalReference: p.ExternalReference, UserID: user, BackURLs: paymentsapi.BackURLs{Success: "https://vamo.invalid/payment/success", Failure: "https://vamo.invalid/payment/failure", Pending: "https://vamo.invalid/payment/pending"}})
	if e != nil {
		return Data{}, apperrors.New("payments:upstream-unavailable", 502, "payments service unavailable")
	}
	if e = u.Repo.SetProvider(ctx, p.ID, providerID(res), "pending", res.InitPoint); e != nil {
		return Data{}, mapErr(e)
	}
	out.Status = "pending"
	out.InitPoint = res.InitPoint
	out.SandboxInitPoint = res.SandboxInitPoint
	return out, nil
}
func providerID(r paymentsapi.PreferenceResponse) string {
	if r.PaymentID == nil {
		return r.PreferenceID
	}
	return fmt.Sprint(*r.PaymentID)
}
func (u Usecases) Webhook(ctx context.Context, bookingID, providerID, status string) *apperrors.ApplicationError {
	allowed := map[string]bool{"pending": true, "approved": true, "rejected": true, "cancelled": true, "refunded": true}
	if !allowed[status] {
		return apperrors.New("payments:invalid-status", 400, "invalid payment status")
	}
	if e := u.Repo.UpdateStatus(ctx, bookingID, providerID, status); e != nil {
		return mapErr(e)
	}
	return nil
}
