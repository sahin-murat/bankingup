package account

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	accountdomain "github.com/sahin-murat/bankingup/internal/account"
	currencydomain "github.com/sahin-murat/bankingup/internal/currency"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
	"gorm.io/gorm"
)

func RegisterAccountRoutes(app *fiber.App, db *gorm.DB) error {
	accountRepository, err := accountdomain.NewRepository(db)
	if err != nil {
		return fmt.Errorf("can not create account repository: %w", err)
	}

	customerRepository, err := customerdomain.NewRepository(db)
	if err != nil {
		return fmt.Errorf("can not create customer repository for accounts: %w", err)
	}

	currencyRepository, err := currencydomain.NewRepository(db)
	if err != nil {
		return fmt.Errorf("can not create currency repository for accounts: %w", err)
	}

	currencyService, err := currencydomain.NewService(currencyRepository)
	if err != nil {
		return fmt.Errorf("can not create currency service for accounts: %w", err)
	}

	service, err := accountdomain.NewService(accountRepository, customerRepository, currencyService)
	if err != nil {
		return fmt.Errorf("can not create account service: %w", err)
	}

	createHandler, err := NewCreateAccountHandler(service)
	if err != nil {
		return fmt.Errorf("can not create create account handler: %w", err)
	}

	getHandler, err := NewGetAccountHandler(service)
	if err != nil {
		return fmt.Errorf("can not create get account handler: %w", err)
	}

	listHandler, err := NewListAccountsHandler(service)
	if err != nil {
		return fmt.Errorf("can not create list accounts handler: %w", err)
	}

	updateHandler, err := NewUpdateAccountHandler(service)
	if err != nil {
		return fmt.Errorf("can not create update account handler: %w", err)
	}

	accounts := app.Group("/accounts")
	accounts.Post("", createHandler.Handle)
	accounts.Get("", listHandler.Handle)
	accounts.Get("/:account_id", getHandler.Handle)
	accounts.Patch("/:account_id", updateHandler.Handle)

	return nil
}
