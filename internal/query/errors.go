package query

import "errors"

// Cross-domain sentinels shared by build and validate phases.
var (
	// ErrInvalidPopulate indicates a populate parameter is malformed.
	ErrInvalidPopulate = errors.New("hush: invalid populate")
	// ErrInvalidPagination indicates a pagination parameter is malformed.
	ErrInvalidPagination = errors.New("hush: invalid pagination")
	// ErrInvalidAggregation indicates an aggregation parameter is malformed.
	ErrInvalidAggregation = errors.New("hush: invalid aggregation")
)

// Error is a structured validation error with context about what failed.
type Error struct {
	Kind     error
	Path     Path
	Field    string
	Operator Operator
	Message  string
}

// Error returns the error message, prefixed with the kind if no extra message.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	if e.Message == "" {
		return e.Kind.Error()
	}

	return e.Kind.Error() + ": " + e.Message
}

// Unwrap returns the underlying kind error.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Kind
}

// QueryError creates an Error with a kind and message.
func QueryError(kind error, message string) error {
	return &Error{Kind: kind, Message: message}
}

// PathError creates an Error with a kind, path, and message.
func PathError(kind error, path Path, message string) error {
	return &Error{Kind: kind, Path: append(Path(nil), path...), Message: message}
}

// FieldError creates an Error with a kind, field name, and message.
func FieldError(kind error, field, message string) error {
	return &Error{Kind: kind, Field: field, Message: message}
}

// OperatorError creates an Error with a kind, field, operator, and message.
func OperatorError(kind error, field string, op Operator, message string) error {
	return &Error{Kind: kind, Field: field, Operator: op, Message: message}
}
