package build

import "errors"

var (
	// ErrInvalidFields indicates a fields parameter is malformed.
	ErrInvalidFields = errors.New("hush: invalid fields")
	// ErrInvalidSort indicates a sort parameter is malformed.
	ErrInvalidSort = errors.New("hush: invalid sort")
	// ErrInvalidFilters indicates a filter parameter is malformed.
	ErrInvalidFilters = errors.New("hush: invalid filters")
	// ErrInvalidGroupBy indicates a groupBy parameter is malformed.
	ErrInvalidGroupBy = errors.New("hush: invalid groupBy")
)
