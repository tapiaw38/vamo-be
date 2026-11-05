package domain

import "time"

type BookingPayment struct {
	ID, BookingID, Provider, ExternalReference, ProviderPaymentID, Amount, Currency, Status, IdempotencyKey string
	CreatedAt, UpdatedAt                                                                                    time.Time
}
