package pipeline

import (
	"context"

	"github.com/spf13/cobra"
)

type Executable interface {
	AddMiddleware(middlewares ...Middleware)
	Execute(ctx context.Context, args []string) int
}

type Middleware interface {
	preRun(cmd *cobra.Command, args []string)
	postRun(cmd *cobra.Command, args []string, runErr error)
}

func New(cmd *cobra.Command) Executable {
	return &executable{
		cmd:         cmd,
		middlewares: []Middleware{},
	}
}

type executable struct {
	cmd *cobra.Command

	middlewares []Middleware
}

var _ Executable = (*executable)(nil)

func (ex *executable) AddMiddleware(middlewares ...Middleware) {
	ex.middlewares = append(ex.middlewares, middlewares...)
}

func (ex *executable) Execute(ctx context.Context, args []string) int {
	// Ensure cobra uses the same arguments
	ex.cmd.SetContext(ctx)
	_ = ex.cmd.ParseFlags(args)

	// Run the 'pre' hooks
	for _, m := range ex.middlewares {
		m.preRun(ex.cmd, args)
	}

	// set args (needed in case caller transforms args in any way)
	ex.cmd.SetArgs(args)

	// Execute the cobra command:
	err := ex.cmd.Execute()

	// Run the 'post' hooks. Note that unlike the default PostRun cobra functionality these
	// run even if the command resulted in an error. This is useful when we still want to clean up
	// before the program exists, or we want to log something. The error, if any, gets passed
	// to the post hook.
	for i := len(ex.middlewares) - 1; i >= 0; i-- {
		ex.middlewares[i].postRun(ex.cmd, args, err)
	}

	if err != nil {
		// Logging is handled by the telemetry middleware
		return 1
	}

	return 0
}
