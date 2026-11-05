package vehicle

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

type Repository interface {
	Create(context.Context, domain.Vehicle) (*domain.Vehicle, error)
	ListByDriver(context.Context, string) ([]domain.Vehicle, error)
	GetByIDAndDriver(context.Context, string, string) (*domain.Vehicle, error)
	Update(context.Context, domain.Vehicle) (*domain.Vehicle, error)
	CreateForCompany(context.Context, domain.Vehicle) (*domain.Vehicle, error)
	ListByCompany(context.Context, string) ([]domain.Vehicle, error)
	GetByIDAndCompany(context.Context, string, string) (*domain.Vehicle, error)
}

type repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) Repository { return &repository{db: db} }
