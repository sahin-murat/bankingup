package account

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidAmount       = errors.New("invalid account transaction amount")
	ErrAccountStatus       = errors.New("account status does not allow this transaction")
	ErrInsufficientFunds   = errors.New("account has insufficient funds")
	ErrIdempotencyConflict = errors.New("idempotency key was used for a different request")
)

type MovementInput struct {
	AccountID      uuid.UUID
	Amount         decimal.Decimal
	IdempotencyKey uuid.UUID
}

type MovementResult struct {
	Transaction *Transaction
	Replay      bool
}

type TransactionService interface {
	Deposit(context.Context, MovementInput) (*MovementResult, error)
	Withdraw(context.Context, MovementInput) (*MovementResult, error)
}

type transactionService struct {
	repository     TransactionRepository
	currencyGetter CurrencyGetter
}

func NewTransactionService(
	repository TransactionRepository,
	currencyGetter CurrencyGetter,
) (*transactionService, error) {
	if repository == nil {
		return nil, errors.New("account transaction repository is required")
	}
	if currencyGetter == nil {
		return nil, errors.New("currency getter is required")
	}

	return &transactionService{repository: repository, currencyGetter: currencyGetter}, nil
}

func (s *transactionService) Deposit(ctx context.Context, input MovementInput) (*MovementResult, error) {
	if err := validateMovementInput(input); err != nil {
		return nil, err
	}

	return s.executeMovement(
		ctx,
		input,
		TransactionTypeDeposit,
		validateDepositStatus,
		depositBalance,
	)
}

func (s *transactionService) Withdraw(ctx context.Context, input MovementInput) (*MovementResult, error) {
	if err := validateMovementInput(input); err != nil {
		return nil, err
	}

	return s.executeMovement(
		ctx,
		input,
		TransactionTypeWithdrawal,
		validateWithdrawalStatus,
		withdrawBalance,
	)
}

func (s *transactionService) executeMovement(
	ctx context.Context,
	input MovementInput,
	typeOfTransaction TransactionType,
	validateStatus func(Status) error,
	calculateBalance func(decimal.Decimal, decimal.Decimal) (decimal.Decimal, error),
) (*MovementResult, error) {
	if existing, err := s.repository.FindByIdempotencyKey(ctx, input.IdempotencyKey); err == nil {
		return movementReplay(existing, input, typeOfTransaction)
	} else if !errors.Is(err, ErrTransactionNotFound) {
		return nil, err
	}

	var result *MovementResult
	err := s.repository.WithinTransaction(ctx, func(store TransactionStore) error {
		lockedAccount, err := store.GetAccountForUpdate(ctx, input.AccountID)
		if err != nil {
			return err
		}

		existing, err := store.FindByIdempotencyKey(ctx, input.IdempotencyKey)
		if err == nil {
			result, err = movementReplay(existing, input, typeOfTransaction)
			return err
		}
		if !errors.Is(err, ErrTransactionNotFound) {
			return err
		}

		if err := validateStatus(lockedAccount.Status); err != nil {
			return err
		}

		currency, err := s.currencyGetter.GetByCode(ctx, lockedAccount.Currency)
		if err != nil {
			return fmt.Errorf("get account transaction currency: %w", err)
		}
		if err := validateAmount(input.Amount, currency.DecimalPlaces); err != nil {
			return err
		}

		balanceAfter, err := calculateBalance(lockedAccount.Balance, input.Amount)
		if err != nil {
			return err
		}

		if err := store.UpdateBalance(ctx, input.AccountID, balanceAfter); err != nil {
			return err
		}

		created, err := store.CreateTransaction(ctx, Transaction{
			AccountID:      input.AccountID,
			Type:           typeOfTransaction,
			Amount:         input.Amount,
			Currency:       lockedAccount.Currency,
			BalanceAfter:   balanceAfter,
			IdempotencyKey: input.IdempotencyKey,
		})
		if err != nil {
			return err
		}

		result = &MovementResult{Transaction: created}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrIdempotencyKeyAlreadyExists) {
			existing, findErr := s.repository.FindByIdempotencyKey(ctx, input.IdempotencyKey)
			if findErr != nil {
				return nil, err
			}
			return movementReplay(existing, input, typeOfTransaction)
		}

		return nil, err
	}

	return result, nil
}

func validateMovementInput(input MovementInput) error {
	if input.AccountID == uuid.Nil {
		return fmt.Errorf("%w: account_id is required", ErrInvalidInput)
	}
	if input.IdempotencyKey == uuid.Nil {
		return fmt.Errorf("%w: idempotency key is required", ErrInvalidInput)
	}
	if !input.Amount.IsPositive() || !input.Amount.LessThan(maximumBalance) {
		return fmt.Errorf("%w: amount must be positive and within the supported range", ErrInvalidAmount)
	}

	return nil
}

func validateDepositStatus(status Status) error {
	if status != StatusActive && status != StatusBlocked {
		return fmt.Errorf("%w: account status is %s", ErrAccountStatus, status)
	}

	return nil
}

func validateWithdrawalStatus(status Status) error {
	if status != StatusActive {
		return fmt.Errorf("%w: account status is %s", ErrAccountStatus, status)
	}

	return nil
}

func depositBalance(balance decimal.Decimal, amount decimal.Decimal) (decimal.Decimal, error) {
	balanceAfter := balance.Add(amount)
	if !balanceAfter.LessThan(maximumBalance) {
		return decimal.Zero, fmt.Errorf("%w: resulting balance is too large", ErrInvalidAmount)
	}

	return balanceAfter, nil
}

func withdrawBalance(balance decimal.Decimal, amount decimal.Decimal) (decimal.Decimal, error) {
	if balance.LessThan(amount) {
		return decimal.Zero, ErrInsufficientFunds
	}

	return balance.Sub(amount), nil
}

func movementReplay(
	existing *Transaction,
	input MovementInput,
	typeOfTransaction TransactionType,
) (*MovementResult, error) {
	if existing.AccountID != input.AccountID ||
		existing.Type != typeOfTransaction ||
		!existing.Amount.Equal(input.Amount) {
		return nil, ErrIdempotencyConflict
	}

	return &MovementResult{Transaction: existing, Replay: true}, nil
}

func validateAmount(amount decimal.Decimal, decimalPlaces int16) error {
	if amount.Exponent() < -int32(decimalPlaces) {
		return fmt.Errorf(
			"%w: amount allows at most %d decimal places",
			ErrInvalidAmount,
			decimalPlaces,
		)
	}

	return nil
}
