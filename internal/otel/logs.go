package otel

import (
	"context"
	"fmt"
	"log/slog"

	// stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func initLogger(ctx context.Context, resource *resource.Resource, serviceName string, cfg OpenTelemetrySignalConfig, log *slog.Logger) (*sdklog.LoggerProvider, error) {
	var otlpExporter sdklog.Exporter
	var err error
	log.Debug("OpenTelemetry Logs signals setup")
	switch cfg.Protocol {
	case ProtocolHTTP:
		otlpExporter, err = otlploghttp.New(
			ctx,
			otlploghttp.WithHeaders(buildHeaders(cfg)),
			otlploghttp.WithEndpointURL(cfg.Endpoint))
		if err != nil {
			return nil, err
		}
	case ProtocolGRPC:
		otlpExporter, err = otlploggrpc.New(
			ctx,
			otlploggrpc.WithHeaders(buildHeaders(cfg)),
			otlploggrpc.WithEndpointURL(cfg.Endpoint))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.Protocol)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(
			sdklog.NewBatchProcessor(otlpExporter),
		),
		sdklog.WithResource(resource),
	)

	defer lp.Shutdown(ctx)

	// global.SetLoggerProvider(lp)
	// logger := otelslog.NewLogger(serviceName)
	// logger.Debug("OpenTelemetry logging configured")
	return lp, nil
}
