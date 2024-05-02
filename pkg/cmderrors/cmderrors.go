package cmderrors

type NotSignedInError struct {
	Message string
}

func (e *NotSignedInError) Error() string {
	return e.Message
}

//goland:noinspection GoUnusedExportedFunction
func NewNotSignedInError(message string) error {
	return &NotSignedInError{
		Message: message,
	}
}
