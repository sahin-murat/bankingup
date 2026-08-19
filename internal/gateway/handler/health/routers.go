package health

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/sahin-murat/bankingup/internal/database"
)

func RegisterHealthRoutes(app *fiber.App, db database.Database) error {
	healthGroup := app.Group("/health")

	healthHandler, err := NewHealthHandler(db)
	if err != nil {
		return fmt.Errorf("can not create health handler: %w", err)
	}

	healthGroup.Get("", healthHandler.Handle)
	healthGroup.Get("/ready", healthHandler.HandleReadiness)

	return nil
}
