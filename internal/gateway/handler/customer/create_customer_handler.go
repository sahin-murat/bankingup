package customer

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
)

type CustomerCreator interface {
	Create(context.Context, customerdomain.CreateInput) (*customerdomain.Customer, error)
}

type createCustomerHandler struct {
	service CustomerCreator
}

type createCustomerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

func NewCreateCustomerHandler(service CustomerCreator) (*createCustomerHandler, error) {
	if service == nil {
		return nil, errors.New("customer creator is required")
	}

	return &createCustomerHandler{service: service}, nil
}

func (h *createCustomerHandler) Handle(requestCtx fiber.Ctx) error {
	var request createCustomerRequest
	if err := requestCtx.Bind().Body(&request); err != nil {
		return writeBadRequest(requestCtx, "invalid request body")
	}

	created, err := h.service.Create(requestCtx.Context(), customerdomain.CreateInput{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
	})
	if err != nil {
		return writeCustomerError(requestCtx, err)
	}

	return requestCtx.Status(fiber.StatusCreated).JSON(newCustomerResponse(*created))
}
