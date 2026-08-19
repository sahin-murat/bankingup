package currency

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	currencydomain "github.com/sahin-murat/bankingup/internal/currency"
)

type currencyCreatorFunc func(context.Context, currencydomain.CreateInput) (*currencydomain.Currency, error)

func (f currencyCreatorFunc) Create(ctx context.Context, input currencydomain.CreateInput) (*currencydomain.Currency, error) {
	return f(ctx, input)
}

func TestCreateCurrencyHandler(t *testing.T) {
	want := currencydomain.Currency{Code: "CAD", Name: "Canadian Dollar", DecimalPlaces: 2}
	handler, err := NewCreateCurrencyHandler(currencyCreatorFunc(
		func(_ context.Context, input currencydomain.CreateInput) (*currencydomain.Currency, error) {
			if input.Code != "cad" || input.Name != "Canadian Dollar" || input.DecimalPlaces != 2 {
				t.Errorf("Create() input = %+v, want request values", input)
			}
			return &want, nil
		},
	))
	if err != nil {
		t.Fatalf("NewCreateCurrencyHandler() error = %v", err)
	}

	app := fiber.New()
	app.Post("/currencies", handler.Handle)
	request := httptest.NewRequest(http.MethodPost, "/currencies", bytes.NewBufferString(
		`{"code":"cad","name":"Canadian Dollar","decimal_places":2}`,
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

	var body currencyResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body != newCurrencyResponse(want) {
		t.Errorf("response = %+v, want %+v", body, newCurrencyResponse(want))
	}
}

func TestCreateCurrencyHandlerRequiresDecimalPlaces(t *testing.T) {
	handler, err := NewCreateCurrencyHandler(currencyCreatorFunc(
		func(context.Context, currencydomain.CreateInput) (*currencydomain.Currency, error) {
			t.Fatal("Create() called for an invalid request")
			return nil, nil
		},
	))
	if err != nil {
		t.Fatalf("NewCreateCurrencyHandler() error = %v", err)
	}

	app := fiber.New()
	app.Post("/currencies", handler.Handle)
	request := httptest.NewRequest(http.MethodPost, "/currencies", bytes.NewBufferString(
		`{"code":"JPY","name":"Japanese Yen"}`,
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

func TestCreateCurrencyHandlerDuplicateCode(t *testing.T) {
	handler, err := NewCreateCurrencyHandler(currencyCreatorFunc(
		func(context.Context, currencydomain.CreateInput) (*currencydomain.Currency, error) {
			return nil, currencydomain.ErrCodeAlreadyExists
		},
	))
	if err != nil {
		t.Fatalf("NewCreateCurrencyHandler() error = %v", err)
	}

	app := fiber.New()
	app.Post("/currencies", handler.Handle)
	request := httptest.NewRequest(http.MethodPost, "/currencies", bytes.NewBufferString(
		`{"code":"EUR","name":"Euro","decimal_places":2}`,
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

func TestCreateCurrencyHandlerUnexpectedError(t *testing.T) {
	handler, err := NewCreateCurrencyHandler(currencyCreatorFunc(
		func(context.Context, currencydomain.CreateInput) (*currencydomain.Currency, error) {
			return nil, errors.New("database failure")
		},
	))
	if err != nil {
		t.Fatalf("NewCreateCurrencyHandler() error = %v", err)
	}

	app := fiber.New()
	app.Post("/currencies", handler.Handle)
	request := httptest.NewRequest(http.MethodPost, "/currencies", bytes.NewBufferString(
		`{"code":"CAD","name":"Canadian Dollar","decimal_places":2}`,
	))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusInternalServerError)
	}
}
