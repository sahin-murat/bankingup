package account

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
	"github.com/shopspring/decimal"
)

type accountDepositorFunc func(context.Context, accountdomain.MovementInput) (*accountdomain.MovementResult, error)

func (f accountDepositorFunc) Deposit(
	ctx context.Context,
	input accountdomain.MovementInput,
) (*accountdomain.MovementResult, error) {
	return f(ctx, input)
}

func TestDepositAccountHandler(t *testing.T) {
	accountID := uuid.New()
	key := uuid.New()
	want := accountdomain.Transaction{
		ID:             uuid.New(),
		AccountID:      accountID,
		Type:           accountdomain.TransactionTypeDeposit,
		Amount:         decimal.RequireFromString("25.50"),
		Currency:       "EUR",
		BalanceAfter:   decimal.RequireFromString("125.50"),
		IdempotencyKey: key,
	}
	handler, err := NewDepositAccountHandler(accountDepositorFunc(
		func(_ context.Context, input accountdomain.MovementInput) (*accountdomain.MovementResult, error) {
			if input.AccountID != accountID || input.IdempotencyKey != key || !input.Amount.Equal(want.Amount) {
				t.Errorf("Deposit() input = %+v, want request values", input)
			}
			return &accountdomain.MovementResult{Transaction: &want}, nil
		},
	))
	if err != nil {
		t.Fatalf("NewDepositAccountHandler() error = %v", err)
	}

	response := performMovementRequest(t, handler.Handle, "/accounts/"+accountID.String()+"/deposits", key, "25.50")
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusCreated)
	}

	var body transactionResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body != newTransactionResponse(want) {
		t.Errorf("response = %+v, want %+v", body, newTransactionResponse(want))
	}
}

func TestDepositAccountHandlerReplay(t *testing.T) {
	accountID := uuid.New()
	key := uuid.New()
	transaction := &accountdomain.Transaction{ID: uuid.New(), AccountID: accountID}
	handler, err := NewDepositAccountHandler(accountDepositorFunc(
		func(context.Context, accountdomain.MovementInput) (*accountdomain.MovementResult, error) {
			return &accountdomain.MovementResult{Transaction: transaction, Replay: true}, nil
		},
	))
	if err != nil {
		t.Fatalf("NewDepositAccountHandler() error = %v", err)
	}

	response := performMovementRequest(t, handler.Handle, "/accounts/"+accountID.String()+"/deposits", key, "10.00")
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusOK {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
	}
}

func TestDepositAccountHandlerRejectsInvalidIdempotencyKey(t *testing.T) {
	handler, err := NewDepositAccountHandler(accountDepositorFunc(
		func(context.Context, accountdomain.MovementInput) (*accountdomain.MovementResult, error) {
			t.Fatal("Deposit() called for invalid idempotency key")
			return nil, nil
		},
	))
	if err != nil {
		t.Fatalf("NewDepositAccountHandler() error = %v", err)
	}

	app := fiber.New()
	app.Post("/accounts/:account_id/deposits", handler.Handle)
	request := httptest.NewRequest(
		http.MethodPost,
		"/accounts/"+uuid.NewString()+"/deposits",
		bytes.NewBufferString(`{"amount":"10.00"}`),
	)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusBadRequest)
	}
}

func performMovementRequest(
	t *testing.T,
	handler fiber.Handler,
	path string,
	key uuid.UUID,
	amount string,
) *http.Response {
	t.Helper()

	app := fiber.New()
	app.Post("/accounts/:account_id/:operation", handler)
	request := httptest.NewRequest(
		http.MethodPost,
		path,
		bytes.NewBufferString(`{"amount":"`+amount+`"}`),
	)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	request.Header.Set("Idempotency-Key", key.String())
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}

	return response
}
