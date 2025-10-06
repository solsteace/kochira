package oops

// An error that indicates certain user-inputted data isn't valid from
// business rules point of view
type BadValues struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e BadValues) Error() string {
	if e.Msg == "" {
		return "A data that doesn't comply with our standards had found"
	}
	return e.Msg
}
