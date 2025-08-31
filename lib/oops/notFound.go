package oops

type NotFound struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e NotFound) Error() string {
	if e.Msg == "" {
		return "Data yang dicari tidak ditemukan"
	}
	return e.Msg
}
