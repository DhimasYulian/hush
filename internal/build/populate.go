package build

import (
	"strconv"

	"github.com/DhimasYulian/hush/internal/query"
)

// populateSyntax classifies how populate parameters are structured.
type populateSyntax int

const (
	populateSyntaxInvalid  populateSyntax = iota
	populateSyntaxIndexed                 // shorthand: populate[0]=author
	populateSyntaxRelation                // relation-keyed: populate[author][fields][0]=name
)

// BuildPopulate dispatches populate parameter building based on syntax mode.
// Returns the built populates, whether wildcard (*) was specified, and any error.
func BuildPopulate(params []query.Param) ([]query.Populate, bool, error) {
	wildcard, syntax := scanPopulateSyntax(params)

	if wildcard {
		if len(params) != 1 {
			return nil, false, query.QueryError(query.ErrInvalidPopulate, "populate=* must be used alone")
		}
		return nil, true, nil
	}

	switch syntax {
	case populateSyntaxIndexed:
		populates, err := buildPopulateIndexed(params)
		return populates, false, err
	case populateSyntaxRelation:
		populates, err := buildPopulateRelation(params)
		return populates, false, err
	default:
		return nil, false, query.QueryError(query.ErrInvalidPopulate, "cannot mix shorthand/indexed and relation-keyed syntax")
	}
}

// scanPopulateSyntax reports whether any param is the wildcard (*) and
// classifies the params as indexed or relation-keyed in a single pass.
// Wildcard detection takes precedence over syntax classification, matching the
// previous two-pass behavior.
func scanPopulateSyntax(params []query.Param) (wildcard bool, syntax populateSyntax) {
	if len(params) == 0 {
		return false, populateSyntaxIndexed
	}

	var hasIndexed, hasRelation bool

	for _, p := range params {
		switch {
		case len(p.Path) == 1:
			hasIndexed = true
			if p.Value == "*" {
				wildcard = true
			}

		default:
			if _, err := strconv.Atoi(p.Path[1]); err != nil {
				hasRelation = true
				continue
			}
			hasIndexed = true
			if len(p.Path) == 2 && p.Value == "*" {
				wildcard = true
			}
		}
	}

	switch {
	case hasIndexed && hasRelation:
		return wildcard, populateSyntaxInvalid
	case hasIndexed:
		return wildcard, populateSyntaxIndexed
	default:
		return wildcard, populateSyntaxRelation
	}
}

// buildPopulateIndexed handles the shorthand syntax: populate[0]=author.
func buildPopulateIndexed(params []query.Param) ([]query.Populate, error) {
	return BuildIndexed(params, query.ErrInvalidPopulate, func(value string) (query.Populate, error) {
		if value == "" {
			return query.Populate{}, query.QueryError(query.ErrInvalidPopulate, "empty relation")
		}
		return query.Populate{Relation: value}, nil
	})
}
