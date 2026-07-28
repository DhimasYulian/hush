package parse

import (
	"github.com/DhimasYulian/hush/internal/query"
)

// ParsePath splits a bracket-notation key into path segments.
// For example, "filters[name][$eq]" becomes ["filters", "name", "$eq"].
func ParsePath(key string) ([]string, error) {
	if key == "" {
		return nil, query.QueryError(ErrEmptyKey, "key must not be empty")
	}

	var (
		segments  []string
		current   []rune
		inBracket bool
	)

	for _, r := range key {
		switch r {
		case '[':
			if inBracket {
				return nil, query.QueryError(ErrInvalidSyntax, "unexpected '['")
			}

			if len(segments) == 0 {
				if len(current) == 0 {
					return nil, query.QueryError(ErrEmptySegment, "empty root segment")
				}

				segments = append(segments, string(current))
				current = current[:0]
			}

			inBracket = true

		case ']':
			if !inBracket {
				return nil, query.QueryError(ErrInvalidSyntax, "unexpected ']'")
			}

			if len(current) == 0 {
				return nil, query.QueryError(ErrEmptySegment, "empty bracket segment")
			}

			segments = append(segments, string(current))
			current = current[:0]
			inBracket = false

		default:
			if !inBracket && len(segments) > 0 {
				return nil, query.QueryError(ErrUnexpectedCharacter, "unexpected character")
			}

			current = append(current, r)
		}
	}

	if inBracket {
		return nil, query.QueryError(ErrInvalidSyntax, "missing closing bracket")
	}

	if len(current) > 0 {
		if len(segments) > 0 {
			return nil, query.QueryError(ErrUnexpectedCharacter, "unexpected trailing text")
		}

		segments = append(segments, string(current))
	}

	if len(segments) == 0 {
		return nil, query.QueryError(ErrEmptyKey, "key must not be empty")
	}

	return segments, nil
}
