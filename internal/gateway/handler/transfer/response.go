package transfer

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	transferdomain "github.com/sahin-murat/bankingup/internal/transfer"
)

type transferResponse struct {
	ID                      string    `json:"id"`
	SourceAccountID         string    `json:"source_account_id"`
	DestinationAccountID    string    `json:"destination_account_id"`
	Amount                  string    `json:"amount"`
	Currency                string    `json:"currency"`
	SourceBalanceAfter      string    `json:"source_balance_after"`
	DestinationBalanceAfter string    `json:"destination_balance_after"`
	CreatedAt               time.Time `json:"created_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func newTransferResponse(input transferdomain.Transfer) transferResponse {
	return transferResponse{
		ID:                      input.ID.String(),
		SourceAccountID:         input.SourceAccountID.String(),
		DestinationAccountID:    input.DestinationAccountID.String(),
		Amount:                  input.Amount.String(),
		Currency:                input.Currency,
		SourceBalanceAfter:      input.SourceBalanceAfter.String(),
		DestinationBalanceAfter: input.DestinationBalanceAfter.String(),
		CreatedAt:               input.CreatedAt,
	}
}

func writeTransferError(requestCtx fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, transferdomain.ErrInvalidInput),
		errors.Is(err, transferdomain.ErrInvalidAmount),
		errors.Is(err, transferdomain.ErrSameAccount):
		return requestCtx.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: err.Error()})
	case errors.Is(err, transferdomain.ErrAccountNotFound):
		return requestCtx.Status(fiber.StatusNotFound).JSON(errorResponse{Error: err.Error()})
	case errors.Is(err, transferdomain.ErrCurrencyMismatch),
		errors.Is(err, transferdomain.ErrSourceStatus),
		errors.Is(err, transferdomain.ErrDestinationStatus),
		errors.Is(err, transferdomain.ErrInsufficientFunds),
		errors.Is(err, transferdomain.ErrIdempotencyConflict):
		return requestCtx.Status(fiber.StatusConflict).JSON(errorResponse{Error: err.Error()})
	default:
		log.Errorf("transfer request failed: %v", err)
		return requestCtx.Status(fiber.StatusInternalServerError).JSON(errorResponse{Error: "internal server error"})
	}
}

func writeBadRequest(requestCtx fiber.Ctx, message string) error {
	return requestCtx.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: message})
}
