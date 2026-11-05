package trip

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

type Repository interface {
	Create(context.Context, domain.Trip) (*domain.Trip, error)
	Search(context.Context, domain.TripSearch) ([]domain.Trip, error)
	GetPublic(context.Context, string) (*domain.Trip, error)
	GetOwned(context.Context, string, string) (*domain.Trip, error)
	Update(context.Context, domain.Trip) (*domain.Trip, error)
	Cancel(context.Context, string, string) (*domain.Trip, error)
	ListMine(context.Context, string, string) ([]domain.Trip, error)
	ListMineForUser(context.Context, string, string) ([]domain.Trip, error)
}

type repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) Repository { return &repository{db: db} }
