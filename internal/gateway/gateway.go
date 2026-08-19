package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sahin-murat/bankingup/internal/config"
	"gorm.io/gorm"
)

const serverReadTimeout = 10 * time.Second

type database interface {
	Ping(context.Context) error
	GormDB() *gorm.DB
}

type gateway struct {
	app          *fiber.App
	serverAdress string
}

func New(cfg config.Config, db database) (*gateway, error) {
	serverAddress := cfg.GetServerAddress()

	app := fiber.New(fiber.Config{
		ReadTimeout: serverReadTimeout,
	})

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

func (gw *gateway) Shutdown(ctx context.Context) error {
	if err := gw.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("can not shut down gateway server: %w", err)
	}

	return nil
}
