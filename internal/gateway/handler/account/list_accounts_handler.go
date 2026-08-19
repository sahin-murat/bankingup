package account

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
)

type AccountLister interface {
	List(context.Context, *uuid.UUID) ([]accountdomain.Account, error)
}

type listAccountsHandler struct {
	service AccountLister
}

func NewListAccountsHandler(service AccountLister) (*listAccountsHandler, error) {
	if service == nil {
		return nil, errors.New("account lister is required")
	}

	return &listAccountsHandler{service: service}, nil
}

func (h *listAccountsHandler) Handle(requestCtx fiber.Ctx) error {
	var customerID *uuid.UUID
	if value := requestCtx.Query("customer_id"); value != "" {
		parsed, err := uuid.Parse(value)
		if err != nil {
			return writeBadRequest(requestCtx, "invalid customer_id")
		}
		customerID = &parsed
	}

	accounts, err := h.service.List(requestCtx.Context(), customerID)
	if err != nil {
		return writeAccountError(requestCtx, err)
	}

	response := make([]accountResponse, len(accounts))
	for index, item := range accounts {
		response[index] = newAccountResponse(item)
	}

	return requestCtx.Status(fiber.StatusOK).JSON(response)
}
