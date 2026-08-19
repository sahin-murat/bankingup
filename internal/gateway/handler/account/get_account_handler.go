package account

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
)

type AccountGetter interface {
	GetByID(context.Context, uuid.UUID) (*accountdomain.Account, error)
}

type getAccountHandler struct {
	service AccountGetter
}

func NewGetAccountHandler(service AccountGetter) (*getAccountHandler, error) {
	if service == nil {
		return nil, errors.New("account getter is required")
	}

	return &getAccountHandler{service: service}, nil
}

func (h *getAccountHandler) Handle(requestCtx fiber.Ctx) error {
	id, err := uuid.Parse(requestCtx.Params("account_id"))
	if err != nil {
		return writeBadRequest(requestCtx, "invalid account_id")
	}

	found, err := h.service.GetByID(requestCtx.Context(), id)
	if err != nil {
		return writeAccountError(requestCtx, err)
	}

	return requestCtx.Status(fiber.StatusOK).JSON(newAccountResponse(*found))
}
