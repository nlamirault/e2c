// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package featureflags

import (
	"fmt"
	"log/slog"
	"time"

	dvc "github.com/devcyclehq/go-server-sdk/v2"
	"github.com/open-feature/go-sdk/pkg/openfeature"
)

// DevCycleConfig holds the configuration for DevCycle
type DevCycleConfig struct {
	// Server SDK key for DevCycle
	ServerKey string `mapstructure:"server_key"`
	// Optional configuration options
	EnableEdgeDB bool `mapstructure:"enable_edge_db"`
	// Optional configuration to enable cloud bucketing
	EnableCloudBucketing bool `mapstructure:"enable_cloud_bucketing"`
	// Optional timeout in seconds for DevCycle requests
	TimeoutSeconds int `mapstructure:"timeout_seconds"`
	// Optional interval in seconds for config polling
	ConfigPollingIntervalSeconds int `mapstructure:"config_polling_interval_seconds"`
	// Optional event flush interval in seconds
	EventFlushIntervalSeconds int `mapstructure:"event_flush_interval_seconds"`
	// Optional flag to disable automatic event logging
	DisableAutomaticEventLogging bool `mapstructure:"disable_automatic_event_logging"`
	// Optional flag to disable custom event logging
	DisableCustomEventLogging bool `mapstructure:"disable_custom_event_logging"`
}

// NewDevCycleProvider creates and returns a new DevCycle provider
func NewDevCycleProvider(log *slog.Logger, config DevCycleConfig) (openfeature.FeatureProvider, error) {
	if config.ServerKey == "" {
		return nil, fmt.Errorf("DevCycle server key is required")
	}

	log.Info("Initializing DevCycle provider",
		"server_key_length", len(config.ServerKey),
		"enable_edge_db", config.EnableEdgeDB,
		"timeout", config.TimeoutSeconds,
		"polling_interval", config.ConfigPollingIntervalSeconds)

	// Create DevCycle options
	options := dvc.Options{
		EnableEdgeDB:         config.EnableEdgeDB,
		EnableCloudBucketing: config.EnableCloudBucketing,
		DisableAutomaticEventLogging: config.DisableAutomaticEventLogging,
		DisableCustomEventLogging:    config.DisableCustomEventLogging,
	}

	// Set timeouts and intervals if configured
	if config.TimeoutSeconds > 0 {
		options.RequestTimeout = time.Duration(config.TimeoutSeconds) * time.Second
	}

	if config.ConfigPollingIntervalSeconds > 0 {
		options.ConfigPollingIntervalMS = time.Duration(config.ConfigPollingIntervalSeconds) * time.Second
	}

	if config.EventFlushIntervalSeconds > 0 {
		options.EventFlushIntervalMS = time.Duration(config.EventFlushIntervalSeconds) * time.Second
	}

	// Initialize DevCycle client
	dvcClient, err := dvc.NewClient(config.ServerKey, &options)
	if err != nil {
		return nil, fmt.Errorf("failed to create DevCycle client: %w", err)
	}

	// Create DevCycle OpenFeature provider
	provider := dvcClient.OpenFeatureProvider()

	return provider, nil
}