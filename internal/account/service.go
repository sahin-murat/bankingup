package account

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	currencydomain "github.com/sahin-murat/bankingup/internal/currency"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidInput            = errors.New("invalid account input")
	ErrCustomerNotActive       = errors.New("account customer is not active")
	ErrUnsupportedCurrency     = errors.New("unsupported account currency")
	ErrInvalidStatusTransition = errors.New("invalid account status transition")
	ErrNonZeroBalance          = errors.New("account balance must be zero to close")
)

var maximumBalance = decimal.New(1, 15)

type CustomerGetter interface {
	GetByID(context.Context, uuid.UUID) (*customerdomain.Customer, error)
}

type CurrencyGetter interface {
	GetByCode(context.Context, string) (*currencydomain.Currency, error)
}

type CreateInput struct {
	CustomerID     uuid.UUID
	Currency       string
	InitialBalance decimal.Decimal
}

type UpdateStatusInput struct {
	ID     uuid.UUID
	Status Status
}

type Service interface {
	Create(context.Context, CreateInput) (*Account, error)
	GetByID(context.Context, uuid.UUID) (*Account, error)
	List(context.Context, *uuid.UUID) ([]Account, error)
	UpdateStatus(context.Context, UpdateStatusInput) (*Account, error)
}

type service struct {
	repository     Repository
	customerGetter CustomerGetter
	currencyGetter CurrencyGetter
}

func NewService(
	repository Repository,
	customerGetter CustomerGetter,
	currencyGetter CurrencyGetter,
) (*service, error) {
	if repository == nil {
		return nil, errors.New("account repository is required")
	}
	if customerGetter == nil {
		return nil, errors.New("customer getter is required")
	}
	if currencyGetter == nil {
		return nil, errors.New("currency getter is required")
	}

	return &service{
		repository:     repository,
		customerGetter: customerGetter,
		currencyGetter: currencyGetter,
	}, nil
}

func (s *service) Create(ctx context.Context, input CreateInput) (*Account, error) {
	if input.CustomerID == uuid.Nil {
		return nil, fmt.Errorf("%w: customer_id is required", ErrInvalidInput)
	}

	customer, err := s.customerGetter.GetByID(ctx, input.CustomerID)
	if err != nil {
		if errors.Is(err, customerdomain.ErrNotFound) {
			return nil, fmt.Errorf("create account: %w", ErrCustomerNotFound)
		}

		return nil, fmt.Errorf("get account customer: %w", err)
	}
	if customer.Status != customerdomain.StatusActive {
		return nil, fmt.Errorf("%w: customer status is %s", ErrCustomerNotActive, customer.Status)
	}

	currency, err := s.currencyGetter.GetByCode(ctx, input.Currency)
	if err != nil {
		if errors.Is(err, currencydomain.ErrInvalidCode) ||
			errors.Is(err, currencydomain.ErrUnsupportedCurrency) ||
			errors.Is(err, currencydomain.ErrNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedCurrency, input.Currency)
		}

		return nil, fmt.Errorf("get account currency: %w", err)
	}

	if err := validateBalance(input.InitialBalance, currency.DecimalPlaces); err != nil {
		return nil, err
	}

	account := Account{
		CustomerID: input.CustomerID,
		Currency:   currency.Code,
		Balance:    input.InitialBalance,
		Status:     StatusActive,
	}

	var openingTransaction *Transaction
	if !input.InitialBalance.IsZero() {
		openingTransaction = &Transaction{
			Type:           TransactionTypeDeposit,
			Amount:         input.InitialBalance,
			Currency:       currency.Code,
			BalanceAfter:   input.InitialBalance,
			IdempotencyKey: uuid.New(),
		}
	}

	return s.repository.Create(ctx, account, openingTransaction)
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: account_id is required", ErrInvalidInput)
	}

	return s.repository.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, customerID *uuid.UUID) ([]Account, error) {
	if customerID != nil && *customerID == uuid.Nil {
		return nil, fmt.Errorf("%w: customer_id is invalid", ErrInvalidInput)
	}

	return s.repository.List(ctx, customerID)
}

func (s *service) UpdateStatus(ctx context.Context, input UpdateStatusInput) (*Account, error) {
	if input.ID == uuid.Nil {
		return nil, fmt.Errorf("%w: account_id is required", ErrInvalidInput)
	}
	if !validStatus(input.Status) {
		return nil, fmt.Errorf("%w: unknown status %q", ErrInvalidInput, input.Status)
	}

	existing, err := s.repository.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if !canTransition(existing.Status, input.Status) {
		return nil, fmt.Errorf(
			"%w: %s to %s",
			ErrInvalidStatusTransition,
			existing.Status,
			input.Status,
		)
	}
	if input.Status == StatusClosed && !existing.Balance.IsZero() {
		return nil, ErrNonZeroBalance
	}

	return s.repository.UpdateStatus(
		ctx,
		input.ID,
		existing.Status,
		input.Status,
		input.Status == StatusClosed,
	)
}

func validateBalance(balance decimal.Decimal, decimalPlaces int16) error {
	if balance.IsNegative() {
		return fmt.Errorf("%w: initial_balance cannot be negative", ErrInvalidBalance)
	}
	if !balance.LessThan(maximumBalance) {
		return fmt.Errorf("%w: initial_balance is too large", ErrInvalidBalance)
	}

	allowedExponent := -int32(decimalPlaces)
	if balance.Exponent() < allowedExponent {
		return fmt.Errorf(
			"%w: initial_balance allows at most %d decimal places",
			ErrInvalidBalance,
			decimalPlaces,
		)
	}

	return nil
}

func validStatus(status Status) bool {
	switch status {
	case StatusActive, StatusBlocked, StatusFrozen, StatusClosed:
		return true
	default:
		return false
	}
}

func canTransition(from Status, to Status) bool {
	if from == to {
		return true
	}
	if from == StatusClosed {
		return false
	}

	return validStatus(from) && validStatus(to)
}
