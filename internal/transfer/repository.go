package transfer

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const transfersIdempotencyConstraint = "transfers_idempotency_key_unique"

var (
	ErrNotFound                    = errors.New("transfer not found")
	ErrAccountNotFound             = errors.New("transfer account not found")
	ErrIdempotencyKeyAlreadyExists = errors.New("transfer idempotency key already exists")
)

type Store interface {
	FindByIdempotencyKey(context.Context, uuid.UUID) (*Transfer, error)
	GetAccountsForUpdate(context.Context, uuid.UUID, uuid.UUID) ([]accountdomain.Account, error)
	UpdateBalance(context.Context, uuid.UUID, decimal.Decimal) error
	CreateTransfer(context.Context, Transfer) (*Transfer, error)
	CreateAccountTransaction(context.Context, accountdomain.Transaction) error
}

type Repository interface {
	Store
	WithinTransaction(context.Context, func(Store) error) error
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

func (r *repository) WithinTransaction(ctx context.Context, operation func(Store) error) error {
	if operation == nil {
		return errors.New("transaction operation is required")
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return operation(&repository{db: tx})
	})
	if err != nil {
		return fmt.Errorf("execute transfer: %w", err)
	}

	return nil
}

func (r *repository) FindByIdempotencyKey(ctx context.Context, key uuid.UUID) (*Transfer, error) {
	var found Transfer

	if err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).Take(&found).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get transfer by idempotency key: %w", ErrNotFound)
		}

		return nil, fmt.Errorf("get transfer by idempotency key: %w", err)
	}

	return &found, nil
}

func (r *repository) GetAccountsForUpdate(
	ctx context.Context,
	firstID uuid.UUID,
	secondID uuid.UUID,
) ([]accountdomain.Account, error) {
	accounts := make([]accountdomain.Account, 0, 2)
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id IN ?", []uuid.UUID{firstID, secondID}).
		Order("id ASC").
		Find(&accounts).Error
	if err != nil {
		return nil, fmt.Errorf("get transfer accounts for update: %w", err)
	}
	if len(accounts) != 2 {
		return nil, ErrAccountNotFound
	}

	return accounts, nil
}

func (r *repository) UpdateBalance(ctx context.Context, id uuid.UUID, balance decimal.Decimal) error {
	result := r.db.WithContext(ctx).
		Model(&accountdomain.Account{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"balance":    balance,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		return fmt.Errorf("update transfer account balance: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAccountNotFound
	}

	return nil
}

func (r *repository) CreateTransfer(ctx context.Context, input Transfer) (*Transfer, error) {
	created := input

	if err := r.db.WithContext(ctx).Create(&created).Error; err != nil {
		return nil, repositoryError("create transfer", err)
	}

	return &created, nil
}

func (r *repository) CreateAccountTransaction(
	ctx context.Context,
	input accountdomain.Transaction,
) error {
	if err := r.db.WithContext(ctx).Create(&input).Error; err != nil {
		return fmt.Errorf("create transfer account transaction: %w", err)
	}

	return nil
}

func repositoryError(operation string, err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == transfersIdempotencyConstraint {
		return fmt.Errorf("%s: %w", operation, ErrIdempotencyKeyAlreadyExists)
	}

	return fmt.Errorf("%s: %w", operation, err)
}
