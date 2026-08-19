package account

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
)

type AccountWithdrawer interface {
	Withdraw(context.Context, accountdomain.MovementInput) (*accountdomain.MovementResult, error)
}

type withdrawAccountHandler struct {
	service AccountWithdrawer
}

func NewWithdrawAccountHandler(service AccountWithdrawer) (*withdrawAccountHandler, error) {
	if service == nil {
		return nil, errors.New("account withdrawer is required")
	}

	return &withdrawAccountHandler{service: service}, nil
}

func (h *withdrawAccountHandler) Handle(requestCtx fiber.Ctx) error {
	input, err := parseMovementInput(requestCtx)
	if err != nil {
		return err
	}
	if input == nil {
		return nil
	}

	result, err := h.service.Withdraw(requestCtx.Context(), *input)
	if err != nil {
		return writeAccountError(requestCtx, err)
	}

	status := fiber.StatusCreated
	if result.Replay {
		status = fiber.StatusOK
	}

	return requestCtx.Status(status).JSON(newTransactionResponse(*result.Transaction))
}
