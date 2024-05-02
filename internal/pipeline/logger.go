package pipeline

import (
	"github.com/spf13/cobra"

	"github.com/intility/minctl/internal/ux"
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
		ux.Ferror(cmd.ErrOrStderr(), err.Error())
	}

	return err
}
