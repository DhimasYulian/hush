package validate

import "errors"

var (
	// ErrInvalidPath indicates a filter or sort path is invalid.
	ErrInvalidPath = errors.New("hush: invalid path")
	// ErrUnknownField indicates a field is not declared in the schema.
	ErrUnknownField = errors.New("hush: unknown field")
	// ErrOperatorNotAllowed indicates an operator is not allowed on a field.
	ErrOperatorNotAllowed = errors.New("hush: operator not allowed")
	// ErrNestingTooDeep indicates relation nesting exceeds the allowed depth.
	ErrNestingTooDeep = errors.New("hush: nesting too deep")
	// ErrInvalidValue indicates a filter value doesn't match the field type.
	ErrInvalidValue = errors.New("hush: invalid value")
	// ErrUnknownFilterNode indicates a filter node type is not recognized.
	ErrUnknownFilterNode = errors.New("hush: unknown filter node")
	// ErrMissingSchema indicates the root schema is nil.
	ErrMissingSchema = errors.New("hush: schema must not be nil")
	// ErrUnknownGroupBy indicates a groupBy field is not declared in the schema.
	ErrUnknownGroupBy = errors.New("hush: unknown groupBy field")
)
