package health

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sahin-murat/bankingup/internal/config"
)

type healthHandler struct {
}

func NewHealthHandler(cfg config.Config) (*healthHandler, error) {
	return &healthHandler{}, nil
}

func (h *healthHandler) Handle(requestCtx fiber.Ctx) error {
	return requestCtx.SendStatus(fiber.StatusOK)
}
