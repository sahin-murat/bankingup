package transfer

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	transferdomain "github.com/sahin-murat/bankingup/internal/transfer"
	"github.com/shopspring/decimal"
)

type TransferCreator interface {
	Create(context.Context, transferdomain.CreateInput) (*transferdomain.CreateResult, error)
}

type createTransferHandler struct {
	service TransferCreator
}

type createTransferRequest struct {
	SourceAccountID      string `json:"source_account_id"`
	DestinationAccountID string `json:"destination_account_id"`
	Amount               string `json:"amount"`
}

func NewCreateTransferHandler(service TransferCreator) (*createTransferHandler, error) {
	if service == nil {
		return nil, errors.New("transfer creator is required")
	}

	return &createTransferHandler{service: service}, nil
}

func (h *createTransferHandler) Handle(requestCtx fiber.Ctx) error {
	idempotencyKey, err := uuid.Parse(requestCtx.Get("Idempotency-Key"))
	if err != nil {
		return writeBadRequest(requestCtx, "invalid Idempotency-Key")
	}

	var request createTransferRequest
	if err := requestCtx.Bind().Body(&request); err != nil {
		return writeBadRequest(requestCtx, "invalid request body")
	}

	sourceAccountID, err := uuid.Parse(request.SourceAccountID)
	if err != nil {
		return writeBadRequest(requestCtx, "invalid source_account_id")
	}
	destinationAccountID, err := uuid.Parse(request.DestinationAccountID)
	if err != nil {
		return writeBadRequest(requestCtx, "invalid destination_account_id")
	}
	amount, err := decimal.NewFromString(request.Amount)
	if err != nil {
		return writeBadRequest(requestCtx, "invalid amount")
	}

	result, err := h.service.Create(requestCtx.Context(), transferdomain.CreateInput{
		SourceAccountID:      sourceAccountID,
		DestinationAccountID: destinationAccountID,
		Amount:               amount,
		IdempotencyKey:       idempotencyKey,
	})
	if err != nil {
		return writeTransferError(requestCtx, err)
	}

	status := fiber.StatusCreated
	if result.Replay {
		status = fiber.StatusOK
	}

	return requestCtx.Status(status).JSON(newTransferResponse(*result.Transfer))
}
