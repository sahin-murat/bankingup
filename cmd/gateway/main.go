package main

import (
	"github.com/gofiber/fiber/v3/log"
	"github.com/sahin-murat/bankingup/internal/config"
	"github.com/sahin-murat/bankingup/internal/gateway"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Errorf("can not create config: %w", err)
		return
	}

	application, err := gateway.New(cfg)
	if err != nil {
		log.Errorf("can not gateway application: %w", err)
		return
	}

	err = application.Listen()
	if err != nil {
		log.Errorf("can not create config: %w", err)
		return
	}
}
