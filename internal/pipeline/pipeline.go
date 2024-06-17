package pipeline

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/intility/icpctl/internal/build"
	"github.com/intility/icpctl/internal/env"
	"github.com/intility/icpctl/internal/telemetry"
	"github.com/intility/icpctl/internal/telemetry/exporters"
)

type Executable interface {
	AddMiddleware(middlewares ...Middleware)
	Execute(ctx context.Context, args []string) int
}

type NextFunc func(cmd *cobra.Command, args []string) error

type Middleware interface {
	Handle(cmd *cobra.Command, args []string, next NextFunc) error
}

func New(cmd *cobra.Command) Executable {
	return &executable{
		cmd:            cmd,
		tracerProvider: nil,
		middlewares:    []Middleware{},
	}
}

type executable struct {
	cmd *cobra.Command

	tracerProvider *tracesdk.TracerProvider

	middlewares []Middleware
}

var _ Executable = (*executable)(nil)

func (ex *executable) AddMiddleware(middlewares ...Middleware) {
	ex.middlewares = append(ex.middlewares, middlewares...)
}

func (ex *executable) Execute(ctx context.Context, args []string) int {
	err := ex.executeInstrumented(ctx, args)
	// logging in handled in the middlewares
	if err != nil {
		return 1
	}

	return 0
}

func (ex *executable) execute(ctx context.Context, args []string) error {
	ex.cmd.SetContext(ctx)
	_ = ex.cmd.ParseFlags(args)

	// wrap the cmd.Execute as the innermost middleware
	next := func(cm *cobra.Command, args []string) error {
		return cm.Execute()
	}

	// compose the middleware chain
	for i := len(ex.middlewares) - 1; i >= 0; i-- {
		next = func(next NextFunc, middleware Middleware) NextFunc {
			return func(cmd *cobra.Command, args []string) error {
				return middleware.Handle(cmd, args, next)
			}
		}(next, ex.middlewares[i])
	}

	// set args (needed in case caller transforms args in any way)
	ex.cmd.SetArgs(args)

	// Execute the pipeline
	return next(ex.cmd, args)
}

func (ex *executable) executeInstrumented(ctx context.Context, args []string) error {
	telemetry.Start()
	defer telemetry.Stop()

	tracer, shutdown, err := ex.initTracer(ctx, semconv.ProcessCommandArgs(args...))
	if err != nil {
		if build.IsDev {
			panic(err)
		}

		defer func() { _ = shutdown(ctx) }()

		return ex.execute(ctx, args)
	}

	defer func() { _ = shutdown(ctx) }()

	var span trace.Span
	ctx, span = tracer.Start(ctx, "cli.command")

	defer span.End()

	ctx = telemetry.WithTracer(ctx, tracer)

	err = ex.execute(ctx, args)

	switch err {
	case nil:
		span.SetStatus(codes.Ok, "Command succeeded")
	default:
		span.SetStatus(codes.Error, "Command failed")
		span.RecordError(err)
	}

	span.AddEvent("completed")

	return err
}

type ShutdownFunc func(context.Context) error

func (ex *executable) initTracer(ctx context.Context, attrs ...attribute.KeyValue) (trace.Tracer, ShutdownFunc, error) {
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
		return nil, nil, fmt.Errorf("failed to initialize stdouttrace exporter: %w", err)
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
