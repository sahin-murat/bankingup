package account

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
)

type AccountStatusUpdater interface {
	UpdateStatus(context.Context, accountdomain.UpdateStatusInput) (*accountdomain.Account, error)
}

type updateAccountHandler struct {
	service AccountStatusUpdater
}

type updateAccountRequest struct {
	Status *accountdomain.Status `json:"status"`
}

func NewUpdateAccountHandler(service AccountStatusUpdater) (*updateAccountHandler, error) {
	if service == nil {
		return nil, errors.New("account status updater is required")
	}

	return &updateAccountHandler{service: service}, nil
}

func (h *updateAccountHandler) Handle(requestCtx fiber.Ctx) error {
	id, err := uuid.Parse(requestCtx.Params("account_id"))
	if err != nil {
		return writeBadRequest(requestCtx, "invalid account_id")
	}

	var request updateAccountRequest
	if err := requestCtx.Bind().Body(&request); err != nil {
		return writeBadRequest(requestCtx, "invalid request body")
	}
	if request.Status == nil {
		return writeBadRequest(requestCtx, "status is required")
	}

	updated, err := h.service.UpdateStatus(requestCtx.Context(), accountdomain.UpdateStatusInput{
		ID:     id,
		Status: *request.Status,
	})
	if err != nil {
		return writeAccountError(requestCtx, err)
	}

	return requestCtx.Status(fiber.StatusOK).JSON(newAccountResponse(*updated))
}
