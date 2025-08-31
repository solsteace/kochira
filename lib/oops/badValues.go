package oops

type BadValues struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e BadValues) Error() string {
	if e.Msg == "" {
		return "Terdapat data yang tidak sesuai dengan ketentuan"
	}
	return e.Msg
}
