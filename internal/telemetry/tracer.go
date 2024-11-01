package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/intility/indev/internal/build"
	"github.com/intility/indev/internal/env"
	"github.com/intility/indev/internal/telemetry/exporters"
)

var tracerKey = struct{}{}

// TracerFromContext returns the tracer stored in the provided context.
func TracerFromContext(ctx context.Context) (trace.Tracer, bool) { //nolint:ireturn
	tracer, ok := ctx.Value(tracerKey).(trace.Tracer)
	return tracer, ok
}

// ContextWithTracer returns a new context.Context that contains the provided tracer.
func ContextWithTracer(ctx context.Context, tracer trace.Tracer) context.Context {
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

type ShutdownFunc func(context.Context) error

func InitTracer(ctx context.Context, attrs ...attribute.KeyValue) (trace.Tracer, ShutdownFunc, error) { //nolint:ireturn
	if env.DoNotTrack() {
		tp := noop.NewTracerProvider()
		otel.SetTracerProvider(tp)

		return tp.Tracer(build.AppName), func(ctx context.Context) error { return nil }, nil
	}

	attrs = append(attrs,
		semconv.ServiceName(build.AppName),
		semconv.ServiceVersionKey.String(build.Version),
	)

	r, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcessExecutableName(),
		resource.WithProcessRuntimeName(),
		resource.WithProcessRuntimeVersion(),
		resource.WithProcessRuntimeDescription(),
		resource.WithOS(),
		resource.WithProcessRuntimeName(),
		resource.WithProcessRuntimeVersion(),
		// resource.WithProcessCommandArgs(), // exposes sensitive data
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize resource: %w", err)
	}

	exp, err := exporters.NewTraceExporter()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize trace exporter: %w", err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithResource(r),
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exp),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Tracer(build.AppName), tp.Shutdown, nil
}
