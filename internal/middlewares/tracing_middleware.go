package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"doit/internal/web"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "doit-api/http"

// Tracing creates a middleware that starts a span for each HTTP request.
// It extracts trace context from incoming headers and injects it into the response.
func Tracing() web.MiddleWare {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(handler web.Handler) web.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			// Extract trace context from incoming request headers
			// This allows distributed tracing across services
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Create span name: "HTTP GET /api/v1/todos"
			spanName := fmt.Sprintf("HTTP %s %s", r.Method, r.URL.Path)

			// Start a new span
			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.path", r.URL.Path),
					attribute.String("http.host", r.Host),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("http.remote_addr", r.RemoteAddr),
				),
			)
			defer span.End()

			// Store trace ID in context for logging correlation
			traceID := span.SpanContext().TraceID().String()
			ctx = web.SetTraceID(ctx, traceID)

			// Inject trace context into response headers
			// This helps with debugging and client-side correlation
			propagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))

			// Call the next handler with the traced context
			err := handler(w, r.WithContext(ctx))

			// Record the response status
			statusCode := web.GetStatusCode(ctx)
			span.SetAttributes(attribute.Int("http.status_code", statusCode))

			// Mark span as error if status >= 400
			if statusCode >= 400 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
			} else {
				span.SetStatus(codes.Ok, "")
			}

			// Record error if present
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}

			return err
		}
	}
}

// SpanFromContext returns the current span from context.
// Use this in handlers to add attributes or create child spans.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanAttributes adds attributes to the current span.
// Use this to add business-specific information to traces.
func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// AddSpanEvent adds an event to the current span.
// Events are like log entries within a span.
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}
