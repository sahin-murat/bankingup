package account

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
)

type accountWithdrawerFunc func(context.Context, accountdomain.MovementInput) (*accountdomain.MovementResult, error)

func (f accountWithdrawerFunc) Withdraw(
	ctx context.Context,
	input accountdomain.MovementInput,
) (*accountdomain.MovementResult, error) {
	return f(ctx, input)
}

func TestWithdrawAccountHandler(t *testing.T) {
	accountID := uuid.New()
	key := uuid.New()
	transaction := &accountdomain.Transaction{ID: uuid.New(), AccountID: accountID}
	handler, err := NewWithdrawAccountHandler(accountWithdrawerFunc(
		func(context.Context, accountdomain.MovementInput) (*accountdomain.MovementResult, error) {
			return &accountdomain.MovementResult{Transaction: transaction}, nil
		},
	))
	if err != nil {
		t.Fatalf("NewWithdrawAccountHandler() error = %v", err)
	}

	response := performMovementRequest(t, handler.Handle, "/accounts/"+accountID.String()+"/withdrawals", key, "10.00")
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusCreated {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusCreated)
	}
}

func TestWithdrawAccountHandlerInsufficientFunds(t *testing.T) {
	handler, err := NewWithdrawAccountHandler(accountWithdrawerFunc(
		func(context.Context, accountdomain.MovementInput) (*accountdomain.MovementResult, error) {
			return nil, accountdomain.ErrInsufficientFunds
		},
	))
	if err != nil {
		t.Fatalf("NewWithdrawAccountHandler() error = %v", err)
	}

	response := performMovementRequest(
		t,
		handler.Handle,
		"/accounts/"+uuid.NewString()+"/withdrawals",
		uuid.New(),
		"1000.00",
	)
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusConflict {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusConflict)
	}
}
