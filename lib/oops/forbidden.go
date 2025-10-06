package oops

// An error equivalent to 403 Forbidden HTTP error.
type Forbidden struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e Forbidden) Error() string {
	if e.Msg == "" {
		return "You don't have the permission to do this action"
	}
	return e.Msg
}
