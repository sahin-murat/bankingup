package health

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/sahin-murat/bankingup/internal/config"
)

func RegisterHealthRoutes(app *fiber.App, cfg config.Config) error {
	healthGroup := app.Group("/health")

	healthHandler, err := NewHealthHandler(cfg)
	if err != nil {
		return fmt.Errorf("can not create health handler: %w", err)
	}

	healthGroup.Get("", healthHandler.Handle)

	return nil
}
