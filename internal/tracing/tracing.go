// Package tracing provides OpenTelemetry distributed tracing configuration.
package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0" // Match the SDK version
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds the tracing configuration.
type Config struct {
	ServiceName    string  // Name of your service (e.g., "doit-api")
	ServiceVersion string  // Version of your service (e.g., "1.0.0")
	Environment    string  // Environment (e.g., "development", "production")
	OTLPEndpoint   string  // Jaeger OTLP endpoint (e.g., "localhost:4317")
	SampleRate     float64 // Sampling rate (1.0 = 100%, 0.1 = 10%)
	Enabled        bool    // Enable/disable tracing
}

// Provider wraps the OpenTelemetry TracerProvider.
type Provider struct {
	provider *sdktrace.TracerProvider
}

// NewProvider creates and configures a new tracing provider.
func NewProvider(ctx context.Context, cfg Config) (*Provider, error) {
	// If tracing is disabled, return a no-op provider
	if !cfg.Enabled {
		return &Provider{provider: nil}, nil
	}

	// Create OTLP exporter (sends traces to Jaeger)
	conn, err := grpc.NewClient(
		cfg.OTLPEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource (metadata about this service)
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentName(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create sampler based on config
	var sampler sdktrace.Sampler
	if cfg.SampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SampleRate <= 0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
	}

	// Create TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global TracerProvider and Propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // W3C Trace Context
		propagation.Baggage{},      // W3C Baggage
	))

	return &Provider{provider: tp}, nil
}

// Tracer returns a named tracer for creating spans.
func (p *Provider) Tracer(name string) trace.Tracer {
	if p.provider == nil {
		return otel.Tracer(name) // Returns no-op tracer
	}
	return p.provider.Tracer(name)
}

// Shutdown gracefully shuts down the tracer provider.
// Call this when your application is shutting down.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.provider == nil {
		return nil
	}
	return p.provider.Shutdown(ctx)
}
