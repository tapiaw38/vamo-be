package driver

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

type Repository interface {
	GetByUserID(context.Context, string) (*domain.DriverProfile, error)
	Save(context.Context, domain.DriverProfile) (*domain.DriverProfile, error)
	SetAvailability(context.Context, string, bool) (*domain.DriverProfile, error)
}

type repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) Repository { return &repository{db: db} }
