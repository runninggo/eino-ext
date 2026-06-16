package tls

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/callbacks/tls/internal/otlptraceexporter"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type OtelProvider struct {
	TracerProvider *sdktrace.TracerProvider
}

func (p *OtelProvider) Shutdown(ctx context.Context) error {
	if p == nil || p.TracerProvider == nil {
		return nil
	}

	return p.TracerProvider.Shutdown(ctx)
}

func newTraceProvider(cfg *TLSConfig) (*OtelProvider, error) {
	traceExp, err := otlptraceexporter.New(cfg.TLSEndpoint, cfg.TLSOTLPHeader, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create otlp trace exporter: %w", err)
	}

	res := newTraceResource(cfg)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(traceExp)),
	)

	return &OtelProvider{TracerProvider: tracerProvider}, nil
}

func newTraceResource(cfg *TLSConfig) *resource.Resource {
	attrs := []attribute.KeyValue{
		attribute.String("service.name", cfg.AppName),
		attribute.String("tls.business_type", "gen_ai"),
	}
	if cfg.Release != "" {
		attrs = append(attrs, attribute.String("service.version", cfg.Release))
	}

	res, err := resource.New(
		context.Background(),
		resource.WithHost(),
		resource.WithFromEnv(),
		resource.WithProcessPID(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		return resource.Default()
	}

	return res
}
