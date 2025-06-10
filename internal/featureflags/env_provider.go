// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package featureflags

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/open-feature/go-sdk/pkg/openfeature"
)

// EnvConfig holds the configuration for the environment variable provider
type EnvConfig struct {
	// Optional prefix for environment variables
	Prefix string `mapstructure:"prefix"`
	// Optional case-sensitivity flag
	CaseSensitive bool `mapstructure:"case_sensitive"`
}

// envProvider is a simple implementation of the OpenFeature provider interface
// that reads feature flags from environment variables
type envProvider struct {
	prefix        string
	caseSensitive bool
	log           *slog.Logger
}

// Metadata returns provider metadata
func (p *envProvider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{
		Name: "env",
	}
}

// Hooks returns provider hooks
func (p *envProvider) Hooks() []openfeature.Hook {
	return nil
}

// getEnvVarName returns the environment variable name for a flag
func (p *envProvider) getEnvVarName(flag string) string {
	envName := p.prefix + flag
	if !p.caseSensitive {
		envName = strings.ToUpper(envName)
	}
	return envName
}

// BooleanEvaluation evaluates a boolean flag
func (p *envProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
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
func (p *envProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
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

// IntEvaluation evaluates an integer flag
func (p *envProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	envVar := os.Getenv(p.getEnvVarName(flag))
	if envVar == "" {
		return openfeature.IntResolutionDetail{
			Value: defaultValue,
		}
	}

	value, err := strconv.ParseInt(envVar, 10, 64)
	if err != nil {
		p.log.Warn("Failed to parse int environment variable", "key", flag, "value", envVar, "error", err)
		return openfeature.IntResolutionDetail{
			Value: defaultValue,
		}
	}

	return openfeature.IntResolutionDetail{
		Value: value,
	}
}

// FloatEvaluation evaluates a float flag
func (p *envProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	envVar := os.Getenv(p.getEnvVarName(flag))
	if envVar == "" {
		return openfeature.FloatResolutionDetail{
			Value: defaultValue,
		}
	}

	value, err := strconv.ParseFloat(envVar, 64)
	if err != nil {
		p.log.Warn("Failed to parse float environment variable", "key", flag, "value", envVar, "error", err)
		return openfeature.FloatResolutionDetail{
			Value: defaultValue,
		}
	}

	return openfeature.FloatResolutionDetail{
		Value: value,
	}
}

// ObjectEvaluation evaluates an object flag
func (p *envProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	// Environment variables can't represent complex objects directly
	return openfeature.InterfaceResolutionDetail{
		Value: defaultValue,
	}
}

// NewEnvProvider creates and returns a new environment variable provider
func NewEnvProvider(log *slog.Logger, config EnvConfig) (openfeature.FeatureProvider, error) {
	log.Info("Initializing environment variable provider", "prefix", config.Prefix, "case_sensitive", config.CaseSensitive)

	provider := &envProvider{
		prefix:        config.Prefix,
		caseSensitive: config.CaseSensitive,
		log:           log,
	}
	
	return provider, nil
}