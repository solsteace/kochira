package oops

type Uncategorized struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e Uncategorized) Error() string {
	if e.Msg == "" {
		return "Terjadi kesalahan yang belum dikategorikan"
	}
	return e.Msg
}
