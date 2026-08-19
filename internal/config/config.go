package config

import (
	"os"
	"strings"
)

const defaultServerAddress = ":8080"

type Config interface {
	GetServerAddress() string
}

type config struct {
	serverAddress string
}

func New() (*config, error) {
	serverAddress := strings.TrimSpace(os.Getenv("SERVER_ADDRESS"))
	if serverAddress == "" {
		serverAddress = defaultServerAddress
	}

	return &config{
		serverAddress: serverAddress,
	}, nil
}

func (c *config) GetServerAddress() string {
	return c.serverAddress
}
