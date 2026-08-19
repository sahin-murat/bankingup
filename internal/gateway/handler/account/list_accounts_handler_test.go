package account

import (
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

type accountListerFunc func(context.Context, *uuid.UUID) ([]accountdomain.Account, error)

func (f accountListerFunc) List(ctx context.Context, customerID *uuid.UUID) ([]accountdomain.Account, error) {
	return f(ctx, customerID)
}

func TestListAccountsHandler(t *testing.T) {
	customerID := uuid.New()
	want := []accountdomain.Account{{
		ID:         uuid.New(),
		CustomerID: customerID,
		Currency:   "EUR",
		Balance:    decimal.Zero,
		Status:     accountdomain.StatusActive,
	}}
	handler, err := NewListAccountsHandler(accountListerFunc(
		func(_ context.Context, filter *uuid.UUID) ([]accountdomain.Account, error) {
			if filter == nil || *filter != customerID {
				t.Errorf("List() customerID = %v, want %v", filter, customerID)
			}
			return want, nil
		},
	))
	if err != nil {
		t.Fatalf("NewListAccountsHandler() error = %v", err)
	}

	app := fiber.New()
	app.Get("/accounts", handler.Handle)
	request := httptest.NewRequest(http.MethodGet, "/accounts?customer_id="+customerID.String(), http.NoBody)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
	}

	var body []accountResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(body) != 1 || body[0] != newAccountResponse(want[0]) {
		t.Errorf("response = %+v, want account %+v", body, newAccountResponse(want[0]))
	}
}

func TestListAccountsHandlerRejectsInvalidCustomerID(t *testing.T) {
	handler, err := NewListAccountsHandler(accountListerFunc(
		func(context.Context, *uuid.UUID) ([]accountdomain.Account, error) {
			t.Fatal("List() called for invalid customer_id")
			return nil, nil
		},
	))
	if err != nil {
		t.Fatalf("NewListAccountsHandler() error = %v", err)
	}

	app := fiber.New()
	app.Get("/accounts", handler.Handle)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/accounts?customer_id=invalid", http.NoBody))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusBadRequest)
	}
}
