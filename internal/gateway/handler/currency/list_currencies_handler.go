package currency

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	currencydomain "github.com/sahin-murat/bankingup/internal/currency"
)

type CurrencyLister interface {
	List(context.Context) ([]currencydomain.Currency, error)
}

type listCurrenciesHandler struct {
	service CurrencyLister
}

func NewListCurrenciesHandler(service CurrencyLister) (*listCurrenciesHandler, error) {
	if service == nil {
		return nil, errors.New("currency lister is required")
	}

	return &listCurrenciesHandler{service: service}, nil
}

func (h *listCurrenciesHandler) Handle(requestCtx fiber.Ctx) error {
	currencies, err := h.service.List(requestCtx.Context())
	if err != nil {
		return writeCurrencyError(requestCtx, err)
	}

	response := make([]currencyResponse, len(currencies))
	for index, item := range currencies {
		response[index] = newCurrencyResponse(item)
	}

	return requestCtx.Status(fiber.StatusOK).JSON(response)
}
