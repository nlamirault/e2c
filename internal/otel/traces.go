package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	// stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer(ctx context.Context, resource *resource.Resource, cfg OpenTelemetrySignalConfig) (*sdktrace.TracerProvider, error) {
	// stdoutExporter, err := stdout.New(stdout.WithPrettyPrint())
	// if err != nil {
	// 	return nil, err
	// }

	var otlpExporter sdktrace.SpanExporter
	var err error
	switch cfg.Protocol {
	case ProtocolHTTP:
		otlpExporter, err = otlptracehttp.New(ctx)
		if err != nil {
			return nil, err
		}
	case ProtocolGRPC:
		otlpExporter, err = otlptracegrpc.New(ctx)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.Protocol)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(otlpExporter),
		// sdktrace.WithBatcher(stdoutExporter),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}
