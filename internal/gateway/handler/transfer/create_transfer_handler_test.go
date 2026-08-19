package transfer

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	transferdomain "github.com/sahin-murat/bankingup/internal/transfer"
	"github.com/shopspring/decimal"
)

type transferCreatorFunc func(context.Context, transferdomain.CreateInput) (*transferdomain.CreateResult, error)

func (f transferCreatorFunc) Create(
	ctx context.Context,
	input transferdomain.CreateInput,
) (*transferdomain.CreateResult, error) {
	return f(ctx, input)
}

func TestCreateTransferHandler(t *testing.T) {
	sourceID := uuid.New()
	destinationID := uuid.New()
	key := uuid.New()
	want := transferdomain.Transfer{
		ID:                      uuid.New(),
		SourceAccountID:         sourceID,
		DestinationAccountID:    destinationID,
		Amount:                  decimal.NewFromInt(15),
		Currency:                "EUR",
		SourceBalanceAfter:      decimal.RequireFromString("100.75"),
		DestinationBalanceAfter: decimal.NewFromInt(15),
	}
	handler, err := NewCreateTransferHandler(transferCreatorFunc(
		func(_ context.Context, input transferdomain.CreateInput) (*transferdomain.CreateResult, error) {
			if input.SourceAccountID != sourceID || input.DestinationAccountID != destinationID ||
				input.IdempotencyKey != key || !input.Amount.Equal(want.Amount) {
				t.Errorf("Create() input = %+v, want request values", input)
			}
			return &transferdomain.CreateResult{Transfer: &want}, nil
		},
	))
	if err != nil {
		t.Fatalf("NewCreateTransferHandler() error = %v", err)
	}

	response := performRequest(t, handler.Handle, sourceID, destinationID, key)
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusCreated)
	}

	var body transferResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body != newTransferResponse(want) {
		t.Errorf("response = %+v, want %+v", body, newTransferResponse(want))
	}
}

func TestCreateTransferHandlerReplay(t *testing.T) {
	transfer := &transferdomain.Transfer{ID: uuid.New()}
	handler, err := NewCreateTransferHandler(transferCreatorFunc(
		func(context.Context, transferdomain.CreateInput) (*transferdomain.CreateResult, error) {
			return &transferdomain.CreateResult{Transfer: transfer, Replay: true}, nil
		},
	))
	if err != nil {
		t.Fatalf("NewCreateTransferHandler() error = %v", err)
	}

	response := performRequest(t, handler.Handle, uuid.New(), uuid.New(), uuid.New())
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusOK {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
	}
}

func TestCreateTransferHandlerInsufficientFunds(t *testing.T) {
	handler, err := NewCreateTransferHandler(transferCreatorFunc(
		func(context.Context, transferdomain.CreateInput) (*transferdomain.CreateResult, error) {
			return nil, transferdomain.ErrInsufficientFunds
		},
	))
	if err != nil {
		t.Fatalf("NewCreateTransferHandler() error = %v", err)
	}

	response := performRequest(t, handler.Handle, uuid.New(), uuid.New(), uuid.New())
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusConflict {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusConflict)
	}
}

func performRequest(
	t *testing.T,
	handler fiber.Handler,
	sourceID uuid.UUID,
	destinationID uuid.UUID,
	key uuid.UUID,
) *http.Response {
	t.Helper()

	app := fiber.New()
	app.Post("/transfers", handler)
	request := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBufferString(
		`{"source_account_id":"`+sourceID.String()+`","destination_account_id":"`+
			destinationID.String()+`","amount":"15.00"}`,
	))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	request.Header.Set("Idempotency-Key", key.String())
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}

	return response
}
