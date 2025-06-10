// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package featureflags

import (
	"log/slog"
	"strings"

	fromenv "github.com/open-feature/go-sdk-contrib/providers/from-env/pkg"
	"github.com/open-feature/go-sdk/pkg/openfeature"
)

// EnvConfig holds the configuration for the environment variable provider
type EnvConfig struct {
	// Optional prefix for environment variables
	Prefix string `mapstructure:"prefix"`
	// Optional case-sensitivity flag
	CaseSensitive bool `mapstructure:"case_sensitive"`
}

// NewEnvProvider creates and returns a new environment variable provider
func NewEnvProvider(log *slog.Logger, config EnvConfig) (openfeature.FeatureProvider, error) {
	log.Info("Initializing environment variable provider", "prefix", config.Prefix, "case_sensitive", config.CaseSensitive)

	// Configure the environment variable provider options
	options := []fromenv.ProviderOption{}

	// If prefix is configured, create a custom mapper function
	if config.Prefix != "" {
		options = append(options, fromenv.WithFlagToEnvMapper(func(flagKey string) string {
			envKey := config.Prefix + flagKey
			if !config.CaseSensitive {
				envKey = strings.ToUpper(envKey)
			}
			return envKey
		}))
	}

	// Create and return the environment variable provider
	provider := fromenv.NewProvider(options...)
	return provider, nil
}