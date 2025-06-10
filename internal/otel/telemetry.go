package otel

import (
	"context"
	"log/slog"
	"time"

	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/host"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	logglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/nlamirault/e2c/internal/utils"
	"github.com/nlamirault/e2c/internal/version"
)

// createResource creates a new OpenTelemetry resource with the application attributes
func createResource(ctx context.Context, cfg OpenTelemetryConfig) (*resource.Resource, error) {
	extraResources, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(version.GetVersion()),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
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

func buildHeaders(cfg OpenTelemetrySignalConfig) map[string]string {
	if len(cfg.Headers) == 0 {
		return nil
	}
	headers := make(map[string]string)
	for k, v := range cfg.Headers {
		headers[k] = v
	}
	return headers
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
		log.Info("Setup OpenTelemetry for logs")
		lp, err := initLogger(ctx, res, cfg.ServiceName, cfg.Logs, log)
		if err != nil {
			return err
		}
		defer func() {
			if err := lp.Shutdown(context.Background()); err != nil {
				log.Warn("Error shutting down OpenTelemtry logger provider", "error", err)
			}
		}()
		handlers := []slog.Handler{
			slog.Default().Handler(),
			otelslog.NewHandler(utils.APP_NAME),
		}
		slog.SetDefault(slog.New(slogmulti.Fanout(handlers...)))
		logglobal.SetLoggerProvider(lp)
	}

	if cfg.Traces.Enabled {
		log.Info("Setup OpenTelemetry for traces")
		tp, err := initTracer(ctx, res, cfg.Traces)
		if err != nil {
			return err
		}
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Warn("Error shutting down OpenTelemetry tracer provider", "error", err)
			}
		}()
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	}

	if cfg.Metrics.Enabled {
		log.Info("Setup OpenTelemetry for metrics")
		mp, err := initMeter(ctx, res, cfg.Metrics)
		if err != nil {
			return err
		}
		defer func() {
			if err := mp.Shutdown(context.Background()); err != nil {
				log.Warn("Error shutting down OpenTelemetry meter provider", "error", err)
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
		otel.SetMeterProvider(mp)
	}

	log.Debug("OpenTelemetry providers are setup")
	return nil
}

// Shutdown gracefully shuts down the OpenTelemetry SDK
func Shutdown(ctx context.Context, log *slog.Logger) {
	log.Info("Shutting down OpenTelemetry")

	if gtp, ok := otel.GetTracerProvider().(*tracesdk.TracerProvider); ok {
		log.Debug("Shutting down OpenTelemetry Log")
		gtp.ForceFlush(ctx)
		gtp.Shutdown(ctx)
	}

	if gmp, ok := otel.GetMeterProvider().(*metricsdk.MeterProvider); ok {
		log.Debug("Shutting down OpenTelemetry Metric")
		gmp.ForceFlush(ctx)
		gmp.Shutdown(ctx)
	}

	if glp, ok := logglobal.GetLoggerProvider().(*logsdk.LoggerProvider); ok {
		log.Debug("Shutting down OpenTelemetry Trace")
		glp.ForceFlush(ctx)
		glp.Shutdown(ctx)
	}
}
