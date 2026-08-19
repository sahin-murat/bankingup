package customer

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
)

type customerResponse struct {
	ID        string                `json:"id"`
	FirstName string                `json:"first_name"`
	LastName  string                `json:"last_name"`
	Email     string                `json:"email"`
	Status    customerdomain.Status `json:"status"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func newCustomerResponse(input customerdomain.Customer) customerResponse {
	return customerResponse{
		ID:        input.ID.String(),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Status:    input.Status,
		CreatedAt: input.CreatedAt,
		UpdatedAt: input.UpdatedAt,
	}
}

func writeCustomerError(requestCtx fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, customerdomain.ErrInvalidInput),
		errors.Is(err, customerdomain.ErrInvalidStatusTransition):
		return requestCtx.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: err.Error()})
	case errors.Is(err, customerdomain.ErrNotFound):
		return requestCtx.Status(fiber.StatusNotFound).JSON(errorResponse{Error: customerdomain.ErrNotFound.Error()})
	case errors.Is(err, customerdomain.ErrEmailAlreadyExists):
		return requestCtx.Status(fiber.StatusConflict).JSON(errorResponse{Error: customerdomain.ErrEmailAlreadyExists.Error()})
	default:
		log.Errorf("customer request failed: %v", err)
		return requestCtx.Status(fiber.StatusInternalServerError).JSON(errorResponse{Error: "internal server error"})
	}
}

func writeBadRequest(requestCtx fiber.Ctx, message string) error {
	return requestCtx.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: message})
}
