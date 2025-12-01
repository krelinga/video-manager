package page

import "errors"

// These are all errors that this package can panic with.
var (
	ErrMarshal = errors.New("failed to marshal page data")
	ErrDefSize = errors.New("invalid default page size")
	ErrOpts    = errors.New("invalid options")
	ErrType    = errors.New("invalid type")
)
