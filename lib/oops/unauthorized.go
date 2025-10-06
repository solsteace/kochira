package oops

// An error equivalent to 401 Unauthorized HTTP error.
type Unauthorized struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e Unauthorized) Error() string {
	if e.Msg == "" {
		return "You need to login first in order to do this action"
	}
	return e.Msg
}
