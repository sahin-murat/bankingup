package gateway

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/sahin-murat/bankingup/internal/config"
	"github.com/sahin-murat/bankingup/internal/database"
)

type gateway struct {
	app          *fiber.App
	serverAdress string
}

func New(cfg config.Config, db database.Database) (*gateway, error) {
	serverAddress := cfg.GetServerAddress()

	app := fiber.New()

	gw := &gateway{
		app:          app,
		serverAdress: serverAddress,
	}

	if err := gw.DefineRoutes(db); err != nil {
		return nil, fmt.Errorf("can not define gateway routes: %w", err)
	}

	return gw, nil
}

func (gw *gateway) Listen() error {
	err := gw.app.Listen(gw.serverAdress)
	if err != nil {
		return fmt.Errorf("can not start listening server on port: %s, %w", gw.serverAdress, err)
	}

	return nil
}
