// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

// Package featureflags provides feature flag functionality using OpenFeature.
// It supports multiple providers including ConfigCat and environment variables.
package featureflags

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/open-feature/go-sdk/pkg/openfeature"
)

// ProviderType represents the type of feature flag provider to use
type ProviderType string

const (
	// ConfigCatProvider represents the ConfigCat provider
	ConfigCatProvider ProviderType = "configcat"
	// EnvProvider represents the environment variable provider
	EnvProvider ProviderType = "env"
)

var (
	client     *openfeature.Client
	clientOnce sync.Once
)

// FeatureFlagsConfig holds the configuration for feature flags
type FeatureFlagsConfig struct {
	// The provider to use (configcat, env)
	Provider ProviderType `mapstructure:"provider"`
	// ConfigCat-specific configuration
	ConfigCat ConfigCatConfig `mapstructure:"configcat"`
	// Environment variable provider configuration
	Env EnvConfig `mapstructure:"env"`
	// Enabled state for feature flags functionality
	Enabled bool `mapstructure:"enabled"`
}

// ConfigCatConfig holds the configuration for ConfigCat
type ConfigCatConfig struct {
	// SDK Key for ConfigCat
	SDKKey string `mapstructure:"sdk_key"`
	// Optional flag override environment
	Environment string `mapstructure:"environment"`
	// Optional ConfigCat base URL (for on-premise installations)
	BaseURL string `mapstructure:"base_url"`
	// Optional cache TTL in seconds
	CacheTTLSeconds int `mapstructure:"cache_ttl_seconds"`
	// Optional polling interval in seconds
	PollingIntervalSeconds int `mapstructure:"polling_interval_seconds"`
}

// EnvConfig holds the configuration for the environment variable provider
type EnvConfig struct {
	// Optional prefix for environment variables
	Prefix string `mapstructure:"prefix"`
	// Optional case-sensitivity flag
	CaseSensitive bool `mapstructure:"case_sensitive"`
}

// InitializeClient initializes the OpenFeature client with the specified provider
func InitializeClient(log *slog.Logger, config FeatureFlagsConfig) error {
	if !config.Enabled {
		log.Info("Feature flags are disabled, skipping initialization")
		return nil
	}

	// Reset the client if it was already initialized
	ResetClient()

	var err error
	clientOnce.Do(func() {
		log.Info("Initializing feature flag client", "provider", config.Provider)

		var provider openfeature.FeatureProvider
		var providerErr error

		switch config.Provider {
		case ConfigCatProvider:
			provider, providerErr = initializeConfigCatProvider(log, config.ConfigCat)
		case EnvProvider:
			provider, providerErr = initializeEnvProvider(log, config.Env)
		default:
			providerErr = fmt.Errorf("unsupported provider type: %s", config.Provider)
		}

		if providerErr != nil {
			err = providerErr
			return
		}

		// Set the provider at the global level
		setErr := openfeature.SetProvider(provider)
		if setErr != nil {
			err = fmt.Errorf("failed to set OpenFeature provider: %w", setErr)
			return
		}

		// Create a named client
		client = openfeature.NewClient("e2c")
		log.Info("Feature flag client initialized successfully", "provider", config.Provider)
	})

	return err
}

// initializeConfigCatProvider initializes and returns a ConfigCat provider
func initializeConfigCatProvider(log *slog.Logger, config ConfigCatConfig) (openfeature.FeatureProvider, error) {
	if config.SDKKey == "" {
		return nil, fmt.Errorf("ConfigCat SDK key is required")
	}

	log.Info("Initializing ConfigCat provider", "environment", config.Environment)

	// For now, return a stub provider that just returns the default values
	// The actual ConfigCat provider should be implemented with the proper SDK
	return &envVarProvider{
		prefix:        "CC_", // Placeholder prefix
		caseSensitive: false,
		providerName:  "configcat",
	}, nil
}

// initializeEnvProvider initializes and returns an environment variable provider
func initializeEnvProvider(log *slog.Logger, config EnvConfig) (openfeature.FeatureProvider, error) {
	log.Info("Initializing environment variable provider", "prefix", config.Prefix, "case_sensitive", config.CaseSensitive)

	return &envVarProvider{
		prefix:        config.Prefix,
		caseSensitive: config.CaseSensitive,
		providerName:  "env",
	}, nil
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

	value, err := client.BooleanValue(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("", nil))
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

	value, err := client.StringValue(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("", nil))
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

	value, err := client.IntValue(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("", nil))
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

	value, err := client.FloatValue(ctx, flagKey, defaultValue, openfeature.NewEvaluationContext("", nil))
	if err != nil {
		slog.Warn("Failed to retrieve feature flag value", "key", flagKey, "error", err)
		return defaultValue
	}

	return value
}

// envVarProvider is a simple implementation of the openfeature.FeatureProvider interface
// that reads feature flags from environment variables
type envVarProvider struct {
	prefix        string
	caseSensitive bool
	providerName  string
}

// Metadata returns the provider metadata
func (p *envVarProvider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{
		Name: p.providerName,
	}
}

// Hooks returns provider hooks
func (p *envVarProvider) Hooks() []openfeature.Hook {
	return nil
}

// getEnvVarName returns the environment variable name for a flag
func (p *envVarProvider) getEnvVarName(flag string) string {
	envName := p.prefix + flag
	if !p.caseSensitive {
		envName = strings.ToUpper(envName)
	}
	return envName
}

// BooleanEvaluation evaluates a boolean flag
func (p *envVarProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
	envVar := os.Getenv(p.getEnvVarName(flag))
	if envVar == "" {
		return openfeature.BoolResolutionDetail{
			Value: defaultValue,
		}
	}

	value := false
	if strings.ToLower(envVar) == "true" || envVar == "1" {
		value = true
	}

	return openfeature.BoolResolutionDetail{
		Value: value,
	}
}

// StringEvaluation evaluates a string flag
func (p *envVarProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
	envVar := os.Getenv(p.getEnvVarName(flag))
	if envVar == "" {
		return openfeature.StringResolutionDetail{
			Value: defaultValue,
		}
	}

	return openfeature.StringResolutionDetail{
		Value: envVar,
	}
}

// FloatEvaluation evaluates a float flag
func (p *envVarProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	envVar := os.Getenv(p.getEnvVarName(flag))
	if envVar == "" {
		return openfeature.FloatResolutionDetail{
			Value: defaultValue,
		}
	}

	value, err := strconv.ParseFloat(envVar, 64)
	if err != nil {
		return openfeature.FloatResolutionDetail{
			Value: defaultValue,
		}
	}

	return openfeature.FloatResolutionDetail{
		Value: value,
	}
}

// IntEvaluation evaluates an integer flag
func (p *envVarProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	envVar := os.Getenv(p.getEnvVarName(flag))
	if envVar == "" {
		return openfeature.IntResolutionDetail{
			Value: defaultValue,
		}
	}

	value, err := strconv.ParseInt(envVar, 10, 64)
	if err != nil {
		return openfeature.IntResolutionDetail{
			Value: defaultValue,
		}
	}

	return openfeature.IntResolutionDetail{
		Value: value,
	}
}

// ObjectEvaluation evaluates an object flag
func (p *envVarProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	// Environment variables can't represent complex objects directly
	return openfeature.InterfaceResolutionDetail{
		Value: defaultValue,
	}
}