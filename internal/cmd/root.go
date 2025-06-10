// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/hooks"
	"github.com/spf13/cobra"

	"github.com/nlamirault/e2c/internal/aws"
	"github.com/nlamirault/e2c/internal/config"
	"github.com/nlamirault/e2c/internal/featureflags"
	"github.com/nlamirault/e2c/internal/logger"
	"github.com/nlamirault/e2c/internal/ui"
	"github.com/nlamirault/e2c/internal/version"
)

// NewRootCommand creates the root command for e2c
func NewRootCommand(log *slog.Logger) *cobra.Command {
	var (
		cfgFile             string
		profile             string
		region              string
		featureFlagProvider string
		openfeatureClient   *openfeature.Client
	)

	cmd := &cobra.Command{
		Use:   "e2c",
		Short: "AWS EC2 Terminal UI Manager",
		Long: `e2c is a terminal-based UI application for managing AWS EC2 instances,
inspired by k9s for Kubernetes and e1s for ECS.

It provides a simple, intuitive interface for managing EC2 instances
across multiple regions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.LoadConfig(log)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Override feature flag provider if specified
			if featureFlagProvider != "" {
				log.Info("Overriding feature flag provider", "provider", featureFlagProvider)
				cfg.OverrideFeatureFlags(featureFlagProvider)

				// Make sure feature flags are enabled when provider is specified
				if !cfg.FeatureFlags.Enabled {
					log.Info("Enabling feature flags as provider was specified")
					cfg.FeatureFlags.Enabled = true
				}

				// Initialize the feature flags client with the new provider
				openfeatureClient, err = featureflags.InitializeClient(log, cfg.FeatureFlags)
				if err != nil {
					log.Warn("Failed to initialize feature flags client", "error", err)
				}
			}

			ctx := context.Background()

			logging, err := openfeatureClient.BooleanValue(ctx, "logging", false, openfeature.EvaluationContext{})
			if err != nil {
				log.Warn("Feature flag error while getting logging value", "error", err)
			} else {
				if !logging {
					log.Info("Feature flag set to disable logging")
					logger.SetAsDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
				}
			}

			// Override with CLI flags
			cfg.Override(profile, region)

			// Register a logging hook globally to run on all evaluations
			loggingHook := hooks.NewLoggingHook(false, log)
			openfeature.AddHooks(loggingHook)

			// Create AWS EC2 client
			ec2Client, err := aws.NewEC2Client(log, cfg.AWS.DefaultRegion, cfg.AWS.Profile)
			if err != nil {
				return fmt.Errorf("failed to create EC2 client: %w", err)
			}

			// Create and start UI
			app := ui.NewUI(log, ec2Client, openfeatureClient, cfg)
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
	cmd.PersistentFlags().StringVar(&featureFlagProvider, "openfeature-provider", "env", "feature flag provider to use (configcat, env, devcycle)")

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
