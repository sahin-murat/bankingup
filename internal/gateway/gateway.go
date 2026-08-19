package gateway

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/sahin-murat/bankingup/internal/config"
)

type gateway struct {
	app          *fiber.App
	serverAdress string
}

func New(cfg config.Config) (*gateway, error) {
	serverAddress := cfg.GetServerAddress()

	app := fiber.New()

	gw := &gateway{
		app:          app,
		serverAdress: serverAddress,
	}

	gw.DefineRoutes(cfg)

	return gw, nil
}

func (gw *gateway) Listen() error {
	err := gw.app.Listen(gw.serverAdress)
	if err != nil {
		return fmt.Errorf("can not start listening server on port: %s, %w", gw.serverAdress, err)
	}

	return nil
}
