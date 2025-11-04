package errors

// HttpError adalah custom error dengan status kode HTTP
type HttpError struct {
	Code    int
	Message string
}

func (e *HttpError) Error() string {
	return e.Message
}

// NewHttpError membantu membuat instance HttpError dengan mudah
func NewHttpError(code int, message string) *HttpError {
	return &HttpError{
		Code:    code,
		Message: message,
	}
}
