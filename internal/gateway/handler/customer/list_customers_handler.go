package customer

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
)

type CustomerLister interface {
	List(context.Context) ([]customerdomain.Customer, error)
}

type listCustomersHandler struct {
	service CustomerLister
}

func NewListCustomersHandler(service CustomerLister) (*listCustomersHandler, error) {
	if service == nil {
		return nil, errors.New("customer lister is required")
	}

	return &listCustomersHandler{service: service}, nil
}

func (h *listCustomersHandler) Handle(requestCtx fiber.Ctx) error {
	customers, err := h.service.List(requestCtx.Context())
	if err != nil {
		return writeCustomerError(requestCtx, err)
	}

	response := make([]customerResponse, len(customers))
	for index, item := range customers {
		response[index] = newCustomerResponse(item)
	}

	return requestCtx.Status(fiber.StatusOK).JSON(response)
}
