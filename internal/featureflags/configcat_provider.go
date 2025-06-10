// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package featureflags

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	configcatsdk "github.com/configcat/go-sdk/v9"
	configcat "github.com/open-feature/go-sdk-contrib/providers/configcat/pkg"
	"github.com/open-feature/go-sdk/openfeature"
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

// NewConfigCatProvider creates and returns a new ConfigCat provider
func NewConfigCatProvider(log *slog.Logger, config ConfigCatConfig) (openfeature.FeatureProvider, error) {
	// Validate SDK key presence
	if config.SDKKey == "" {
		log.Error("ConfigCat initialization failed: empty SDK key", "provider", "configcat")
		return nil, fmt.Errorf("ConfigCat SDK key is required")
	}
	
	// Validate SDK key format (basic check for non-whitespace characters)
	trimmedKey := strings.TrimSpace(config.SDKKey)
	if trimmedKey == "" || len(trimmedKey) < 10 {
		log.Error("ConfigCat initialization failed: invalid SDK key format", "provider", "configcat")
		return nil, fmt.Errorf("ConfigCat SDK key format is invalid")
	}

	log.Info("Initializing ConfigCat provider",
		"sdk_key_length", len(config.SDKKey),
		"environment", config.Environment,
		"base_url", config.BaseURL,
		"cache_ttl", config.CacheTTLSeconds,
		"polling_interval", config.PollingIntervalSeconds)

	// Create ConfigCat client config
	clientConfig := configcatsdk.Config{
		SDKKey: config.SDKKey,
		Logger: configcatsdk.DefaultLogger(),
	}

	// Add optional configurations if provided
	if config.BaseURL != "" {
		clientConfig.BaseURL = config.BaseURL
	}

	if config.PollingIntervalSeconds > 0 {
		clientConfig.PollInterval = time.Duration(config.PollingIntervalSeconds) * time.Second
	}

	// Create ConfigCat client
	client := configcatsdk.NewCustomClient(clientConfig)

	// Create the ConfigCat provider
	provider := configcat.NewProvider(client)
	
	log.Info("ConfigCat provider initialized successfully", 
		"sdk_key_masked", maskSDKKey(config.SDKKey),
		"environment", config.Environment)

	return provider, nil
}

// maskSDKKey returns a masked version of the SDK key for safe logging
func maskSDKKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	// Show first 4 and last 4 characters, mask the rest
	return key[:4] + "..." + key[len(key)-4:]
}
