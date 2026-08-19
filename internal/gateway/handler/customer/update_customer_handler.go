package customer

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
)

type CustomerUpdater interface {
	Update(context.Context, customerdomain.UpdateInput) (*customerdomain.Customer, error)
}

type updateCustomerHandler struct {
	service CustomerUpdater
}

type updateCustomerRequest struct {
	FirstName *string                `json:"first_name"`
	LastName  *string                `json:"last_name"`
	Email     *string                `json:"email"`
	Status    *customerdomain.Status `json:"status"`
}

func NewUpdateCustomerHandler(service CustomerUpdater) (*updateCustomerHandler, error) {
	if service == nil {
		return nil, errors.New("customer updater is required")
	}

	return &updateCustomerHandler{service: service}, nil
}

func (h *updateCustomerHandler) Handle(requestCtx fiber.Ctx) error {
	id, err := uuid.Parse(requestCtx.Params("customer_id"))
	if err != nil {
		return writeBadRequest(requestCtx, "invalid customer_id")
	}

	var request updateCustomerRequest
	if err := requestCtx.Bind().Body(&request); err != nil {
		return writeBadRequest(requestCtx, "invalid request body")
	}

	updated, err := h.service.Update(requestCtx.Context(), customerdomain.UpdateInput{
		ID:        id,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
		Status:    request.Status,
	})
	if err != nil {
		return writeCustomerError(requestCtx, err)
	}

	return requestCtx.Status(fiber.StatusOK).JSON(newCustomerResponse(*updated))
}
