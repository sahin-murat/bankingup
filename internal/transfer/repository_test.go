package transfer

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestNewRepositoryRequiresDatabase(t *testing.T) {
	repository, err := NewRepository(nil)
	if err == nil {
		t.Fatal("NewRepository() error = nil, want an error")
	}
	if repository != nil {
		t.Errorf("NewRepository() repository = %v, want nil", repository)
	}
}

func TestRepositoryErrorMapsDuplicateIdempotencyKey(t *testing.T) {
	err := repositoryError("test operation", &pgconn.PgError{
		Code:           "23505",
		ConstraintName: transfersIdempotencyConstraint,
	})
	if !errors.Is(err, ErrIdempotencyKeyAlreadyExists) {
		t.Errorf("repositoryError() = %v, want %v", err, ErrIdempotencyKeyAlreadyExists)
	}
}
