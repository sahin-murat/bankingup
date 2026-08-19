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

type accountCreatorFunc func(context.Context, accountdomain.CreateInput) (*accountdomain.Account, error)

func (f accountCreatorFunc) Create(ctx context.Context, input accountdomain.CreateInput) (*accountdomain.Account, error) {
	return f(ctx, input)
}

func TestCreateAccountHandler(t *testing.T) {
	customerID := uuid.New()
	want := accountdomain.Account{
		ID:         uuid.New(),
		CustomerID: customerID,
		Currency:   "EUR",
		Balance:    decimal.RequireFromString("100.25"),
		Status:     accountdomain.StatusActive,
	}
	handler, err := NewCreateAccountHandler(accountCreatorFunc(
		func(_ context.Context, input accountdomain.CreateInput) (*accountdomain.Account, error) {
			if input.CustomerID != customerID || input.Currency != "EUR" || !input.InitialBalance.Equal(want.Balance) {
				t.Errorf("Create() input = %+v, want request values", input)
			}
			return &want, nil
		},
	))
	if err != nil {
		t.Fatalf("NewCreateAccountHandler() error = %v", err)
	}

	app := fiber.New()
	app.Post("/accounts", handler.Handle)
	request := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(
		`{"customer_id":"`+customerID.String()+`","currency":"EUR","initial_balance":"100.25"}`,
	))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusCreated)
	}

	var body accountResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body != newAccountResponse(want) {
		t.Errorf("response = %+v, want %+v", body, newAccountResponse(want))
	}
}

func TestCreateAccountHandlerRejectsInvalidBalance(t *testing.T) {
	handler, err := NewCreateAccountHandler(accountCreatorFunc(
		func(context.Context, accountdomain.CreateInput) (*accountdomain.Account, error) {
			t.Fatal("Create() called for invalid balance")
			return nil, nil
		},
	))
	if err != nil {
		t.Fatalf("NewCreateAccountHandler() error = %v", err)
	}

	app := fiber.New()
	app.Post("/accounts", handler.Handle)
	request := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(
		`{"customer_id":"`+uuid.NewString()+`","currency":"EUR","initial_balance":"not-a-number"}`,
	))
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

func TestCreateAccountHandlerInactiveCustomer(t *testing.T) {
	handler, err := NewCreateAccountHandler(accountCreatorFunc(
		func(context.Context, accountdomain.CreateInput) (*accountdomain.Account, error) {
			return nil, accountdomain.ErrCustomerNotActive
		},
	))
	if err != nil {
		t.Fatalf("NewCreateAccountHandler() error = %v", err)
	}

	app := fiber.New()
	app.Post("/accounts", handler.Handle)
	request := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(
		`{"customer_id":"`+uuid.NewString()+`","currency":"EUR"}`,
	))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusConflict {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusConflict)
	}
}
