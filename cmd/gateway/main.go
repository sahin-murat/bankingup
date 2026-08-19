package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/sahin-murat/bankingup/internal/config"
	"github.com/sahin-murat/bankingup/internal/database"
	"github.com/sahin-murat/bankingup/internal/gateway"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		log.Errorf("gateway application stopped: %v", err)
		os.Exit(1)
	}
}

func run() (runErr error) {
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("can not create config: %w", err)
	}

	db, err := database.New(cfg)
	if err != nil {
		return fmt.Errorf("can not create database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			runErr = errors.Join(runErr, err)
		}
	}()

	application, err := gateway.New(cfg, db)
	if err != nil {
		return fmt.Errorf("can not create gateway application: %w", err)
	}

	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	listenErr := make(chan error, 1)
	go func() {
		listenErr <- application.Listen()
	}()

	select {
	case err := <-listenErr:
		if err != nil {
			return fmt.Errorf("gateway listener stopped: %w", err)
		}
		return nil
	case <-signalCtx.Done():
		stopSignals()
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()

	if err := application.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("can not gracefully shut down gateway: %w", err)
	}

	if err := <-listenErr; err != nil {
		return fmt.Errorf("gateway listener stopped during shutdown: %w", err)
	}

	return nil
}
