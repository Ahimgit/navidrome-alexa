package httpclient

import "github.com/pkg/errors"

type HttpError struct {
	Status     string
	StatusCode int
	message    string
	cause      error
}

func NewHttpError(message string, cause error) error {
	return errors.WithStack(&HttpError{
		message: message,
		cause:   cause,
	})
}

func NewHttpErrorWithStatus(message string, status string, statusCode int) error {
	return errors.WithStack(&HttpError{
		message:    message,
		Status:     status,
		StatusCode: statusCode,
	})
}

func (he *HttpError) Error() string {
	if he.cause != nil {
		return he.message + ": " + he.cause.Error()
	} else {
		return he.message
	}
}

func (he *HttpError) Unwrap() error { return he.cause }

func IsAuthError(err error) bool {
	var httpError *HttpError
	return errors.As(err, &httpError) && httpError.StatusCode == 401
}
