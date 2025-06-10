// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package featureflags

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/open-feature/go-sdk/pkg/openfeature"
)

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

// configCatProvider is a simple implementation of the OpenFeature provider interface
// that uses ConfigCat as the underlying feature flag system
type configCatProvider struct {
	log    *slog.Logger
	config ConfigCatConfig
}

// Metadata returns provider metadata
func (p *configCatProvider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{
		Name: "configcat",
	}
}

// Hooks returns provider hooks
func (p *configCatProvider) Hooks() []openfeature.Hook {
	return nil
}

// BooleanEvaluation evaluates a boolean flag
func (p *configCatProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
	// This is a stub implementation that just returns the default value
	// In a real implementation, this would call the ConfigCat SDK to get the actual value
	p.log.Debug("ConfigCat flag evaluation", "flag", flag, "type", "boolean", "default", defaultValue)
	
	return openfeature.BoolResolutionDetail{
		Value: defaultValue,
	}
}

// StringEvaluation evaluates a string flag
func (p *configCatProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
	// This is a stub implementation that just returns the default value
	// In a real implementation, this would call the ConfigCat SDK to get the actual value
	p.log.Debug("ConfigCat flag evaluation", "flag", flag, "type", "string", "default", defaultValue)
	
	return openfeature.StringResolutionDetail{
		Value: defaultValue,
	}
}

// IntEvaluation evaluates an integer flag
func (p *configCatProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	// This is a stub implementation that just returns the default value
	// In a real implementation, this would call the ConfigCat SDK to get the actual value
	p.log.Debug("ConfigCat flag evaluation", "flag", flag, "type", "int", "default", defaultValue)
	
	return openfeature.IntResolutionDetail{
		Value: defaultValue,
	}
}

// FloatEvaluation evaluates a float flag
func (p *configCatProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	// This is a stub implementation that just returns the default value
	// In a real implementation, this would call the ConfigCat SDK to get the actual value
	p.log.Debug("ConfigCat flag evaluation", "flag", flag, "type", "float", "default", defaultValue)
	
	return openfeature.FloatResolutionDetail{
		Value: defaultValue,
	}
}

// ObjectEvaluation evaluates an object flag
func (p *configCatProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	// This is a stub implementation that just returns the default value
	// In a real implementation, this would call the ConfigCat SDK to get the actual value
	p.log.Debug("ConfigCat flag evaluation", "flag", flag, "type", "object", "default", fmt.Sprintf("%v", defaultValue))
	
	return openfeature.InterfaceResolutionDetail{
		Value: defaultValue,
	}
}

// NewConfigCatProvider creates and returns a new ConfigCat provider
func NewConfigCatProvider(log *slog.Logger, config ConfigCatConfig) (openfeature.FeatureProvider, error) {
	if config.SDKKey == "" {
		return nil, fmt.Errorf("ConfigCat SDK key is required")
	}

	log.Info("Initializing ConfigCat provider", 
		"sdk_key_length", len(config.SDKKey),
		"environment", config.Environment,
		"base_url", config.BaseURL,
		"cache_ttl", config.CacheTTLSeconds,
		"polling_interval", config.PollingIntervalSeconds)

	// Create a stub provider implementation
	// In a real implementation, this would initialize the ConfigCat SDK client
	provider := &configCatProvider{
		log:    log,
		config: config,
	}

	return provider, nil
}