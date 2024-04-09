package cli

import (
	"os"

	"sigs.k8s.io/kind/pkg/log"
)

func Operation(operation string, function func()) {
	spinner := NewSpinner(os.Stderr)
	logger := NewLogger(spinner, log.Level(1))
	status := StatusForLogger(logger)
	status.Start(operation)
	function()
	status.End(true)
}
