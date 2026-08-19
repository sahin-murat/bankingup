package config

import (
	"os"
	"strings"
)

const (
	defaultServerAddress = ":8080"
	defaultDatabaseURL   = "postgresql://postgres:postgres@localhost:5432/postgres"
)

type Config interface {
	GetServerAddress() string
	GetDatabaseURL() string
}

type config struct {
	serverAddress string
	databaseURL   string
}

func New() (*config, error) {
	serverAddress := strings.TrimSpace(os.Getenv("SERVER_ADDRESS"))
	if serverAddress == "" {
		serverAddress = defaultServerAddress
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
	}

	return &config{
		serverAddress: serverAddress,
		databaseURL:   databaseURL,
	}, nil
}

func (c *config) GetServerAddress() string {
	return c.serverAddress
}

func (c *config) GetDatabaseURL() string {
	return c.databaseURL
}
