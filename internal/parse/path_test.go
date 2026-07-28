package parse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePath(t *testing.T) {
	testCases := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "single segment",
			path: "fields",
			want: []string{
				"fields",
			},
		},
		{
			name: "array index",
			path: "fields[0]",
			want: []string{
				"fields",
				"0",
			},
		},
		{
			name: "simple filter",
			path: "filters[name][$eq]",
			want: []string{
				"filters",
				"name",
				"$eq",
			},
		},
		{
			name: "nested populate",
			path: "populate[department][fields][0]",
			want: []string{
				"populate",
				"department",
				"fields",
				"0",
			},
		},
		{
			name: "deep nesting",
			path: "filters[$or][0][department][company][name][$eq]",
			want: []string{
				"filters",
				"$or",
				"0",
				"department",
				"company",
				"name",
				"$eq",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParsePath(tc.path)

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestParsePath_Errors(t *testing.T) {
	testCases := []struct {
		name string
		path string
	}{
		{
			name: "empty string",
			path: "",
		},
		{
			name: "empty bracket",
			path: "filters[]",
		},
		{
			name: "double empty bracket",
			path: "filters[][]",
		},
		{
			name: "unclosed bracket",
			path: "filters[name",
		},
		{
			name: "unexpected closing bracket",
			path: "filters]",
		},
		{
			name: "double opening bracket",
			path: "filters[[name]",
		},
		{
			name: "double closing bracket",
			path: "filters[name]]",
		},
		{
			name: "empty nested bracket",
			path: "filters[][name]",
		},
		{
			name: "text after bracket",
			path: "filters[name]foo",
		},
		{
			name: "text after nested bracket",
			path: "filters[name][foo]bar",
		},
		{
			name: "starts with closing bracket",
			path: "][",
		},
		{
			name: "only brackets",
			path: "[]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParsePath(tc.path)

			require.Error(t, err)
		})
	}
}
