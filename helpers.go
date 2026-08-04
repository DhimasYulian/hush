package hush

import "strings"

// EscapeLike escapes LIKE wildcard characters in a literal value so that user
// input is matched literally rather than interpreted as a pattern. It escapes
// the backslash, percent, and underscore characters using a backslash escape.
//
// Adapters must wrap the escaped value with the appropriate '%' placeholders
// for the pattern operators:
//
//	$contains:   "%" + EscapeLike(v) + "%"
//	$startsWith: EscapeLike(v) + "%"
//	$endsWith:   "%" + EscapeLike(v)
func EscapeLike(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\', '%', '_':
			sb.WriteByte('\\')
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}

// IsNullOperator reports whether the operator is $null, which maps to SQL's IS
// NULL. NULL operators carry no coerced values (Condition.Values is nil).
func IsNullOperator(op Operator) bool {
	return op == OpNull
}

// IsNotNullOperator reports whether the operator is $notNull, which maps to
// SQL's IS NOT NULL. NULL operators carry no coerced values
// (Condition.Values is nil).
func IsNotNullOperator(op Operator) bool {
	return op == OpNotNull
}
