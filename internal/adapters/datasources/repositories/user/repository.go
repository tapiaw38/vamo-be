package user

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

type Repository interface {
	Get(context.Context, string) (*domain.User, error)
	Save(context.Context, domain.User) (*domain.User, error)
	Search(context.Context, string) ([]domain.User, error)
}
type repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) Repository { return &repository{db: db} }
