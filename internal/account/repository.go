package account

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const (
	accountsBalanceConstraint  = "accounts_balance_not_negative"
	accountsCurrencyConstraint = "accounts_currency_fkey"
	accountsCustomerConstraint = "accounts_customer_fkey"
)

var (
	ErrNotFound         = errors.New("account not found")
	ErrCustomerNotFound = errors.New("account customer not found")
	ErrCurrencyNotFound = errors.New("account currency not found")
	ErrInvalidBalance   = errors.New("account balance is invalid")
)

type Repository interface {
	Create(context.Context, Account) (*Account, error)
	GetByID(context.Context, uuid.UUID) (*Account, error)
	List(context.Context, *uuid.UUID) ([]Account, error)
	UpdateStatus(context.Context, uuid.UUID, Status) (*Account, error)
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

func (r *repository) Create(ctx context.Context, input Account) (*Account, error) {
	created := input

	if err := r.db.WithContext(ctx).Create(&created).Error; err != nil {
		return nil, repositoryError("create account", err)
	}

	return &created, nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	var found Account

	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&found).Error; err != nil {
		return nil, repositoryError("get account by id", err)
	}

	return &found, nil
}

func (r *repository) List(ctx context.Context, customerID *uuid.UUID) ([]Account, error) {
	accounts := make([]Account, 0)
	query := r.db.WithContext(ctx).Order("id ASC")
	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}

	if err := query.Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	return accounts, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id uuid.UUID, status Status) (*Account, error) {
	result := r.db.WithContext(ctx).
		Model(&Account{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     status,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})

	if result.Error != nil {
		return nil, repositoryError("update account status", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("update account status: %w", ErrNotFound)
	}

	return r.GetByID(ctx, id)
}

func repositoryError(operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%s: %w", operation, ErrNotFound)
	}

	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return fmt.Errorf("%s: %w", operation, err)
	}

	switch {
	case postgresError.Code == "23503" && postgresError.ConstraintName == accountsCustomerConstraint:
		return fmt.Errorf("%s: %w", operation, ErrCustomerNotFound)
	case postgresError.Code == "23503" && postgresError.ConstraintName == accountsCurrencyConstraint:
		return fmt.Errorf("%s: %w", operation, ErrCurrencyNotFound)
	case postgresError.Code == "23514" && postgresError.ConstraintName == accountsBalanceConstraint:
		return fmt.Errorf("%s: %w", operation, ErrInvalidBalance)
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}
