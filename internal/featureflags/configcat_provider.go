// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package featureflags

import (
	"fmt"
	"log/slog"
	"time"

	configcatsdk "github.com/configcat/go-sdk/v9"
	"github.com/open-feature/go-sdk-contrib/providers/configcat/pkg"
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

	// Create ConfigCat client config
	clientConfig := configcatsdk.Config{
		SDKKey: config.SDKKey,
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

	return provider, nil
}