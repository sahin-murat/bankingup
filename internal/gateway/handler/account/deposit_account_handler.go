package account

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
	"github.com/shopspring/decimal"
)

type AccountDepositor interface {
	Deposit(context.Context, accountdomain.MovementInput) (*accountdomain.MovementResult, error)
}

type depositAccountHandler struct {
	service AccountDepositor
}

type movementRequest struct {
	Amount string `json:"amount"`
}

func NewDepositAccountHandler(service AccountDepositor) (*depositAccountHandler, error) {
	if service == nil {
		return nil, errors.New("account depositor is required")
	}

	return &depositAccountHandler{service: service}, nil
}

func (h *depositAccountHandler) Handle(requestCtx fiber.Ctx) error {
	input, err := parseMovementInput(requestCtx)
	if err != nil {
		return err
	}
	if input == nil {
		return nil
	}

	result, err := h.service.Deposit(requestCtx.Context(), *input)
	if err != nil {
		return writeAccountError(requestCtx, err)
	}

	status := fiber.StatusCreated
	if result.Replay {
		status = fiber.StatusOK
	}

	return requestCtx.Status(status).JSON(newTransactionResponse(*result.Transaction))
}

func parseMovementInput(requestCtx fiber.Ctx) (*accountdomain.MovementInput, error) {
	accountID, err := uuid.Parse(requestCtx.Params("account_id"))
	if err != nil {
		return nil, writeBadRequest(requestCtx, "invalid account_id")
	}

	idempotencyKey, err := uuid.Parse(requestCtx.Get("Idempotency-Key"))
	if err != nil {
		return nil, writeBadRequest(requestCtx, "invalid Idempotency-Key")
	}

	var request movementRequest
	if err := requestCtx.Bind().Body(&request); err != nil {
		return nil, writeBadRequest(requestCtx, "invalid request body")
	}

	amount, err := decimal.NewFromString(request.Amount)
	if err != nil {
		return nil, writeBadRequest(requestCtx, "invalid amount")
	}

	return &accountdomain.MovementInput{
		AccountID:      accountID,
		Amount:         amount,
		IdempotencyKey: idempotencyKey,
	}, nil
}
