package page

import "errors"

var (
	ErrMarshal = errors.New("failed to marshal page data")
	ErrUnmarshal = errors.New("failed to unmarshal page data")
	ErrDefSize = errors.New("invalid default page size")
	ErrListQuery = errors.New("invalid list query")
	ErrList = errors.New("failed to list items")
	ErrListOpts = errors.New("invalid list options")
	ErrListType = errors.New("invalid list type")
)