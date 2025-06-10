// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

// Package otel provides OpenTelemetry integration for the e2c application.
package otel

import (
	"time"
)

// Protocol specifies the OTLP exporter protocol
type Protocol string

const (
	// ProtocolGRPC uses gRPC for the OTLP exporter
	ProtocolGRPC Protocol = "grpc"
	// ProtocolHTTP uses HTTP for the OTLP exporter
	ProtocolHTTP Protocol = "http"
)

type OpenTelemetrySignalConfig struct {
	// If this OpenTelemetry signal is enabled
	Enabled bool `bool:"enabled"`
	// Protocol is the OTLP exporter protocol (grpc or http)
	Protocol Protocol `mapstructure:"protocol"`
	// Endpoint is the OTLP exporter endpoint
	Endpoint string `mapstructure:"endpoint"`
	// Insecure disables TLS for the OTLP exporter
	Insecure bool `bool:"insecure"`
	// Headers are additional headers to send with the OTLP exporter
	Headers map[string]string `mapstructure:"headers"`
	// Timeout is the timeout for OTLP exporter operations
	Timeout time.Duration `mapstructure:"timeout"`
}

// OpenTelemetryConfig holds the configuration for OpenTelemetry
type OpenTelemetryConfig struct {
	// ServiceName is the name of the service
	ServiceName string `mapstructure:"service_name"`
	// Environment is the environment the service is running in
	Environment string                    `mapstructure:"environment"`
	Logs        OpenTelemetrySignalConfig `mapstructure:"logs"`
	Metrics     OpenTelemetrySignalConfig `mapstructure:"metrics"`
	Traces      OpenTelemetrySignalConfig `mapstructure:"traces"`
}
