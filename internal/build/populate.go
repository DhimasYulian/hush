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
	if hasWildcard(params) {
		if len(params) != 1 {
			return nil, false, query.QueryError(query.ErrInvalidPopulate, "populate=* must be used alone")
		}
		return nil, true, nil
	}

	switch classifyPopulateSyntax(params) {
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

// hasWildcard reports whether any param represents a wildcard populate (*).
func hasWildcard(params []query.Param) bool {
	for _, p := range params {
		if p.Value != "*" {
			continue
		}

		if len(p.Path) == 1 {
			return true
		}

		if len(p.Path) == 2 {
			if _, err := strconv.Atoi(p.Path[1]); err == nil {
				return true
			}
		}
	}

	return false
}

// classifyPopulateSyntax determines whether params use indexed or relation-keyed syntax.
func classifyPopulateSyntax(params []query.Param) populateSyntax {
	if len(params) == 0 {
		return populateSyntaxIndexed
	}

	var hasIndexed, hasRelation bool

	for _, p := range params {
		switch {
		case len(p.Path) == 1:
			hasIndexed = true

		default:
			if _, err := strconv.Atoi(p.Path[1]); err == nil {
				hasIndexed = true
			} else {
				hasRelation = true
			}
		}

		if hasIndexed && hasRelation {
			return populateSyntaxInvalid
		}
	}

	if hasIndexed {
		return populateSyntaxIndexed
	}

	return populateSyntaxRelation
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
