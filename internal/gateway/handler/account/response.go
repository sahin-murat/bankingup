package account

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
)

type accountResponse struct {
	ID         string               `json:"id"`
	CustomerID string               `json:"customer_id"`
	Currency   string               `json:"currency"`
	Balance    string               `json:"balance"`
	Status     accountdomain.Status `json:"status"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
}

type transactionResponse struct {
	ID           string                        `json:"id"`
	AccountID    string                        `json:"account_id"`
	Type         accountdomain.TransactionType `json:"type"`
	Amount       string                        `json:"amount"`
	Currency     string                        `json:"currency"`
	BalanceAfter string                        `json:"balance_after"`
	CreatedAt    time.Time                     `json:"created_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func newAccountResponse(input accountdomain.Account) accountResponse {
	return accountResponse{
		ID:         input.ID.String(),
		CustomerID: input.CustomerID.String(),
		Currency:   input.Currency,
		Balance:    input.Balance.String(),
		Status:     input.Status,
		CreatedAt:  input.CreatedAt,
		UpdatedAt:  input.UpdatedAt,
	}
}

func newTransactionResponse(input accountdomain.Transaction) transactionResponse {
	return transactionResponse{
		ID:           input.ID.String(),
		AccountID:    input.AccountID.String(),
		Type:         input.Type,
		Amount:       input.Amount.String(),
		Currency:     input.Currency,
		BalanceAfter: input.BalanceAfter.String(),
		CreatedAt:    input.CreatedAt,
	}
}

func writeAccountError(requestCtx fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accountdomain.ErrInvalidInput),
		errors.Is(err, accountdomain.ErrInvalidBalance),
		errors.Is(err, accountdomain.ErrCurrencyNotFound),
		errors.Is(err, accountdomain.ErrUnsupportedCurrency),
		errors.Is(err, accountdomain.ErrInvalidStatusTransition),
		errors.Is(err, accountdomain.ErrNonZeroBalance),
		errors.Is(err, accountdomain.ErrInvalidAmount):
		return requestCtx.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: err.Error()})
	case errors.Is(err, accountdomain.ErrNotFound),
		errors.Is(err, accountdomain.ErrCustomerNotFound):
		return requestCtx.Status(fiber.StatusNotFound).JSON(errorResponse{Error: err.Error()})
	case errors.Is(err, accountdomain.ErrCustomerNotActive),
		errors.Is(err, accountdomain.ErrAccountStatus),
		errors.Is(err, accountdomain.ErrInsufficientFunds),
		errors.Is(err, accountdomain.ErrIdempotencyConflict),
		errors.Is(err, accountdomain.ErrConcurrentUpdate):
		return requestCtx.Status(fiber.StatusConflict).JSON(errorResponse{Error: err.Error()})
	default:
		log.Errorf("account request failed: %v", err)
		return requestCtx.Status(fiber.StatusInternalServerError).JSON(errorResponse{Error: "internal server error"})
	}
}

func writeBadRequest(requestCtx fiber.Ctx, message string) error {
	return requestCtx.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: message})
}
