package account

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
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

func TestRepositoryError(t *testing.T) {
	databaseError := errors.New("database failure")

	tests := []struct {
		name   string
		err    error
		target error
	}{
		{
			name:   "account not found",
			err:    gorm.ErrRecordNotFound,
			target: ErrNotFound,
		},
		{
			name: "customer not found",
			err: &pgconn.PgError{
				Code:           "23503",
				ConstraintName: accountsCustomerConstraint,
			},
			target: ErrCustomerNotFound,
		},
		{
			name: "currency not found",
			err: &pgconn.PgError{
				Code:           "23503",
				ConstraintName: accountsCurrencyConstraint,
			},
			target: ErrCurrencyNotFound,
		},
		{
			name: "invalid balance",
			err: &pgconn.PgError{
				Code:           "23514",
				ConstraintName: accountsBalanceConstraint,
			},
			target: ErrInvalidBalance,
		},
		{
			name:   "other database error",
			err:    databaseError,
			target: databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := repositoryError("test operation", test.err)
			if !errors.Is(err, test.target) {
				t.Errorf("repositoryError() = %v, want error matching %v", err, test.target)
			}
		})
	}
}
