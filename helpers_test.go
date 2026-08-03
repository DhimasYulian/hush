package hush

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func intPtr(v int) *int    { return &v }
func boolPtr(v bool) *bool { return &v }

func TestEscapeLike(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "go", want: "go"},
		{name: "percent", in: "50%_off", want: `50\%\_off`},
		{name: "underscore", in: "a_b", want: `a\_b`},
		{name: "backslash", in: `a\b`, want: `a\\b`},
		{name: "all wildcards", in: `%_\`, want: `\%\_\\`},
		{name: "empty", in: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, EscapeLike(tt.in))
		})
	}
}

func TestNullSemantics(t *testing.T) {
	require.True(t, IsNullOperator(OpNull))
	require.False(t, IsNullOperator(OpNotNull))

	require.True(t, IsNotNullOperator(OpNotNull))
	require.False(t, IsNotNullOperator(OpNull))

	for _, op := range []Operator{OpEq, OpNe, OpIn, OpContains, OpGt} {
		require.False(t, IsNullOperator(op), op)
		require.False(t, IsNotNullOperator(op), op)
	}
}
