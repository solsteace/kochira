package oops

type Internal struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e Internal) Error() string {
	if e.Msg == "" {
		return "Mohon maaf, telah terjadi kesalahan pada sistem kami"
	}
	return e.Msg
}
