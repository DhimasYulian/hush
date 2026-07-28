package parse

import "errors"

var (
	// ErrEmptyKey indicates a query string key is empty.
	ErrEmptyKey = errors.New("hush: empty key")
	// ErrEmptySegment indicates a path segment is empty.
	ErrEmptySegment = errors.New("hush: empty path segment")
	// ErrInvalidSyntax indicates a bracket path has invalid syntax.
	ErrInvalidSyntax = errors.New("hush: invalid LHS syntax")
	// ErrUnexpectedCharacter indicates an unexpected character in a path.
	ErrUnexpectedCharacter = errors.New("hush: unexpected character")
)
