package transfer

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
	currencydomain "github.com/sahin-murat/bankingup/internal/currency"
	"github.com/shopspring/decimal"
)

type fakeRepository struct {
	find          func(context.Context, uuid.UUID) (*Transfer, error)
	accounts      func(context.Context, uuid.UUID, uuid.UUID) ([]accountdomain.Account, error)
	update        func(context.Context, uuid.UUID, decimal.Decimal) error
	create        func(context.Context, Transfer) (*Transfer, error)
	createAccount func(context.Context, accountdomain.Transaction) error
}

func (r *fakeRepository) WithinTransaction(ctx context.Context, operation func(Store) error) error {
	return operation(r)
}
func (r *fakeRepository) FindByIdempotencyKey(ctx context.Context, key uuid.UUID) (*Transfer, error) {
	return r.find(ctx, key)
}
func (r *fakeRepository) GetAccountsForUpdate(ctx context.Context, first uuid.UUID, second uuid.UUID) ([]accountdomain.Account, error) {
	return r.accounts(ctx, first, second)
}
func (r *fakeRepository) UpdateBalance(ctx context.Context, id uuid.UUID, balance decimal.Decimal) error {
	return r.update(ctx, id, balance)
}
func (r *fakeRepository) CreateTransfer(ctx context.Context, input Transfer) (*Transfer, error) {
	return r.create(ctx, input)
}
func (r *fakeRepository) CreateAccountTransaction(ctx context.Context, input accountdomain.Transaction) error {
	return r.createAccount(ctx, input)
}

type currencyGetterFunc func(context.Context, string) (*currencydomain.Currency, error)

func (f currencyGetterFunc) GetByCode(ctx context.Context, code string) (*currencydomain.Currency, error) {
	return f(ctx, code)
}

func euroGetter() currencyGetterFunc {
	return func(context.Context, string) (*currencydomain.Currency, error) {
		return &currencydomain.Currency{Code: "EUR", DecimalPlaces: 2}, nil
	}
}

func TestServiceCreate(t *testing.T) {
	sourceID := uuid.New()
	destinationID := uuid.New()
	key := uuid.New()
	amount := decimal.NewFromInt(25)
	findCalls := 0
	updates := make(map[uuid.UUID]decimal.Decimal)
	accountTransactions := make([]accountdomain.Transaction, 0, 2)
	repository := &fakeRepository{
		find: func(context.Context, uuid.UUID) (*Transfer, error) {
			findCalls++
			return nil, ErrNotFound
		},
		accounts: func(context.Context, uuid.UUID, uuid.UUID) ([]accountdomain.Account, error) {
			return []accountdomain.Account{
				{ID: destinationID, Currency: "EUR", Balance: decimal.NewFromInt(10), Status: accountdomain.StatusBlocked},
				{ID: sourceID, Currency: "EUR", Balance: decimal.NewFromInt(100), Status: accountdomain.StatusActive},
			}, nil
		},
		update: func(_ context.Context, id uuid.UUID, balance decimal.Decimal) error {
			updates[id] = balance
			return nil
		},
		create: func(_ context.Context, input Transfer) (*Transfer, error) {
			input.ID = uuid.New()
			return &input, nil
		},
		createAccount: func(_ context.Context, input accountdomain.Transaction) error {
			accountTransactions = append(accountTransactions, input)
			return nil
		},
	}
	service, err := NewService(repository, euroGetter())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Create(t.Context(), CreateInput{
		SourceAccountID:      sourceID,
		DestinationAccountID: destinationID,
		Amount:               amount,
		IdempotencyKey:       key,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.Replay || result.Transfer == nil {
		t.Errorf("Create() result = %+v, want new transfer", result)
	}
	if !updates[sourceID].Equal(decimal.NewFromInt(75)) || !updates[destinationID].Equal(decimal.NewFromInt(35)) {
		t.Errorf("UpdateBalance() values = %+v, want source 75 and destination 35", updates)
	}
	if len(accountTransactions) != 2 {
		t.Fatalf("CreateAccountTransaction() calls = %d, want 2", len(accountTransactions))
	}
	for _, transaction := range accountTransactions {
		if transaction.Type != accountdomain.TransactionTypeTransfer || transaction.TransferID == nil {
			t.Errorf("account transaction = %+v, want linked transfer transaction", transaction)
		}
	}
	if findCalls != 2 {
		t.Errorf("FindByIdempotencyKey() calls = %d, want 2", findCalls)
	}
}

func TestServiceReplay(t *testing.T) {
	want := &Transfer{
		ID:                   uuid.New(),
		SourceAccountID:      uuid.New(),
		DestinationAccountID: uuid.New(),
		Amount:               decimal.NewFromInt(10),
	}
	repository := &fakeRepository{
		find: func(context.Context, uuid.UUID) (*Transfer, error) { return want, nil },
	}
	service, err := NewService(repository, euroGetter())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Create(t.Context(), CreateInput{
		SourceAccountID:      want.SourceAccountID,
		DestinationAccountID: want.DestinationAccountID,
		Amount:               want.Amount,
		IdempotencyKey:       uuid.New(),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !result.Replay || result.Transfer != want {
		t.Errorf("Create() result = %+v, want replay", result)
	}
}

func TestServiceRejectsInvalidTransfer(t *testing.T) {
	id := uuid.New()
	service, err := NewService(&fakeRepository{}, euroGetter())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Create(t.Context(), CreateInput{
		SourceAccountID:      id,
		DestinationAccountID: id,
		Amount:               decimal.NewFromInt(10),
		IdempotencyKey:       uuid.New(),
	})
	if !errors.Is(err, ErrSameAccount) {
		t.Errorf("Create() error = %v, want %v", err, ErrSameAccount)
	}
	if result != nil {
		t.Errorf("Create() result = %+v, want nil", result)
	}
}

func TestServiceRejectsInsufficientFunds(t *testing.T) {
	sourceID := uuid.New()
	destinationID := uuid.New()
	repository := &fakeRepository{
		find: func(context.Context, uuid.UUID) (*Transfer, error) { return nil, ErrNotFound },
		accounts: func(context.Context, uuid.UUID, uuid.UUID) ([]accountdomain.Account, error) {
			return []accountdomain.Account{
				{ID: sourceID, Currency: "EUR", Balance: decimal.NewFromInt(5), Status: accountdomain.StatusActive},
				{ID: destinationID, Currency: "EUR", Balance: decimal.Zero, Status: accountdomain.StatusActive},
			}, nil
		},
	}
	service, err := NewService(repository, euroGetter())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Create(t.Context(), CreateInput{
		SourceAccountID:      sourceID,
		DestinationAccountID: destinationID,
		Amount:               decimal.NewFromInt(10),
		IdempotencyKey:       uuid.New(),
	})
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("Create() error = %v, want %v", err, ErrInsufficientFunds)
	}
	if result != nil {
		t.Errorf("Create() result = %+v, want nil", result)
	}
}

func TestServiceRejectsIneligibleAccounts(t *testing.T) {
	tests := []struct {
		name              string
		sourceStatus      accountdomain.Status
		destinationStatus accountdomain.Status
		destinationCode   string
		target            error
	}{
		{
			name:              "blocked source",
			sourceStatus:      accountdomain.StatusBlocked,
			destinationStatus: accountdomain.StatusActive,
			destinationCode:   "EUR",
			target:            ErrSourceStatus,
		},
		{
			name:              "frozen destination",
			sourceStatus:      accountdomain.StatusActive,
			destinationStatus: accountdomain.StatusFrozen,
			destinationCode:   "EUR",
			target:            ErrDestinationStatus,
		},
		{
			name:              "different currencies",
			sourceStatus:      accountdomain.StatusActive,
			destinationStatus: accountdomain.StatusActive,
			destinationCode:   "USD",
			target:            ErrCurrencyMismatch,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceID := uuid.New()
			destinationID := uuid.New()
			repository := &fakeRepository{
				find: func(context.Context, uuid.UUID) (*Transfer, error) { return nil, ErrNotFound },
				accounts: func(context.Context, uuid.UUID, uuid.UUID) ([]accountdomain.Account, error) {
					return []accountdomain.Account{
						{ID: sourceID, Currency: "EUR", Balance: decimal.NewFromInt(100), Status: test.sourceStatus},
						{ID: destinationID, Currency: test.destinationCode, Balance: decimal.Zero, Status: test.destinationStatus},
					}, nil
				},
			}
			service, err := NewService(repository, euroGetter())
			if err != nil {
				t.Fatalf("NewService() error = %v", err)
			}

			result, err := service.Create(t.Context(), CreateInput{
				SourceAccountID:      sourceID,
				DestinationAccountID: destinationID,
				Amount:               decimal.NewFromInt(10),
				IdempotencyKey:       uuid.New(),
			})
			if !errors.Is(err, test.target) {
				t.Errorf("Create() error = %v, want %v", err, test.target)
			}
			if result != nil {
				t.Errorf("Create() result = %+v, want nil", result)
			}
		})
	}
}
