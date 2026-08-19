package gateway

import (
	"fmt"

	customerhandler "github.com/sahin-murat/bankingup/internal/gateway/handler/customer"
	"github.com/sahin-murat/bankingup/internal/gateway/handler/health"
)

func (gw *gateway) DefineRoutes(db database) error {
	err := health.RegisterHealthRoutes(gw.app, db)
	if err != nil {
		return fmt.Errorf("can not register health group routes: %w", err)
	}

	err = customerhandler.RegisterCustomerRoutes(gw.app, db.GormDB())
	if err != nil {
		return fmt.Errorf("can not register customer group routes: %w", err)
	}

	return nil
}
