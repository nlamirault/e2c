package main

import (
	"os"

	"github.com/nlamirault/e2c/internal/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	
	// Set log level based on environment variable
	logLevel := os.Getenv("E2C_LOG_LEVEL")
	if logLevel != "" {
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			log.WithError(err).Warnf("Invalid log level %s, defaulting to info", logLevel)
			log.SetLevel(logrus.InfoLevel)
		} else {
			log.SetLevel(level)
		}
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	// Execute the root command
	if err := cmd.NewRootCommand(log).Execute(); err != nil {
		log.WithError(err).Fatal("Failed to execute command")
		os.Exit(1)
	}
}