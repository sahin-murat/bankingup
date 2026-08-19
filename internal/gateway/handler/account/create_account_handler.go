package account

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
	"github.com/shopspring/decimal"
)

type AccountCreator interface {
	Create(context.Context, accountdomain.CreateInput) (*accountdomain.Account, error)
}

type createAccountHandler struct {
	service AccountCreator
}

type createAccountRequest struct {
	CustomerID     string  `json:"customer_id"`
	Currency       string  `json:"currency"`
	InitialBalance *string `json:"initial_balance"`
}

func NewCreateAccountHandler(service AccountCreator) (*createAccountHandler, error) {
	if service == nil {
		return nil, errors.New("account creator is required")
	}

	return &createAccountHandler{service: service}, nil
}

func (h *createAccountHandler) Handle(requestCtx fiber.Ctx) error {
	var request createAccountRequest
	if err := requestCtx.Bind().Body(&request); err != nil {
		return writeBadRequest(requestCtx, "invalid request body")
	}

	customerID, err := uuid.Parse(request.CustomerID)
	if err != nil {
		return writeBadRequest(requestCtx, "invalid customer_id")
	}

	initialBalance := decimal.Zero
	if request.InitialBalance != nil {
		initialBalance, err = decimal.NewFromString(*request.InitialBalance)
		if err != nil {
			return writeBadRequest(requestCtx, "invalid initial_balance")
		}
	}

	created, err := h.service.Create(requestCtx.Context(), accountdomain.CreateInput{
		CustomerID:     customerID,
		Currency:       request.Currency,
		InitialBalance: initialBalance,
	})
	if err != nil {
		return writeAccountError(requestCtx, err)
	}

	return requestCtx.Status(fiber.StatusCreated).JSON(newAccountResponse(*created))
}
