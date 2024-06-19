package pipeline

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/intility/icpctl/internal/ux"
)

func Logger() *LoggerMiddleware {
	return &LoggerMiddleware{}
}

type LoggerMiddleware struct{}

var _ Middleware = (*LoggerMiddleware)(nil)

func (m *LoggerMiddleware) Handle(cmd *cobra.Command, args []string, next NextFunc) error {
	err := next(cmd, args)
	if err != nil {
		// We can introduce warnings here if needed.
		switch {
		case errors.Is(err, context.Canceled):
			ux.Fprint(cmd.OutOrStdout(), "Operation was canceled.")
		default:
			ux.Ferror(cmd.ErrOrStderr(), err.Error()+"\n")
		}
	}

	return err
}
