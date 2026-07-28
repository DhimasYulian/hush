package hush

import (
	"net/url"

	"github.com/DhimasYulian/hush/internal/build"
	"github.com/DhimasYulian/hush/internal/parse"
	"github.com/DhimasYulian/hush/internal/validate"
)

// Parse processes URL query values through the three-stage pipeline
// (parse → build → validate) and returns a validated Query.
func Parse(values url.Values, root *Schema) (*Query, error) {
	params, err := parse.ParseParams(values)
	if err != nil {
		return nil, err
	}

	query, err := build.BuildQuery(params)
	if err != nil {
		return nil, err
	}

	if err := validate.Validate(query, root.inner); err != nil {
		return nil, err
	}

	return query, nil
}
