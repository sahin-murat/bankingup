package account

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const accountTransactionsIdempotencyConstraint = "account_transactions_idempotency_key_unique"

var (
	ErrTransactionNotFound         = errors.New("account transaction not found")
	ErrIdempotencyKeyAlreadyExists = errors.New("account transaction idempotency key already exists")
)

type TransactionStore interface {
	FindByIdempotencyKey(context.Context, uuid.UUID) (*Transaction, error)
	GetAccountForUpdate(context.Context, uuid.UUID) (*Account, error)
	UpdateBalance(context.Context, uuid.UUID, decimal.Decimal) error
	CreateTransaction(context.Context, Transaction) (*Transaction, error)
}

type TransactionRepository interface {
	TransactionStore
	WithinTransaction(context.Context, func(TransactionStore) error) error
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) (*transactionRepository, error) {
	if db == nil {
		return nil, errors.New("gorm database is required")
	}

	return &transactionRepository{db: db}, nil
}

func (r *transactionRepository) WithinTransaction(
	ctx context.Context,
	operation func(TransactionStore) error,
) error {
	if operation == nil {
		return errors.New("transaction operation is required")
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return operation(&transactionRepository{db: tx})
	})
	if err != nil {
		return fmt.Errorf("execute account transaction: %w", err)
	}

	return nil
}

func (r *transactionRepository) FindByIdempotencyKey(
	ctx context.Context,
	key uuid.UUID,
) (*Transaction, error) {
	var found Transaction

	if err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).Take(&found).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get account transaction by idempotency key: %w", ErrTransactionNotFound)
		}

		return nil, fmt.Errorf("get account transaction by idempotency key: %w", err)
	}

	return &found, nil
}

func (r *transactionRepository) GetAccountForUpdate(ctx context.Context, id uuid.UUID) (*Account, error) {
	var found Account

	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		Take(&found).Error
	if err != nil {
		return nil, repositoryError("get account for update", err)
	}

	return &found, nil
}

func (r *transactionRepository) UpdateBalance(
	ctx context.Context,
	id uuid.UUID,
	balance decimal.Decimal,
) error {
	result := r.db.WithContext(ctx).
		Model(&Account{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"balance":    balance,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		return repositoryError("update account balance", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update account balance: %w", ErrNotFound)
	}

	return nil
}

func (r *transactionRepository) CreateTransaction(
	ctx context.Context,
	input Transaction,
) (*Transaction, error) {
	created := input

	if err := r.db.WithContext(ctx).Create(&created).Error; err != nil {
		return nil, transactionRepositoryError("create account transaction", err)
	}

	return &created, nil
}

func transactionRepositoryError(operation string, err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == accountTransactionsIdempotencyConstraint {
		return fmt.Errorf("%s: %w", operation, ErrIdempotencyKeyAlreadyExists)
	}

	return fmt.Errorf("%s: %w", operation, err)
}
