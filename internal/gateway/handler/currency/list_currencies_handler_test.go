package currency

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	currencydomain "github.com/sahin-murat/bankingup/internal/currency"
)

type currencyListerFunc func(context.Context) ([]currencydomain.Currency, error)

func (f currencyListerFunc) List(ctx context.Context) ([]currencydomain.Currency, error) {
	return f(ctx)
}

func TestListCurrenciesHandler(t *testing.T) {
	want := []currencydomain.Currency{
		{Code: "EUR", Name: "Euro", DecimalPlaces: 2},
		{Code: "JPY", Name: "Japanese Yen", DecimalPlaces: 0},
	}
	handler, err := NewListCurrenciesHandler(currencyListerFunc(
		func(context.Context) ([]currencydomain.Currency, error) {
			return want, nil
		},
	))
	if err != nil {
		t.Fatalf("NewListCurrenciesHandler() error = %v", err)
	}

	app := fiber.New()
	app.Get("/currencies", handler.Handle)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/currencies", http.NoBody))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
	}

	var body []currencyResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(body) != len(want) {
		t.Fatalf("response length = %d, want %d", len(body), len(want))
	}
	for index := range want {
		if body[index] != newCurrencyResponse(want[index]) {
			t.Errorf("response[%d] = %+v, want %+v", index, body[index], newCurrencyResponse(want[index]))
		}
	}
}

func TestListCurrenciesHandlerUnexpectedError(t *testing.T) {
	handler, err := NewListCurrenciesHandler(currencyListerFunc(
		func(context.Context) ([]currencydomain.Currency, error) {
			return nil, errors.New("database failure")
		},
	))
	if err != nil {
		t.Fatalf("NewListCurrenciesHandler() error = %v", err)
	}

	app := fiber.New()
	app.Get("/currencies", handler.Handle)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/currencies", http.NoBody))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusInternalServerError)
	}
}
