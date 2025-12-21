// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/nlamirault/e2c/internal/aws"
	"github.com/nlamirault/e2c/internal/config"
	"github.com/nlamirault/e2c/internal/logger"
	"github.com/nlamirault/e2c/internal/ui"
	"github.com/nlamirault/e2c/internal/version"
)

// NewRootCommand creates the root command for e2c
func NewRootCommand(log *slog.Logger) *cobra.Command {
	var (
		cfgFile   string
		profile   string
		region    string
		logFormat string
		logLevel  string
		expert    bool
	)

	cmd := &cobra.Command{
		Use:   "e2c",
		Short: "AWS EC2 Terminal UI Manager",
		Long: `e2c is a terminal-based UI application for managing AWS EC2 instances,
inspired by k9s for Kubernetes and e1s for ECS.

It provides a simple, intuitive interface for managing EC2 instances
across multiple regions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Configure logging if requested via flags
			if logFormat != "" || logLevel != "" {
				logConfig := logger.NewConfig()

				// Set format if specified
				if logFormat != "" {
					logConfig.Format = logger.ParseFormat(logFormat)
				}

				// Set level if specified
				if logLevel != "" {
					logConfig.Level = logger.ParseLevel(logLevel)
				}

				// Create and set the new logger
				log = logger.New(logConfig)
				logger.SetAsDefault(log)
			}

			// Load configuration
			cfg, err := config.LoadConfig(log)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Override with CLI flags
			cfg.Override(profile, region)

			// Enable expert mode if requested via flag
			if expert {
				cfg.UI.ExpertMode = true
			}

			// Create AWS EC2 client
			ec2Client, err := aws.NewEC2Client(log, cfg.AWS.DefaultRegion, cfg.AWS.Profile)
			if err != nil {
				return fmt.Errorf("failed to create EC2 client: %w", err)
			}

			// Create and start UI
			app := ui.NewUI(log, ec2Client, cfg)
			if err := app.Start(); err != nil {
				return fmt.Errorf("UI error: %w", err)
			}

			return nil
		},
		Version: version.GetVersion(),
	}

	// Add flags
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/e2c/config.yaml)")
	cmd.PersistentFlags().StringVar(&profile, "profile", "", "AWS profile to use")
	cmd.PersistentFlags().StringVar(&region, "region", "", "AWS region to use")
	cmd.PersistentFlags().StringVar(&logFormat, "log-format", "", "set log format (json, text)")
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "set logging level (debug, info, warn, error)")
	cmd.PersistentFlags().BoolVar(&expert, "expert-mode", false, "enable expert mode features (protection management)")

	// Add version command
	cmd.AddCommand(newVersionCommand())

	return cmd
}

// newVersionCommand creates a version command
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("e2c version %s\n", version.GetVersion())
		},
	}
}

// Execute executes the root command
func Execute() {
	log := logger.New(nil) // Use default logger configuration
	if err := NewRootCommand(log).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
