package oops

// An error equivalent to 500 Internal Server Error HTTP error.
type Internal struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e Internal) Error() string {
	if e.Msg == "" {
		return "Sorry, an unidentified error just happened on our system"
	}
	return e.Msg
}
