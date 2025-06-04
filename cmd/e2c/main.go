package main

import (
	"log/slog"
	"os"

	"github.com/nlamirault/e2c/internal/cmd"
	"github.com/nlamirault/e2c/internal/logger"
)

func main() {
	// Configure logger from environment variables
	logConfig := logger.NewConfig()
	
	// Set log level from environment variable
	if envLevel := os.Getenv("E2C_LOG_LEVEL"); envLevel != "" {
		logConfig.Level = logger.ParseLevel(envLevel)
	}

	// Set log format from environment variable
	if envFormat := os.Getenv("E2C_LOG_FORMAT"); envFormat != "" {
		logConfig.Format = logger.ParseFormat(envFormat)
	}

	// Create and set default logger
	log := logger.New(logConfig)
	logger.SetAsDefault(log)

	// Execute the root command
	if err := cmd.NewRootCommand(log).Execute(); err != nil {
		slog.Error("Failed to execute command", "error", err)
		os.Exit(1)
	}
}