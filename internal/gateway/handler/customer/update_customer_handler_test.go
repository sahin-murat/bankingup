package customer

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
)

type customerUpdaterFunc func(context.Context, customerdomain.UpdateInput) (*customerdomain.Customer, error)

func (f customerUpdaterFunc) Update(ctx context.Context, input customerdomain.UpdateInput) (*customerdomain.Customer, error) {
	return f(ctx, input)
}

func TestUpdateCustomerHandler(t *testing.T) {
	id := uuid.New()
	want := customerdomain.Customer{ID: id, FirstName: "Ada", Status: customerdomain.StatusBlocked}
	handler, err := NewUpdateCustomerHandler(customerUpdaterFunc(
		func(_ context.Context, input customerdomain.UpdateInput) (*customerdomain.Customer, error) {
			if input.ID != id {
				t.Errorf("Update() id = %v, want %v", input.ID, id)
			}
			if input.FirstName == nil || *input.FirstName != "Ada" {
				t.Errorf("Update() first_name = %v, want Ada", input.FirstName)
			}
			if input.Status == nil || *input.Status != customerdomain.StatusBlocked {
				t.Errorf("Update() status = %v, want %v", input.Status, customerdomain.StatusBlocked)
			}
			return &want, nil
		},
	))
	if err != nil {
		t.Fatalf("NewUpdateCustomerHandler() error = %v", err)
	}

	app := fiber.New()
	app.Patch("/customers/:customer_id", handler.Handle)
	request := httptest.NewRequest(http.MethodPatch, "/customers/"+id.String(), bytes.NewBufferString(
		`{"first_name":"Ada","status":"blocked"}`,
	))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
	}
}

func TestUpdateCustomerHandlerInvalidTransition(t *testing.T) {
	id := uuid.New()
	handler, err := NewUpdateCustomerHandler(customerUpdaterFunc(
		func(context.Context, customerdomain.UpdateInput) (*customerdomain.Customer, error) {
			return nil, customerdomain.ErrInvalidStatusTransition
		},
	))
	if err != nil {
		t.Fatalf("NewUpdateCustomerHandler() error = %v", err)
	}

	app := fiber.New()
	app.Patch("/customers/:customer_id", handler.Handle)
	request := httptest.NewRequest(http.MethodPatch, "/customers/"+id.String(), bytes.NewBufferString(
		`{"status":"active"}`,
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
