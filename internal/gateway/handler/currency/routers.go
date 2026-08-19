package currency

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	currencydomain "github.com/sahin-murat/bankingup/internal/currency"
	"gorm.io/gorm"
)

func RegisterCurrencyRoutes(app *fiber.App, db *gorm.DB) error {
	repository, err := currencydomain.NewRepository(db)
	if err != nil {
		return fmt.Errorf("can not create currency repository: %w", err)
	}

	service, err := currencydomain.NewService(repository)
	if err != nil {
		return fmt.Errorf("can not create currency service: %w", err)
	}

	createHandler, err := NewCreateCurrencyHandler(service)
	if err != nil {
		return fmt.Errorf("can not create create currency handler: %w", err)
	}

	listHandler, err := NewListCurrenciesHandler(service)
	if err != nil {
		return fmt.Errorf("can not create list currencies handler: %w", err)
	}

	currencies := app.Group("/currencies")
	currencies.Post("", createHandler.Handle)
	currencies.Get("", listHandler.Handle)

	return nil
}
