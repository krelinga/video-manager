package vmerr

import (
	"errors"
	"fmt"
)


type Problem string

const (
	ProblemNotFound      Problem = "/errors/not-found"
	ProblemBadRequest    Problem = "/errors/bad-request"
	ProblemInternalError Problem = "/errors/internal-error"
	ProblemConflict      Problem = "/errors/conflict"
)

type HttpError struct {
	Problem Problem
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
		Problem:    ProblemNotFound,
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
		Problem:    ProblemBadRequest,
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
		Problem:    ProblemInternalError,
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
		Problem:    ProblemConflict,
		StatusCode: 409,
		Wrapped:    err,
	}
}
