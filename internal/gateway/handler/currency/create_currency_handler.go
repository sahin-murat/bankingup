package currency

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	currencydomain "github.com/sahin-murat/bankingup/internal/currency"
)

type CurrencyCreator interface {
	Create(context.Context, currencydomain.CreateInput) (*currencydomain.Currency, error)
}

type createCurrencyHandler struct {
	service CurrencyCreator
}

type createCurrencyRequest struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	DecimalPlaces *int16 `json:"decimal_places"`
}

func NewCreateCurrencyHandler(service CurrencyCreator) (*createCurrencyHandler, error) {
	if service == nil {
		return nil, errors.New("currency creator is required")
	}

	return &createCurrencyHandler{service: service}, nil
}

func (h *createCurrencyHandler) Handle(requestCtx fiber.Ctx) error {
	var request createCurrencyRequest
	if err := requestCtx.Bind().Body(&request); err != nil {
		return writeBadRequest(requestCtx, "invalid request body")
	}
	if request.DecimalPlaces == nil {
		return writeBadRequest(requestCtx, "decimal_places is required")
	}

	created, err := h.service.Create(requestCtx.Context(), currencydomain.CreateInput{
		Code:          request.Code,
		Name:          request.Name,
		DecimalPlaces: *request.DecimalPlaces,
	})
	if err != nil {
		return writeCurrencyError(requestCtx, err)
	}

	return requestCtx.Status(fiber.StatusCreated).JSON(newCurrencyResponse(*created))
}
