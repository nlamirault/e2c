package otel

import (
	"context"
	"log/slog"
	"time"

	"github.com/nlamirault/e2c/internal/version"
	"go.opentelemetry.io/contrib/instrumentation/host"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
)

// OpenTelemetryConfig holds the configuration for OpenTelemetry
// type OpenTelemetryConfig struct {
// 	// The provider to use (configcat, env, devcycle)
// 	Provider LogsConfig `mapstructure:"logs"`
// 	// ConfigCat-specific configuration
// 	ConfigCat ConfigCatConfig `mapstructure:"metrics"`
// 	// Environment variable provider configuration
// 	Env EnvConfig `mapstructure:"env"`
// 	// DevCycle-specific configuration
// 	DevCycle DevCycleConfig `mapstructure:"traces"`
// 	// Enabled state for feature flags functionality
// 	Enabled bool `mapstructure:"enabled"`
// }

// createResource creates a new OpenTelemetry resource with the application attributes
func createResource(ctx context.Context, cfg OpenTelemetryConfig) (*resource.Resource, error) {
	extraResources, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(version.GetVersion()),
			attribute.String("environment", cfg.Environment),
		),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, err
	}
	r, _ := resource.Merge(
		resource.Default(),
		extraResources,
	)
	return r, nil
}

// InitializeTelemetry initializes the OpenTelemetry configuration
func InitializeTelemetry(ctx context.Context, log *slog.Logger, cfg OpenTelemetryConfig) error {
	log.Info("Initializing OpenTelemetry",
		"service", cfg.ServiceName,
		"version", version.GetVersion(),
		"environment", cfg.Environment,
		"logs", cfg.Logs,
		"metrics", cfg.Metrics,
		"traces", cfg.Traces,
	)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if !cfg.Logs.Enabled && !cfg.Metrics.Enabled && !cfg.Traces.Enabled {
		log.Info("OpenTelemetry is enabled but no signals are enabled (metrics, traces, logs)")
		return nil
	}

	res, err := createResource(ctx, cfg)
	if err != nil {
		return err
	}

	if cfg.Logs.Enabled {
		log.Error("OpenTelemetry Logs")
		lp, err := initLogger(ctx, res, cfg.ServiceName, cfg.Logs, log)
		if err != nil {
			return err
		}
		defer func() {
			if err := lp.Shutdown(context.Background()); err != nil {
				log.Warn("Error shutting down OpenTelemtry logger provider: %v", err)
			}
		}()
	}

	if cfg.Traces.Enabled {
		tp, err := initTracer(ctx, res, cfg.Traces)
		if err != nil {
			return err
		}
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Warn("Error shutting down OpenTelemetry tracer provider: %v", err)
			}
		}()
	}

	if cfg.Metrics.Enabled {
		mp, err := initMeter(ctx, res, cfg.Metrics)
		if err != nil {
			return err
		}
		defer func() {
			if err := mp.Shutdown(context.Background()); err != nil {
				log.Warn("Error shutting down OpenTelemetry meter provider: %v", err)
			}
		}()

		if err = otelruntime.Start(
			otelruntime.WithMinimumReadMemStatsInterval(time.Second),
			otelruntime.WithMeterProvider(mp),
		); err != nil {
			return err
		}

		if err = host.Start(host.WithMeterProvider(mp)); err != nil {
			return err
		}
	}

	log.Debug("OpenTelemetry providers are setup")
	return nil
}

// Shutdown gracefully shuts down the OpenTelemetry SDK
func Shutdown(ctx context.Context, log *slog.Logger) {
	log.Info("Shutting down OpenTelemetry")
}
