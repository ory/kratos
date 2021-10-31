package oidc

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	for k, tc := range []struct {
		schema string
		form   json.RawMessage
		op     json.RawMessage
		expect json.RawMessage
	}{
		{
			form:   json.RawMessage("{}"),
			op:     json.RawMessage("{}"),
			expect: json.RawMessage("{}"),
		},
		{
			form:   json.RawMessage(`{"foo": "bar", "bool": true, "opv": "blubb"}`),
			op:     json.RawMessage(`{"baz":"bar","opv":"bla"}`),
			expect: json.RawMessage(`{"foo":"bar","baz":"bar","bool":true,"opv":"blubb"}`),
		},
		{
			form:   json.RawMessage(`{"bool": true, "opv": "blubb"}`),
			op:     json.RawMessage(`{"foo":"bar","baz":"bar","opv":"bla"}`),
			expect: json.RawMessage(`{"foo":"bar","baz":"bar","bool":true,"opv":"blubb"}`),
		},
		{
			form:   json.RawMessage(`{"bool": true, "opv": "blubb"}`),
			op:     json.RawMessage(`{"foo":"bar","baz":"bar","opv":"bla"}`),
			expect: json.RawMessage(`{"foo":"bar","baz":"bar","bool":true,"opv":"blubb"}`),
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			got, err := merge(tc.form, tc.op)
			require.NoError(t, err)
			assert.JSONEq(t, string(tc.expect), string(got))
		})
	}
}
