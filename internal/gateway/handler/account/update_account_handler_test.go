package account

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
	"github.com/shopspring/decimal"
)

type accountStatusUpdaterFunc func(context.Context, accountdomain.UpdateStatusInput) (*accountdomain.Account, error)

func (f accountStatusUpdaterFunc) UpdateStatus(ctx context.Context, input accountdomain.UpdateStatusInput) (*accountdomain.Account, error) {
	return f(ctx, input)
}

func TestUpdateAccountHandler(t *testing.T) {
	id := uuid.New()
	want := accountdomain.Account{ID: id, Balance: decimal.Zero, Status: accountdomain.StatusBlocked}
	handler, err := NewUpdateAccountHandler(accountStatusUpdaterFunc(
		func(_ context.Context, input accountdomain.UpdateStatusInput) (*accountdomain.Account, error) {
			if input.ID != id || input.Status != accountdomain.StatusBlocked {
				t.Errorf("UpdateStatus() input = %+v, want id %v and status %s", input, id, accountdomain.StatusBlocked)
			}
			return &want, nil
		},
	))
	if err != nil {
		t.Fatalf("NewUpdateAccountHandler() error = %v", err)
	}

	app := fiber.New()
	app.Patch("/accounts/:account_id", handler.Handle)
	request := httptest.NewRequest(http.MethodPatch, "/accounts/"+id.String(), bytes.NewBufferString(
		`{"status":"blocked"}`,
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

func TestUpdateAccountHandlerInvalidTransition(t *testing.T) {
	handler, err := NewUpdateAccountHandler(accountStatusUpdaterFunc(
		func(context.Context, accountdomain.UpdateStatusInput) (*accountdomain.Account, error) {
			return nil, accountdomain.ErrInvalidStatusTransition
		},
	))
	if err != nil {
		t.Fatalf("NewUpdateAccountHandler() error = %v", err)
	}

	app := fiber.New()
	app.Patch("/accounts/:account_id", handler.Handle)
	request := httptest.NewRequest(http.MethodPatch, "/accounts/"+uuid.NewString(), bytes.NewBufferString(
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
