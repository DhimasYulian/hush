package validate

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func cond(path query.Path, op query.Operator, value ...string) query.Condition {
	return query.Condition{Path: path, Operator: op, Value: value}
}

func wrapNot(f query.Filter, n int) query.Filter {
	for i := 0; i < n; i++ {
		f = query.Not{Filter: f}
	}
	return f
}

func TestValidateFilter(t *testing.T) {
	article, _, _ := fixtureSchemas(t)

	testCases := []struct {
		name    string
		filter  query.Filter
		wantErr error
	}{
		{
			name:   "nil filter is a no-op",
			filter: nil,
		},
		{
			name:   "valid top-level condition",
			filter: cond(query.Path{"title"}, query.OpEq, "hello"),
		},
		{
			name:   "valid condition with a second allowed operator",
			filter: cond(query.Path{"title"}, query.OpContains, "hel"),
		},
		{
			name:    "operator not allowed for field",
			filter:  cond(query.Path{"views"}, query.OpContains, "5"),
			wantErr: ErrOperatorNotAllowed,
		},
		{
			name:    "unknown top-level field",
			filter:  cond(query.Path{"nonexistent"}, query.OpEq, "x"),
			wantErr: ErrUnknownField,
		},
		{
			name:   "condition through one relation",
			filter: cond(query.Path{"author", "name"}, query.OpEq, "Jane"),
		},
		{
			name:    "condition through relation with unknown nested field",
			filter:  cond(query.Path{"author", "nonexistent"}, query.OpEq, "x"),
			wantErr: ErrUnknownField,
		},
		{
			name:   "condition through two relations",
			filter: cond(query.Path{"author", "profile", "bio"}, query.OpContains, "hi"),
		},
		{
			name: "valid $and of two conditions",
			filter: query.And{Filters: []query.Filter{
				cond(query.Path{"title"}, query.OpEq, "a"),
				cond(query.Path{"views"}, query.OpGt, "1"),
			}},
		},
		{
			name: "invalid condition nested inside $and is caught",
			filter: query.And{Filters: []query.Filter{
				cond(query.Path{"title"}, query.OpEq, "a"),
				cond(query.Path{"views"}, query.OpContains, "1"), // not allowed
			}},
			wantErr: ErrOperatorNotAllowed,
		},
		{
			name: "valid $or nested inside $and",
			filter: query.And{Filters: []query.Filter{
				query.Or{Filters: []query.Filter{
					cond(query.Path{"title"}, query.OpEq, "a"),
					cond(query.Path{"author", "name"}, query.OpEq, "b"),
				}},
				cond(query.Path{"views"}, query.OpGt, "1"),
			}},
		},
		{
			name:   "valid $not",
			filter: query.Not{Filter: cond(query.Path{"title"}, query.OpEq, "a")},
		},
		{
			name:   "$not nesting exactly at the depth cap is fine",
			filter: wrapNot(cond(query.Path{"title"}, query.OpEq, "a"), maxLogicalDepth),
		},
		{
			name:    "$not nesting one past the depth cap is rejected",
			filter:  wrapNot(cond(query.Path{"title"}, query.OpEq, "a"), maxLogicalDepth+1),
			wantErr: ErrNestingTooDeep,
		},
		{
			name:   "valid number value",
			filter: cond(query.Path{"views"}, query.OpGt, "42"),
		},
		{
			name:    "non-numeric value on a number field",
			filter:  cond(query.Path{"views"}, query.OpEq, "abc"),
			wantErr: ErrInvalidValue,
		},
		{
			name:   "valid date value",
			filter: cond(query.Path{"publishedAt"}, query.OpGt, "2024-01-15T00:00:00Z"),
		},
		{
			name:    "non-date value on a date field",
			filter:  cond(query.Path{"publishedAt"}, query.OpEq, "not-a-date"),
			wantErr: ErrInvalidValue,
		},
		{
			name:   "valid bool value",
			filter: cond(query.Path{"active"}, query.OpEq, "true"),
		},
		{
			name:    "non-bool value on a bool field",
			filter:  cond(query.Path{"active"}, query.OpEq, "yes"),
			wantErr: ErrInvalidValue,
		},
		{
			// $null's value is a presence flag, not data of the field's
			// own type, so a date field still expects a bool here.
			name:   "$null on a date field expects a bool flag, not a date",
			filter: cond(query.Path{"publishedAt"}, query.OpNull, "true"),
		},
		{
			name:    "$null with a non-bool flag is rejected",
			filter:  cond(query.Path{"publishedAt"}, query.OpNull, "whenever"),
			wantErr: ErrInvalidValue,
		},
		{
			name:   "string field accepts any value",
			filter: cond(query.Path{"title"}, query.OpEq, "anything at all 123"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateFilter(tc.filter, article)

			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
