package cmderrors

import "github.com/spf13/cobra"

type InvalidUsageError struct {
	Cmd     *cobra.Command
	Message string
}

func NewInvalidUsageError(cmd *cobra.Command, message string) InvalidUsageError {
	return InvalidUsageError{
		Cmd:     cmd,
		Message: message,
	}
}

func (e InvalidUsageError) Error() string {
	return e.Message
}
