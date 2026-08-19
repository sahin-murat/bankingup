package gateway

import (
	"fmt"

	"github.com/sahin-murat/bankingup/internal/config"
	"github.com/sahin-murat/bankingup/internal/gateway/handler/health"
)

func (gw *gateway) DefineRoutes(cfg config.Config) error {

	err := health.RegisterHealthRoutes(gw.app, cfg)
	if err != nil {
		return fmt.Errorf("can not register health group routes: %w", err)
	}

	return nil
}
