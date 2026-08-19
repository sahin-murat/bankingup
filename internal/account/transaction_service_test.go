package account

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type fakeTransactionRepository struct {
	findByIdempotencyKey func(context.Context, uuid.UUID) (*Transaction, error)
	getAccountForUpdate  func(context.Context, uuid.UUID) (*Account, error)
	updateBalance        func(context.Context, uuid.UUID, decimal.Decimal) error
	createTransaction    func(context.Context, Transaction) (*Transaction, error)
}

func (r *fakeTransactionRepository) WithinTransaction(
	ctx context.Context,
	operation func(TransactionStore) error,
) error {
	return operation(r)
}

func (r *fakeTransactionRepository) FindByIdempotencyKey(
	ctx context.Context,
	key uuid.UUID,
) (*Transaction, error) {
	return r.findByIdempotencyKey(ctx, key)
}

func (r *fakeTransactionRepository) GetAccountForUpdate(
	ctx context.Context,
	id uuid.UUID,
) (*Account, error) {
	return r.getAccountForUpdate(ctx, id)
}

func (r *fakeTransactionRepository) UpdateBalance(
	ctx context.Context,
	id uuid.UUID,
	balance decimal.Decimal,
) error {
	return r.updateBalance(ctx, id, balance)
}

func (r *fakeTransactionRepository) CreateTransaction(
	ctx context.Context,
	input Transaction,
) (*Transaction, error) {
	return r.createTransaction(ctx, input)
}

func TestTransactionServiceDeposit(t *testing.T) {
	accountID := uuid.New()
	key := uuid.New()
	amount := decimal.RequireFromString("25.50")
	findCalls := 0
	repository := &fakeTransactionRepository{
		findByIdempotencyKey: func(context.Context, uuid.UUID) (*Transaction, error) {
			findCalls++
			return nil, ErrTransactionNotFound
		},
		getAccountForUpdate: func(context.Context, uuid.UUID) (*Account, error) {
			return &Account{
				ID:       accountID,
				Currency: "EUR",
				Balance:  decimal.RequireFromString("100.00"),
				Status:   StatusBlocked,
			}, nil
		},
		updateBalance: func(_ context.Context, id uuid.UUID, balance decimal.Decimal) error {
			if id != accountID || !balance.Equal(decimal.RequireFromString("125.50")) {
				t.Errorf("UpdateBalance() = (%v, %s), want (%v, 125.50)", id, balance, accountID)
			}
			return nil
		},
		createTransaction: func(_ context.Context, input Transaction) (*Transaction, error) {
			if input.Type != TransactionTypeDeposit || !input.Amount.Equal(amount) ||
				!input.BalanceAfter.Equal(decimal.RequireFromString("125.50")) {
				t.Errorf("CreateTransaction() input = %+v", input)
			}
			input.ID = uuid.New()
			return &input, nil
		},
	}
	service, err := NewTransactionService(repository, euroGetter())
	if err != nil {
		t.Fatalf("NewTransactionService() error = %v", err)
	}

	result, err := service.Deposit(t.Context(), MovementInput{
		AccountID:      accountID,
		Amount:         amount,
		IdempotencyKey: key,
	})
	if err != nil {
		t.Fatalf("Deposit() error = %v", err)
	}
	if result.Replay || result.Transaction == nil {
		t.Errorf("Deposit() result = %+v, want a new transaction", result)
	}
	if findCalls != 2 {
		t.Errorf("FindByIdempotencyKey() calls = %d, want 2", findCalls)
	}
}

func TestTransactionServiceIdempotentReplay(t *testing.T) {
	accountID := uuid.New()
	key := uuid.New()
	amount := decimal.RequireFromString("10.00")
	want := &Transaction{
		ID:             uuid.New(),
		AccountID:      accountID,
		Type:           TransactionTypeDeposit,
		Amount:         amount,
		IdempotencyKey: key,
	}
	repository := &fakeTransactionRepository{
		findByIdempotencyKey: func(context.Context, uuid.UUID) (*Transaction, error) {
			return want, nil
		},
	}
	service, err := NewTransactionService(repository, euroGetter())
	if err != nil {
		t.Fatalf("NewTransactionService() error = %v", err)
	}

	result, err := service.Deposit(t.Context(), MovementInput{
		AccountID:      accountID,
		Amount:         amount,
		IdempotencyKey: key,
	})
	if err != nil {
		t.Fatalf("Deposit() error = %v", err)
	}
	if !result.Replay || result.Transaction != want {
		t.Errorf("Deposit() result = %+v, want replay of %+v", result, want)
	}
}

func TestTransactionServiceIdempotencyConflict(t *testing.T) {
	repository := &fakeTransactionRepository{
		findByIdempotencyKey: func(context.Context, uuid.UUID) (*Transaction, error) {
			return &Transaction{
				AccountID: uuid.New(),
				Type:      TransactionTypeDeposit,
				Amount:    decimal.NewFromInt(10),
			}, nil
		},
	}
	service, err := NewTransactionService(repository, euroGetter())
	if err != nil {
		t.Fatalf("NewTransactionService() error = %v", err)
	}

	result, err := service.Deposit(t.Context(), MovementInput{
		AccountID:      uuid.New(),
		Amount:         decimal.NewFromInt(10),
		IdempotencyKey: uuid.New(),
	})
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Errorf("Deposit() error = %v, want %v", err, ErrIdempotencyConflict)
	}
	if result != nil {
		t.Errorf("Deposit() result = %+v, want nil", result)
	}
}

func TestTransactionServiceWithdrawalRules(t *testing.T) {
	tests := []struct {
		name    string
		account Account
		target  error
	}{
		{
			name:    "blocked account",
			account: Account{Currency: "EUR", Balance: decimal.NewFromInt(100), Status: StatusBlocked},
			target:  ErrAccountStatus,
		},
		{
			name:    "insufficient funds",
			account: Account{Currency: "EUR", Balance: decimal.NewFromInt(5), Status: StatusActive},
			target:  ErrInsufficientFunds,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeTransactionRepository{
				findByIdempotencyKey: func(context.Context, uuid.UUID) (*Transaction, error) {
					return nil, ErrTransactionNotFound
				},
				getAccountForUpdate: func(context.Context, uuid.UUID) (*Account, error) {
					return &test.account, nil
				},
			}
			service, err := NewTransactionService(repository, euroGetter())
			if err != nil {
				t.Fatalf("NewTransactionService() error = %v", err)
			}

			result, err := service.Withdraw(t.Context(), MovementInput{
				AccountID:      uuid.New(),
				Amount:         decimal.NewFromInt(10),
				IdempotencyKey: uuid.New(),
			})
			if !errors.Is(err, test.target) {
				t.Errorf("Withdraw() error = %v, want %v", err, test.target)
			}
			if result != nil {
				t.Errorf("Withdraw() result = %+v, want nil", result)
			}
		})
	}
}

func TestTransactionServiceRejectsInvalidAmount(t *testing.T) {
	service, err := NewTransactionService(&fakeTransactionRepository{}, euroGetter())
	if err != nil {
		t.Fatalf("NewTransactionService() error = %v", err)
	}

	result, err := service.Deposit(t.Context(), MovementInput{
		AccountID:      uuid.New(),
		Amount:         decimal.Zero,
		IdempotencyKey: uuid.New(),
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("Deposit() error = %v, want %v", err, ErrInvalidAmount)
	}
	if result != nil {
		t.Errorf("Deposit() result = %+v, want nil", result)
	}
}
