package main

import (
	"fmt"
	"os"

	"github.com/fiverr/go_platform/core/logging"
)

// provideLogger builds the structured logger shared by the handlers.
// Anything other than GO_ENV=production gets the development config.
func provideLogger() (logging.Logger, error) {
	env := logging.EnvDevelopment
	if os.Getenv("GO_ENV") == logging.EnvProduction {
		env = logging.EnvProduction
	}
	cfg := logging.DefaultConfig(env)
	cfg.ServiceName = "book_service"
	logger, err := logging.NewLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	return logger, nil
}
