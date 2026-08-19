package customer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
)

type customerGetterFunc func(context.Context, uuid.UUID) (*customerdomain.Customer, error)

func (f customerGetterFunc) GetByID(ctx context.Context, id uuid.UUID) (*customerdomain.Customer, error) {
	return f(ctx, id)
}

func TestGetCustomerHandler(t *testing.T) {
	id := uuid.New()
	want := customerdomain.Customer{ID: id, Status: customerdomain.StatusActive}
	handler, err := NewGetCustomerHandler(customerGetterFunc(
		func(_ context.Context, requestedID uuid.UUID) (*customerdomain.Customer, error) {
			if requestedID != id {
				t.Errorf("GetByID() id = %v, want %v", requestedID, id)
			}
			return &want, nil
		},
	))
	if err != nil {
		t.Fatalf("NewGetCustomerHandler() error = %v", err)
	}

	app := fiber.New()
	app.Get("/customers/:customer_id", handler.Handle)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/customers/"+id.String(), http.NoBody))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
	}
}

func TestGetCustomerHandlerErrors(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		service    CustomerGetter
		wantStatus int
	}{
		{
			name: "invalid id",
			path: "/customers/not-a-uuid",
			service: customerGetterFunc(func(context.Context, uuid.UUID) (*customerdomain.Customer, error) {
				t.Fatal("GetByID() should not be called")
				return nil, nil
			}),
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "not found",
			path: "/customers/" + uuid.NewString(),
			service: customerGetterFunc(func(context.Context, uuid.UUID) (*customerdomain.Customer, error) {
				return nil, customerdomain.ErrNotFound
			}),
			wantStatus: fiber.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler, err := NewGetCustomerHandler(test.service)
			if err != nil {
				t.Fatalf("NewGetCustomerHandler() error = %v", err)
			}

			app := fiber.New()
			app.Get("/customers/:customer_id", handler.Handle)
			response, err := app.Test(httptest.NewRequest(http.MethodGet, test.path, http.NoBody))
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != test.wantStatus {
				t.Errorf("status = %d, want %d", response.StatusCode, test.wantStatus)
			}
		})
	}
}
