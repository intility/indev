package pipeline

import (
	"github.com/spf13/cobra"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/intility/icpctl/internal/telemetry"
)

func Trace() *TraceMiddleware {
	return &TraceMiddleware{
		provider:     nil,
		tracerCloser: func() {},
	}
}

type TraceMiddleware struct {
	provider     *tracesdk.TracerProvider
	tracerCloser func()
}

var _ Middleware = (*TraceMiddleware)(nil)

func (m *TraceMiddleware) Handle(cmd *cobra.Command, args []string, next NextFunc) error {
	tracer, ok := telemetry.TracerFromContext(cmd.Context())
	if !ok {
		return next(cmd, args)
	}

	subCmd, _, err := getSubcommand(cmd, args)
	if err != nil {
		return next(cmd, args)
	}

	name := subCmd.CommandPath()
	_, span := tracer.Start(cmd.Context(), name)

	defer span.End()

	return next(cmd, args)
}
