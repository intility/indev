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

func (m *LoggerMiddleware) preRun(_ *cobra.Command, _ []string) {
}

func (m *LoggerMiddleware) postRun(cmd *cobra.Command, args []string, runErr error) {
	if runErr == nil {
		return
	}

	// We can introduce warnings here if needed.

	ux.Ferror(cmd.ErrOrStderr(), runErr.Error())
}
