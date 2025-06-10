// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/open-feature/go-sdk/pkg/openfeature"

	"github.com/nlamirault/e2c/internal/config"
	"github.com/nlamirault/e2c/internal/featureflags"
)

func main() {
	// Configure a simple logger
	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelInfo}))

	// Set some environment variables for feature flags
	os.Setenv("E2C_FEATURE_EXAMPLE_BOOL", "true")
	os.Setenv("E2C_FEATURE_EXAMPLE_STRING", "Hello from environment variable!")
	os.Setenv("E2C_FEATURE_EXAMPLE_NUMBER", "42")

	// Create a configuration with the environment provider
	cfg := &config.Config{
		FeatureFlags: featureflags.FeatureFlagsConfig{
			Enabled:  true,
			Provider: featureflags.EnvProvider,
			Env: featureflags.EnvConfig{
				Prefix:        "E2C_FEATURE_",
				CaseSensitive: false,
			},
		},
	}

	// Initialize the feature flags client
	if err := featureflags.InitializeClient(logger, cfg.FeatureFlags); err != nil {
		logger.Error("Failed to initialize feature flags", "error", err)
		os.Exit(1)
	}

	// Create a context
	ctx := context.Background()

	// Get the feature flag client
	client := featureflags.GetClient()
	if client == nil {
		logger.Error("Failed to get feature flag client")
		os.Exit(1)
	}

	// Get feature flag values using the client directly
	boolValue, err := client.BooleanValue(ctx, "EXAMPLE_BOOL", false, nil)
	if err != nil {
		logger.Error("Failed to get boolean value", "error", err)
	}
	logger.Info("Boolean value from client", "value", boolValue)

	stringValue, err := client.StringValue(ctx, "EXAMPLE_STRING", "default", nil)
	if err != nil {
		logger.Error("Failed to get string value", "error", err)
	}
	logger.Info("String value from client", "value", stringValue)

	intValue, err := client.IntValue(ctx, "EXAMPLE_NUMBER", 0, nil)
	if err != nil {
		logger.Error("Failed to get int value", "error", err)
	}
	logger.Info("Int value from client", "value", intValue)

	// Get feature flag values using the helper functions
	helperBoolValue := featureflags.GetBoolValue(ctx, "EXAMPLE_BOOL", false)
	logger.Info("Boolean value from helper", "value", helperBoolValue)

	helperStringValue := featureflags.GetStringValue(ctx, "EXAMPLE_STRING", "default")
	logger.Info("String value from helper", "value", helperStringValue)

	helperIntValue := featureflags.GetIntValue(ctx, "EXAMPLE_NUMBER", 0)
	logger.Info("Int value from helper", "value", helperIntValue)

	// Try a non-existent feature flag
	nonExistentValue := featureflags.GetStringValue(ctx, "NON_EXISTENT", "default value")
	logger.Info("Non-existent value", "value", nonExistentValue)

	fmt.Println("\nEnvironment variables provider example completed successfully!")
}