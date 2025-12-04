package vmerr

import (
	"errors"
	"fmt"
)

type HttpError struct {
	StatusCode int
	Wrapped    error
}

func (e *HttpError) Error() string {
	return e.Wrapped.Error()
}

func (e *HttpError) Unwrap() error {
	return e.Wrapped
}

var ErrPanicAlreadyWrapped = errors.New("error already wrapped")

func checkAlreadyWrapped(err error) {
	var target *HttpError
	if errors.As(err, &target) {
		panic(fmt.Errorf("%w: %v", ErrPanicAlreadyWrapped, err))
	}
}

func NotFound(err error) error {
	if err == nil {
		return nil
	}
	checkAlreadyWrapped(err)
	return &HttpError{
		StatusCode: 404,
		Wrapped:    err,
	}
}

func BadRequest(err error) error {
	if err == nil {
		return nil
	}
	checkAlreadyWrapped(err)
	return &HttpError{
		StatusCode: 400,
		Wrapped:    err,
	}
}

func InternalError(err error) error {
	if err == nil {
		return nil
	}
	checkAlreadyWrapped(err)
	return &HttpError{
		StatusCode: 500,
		Wrapped:    err,
	}
}

func Conflict(err error) error {
	if err == nil {
		return nil
	}
	checkAlreadyWrapped(err)
	return &HttpError{
		StatusCode: 409,
		Wrapped:    err,
	}
}
