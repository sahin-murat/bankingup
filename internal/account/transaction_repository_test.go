package account

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestNewTransactionRepositoryRequiresDatabase(t *testing.T) {
	repository, err := NewTransactionRepository(nil)
	if err == nil {
		t.Fatal("NewTransactionRepository() error = nil, want an error")
	}
	if repository != nil {
		t.Errorf("NewTransactionRepository() repository = %v, want nil", repository)
	}
}

func TestTransactionRepositoryRequiresOperation(t *testing.T) {
	repository := &transactionRepository{}
	if err := repository.WithinTransaction(t.Context(), nil); err == nil {
		t.Fatal("WithinTransaction() error = nil, want an error")
	}
}

func TestTransactionRepositoryError(t *testing.T) {
	postgresError := &pgconn.PgError{
		Code:           "23505",
		ConstraintName: accountTransactionsIdempotencyConstraint,
	}
	err := transactionRepositoryError("test operation", postgresError)
	if !errors.Is(err, ErrIdempotencyKeyAlreadyExists) {
		t.Errorf("transactionRepositoryError() = %v, want %v", err, ErrIdempotencyKeyAlreadyExists)
	}
}
