package transfer

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
	currencydomain "github.com/sahin-murat/bankingup/internal/currency"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidInput        = errors.New("invalid transfer input")
	ErrInvalidAmount       = errors.New("invalid transfer amount")
	ErrSameAccount         = errors.New("transfer accounts must be different")
	ErrCurrencyMismatch    = errors.New("transfer account currencies do not match")
	ErrSourceStatus        = errors.New("source account status does not allow transfers")
	ErrDestinationStatus   = errors.New("destination account status does not allow transfers")
	ErrInsufficientFunds   = errors.New("source account has insufficient funds")
	ErrIdempotencyConflict = errors.New("idempotency key was used for a different transfer")
)

var maximumAmount = decimal.New(1, 15)

type CurrencyGetter interface {
	GetByCode(context.Context, string) (*currencydomain.Currency, error)
}

type CreateInput struct {
	SourceAccountID      uuid.UUID
	DestinationAccountID uuid.UUID
	Amount               decimal.Decimal
	IdempotencyKey       uuid.UUID
}

type CreateResult struct {
	Transfer *Transfer
	Replay   bool
}

type Service interface {
	Create(context.Context, CreateInput) (*CreateResult, error)
}

type service struct {
	repository     Repository
	currencyGetter CurrencyGetter
}

func NewService(repository Repository, currencyGetter CurrencyGetter) (*service, error) {
	if repository == nil {
		return nil, errors.New("transfer repository is required")
	}
	if currencyGetter == nil {
		return nil, errors.New("currency getter is required")
	}

	return &service{repository: repository, currencyGetter: currencyGetter}, nil
}

func (s *service) Create(ctx context.Context, input CreateInput) (*CreateResult, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}

	if existing, err := s.repository.FindByIdempotencyKey(ctx, input.IdempotencyKey); err == nil {
		return replay(existing, input)
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	var result *CreateResult
	err := s.repository.WithinTransaction(ctx, func(store Store) error {
		accounts, err := store.GetAccountsForUpdate(
			ctx,
			input.SourceAccountID,
			input.DestinationAccountID,
		)
		if err != nil {
			return err
		}

		existing, err := store.FindByIdempotencyKey(ctx, input.IdempotencyKey)
		if err == nil {
			result, err = replay(existing, input)
			return err
		}
		if !errors.Is(err, ErrNotFound) {
			return err
		}

		source, destination := transferAccounts(accounts, input)
		if source == nil || destination == nil {
			return ErrAccountNotFound
		}
		if source.Status != accountdomain.StatusActive {
			return fmt.Errorf("%w: source status is %s", ErrSourceStatus, source.Status)
		}
		if destination.Status != accountdomain.StatusActive &&
			destination.Status != accountdomain.StatusBlocked {
			return fmt.Errorf("%w: destination status is %s", ErrDestinationStatus, destination.Status)
		}
		if source.Currency != destination.Currency {
			return ErrCurrencyMismatch
		}

		currency, err := s.currencyGetter.GetByCode(ctx, source.Currency)
		if err != nil {
			return fmt.Errorf("get transfer currency: %w", err)
		}
		if input.Amount.Exponent() < -int32(currency.DecimalPlaces) {
			return fmt.Errorf(
				"%w: amount allows at most %d decimal places",
				ErrInvalidAmount,
				currency.DecimalPlaces,
			)
		}
		if source.Balance.LessThan(input.Amount) {
			return ErrInsufficientFunds
		}

		sourceBalanceAfter := source.Balance.Sub(input.Amount)
		destinationBalanceAfter := destination.Balance.Add(input.Amount)
		if !destinationBalanceAfter.LessThan(maximumAmount) {
			return fmt.Errorf("%w: destination balance is too large", ErrInvalidAmount)
		}

		if err := store.UpdateBalance(ctx, source.ID, sourceBalanceAfter); err != nil {
			return err
		}
		if err := store.UpdateBalance(ctx, destination.ID, destinationBalanceAfter); err != nil {
			return err
		}

		created, err := store.CreateTransfer(ctx, Transfer{
			SourceAccountID:         source.ID,
			DestinationAccountID:    destination.ID,
			Amount:                  input.Amount,
			Currency:                source.Currency,
			SourceBalanceAfter:      sourceBalanceAfter,
			DestinationBalanceAfter: destinationBalanceAfter,
			IdempotencyKey:          input.IdempotencyKey,
		})
		if err != nil {
			return err
		}

		if err := createAccountTransactions(ctx, store, created); err != nil {
			return err
		}

		result = &CreateResult{Transfer: created}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrIdempotencyKeyAlreadyExists) {
			existing, findErr := s.repository.FindByIdempotencyKey(ctx, input.IdempotencyKey)
			if findErr != nil {
				return nil, err
			}
			return replay(existing, input)
		}

		return nil, err
	}

	return result, nil
}

func validateInput(input CreateInput) error {
	if input.SourceAccountID == uuid.Nil {
		return fmt.Errorf("%w: source_account_id is required", ErrInvalidInput)
	}
	if input.DestinationAccountID == uuid.Nil {
		return fmt.Errorf("%w: destination_account_id is required", ErrInvalidInput)
	}
	if input.SourceAccountID == input.DestinationAccountID {
		return ErrSameAccount
	}
	if input.IdempotencyKey == uuid.Nil {
		return fmt.Errorf("%w: idempotency key is required", ErrInvalidInput)
	}
	if !input.Amount.IsPositive() || !input.Amount.LessThan(maximumAmount) {
		return fmt.Errorf("%w: amount must be positive and within the supported range", ErrInvalidAmount)
	}

	return nil
}

func transferAccounts(
	accounts []accountdomain.Account,
	input CreateInput,
) (*accountdomain.Account, *accountdomain.Account) {
	var source *accountdomain.Account
	var destination *accountdomain.Account
	for index := range accounts {
		switch accounts[index].ID {
		case input.SourceAccountID:
			source = &accounts[index]
		case input.DestinationAccountID:
			destination = &accounts[index]
		}
	}

	return source, destination
}

func createAccountTransactions(ctx context.Context, store Store, transfer *Transfer) error {
	transferID := transfer.ID
	transactions := []accountdomain.Transaction{
		{
			AccountID:      transfer.SourceAccountID,
			Type:           accountdomain.TransactionTypeTransfer,
			Amount:         transfer.Amount,
			Currency:       transfer.Currency,
			BalanceAfter:   transfer.SourceBalanceAfter,
			IdempotencyKey: uuid.New(),
			TransferID:     &transferID,
		},
		{
			AccountID:      transfer.DestinationAccountID,
			Type:           accountdomain.TransactionTypeTransfer,
			Amount:         transfer.Amount,
			Currency:       transfer.Currency,
			BalanceAfter:   transfer.DestinationBalanceAfter,
			IdempotencyKey: uuid.New(),
			TransferID:     &transferID,
		},
	}

	for _, transaction := range transactions {
		if err := store.CreateAccountTransaction(ctx, transaction); err != nil {
			return err
		}
	}

	return nil
}

func replay(existing *Transfer, input CreateInput) (*CreateResult, error) {
	if existing.SourceAccountID != input.SourceAccountID ||
		existing.DestinationAccountID != input.DestinationAccountID ||
		!existing.Amount.Equal(input.Amount) {
		return nil, ErrIdempotencyConflict
	}

	return &CreateResult{Transfer: existing, Replay: true}, nil
}
