package company

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tapiaw38/vamo/backend/internal/domain"
)

var (
	ErrNotFound  = errors.New("company not found")
	ErrForbidden = errors.New("company access forbidden")
	ErrConflict  = errors.New("company conflict")
)

type Repository interface {
	Create(context.Context, domain.Company, string) (*domain.Company, error)
	Get(context.Context, string, string) (*domain.Company, error)
	Update(context.Context, domain.Company, string) (*domain.Company, error)
	ListMembers(context.Context, string, string) ([]domain.CompanyMember, error)
	AddMember(context.Context, string, domain.CompanyMember, string) error
	RemoveMember(context.Context, string, string, string) error
	CanManage(context.Context, string, string) (bool, error)
	CanDispatch(context.Context, string, string) (bool, error)
	DriverBelongs(context.Context, string, string) (bool, error)
	ListMine(context.Context, string) ([]domain.Company, error)
}
type repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) Repository { return &repository{db: db} }
func noRows(err error) bool                     { return errors.Is(err, pgx.ErrNoRows) }
