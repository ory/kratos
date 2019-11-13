package oidc

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	for k, tc := range []struct {
		schema string
		form   string
		op     json.RawMessage
		expect json.RawMessage
	}{
		{
			form:   url.Values{}.Encode(),
			op:     json.RawMessage("{}"),
			expect: json.RawMessage("{}"),
		},
		{
			form:   url.Values{"traits.foo": {"bar"}, "traits.bool": {"false", "true"}, "traits.opv": {"blubb"}}.Encode(),
			op:     json.RawMessage(`{"baz":"bar","opv":"bla"}`),
			expect: json.RawMessage(`{"foo":"bar","baz":"bar","bool":true,"opv":"blubb"}`),
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			option, err := decoderRegistration(fmt.Sprintf("file://stub/merge/%d.schema.json", k))
			require.NoError(t, err)

			got, err := merge(tc.form, tc.op, option)
			require.NoError(t, err)
			assert.JSONEq(t, string(tc.expect), string(got))
		})
	}
}
