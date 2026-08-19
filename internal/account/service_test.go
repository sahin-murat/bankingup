package account

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	currencydomain "github.com/sahin-murat/bankingup/internal/currency"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
	"github.com/shopspring/decimal"
)

type fakeRepository struct {
	create       func(context.Context, Account) (*Account, error)
	getByID      func(context.Context, uuid.UUID) (*Account, error)
	list         func(context.Context, *uuid.UUID) ([]Account, error)
	updateStatus func(context.Context, uuid.UUID, Status) (*Account, error)
}

func (r fakeRepository) Create(ctx context.Context, input Account) (*Account, error) {
	return r.create(ctx, input)
}

func (r fakeRepository) GetByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	return r.getByID(ctx, id)
}

func (r fakeRepository) List(ctx context.Context, customerID *uuid.UUID) ([]Account, error) {
	return r.list(ctx, customerID)
}

func (r fakeRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status Status) (*Account, error) {
	return r.updateStatus(ctx, id, status)
}

type customerGetterFunc func(context.Context, uuid.UUID) (*customerdomain.Customer, error)

func (f customerGetterFunc) GetByID(ctx context.Context, id uuid.UUID) (*customerdomain.Customer, error) {
	return f(ctx, id)
}

type currencyGetterFunc func(context.Context, string) (*currencydomain.Currency, error)

func (f currencyGetterFunc) GetByCode(ctx context.Context, code string) (*currencydomain.Currency, error) {
	return f(ctx, code)
}

func activeCustomerGetter() customerGetterFunc {
	return func(context.Context, uuid.UUID) (*customerdomain.Customer, error) {
		return &customerdomain.Customer{Status: customerdomain.StatusActive}, nil
	}
}

func euroGetter() currencyGetterFunc {
	return func(context.Context, string) (*currencydomain.Currency, error) {
		return &currencydomain.Currency{Code: "EUR", DecimalPlaces: 2}, nil
	}
}

func TestNewServiceRequiresDependencies(t *testing.T) {
	tests := []struct {
		name           string
		repository     Repository
		customerGetter CustomerGetter
		currencyGetter CurrencyGetter
	}{
		{
			name:           "repository",
			customerGetter: activeCustomerGetter(),
			currencyGetter: euroGetter(),
		},
		{
			name:           "customer getter",
			repository:     fakeRepository{},
			currencyGetter: euroGetter(),
		},
		{
			name:           "currency getter",
			repository:     fakeRepository{},
			customerGetter: activeCustomerGetter(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := NewService(test.repository, test.customerGetter, test.currencyGetter)
			if err == nil {
				t.Fatal("NewService() error = nil, want an error")
			}
			if service != nil {
				t.Errorf("NewService() service = %v, want nil", service)
			}
		})
	}
}

func TestServiceCreate(t *testing.T) {
	customerID := uuid.New()
	balance := decimal.RequireFromString("100.25")
	want := Account{
		CustomerID: customerID,
		Currency:   "EUR",
		Balance:    balance,
		Status:     StatusActive,
	}

	repository := fakeRepository{
		create: func(_ context.Context, input Account) (*Account, error) {
			if !accountsEqual(input, want) {
				t.Errorf("Create() input = %+v, want %+v", input, want)
			}
			return &input, nil
		},
	}
	service, err := NewService(repository, activeCustomerGetter(), euroGetter())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	created, err := service.Create(context.Background(), CreateInput{
		CustomerID:     customerID,
		Currency:       " eur ",
		InitialBalance: balance,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !accountsEqual(*created, want) {
		t.Errorf("Create() = %+v, want %+v", *created, want)
	}
}

func TestServiceCreateRequiresActiveCustomer(t *testing.T) {
	service, err := NewService(
		fakeRepository{},
		funcCustomer(customerdomain.StatusBlocked),
		euroGetter(),
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	created, err := service.Create(context.Background(), CreateInput{
		CustomerID: uuid.New(),
		Currency:   "EUR",
	})
	if !errors.Is(err, ErrCustomerNotActive) {
		t.Errorf("Create() error = %v, want %v", err, ErrCustomerNotActive)
	}
	if created != nil {
		t.Errorf("Create() = %+v, want nil", created)
	}
}

func TestServiceCreateRejectsInvalidBalances(t *testing.T) {
	tests := []struct {
		name    string
		balance decimal.Decimal
	}{
		{name: "negative", balance: decimal.RequireFromString("-0.01")},
		{name: "excess decimal places", balance: decimal.RequireFromString("1.001")},
		{name: "too large", balance: decimal.RequireFromString("1000000000000000")},
	}

	service, err := NewService(fakeRepository{}, activeCustomerGetter(), euroGetter())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			created, err := service.Create(context.Background(), CreateInput{
				CustomerID:     uuid.New(),
				Currency:       "EUR",
				InitialBalance: test.balance,
			})
			if !errors.Is(err, ErrInvalidBalance) {
				t.Errorf("Create() error = %v, want %v", err, ErrInvalidBalance)
			}
			if created != nil {
				t.Errorf("Create() = %+v, want nil", created)
			}
		})
	}
}

func TestServiceCreateRejectsUnsupportedCurrency(t *testing.T) {
	service, err := NewService(
		fakeRepository{},
		activeCustomerGetter(),
		currencyGetterFunc(func(context.Context, string) (*currencydomain.Currency, error) {
			return nil, currencydomain.ErrUnsupportedCurrency
		}),
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	created, err := service.Create(context.Background(), CreateInput{
		CustomerID: uuid.New(),
		Currency:   "CAD",
	})
	if !errors.Is(err, ErrUnsupportedCurrency) {
		t.Errorf("Create() error = %v, want %v", err, ErrUnsupportedCurrency)
	}
	if created != nil {
		t.Errorf("Create() = %+v, want nil", created)
	}
}

func TestServiceUpdateStatus(t *testing.T) {
	id := uuid.New()
	existing := Account{ID: id, Status: StatusActive, Balance: decimal.Zero}
	want := existing
	want.Status = StatusClosed

	repository := fakeRepository{
		getByID: func(context.Context, uuid.UUID) (*Account, error) {
			return &existing, nil
		},
		updateStatus: func(_ context.Context, requestedID uuid.UUID, status Status) (*Account, error) {
			if requestedID != id || status != StatusClosed {
				t.Errorf("UpdateStatus() = (%v, %s), want (%v, %s)", requestedID, status, id, StatusClosed)
			}
			return &want, nil
		},
	}
	service, err := NewService(repository, activeCustomerGetter(), euroGetter())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	updated, err := service.UpdateStatus(context.Background(), UpdateStatusInput{ID: id, Status: StatusClosed})
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if !accountsEqual(*updated, want) {
		t.Errorf("UpdateStatus() = %+v, want %+v", *updated, want)
	}
}

func TestServiceUpdateStatusRejectsInvalidChanges(t *testing.T) {
	tests := []struct {
		name     string
		existing Account
		status   Status
		target   error
	}{
		{
			name:     "closed account",
			existing: Account{Status: StatusClosed, Balance: decimal.Zero},
			status:   StatusActive,
			target:   ErrInvalidStatusTransition,
		},
		{
			name:     "non-zero balance",
			existing: Account{Status: StatusActive, Balance: decimal.NewFromInt(1)},
			status:   StatusClosed,
			target:   ErrNonZeroBalance,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := fakeRepository{
				getByID: func(context.Context, uuid.UUID) (*Account, error) {
					return &test.existing, nil
				},
			}
			service, err := NewService(repository, activeCustomerGetter(), euroGetter())
			if err != nil {
				t.Fatalf("NewService() error = %v", err)
			}

			updated, err := service.UpdateStatus(context.Background(), UpdateStatusInput{
				ID:     uuid.New(),
				Status: test.status,
			})
			if !errors.Is(err, test.target) {
				t.Errorf("UpdateStatus() error = %v, want %v", err, test.target)
			}
			if updated != nil {
				t.Errorf("UpdateStatus() = %+v, want nil", updated)
			}
		})
	}
}

func funcCustomer(status customerdomain.Status) customerGetterFunc {
	return func(context.Context, uuid.UUID) (*customerdomain.Customer, error) {
		return &customerdomain.Customer{Status: status}, nil
	}
}

func accountsEqual(left Account, right Account) bool {
	return left.ID == right.ID &&
		left.CustomerID == right.CustomerID &&
		left.Currency == right.Currency &&
		left.Balance.Equal(right.Balance) &&
		left.Status == right.Status &&
		left.CreatedAt.Equal(right.CreatedAt) &&
		left.UpdatedAt.Equal(right.UpdatedAt)
}
