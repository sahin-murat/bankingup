package transfer

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	currencydomain "github.com/sahin-murat/bankingup/internal/currency"
	transferdomain "github.com/sahin-murat/bankingup/internal/transfer"
	"gorm.io/gorm"
)

func RegisterTransferRoutes(app *fiber.App, db *gorm.DB) error {
	repository, err := transferdomain.NewRepository(db)
	if err != nil {
		return fmt.Errorf("can not create transfer repository: %w", err)
	}

	currencyRepository, err := currencydomain.NewRepository(db)
	if err != nil {
		return fmt.Errorf("can not create currency repository for transfers: %w", err)
	}
	currencyService, err := currencydomain.NewService(currencyRepository)
	if err != nil {
		return fmt.Errorf("can not create currency service for transfers: %w", err)
	}

	service, err := transferdomain.NewService(repository, currencyService)
	if err != nil {
		return fmt.Errorf("can not create transfer service: %w", err)
	}

	createHandler, err := NewCreateTransferHandler(service)
	if err != nil {
		return fmt.Errorf("can not create create transfer handler: %w", err)
	}

	app.Post("/transfers", createHandler.Handle)
	return nil
}
