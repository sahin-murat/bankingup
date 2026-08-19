package currency

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	currencydomain "github.com/sahin-murat/bankingup/internal/currency"
)

type currencyResponse struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	DecimalPlaces int16  `json:"decimal_places"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func newCurrencyResponse(input currencydomain.Currency) currencyResponse {
	return currencyResponse{
		Code:          input.Code,
		Name:          input.Name,
		DecimalPlaces: input.DecimalPlaces,
	}
}

func writeCurrencyError(requestCtx fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, currencydomain.ErrInvalidInput),
		errors.Is(err, currencydomain.ErrInvalidCode):
		return requestCtx.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: err.Error()})
	case errors.Is(err, currencydomain.ErrCodeAlreadyExists):
		return requestCtx.Status(fiber.StatusConflict).JSON(errorResponse{Error: currencydomain.ErrCodeAlreadyExists.Error()})
	default:
		log.Errorf("currency request failed: %v", err)
		return requestCtx.Status(fiber.StatusInternalServerError).JSON(errorResponse{Error: "internal server error"})
	}
}

func writeBadRequest(requestCtx fiber.Ctx, message string) error {
	return requestCtx.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: message})
}
