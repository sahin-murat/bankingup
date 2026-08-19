package currency

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const currenciesPrimaryKeyConstraint = "currencies_pkey"

var (
	ErrNotFound          = errors.New("currency not found")
	ErrCodeAlreadyExists = errors.New("currency code already exists")
)

type Repository interface {
	Create(context.Context, Currency) (*Currency, error)
	List(context.Context) ([]Currency, error)
	GetByCode(context.Context, string) (*Currency, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) (*repository, error) {
	if db == nil {
		return nil, errors.New("gorm database is required")
	}

	return &repository{db: db}, nil
}

func (r *repository) Create(ctx context.Context, input Currency) (*Currency, error) {
	created := input

	if err := r.db.WithContext(ctx).Create(&created).Error; err != nil {
		return nil, repositoryError("create currency", err)
	}

	return &created, nil
}

func (r *repository) List(ctx context.Context) ([]Currency, error) {
	currencies := make([]Currency, 0)

	if err := r.db.WithContext(ctx).Order("code ASC").Find(&currencies).Error; err != nil {
		return nil, fmt.Errorf("list currencies: %w", err)
	}

	return currencies, nil
}

func (r *repository) GetByCode(ctx context.Context, code string) (*Currency, error) {
	var found Currency

	if err := r.db.WithContext(ctx).Where("code = ?", code).Take(&found).Error; err != nil {
		return nil, repositoryError("get currency by code", err)
	}

	return &found, nil
}

func repositoryError(operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%s: %w", operation, ErrNotFound)
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == currenciesPrimaryKeyConstraint {
		return fmt.Errorf("%s: %w", operation, ErrCodeAlreadyExists)
	}

	return fmt.Errorf("%s: %w", operation, err)
}

var _ Repository = (*repository)(nil)
