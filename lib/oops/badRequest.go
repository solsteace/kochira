package oops

// An error equivalent to 400 Bad Request HTTP error.
type BadRequest struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e BadRequest) Error() string {
	if e.Msg == "" {
		return "We cannot process your request as it may had been malformed"
	}
	return e.Msg
}
