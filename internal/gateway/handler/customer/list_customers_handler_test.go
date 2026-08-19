package customer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
)

type customerListerFunc func(context.Context) ([]customerdomain.Customer, error)

func (f customerListerFunc) List(ctx context.Context) ([]customerdomain.Customer, error) {
	return f(ctx)
}

func TestListCustomersHandler(t *testing.T) {
	want := []customerdomain.Customer{{ID: uuid.New(), Status: customerdomain.StatusActive}}
	handler, err := NewListCustomersHandler(customerListerFunc(
		func(context.Context) ([]customerdomain.Customer, error) {
			return want, nil
		},
	))
	if err != nil {
		t.Fatalf("NewListCustomersHandler() error = %v", err)
	}

	app := fiber.New()
	app.Get("/customers", handler.Handle)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/customers", http.NoBody))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
	}

	var body []customerResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(body) != 1 || body[0] != newCustomerResponse(want[0]) {
		t.Errorf("response = %+v, want customer %+v", body, newCustomerResponse(want[0]))
	}
}

func TestListCustomersHandlerUnexpectedError(t *testing.T) {
	handler, err := NewListCustomersHandler(customerListerFunc(
		func(context.Context) ([]customerdomain.Customer, error) {
			return nil, errors.New("database failure")
		},
	))
	if err != nil {
		t.Fatalf("NewListCustomersHandler() error = %v", err)
	}

	app := fiber.New()
	app.Get("/customers", handler.Handle)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/customers", http.NoBody))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusInternalServerError)
	}
}
