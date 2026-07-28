package schema

import "errors"

var (
	// ErrEmptyName indicates a schema or field name is empty.
	ErrEmptyName = errors.New("hush: schema name must not be empty")
	// ErrNoOperators indicates a filterable field has no allowed operators.
	ErrNoOperators = errors.New("hush: field must allow at least one operator")
	// ErrUnknownOperator indicates an operator is not recognized.
	ErrUnknownOperator = errors.New("hush: unknown operator")
	// ErrDuplicateField indicates a field is declared twice in the same schema.
	ErrDuplicateField = errors.New("hush: field already declared")
	// ErrDuplicateRelation indicates a relation is declared twice in the same schema.
	ErrDuplicateRelation = errors.New("hush: relation already declared")
	// ErrNilTarget indicates a relation target schema is nil.
	ErrNilTarget = errors.New("hush: relation target must not be nil")
	// ErrInvalidMaxDepth indicates a relation max depth is less than 1.
	ErrInvalidMaxDepth = errors.New("hush: relation max depth must be >= 1")
	// ErrInvalidMaxLimit indicates a max limit is out of range.
	ErrInvalidMaxLimit = errors.New("hush: max limit must be between 1 and AbsoluteMaxLimit")
	// ErrUnknownFieldType indicates a field type is not recognized.
	ErrUnknownFieldType = errors.New("hush: unknown field type")
)
