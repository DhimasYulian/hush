package coerce

import (
	"testing"
	"time"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func TestCoerce(t *testing.T) {
	now := time.Date(2024, 5, 17, 12, 30, 45, 0, time.UTC)

	testCases := []struct {
		name    string
		typ     query.FieldType
		raw     string
		want    any
		wantErr bool
	}{
		{name: "string", typ: query.TypeString, raw: "hello world", want: "hello world"},
		{name: "number int", typ: query.TypeNumber, raw: "42", want: float64(42)},
		{name: "number float", typ: query.TypeNumber, raw: "3.14", want: 3.14},
		{name: "number invalid", typ: query.TypeNumber, raw: "abc", wantErr: true},
		{name: "bool true", typ: query.TypeBool, raw: "true", want: true},
		{name: "bool 1", typ: query.TypeBool, raw: "1", want: true},
		{name: "bool invalid", typ: query.TypeBool, raw: "yes", wantErr: true},
		{name: "date", typ: query.TypeDate, raw: "2024-05-17T12:30:45Z", want: now},
		{name: "date invalid", typ: query.TypeDate, raw: "2024-05-17", wantErr: true},
		{name: "unknown type", typ: query.FieldType("blob"), raw: "x", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Coerce(tc.typ, tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
