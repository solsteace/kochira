package oops

type Forbidden struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e Forbidden) Error() string {
	if e.Msg == "" {
		return "Anda tidak memiliki izin melakukan aksi ini"
	}
	return e.Msg
}
