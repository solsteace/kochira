package oops

type Unauthorized struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e Unauthorized) Error() string {
	if e.Msg == "" {
		return "Anda perlu log-in untuk melakukan aksi ini"
	}
	return e.Msg
}
