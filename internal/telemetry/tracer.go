package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

var tracerKey = struct{}{}

// TracerFromContext returns the tracer stored in the provided context.
func TracerFromContext(ctx context.Context) (trace.Tracer, bool) { //nolint:ireturn
	tracer, ok := ctx.Value(tracerKey).(trace.Tracer)
	return tracer, ok
}

// WithTracer returns a new context.Context that contains the provided tracer.
func WithTracer(ctx context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(ctx, tracerKey, tracer)
}

// StartSpan creates a span and a context.Context containing the newly-created span.
//
// If the context.Context provided in `ctx` contains a Span then the newly-created
// Span will be a child of that span, otherwise it will be a root span. This behavior
// can be overridden by providing `WithNewRoot()` as a SpanOption, causing the
// newly-created Span to be a root span even if `ctx` contains a Span.
//
// When creating a Span it is recommended to provide all known span attributes using
// the `WithAttributes()` SpanOption as samplers will only have access to the
// attributes provided when a Span is created.
//
// Any Span that is created MUST also be ended. This is the responsibility of the user.
// Implementations of this API may leak memory or other resources if Spans are not ended.
func StartSpan( //nolint:ireturn
	ctx context.Context,
	name string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	tracer, ok := TracerFromContext(ctx)
	if !ok {
		panic("no tracer found in context")
	}

	return tracer.Start(ctx, name, opts...) //nolint:spancheck
}
