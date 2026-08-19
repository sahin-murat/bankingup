package customer

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
)

type customerCreatorFunc func(context.Context, customerdomain.CreateInput) (*customerdomain.Customer, error)

func (f customerCreatorFunc) Create(ctx context.Context, input customerdomain.CreateInput) (*customerdomain.Customer, error) {
	return f(ctx, input)
}

func TestCreateCustomerHandler(t *testing.T) {
	now := time.Now().UTC()
	wantInput := customerdomain.CreateInput{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "ada@example.com",
	}
	wantCustomer := customerdomain.Customer{
		ID:        uuid.New(),
		FirstName: wantInput.FirstName,
		LastName:  wantInput.LastName,
		Email:     wantInput.Email,
		Status:    customerdomain.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	handler, err := NewCreateCustomerHandler(customerCreatorFunc(
		func(_ context.Context, input customerdomain.CreateInput) (*customerdomain.Customer, error) {
			if input != wantInput {
				t.Errorf("Create() input = %+v, want %+v", input, wantInput)
			}
			return &wantCustomer, nil
		},
	))
	if err != nil {
		t.Fatalf("NewCreateCustomerHandler() error = %v", err)
	}

	app := fiber.New()
	app.Post("/customers", handler.Handle)
	request := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBufferString(
		`{"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}`,
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

	var body customerResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body != newCustomerResponse(wantCustomer) {
		t.Errorf("response = %+v, want %+v", body, newCustomerResponse(wantCustomer))
	}
}

func TestCreateCustomerHandlerDuplicateEmail(t *testing.T) {
	handler, err := NewCreateCustomerHandler(customerCreatorFunc(
		func(context.Context, customerdomain.CreateInput) (*customerdomain.Customer, error) {
			return nil, customerdomain.ErrEmailAlreadyExists
		},
	))
	if err != nil {
		t.Fatalf("NewCreateCustomerHandler() error = %v", err)
	}

	app := fiber.New()
	app.Post("/customers", handler.Handle)
	request := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBufferString(
		`{"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com"}`,
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
