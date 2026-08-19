package customer

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
)

type CustomerGetter interface {
	GetByID(context.Context, uuid.UUID) (*customerdomain.Customer, error)
}

type getCustomerHandler struct {
	service CustomerGetter
}

func NewGetCustomerHandler(service CustomerGetter) (*getCustomerHandler, error) {
	if service == nil {
		return nil, errors.New("customer getter is required")
	}

	return &getCustomerHandler{service: service}, nil
}

func (h *getCustomerHandler) Handle(requestCtx fiber.Ctx) error {
	id, err := uuid.Parse(requestCtx.Params("customer_id"))
	if err != nil {
		return writeBadRequest(requestCtx, "invalid customer_id")
	}

	found, err := h.service.GetByID(requestCtx.Context(), id)
	if err != nil {
		return writeCustomerError(requestCtx, err)
	}

	return requestCtx.Status(fiber.StatusOK).JSON(newCustomerResponse(*found))
}
