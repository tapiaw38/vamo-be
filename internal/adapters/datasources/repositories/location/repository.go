package location

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

type Repository interface {
	ListProvinces(context.Context) ([]domain.Province, error)
	ListLocalities(context.Context, string) ([]domain.Locality, error)
	SearchLocalities(context.Context, string) ([]domain.Locality, error)
	NearestLocality(context.Context, float64, float64) (*domain.Locality, error)
}

type repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) Repository { return &repository{db: db} }
