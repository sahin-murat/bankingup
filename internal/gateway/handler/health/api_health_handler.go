package health

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/sahin-murat/bankingup/internal/database"
)

const databasePingTimeout = 2 * time.Second

type healthHandler struct {
	db database.Database
}

func NewHealthHandler(db database.Database) (*healthHandler, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}

	return &healthHandler{db: db}, nil
}

func (h *healthHandler) Handle(requestCtx fiber.Ctx) error {
	return requestCtx.SendStatus(fiber.StatusOK)
}

func (h *healthHandler) HandleReadiness(requestCtx fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(requestCtx.Context(), databasePingTimeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		log.Errorf("database health check failed: %v", err)
		return requestCtx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":   "unavailable",
			"database": "down",
		})
	}

	return requestCtx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":   "ok",
		"database": "up",
	})
}
