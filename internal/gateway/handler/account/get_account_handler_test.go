package account

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
	"github.com/shopspring/decimal"
)

type accountGetterFunc func(context.Context, uuid.UUID) (*accountdomain.Account, error)

func (f accountGetterFunc) GetByID(ctx context.Context, id uuid.UUID) (*accountdomain.Account, error) {
	return f(ctx, id)
}

func TestGetAccountHandler(t *testing.T) {
	want := accountdomain.Account{ID: uuid.New(), Balance: decimal.Zero, Status: accountdomain.StatusActive}
	handler, err := NewGetAccountHandler(accountGetterFunc(
		func(_ context.Context, id uuid.UUID) (*accountdomain.Account, error) {
			if id != want.ID {
				t.Errorf("GetByID() id = %v, want %v", id, want.ID)
			}
			return &want, nil
		},
	))
	if err != nil {
		t.Fatalf("NewGetAccountHandler() error = %v", err)
	}

	app := fiber.New()
	app.Get("/accounts/:account_id", handler.Handle)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/accounts/"+want.ID.String(), http.NoBody))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
	}
}

func TestGetAccountHandlerErrors(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		serviceErr error
		wantStatus int
	}{
		{
			name:       "invalid account id",
			path:       "/accounts/not-a-uuid",
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "account not found",
			path:       "/accounts/" + uuid.NewString(),
			serviceErr: accountdomain.ErrNotFound,
			wantStatus: fiber.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler, err := NewGetAccountHandler(accountGetterFunc(
				func(context.Context, uuid.UUID) (*accountdomain.Account, error) {
					return nil, test.serviceErr
				},
			))
			if err != nil {
				t.Fatalf("NewGetAccountHandler() error = %v", err)
			}

			app := fiber.New()
			app.Get("/accounts/:account_id", handler.Handle)
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
