package oops

// An error equivalent to 404 Not Found HTTP error.
type NotFound struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e NotFound) Error() string {
	if e.Msg == "" {
		return "The data you're looking for wasn't found in our system"
	}
	return e.Msg
}
