package domain

import "time"

type PricingSettings struct {
	BasePrice       string
	PricePerKm      string
	UrbanBasePrice  string
	UrbanPricePerKm string
	Currency        string
	UpdatedAt       time.Time
	UpdatedBy       string
}

type PriceQuote struct {
	DistanceKM      float64
	BasePrice       string
	PricePerKm      string
	UrbanBasePrice  string
	UrbanPricePerKm string
	PricePerSeat    string
	Currency        string
	FareKind        string
}
