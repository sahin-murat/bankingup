package gateway

import (
	"fmt"

	"github.com/sahin-murat/bankingup/internal/database"
	"github.com/sahin-murat/bankingup/internal/gateway/handler/health"
)

func (gw *gateway) DefineRoutes(db database.Database) error {
	err := health.RegisterHealthRoutes(gw.app, db)
	if err != nil {
		return fmt.Errorf("can not register health group routes: %w", err)
	}

	return nil
}
