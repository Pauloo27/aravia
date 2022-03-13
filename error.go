package aravia

type HttpError struct {
	StatusCode HttpStatus
	Message    string
	Data       interface{}
}

func (e HttpError) String() string {
	return e.Message
}

func (e HttpError) Error() string {
	return e.String()
}
