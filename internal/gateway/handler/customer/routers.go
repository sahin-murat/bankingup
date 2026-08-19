package customer

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	customerdomain "github.com/sahin-murat/bankingup/internal/customer"
	"gorm.io/gorm"
)

func RegisterCustomerRoutes(app *fiber.App, db *gorm.DB) error {
	repository, err := customerdomain.NewRepository(db)
	if err != nil {
		return fmt.Errorf("can not create customer repository: %w", err)
	}

	service, err := customerdomain.NewService(repository)
	if err != nil {
		return fmt.Errorf("can not create customer service: %w", err)
	}

	createHandler, err := NewCreateCustomerHandler(service)
	if err != nil {
		return fmt.Errorf("can not create create customer handler: %w", err)
	}

	getHandler, err := NewGetCustomerHandler(service)
	if err != nil {
		return fmt.Errorf("can not create get customer handler: %w", err)
	}

	listHandler, err := NewListCustomersHandler(service)
	if err != nil {
		return fmt.Errorf("can not create list customers handler: %w", err)
	}

	updateHandler, err := NewUpdateCustomerHandler(service)
	if err != nil {
		return fmt.Errorf("can not create update customer handler: %w", err)
	}

	customers := app.Group("/customers")
	customers.Post("", createHandler.Handle)
	customers.Get("", listHandler.Handle)
	customers.Get("/:customer_id", getHandler.Handle)
	customers.Patch("/:customer_id", updateHandler.Handle)

	return nil
}
