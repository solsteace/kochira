package oops

type BadRequest struct {
	// Message to be sent to client
	Msg string

	// Actual error
	Err error
}

func (e BadRequest) Error() string {
	if e.Msg == "" {
		return "Data yang diberikan tidak cukup untuk diproses"
	}
	return e.Msg
}
