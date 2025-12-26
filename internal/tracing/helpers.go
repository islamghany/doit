package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracer names for different components
const (
	TracerHTTP    = "doit-api/http"
	TracerDB      = "doit-api/db"
	TracerCache   = "doit-api/cache"
	TracerService = "doit-api/service"
)

// StartSpan starts a new span with the given name.
// Always defer span.End() after calling this.
func StartSpan(ctx context.Context, tracerName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, spanName, opts...)
}

// StartDBSpan starts a span for database operations.
func StartDBSpan(ctx context.Context, operation, table string) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(TracerDB).Start(ctx, "db."+operation,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", operation),
		attribute.String("db.sql.table", table),
	)
	return ctx, span
}

// StartCacheSpan starts a span for cache operations.
func StartCacheSpan(ctx context.Context, operation, key string) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(TracerCache).Start(ctx, "cache."+operation,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	span.SetAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("db.operation", operation),
		attribute.String("cache.key", key),
	)
	return ctx, span
}

// StartServiceSpan starts a span for service/business logic operations.
func StartServiceSpan(ctx context.Context, service, operation string) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(TracerService).Start(ctx, service+"."+operation,
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	return ctx, span
}

// RecordError records an error on the span and sets the status to error.
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetOK sets the span status to OK.
func SetOK(span trace.Span) {
	span.SetStatus(codes.Ok, "")
}

// SetAttributes adds attributes to a span.
func SetAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// Int creates an int attribute.
func Int(key string, value int) attribute.KeyValue {
	return attribute.Int(key, value)
}

// Int64 creates an int64 attribute.
func Int64(key string, value int64) attribute.KeyValue {
	return attribute.Int64(key, value)
}

// String creates a string attribute.
func String(key, value string) attribute.KeyValue {
	return attribute.String(key, value)
}

// Bool creates a bool attribute.
func Bool(key string, value bool) attribute.KeyValue {
	return attribute.Bool(key, value)
}

// AddEvent adds an event to the span (like a log entry).
func AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// CacheHit records a cache hit.
func CacheHit(span trace.Span) {
	span.SetAttributes(attribute.Bool("cache.hit", true))
}

// CacheMiss records a cache miss.
func CacheMiss(span trace.Span) {
	span.SetAttributes(attribute.Bool("cache.hit", false))
}
