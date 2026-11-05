package pricing

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

type Repository interface {
	Get(context.Context) (domain.PricingSettings, error)
	GetRouteCoordinates(context.Context, string, string) (float64, float64, float64, float64, error)
	Update(context.Context, string, string, string, string, string, string) (domain.PricingSettings, error)
}

type repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) Repository { return &repository{db: db} }
