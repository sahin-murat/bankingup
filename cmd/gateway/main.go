package main

import (
	"github.com/gofiber/fiber/v3/log"
	"github.com/sahin-murat/bankingup/internal/config"
	"github.com/sahin-murat/bankingup/internal/database"
	"github.com/sahin-murat/bankingup/internal/gateway"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Errorf("can not create config: %v", err)
		return
	}

	db, err := database.New(cfg)
	if err != nil {
		log.Errorf("can not create database: %v", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Errorf("can not close database: %v", err)
		}
	}()

	application, err := gateway.New(cfg, db)
	if err != nil {
		log.Errorf("can not create gateway application: %v", err)
		return
	}

	err = application.Listen()
	if err != nil {
		log.Errorf("gateway application stopped: %v", err)
		return
	}
}
