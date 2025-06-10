package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"

	// stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func initMeter(ctx context.Context, resource *resource.Resource, cfg OpenTelemetrySignalConfig) (*sdkmetric.MeterProvider, error) {
	var otlpExporter sdkmetric.Exporter
	var err error
	switch cfg.Protocol {
	case ProtocolHTTP:
		otlpExporter, err = otlpmetrichttp.New(ctx)
		if err != nil {
			return nil, err
		}
	case ProtocolGRPC:
		otlpExporter, err = otlpmetricgrpc.New(ctx)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.Protocol)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(otlpExporter)))
	// sdkmetric.WithReader(sdkmetric.NewPeriodicReader(otlpExporter)))
	if err != nil {
		return nil, err
	}

	otel.SetMeterProvider(mp)
	return mp, nil
}
