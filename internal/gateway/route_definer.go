package gateway

import (
	"fmt"

	accounthandler "github.com/sahin-murat/bankingup/internal/gateway/handler/account"
	currencyhandler "github.com/sahin-murat/bankingup/internal/gateway/handler/currency"
	customerhandler "github.com/sahin-murat/bankingup/internal/gateway/handler/customer"
	"github.com/sahin-murat/bankingup/internal/gateway/handler/health"
	transferhandler "github.com/sahin-murat/bankingup/internal/gateway/handler/transfer"
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

	err = currencyhandler.RegisterCurrencyRoutes(gw.app, db.GormDB())
	if err != nil {
		return fmt.Errorf("can not register currency group routes: %w", err)
	}

	err = accounthandler.RegisterAccountRoutes(gw.app, db.GormDB())
	if err != nil {
		return fmt.Errorf("can not register account group routes: %w", err)
	}

	err = transferhandler.RegisterTransferRoutes(gw.app, db.GormDB())
	if err != nil {
		return fmt.Errorf("can not register transfer group routes: %w", err)
	}

	return nil
}
