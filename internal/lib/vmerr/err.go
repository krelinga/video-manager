package vmerr

import (
	"errors"
	"fmt"
)

type httpError struct {
	statusCode int
	wrapped error
}

func (e *httpError) Error() string {
	return e.wrapped.Error()
}

var ErrPanicAlreadyWrapped = errors.New("error already wrapped")

func checkAlreadyWrapped(err error) {
	var target *httpError
	if errors.As(err, &target) {
		panic(fmt.Errorf("%w: %v", ErrPanicAlreadyWrapped, err))
	}
}

func NotFound(err error) error {
	checkAlreadyWrapped(err)
	return &httpError{
		statusCode: 404,
		wrapped:    err,
	}
}

func BadRequest(err error) error {
	checkAlreadyWrapped(err)
	return &httpError{
		statusCode: 400,
		wrapped:    err,
	}
}

func InternalError(err error) error {
	checkAlreadyWrapped(err)
	return &httpError{
		statusCode: 500,
		wrapped:    err,
	}
}