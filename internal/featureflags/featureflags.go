// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

// Package featureflags provides feature flag functionality using OpenFeature.
// It supports multiple providers including ConfigCat and environment variables.
package featureflags

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/nlamirault/e2c/internal/utils"
	"github.com/open-feature/go-sdk/openfeature"
)

// ProviderType represents the type of feature flag provider to use
type ProviderType string

const (
	// ConfigCatProvider represents the ConfigCat provider
	ConfigCatProvider ProviderType = "configcat"
	// EnvProvider represents the environment variable provider
	EnvProvider ProviderType = "env"
	// DevCycleProvider represents the DevCycle provider
	DevCycleProvider ProviderType = "devcycle"
)

var (
	client     *openfeature.Client
	clientOnce sync.Once
)

// FeatureFlagsConfig holds the configuration for feature flags
type FeatureFlagsConfig struct {
	// The provider to use (configcat, env, devcycle)
	Provider ProviderType `mapstructure:"provider"`
	// ConfigCat-specific configuration
	ConfigCat ConfigCatConfig `mapstructure:"configcat"`
	// Environment variable provider configuration
	Env EnvConfig `mapstructure:"env"`
	// DevCycle-specific configuration
	DevCycle DevCycleConfig `mapstructure:"devcycle"`
	// Enabled state for feature flags functionality
	Enabled bool `mapstructure:"enabled"`
}

// InitializeClient initializes the OpenFeature client with the specified provider
func InitializeClient(log *slog.Logger, config FeatureFlagsConfig) (*openfeature.Client, error) {
	if !config.Enabled {
		log.Info("Feature flags are disabled, skipping initialization")
		return nil, nil
	}

	// Reset the client if it already exists (to allow changing providers at runtime)
	ResetClient()

	// var err error
	// clientOnce.Do(func() {

	log.Info("Initializing feature flag client", "provider", config.Provider)

	var provider openfeature.FeatureProvider
	var providerErr error

	switch config.Provider {
	case ConfigCatProvider:
		provider, providerErr = NewConfigCatProvider(log, config.ConfigCat)
	case DevCycleProvider:
		provider, providerErr = NewDevCycleProvider(log, config.DevCycle)
	case EnvProvider:
		provider, providerErr = NewEnvProvider(log, config.Env)
	default:
		providerErr = fmt.Errorf("unsupported provider type: %s", config.Provider)
	}

	if providerErr != nil {
		// err = providerErr
		return nil, providerErr
	}

	// Set the provider at the global level
	err := openfeature.SetProviderAndWait(provider)
	if err != nil {
		// err = fmt.Errorf("failed to set OpenFeature provider: %w", setErr)
		return nil, err
	}

	// })

	// Create a named client
	client = openfeature.NewClient(utils.APP_NAME)
	log.Info("Feature flag client initialized successfully", "provider", config.Provider, "metadata", provider.Metadata())

	return client, err
}

// ResetClient resets the OpenFeature client to allow reinitialization
func ResetClient() {
	clientOnce = sync.Once{}
	client = nil
}

// GetClient returns the initialized OpenFeature client
func GetClient() *openfeature.Client {
	return client
}

// GetBoolValue retrieves a boolean feature flag value
func GetBoolValue(ctx context.Context, flagKey string, defaultValue bool) bool {
	if client == nil {
		return defaultValue
	}

	evalCtx := openfeature.NewEvaluationContext("", nil)
	value, err := client.BooleanValue(ctx, flagKey, defaultValue, evalCtx)
	if err != nil {
		slog.Warn("Failed to retrieve feature flag value", "key", flagKey, "error", err)
		return defaultValue
	}

	return value
}

// GetStringValue retrieves a string feature flag value
func GetStringValue(ctx context.Context, flagKey string, defaultValue string) string {
	if client == nil {
		return defaultValue
	}

	evalCtx := openfeature.NewEvaluationContext("", nil)
	value, err := client.StringValue(ctx, flagKey, defaultValue, evalCtx)
	if err != nil {
		slog.Warn("Failed to retrieve feature flag value", "key", flagKey, "error", err)
		return defaultValue
	}

	return value
}

// GetIntValue retrieves an integer feature flag value
func GetIntValue(ctx context.Context, flagKey string, defaultValue int64) int64 {
	if client == nil {
		return defaultValue
	}

	evalCtx := openfeature.NewEvaluationContext("", nil)
	value, err := client.IntValue(ctx, flagKey, defaultValue, evalCtx)
	if err != nil {
		slog.Warn("Failed to retrieve feature flag value", "key", flagKey, "error", err)
		return defaultValue
	}

	return value
}

// GetFloatValue retrieves a float feature flag value
func GetFloatValue(ctx context.Context, flagKey string, defaultValue float64) float64 {
	if client == nil {
		return defaultValue
	}

	evalCtx := openfeature.NewEvaluationContext("", nil)
	value, err := client.FloatValue(ctx, flagKey, defaultValue, evalCtx)
	if err != nil {
		slog.Warn("Failed to retrieve feature flag value", "key", flagKey, "error", err)
		return defaultValue
	}

	return value
}

// GetValueWithContext retrieves a feature flag value with a specific evaluation context
func GetValueWithContext(ctx context.Context, flagKey string, defaultValue interface{}, evalCtx openfeature.EvaluationContext) interface{} {
	if client == nil {
		return defaultValue
	}

	// Determine the type of the default value and call the appropriate method
	switch v := defaultValue.(type) {
	case bool:
		value, err := client.BooleanValue(ctx, flagKey, v, evalCtx)
		if err != nil {
			slog.Warn("Failed to retrieve boolean feature flag value", "key", flagKey, "error", err)
			return v
		}
		return value
	case string:
		value, err := client.StringValue(ctx, flagKey, v, evalCtx)
		if err != nil {
			slog.Warn("Failed to retrieve string feature flag value", "key", flagKey, "error", err)
			return v
		}
		return value
	case int:
		value, err := client.IntValue(ctx, flagKey, int64(v), evalCtx)
		if err != nil {
			slog.Warn("Failed to retrieve int feature flag value", "key", flagKey, "error", err)
			return v
		}
		return int(value)
	case int64:
		value, err := client.IntValue(ctx, flagKey, v, evalCtx)
		if err != nil {
			slog.Warn("Failed to retrieve int64 feature flag value", "key", flagKey, "error", err)
			return v
		}
		return value
	case float64:
		value, err := client.FloatValue(ctx, flagKey, v, evalCtx)
		if err != nil {
			slog.Warn("Failed to retrieve float64 feature flag value", "key", flagKey, "error", err)
			return v
		}
		return value
	default:
		// For other types, use ObjectValue (for structs, maps, etc.)
		value, err := client.ObjectValue(ctx, flagKey, v, evalCtx)
		if err != nil {
			slog.Warn("Failed to retrieve object feature flag value", "key", flagKey, "error", err)
			return v
		}
		return value
	}
}

// NewEvaluationContext creates a new evaluation context with user information
func NewEvaluationContext(userID string, attributes map[string]interface{}) openfeature.EvaluationContext {
	return openfeature.NewEvaluationContext(userID, attributes)
}
