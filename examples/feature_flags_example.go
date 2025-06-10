// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/open-feature/go-sdk/pkg/openfeature"

	"github.com/nlamirault/e2c/internal/config"
	"github.com/nlamirault/e2c/internal/featureflags"
)

// Example flags - in a real application these would be defined as constants
const (
	EnableNewUIFlag       = "enable_new_ui"
	MaxConnectionsFlag    = "max_connections"
	RefreshIntervalFlag   = "refresh_interval"
	DefaultRegionFlag     = "default_region"
	WelcomeMessageFlag    = "welcome_message"
)

// Environment variables for env provider (prefix will be added automatically)
const (
	EnvPrefix = "E2C_FEATURE_"
)

func main() {
	// Parse command line flags
	var providerName string
	flag.StringVar(&providerName, "provider", "", "Feature flag provider to use (configcat, env)")
	flag.Parse()

	// Configure a simple logger
	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelInfo}))

	// Load configuration (including feature flags)
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Override provider if specified via command line
	if providerName != "" {
		cfg.FeatureFlags.Provider = featureflags.ProviderType(providerName)
		logger.Info("Using provider from command line", "provider", providerName)
	}

	// Check if feature flags are enabled
	if !cfg.FeatureFlags.Enabled {
		logger.Warn("Feature flags are not enabled. Set feature_flags.enabled=true in config to use them.")
		os.Exit(0)
	}

	// Set up environment variables if using env provider
	if cfg.FeatureFlags.Provider == featureflags.EnvProvider {
		// Set example environment variables (in a real app, these would be set externally)
		os.Setenv(EnvPrefix+"ENABLE_NEW_UI", "true")
		os.Setenv(EnvPrefix+"MAX_CONNECTIONS", "25")
		os.Setenv(EnvPrefix+"REFRESH_INTERVAL", "15.5")
		os.Setenv(EnvPrefix+"DEFAULT_REGION", "eu-west-1")
		os.Setenv(EnvPrefix+"WELCOME_MESSAGE", "Hello from environment variables!")
		
		logger.Info("Using environment variable provider with example values")
	} else {
		logger.Info("Using ConfigCat provider")
	}

	// Create context
	ctx := context.Background()

	// Create evaluation context with user information for targeting
	evalCtx := openfeature.NewEvaluationContext(
		"user-123", // User ID
		map[string]interface{}{
			"role":     "admin",
			"region":   "us-west-1",
			"beta":     true,
			"org":      "engineering",
			"verified": true,
		},
	)

	// Simple usage examples
	// 1. Boolean flag - determines if a new UI feature is enabled
	newUIEnabled := featureflags.GetBoolValue(ctx, EnableNewUIFlag, false)
	fmt.Printf("New UI feature enabled: %v\n", newUIEnabled)

	// 2. Integer flag - determines maximum number of connections
	maxConnections := featureflags.GetIntValue(ctx, MaxConnectionsFlag, 10)
	fmt.Printf("Maximum connections: %d\n", maxConnections)

	// 3. Float flag - determines refresh interval in seconds
	refreshInterval := featureflags.GetFloatValue(ctx, RefreshIntervalFlag, 30.0)
	fmt.Printf("Refresh interval: %.1f seconds\n", refreshInterval)

	// 4. String flag - determines default AWS region
	defaultRegion := featureflags.GetStringValue(ctx, DefaultRegionFlag, "us-west-1")
	fmt.Printf("Default region: %s\n", defaultRegion)

	// 5. Using the client directly with evaluation context
	client := featureflags.GetClient()
	if client != nil {
		// This uses the evaluation context for targeting
		welcomeMessage, err := client.StringValue(ctx, WelcomeMessageFlag, "Welcome to e2c!", evalCtx)
		if err != nil {
			logger.Error("Failed to get welcome message flag", "error", err)
		} else {
			fmt.Printf("Welcome message: %s\n", welcomeMessage)
		}
	}

	// Demonstrate flag dependency (conditional logic based on flag values)
	if newUIEnabled {
		// This feature is only relevant if the new UI is enabled
		showDetailedView := featureflags.GetBoolValue(ctx, "detailed_instance_view", true)
		fmt.Printf("Detailed instance view: %v\n", showDetailedView)
	}

	fmt.Println("\nFeature flag evaluation complete!")
	
	// Print instructions on how to switch providers
	fmt.Println("\nTry running this example with different providers:")
	fmt.Println("  go run examples/feature_flags_example.go --provider=configcat")
	fmt.Println("  go run examples/feature_flags_example.go --provider=env")
}