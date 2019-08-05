package oidc

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/hive/selfservice"
)

func TestToFormValues(t *testing.T) {
	for k, tc := range []struct {
		traits json.RawMessage
		expect url.Values
	}{
		{
			traits: json.RawMessage(`{ "foo": [{ "bar": "baz" }] }`),
			expect: url.Values{"foo.0.bar": {"baz"}},
		},
		{
			traits: json.RawMessage(`{ "foo": [{ "bar": "baz" },{ "bar": "baz" }], "baz": true, "baz1": 1234 }`),
			expect: url.Values{
				"foo.0.bar": {"baz"},
				"foo.1.bar": {"baz"},
				"baz":       {"true"},
				"baz1":      {"1234"},
			},
		},
		{
			traits: json.RawMessage(`[{ "bar": "baz" }]`),
			expect: url.Values{"0.bar": {"baz"}},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			assert.EqualValues(t, tc.expect, toFormValues(tc.traits))
		})
	}
}

func TestMerge(t *testing.T) {
	dc := selfservice.NewBodyDecoder()

	for k, tc := range []struct {
		form   string
		traits json.RawMessage
		expect json.RawMessage
	}{
		{
			form:   url.Values{}.Encode(),
			traits: json.RawMessage("{}"),
			expect: json.RawMessage("{}"),
		},
		{
			form:   url.Values{"traits.foo": {"bar"}, "traits.bool": {"false", "true"}}.Encode(),
			traits: json.RawMessage(`{"baz":"bar"}`),
			expect: json.RawMessage(`{"foo":"bar","baz":"bar","bool":true}`),
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			got, err := merge(dc, tc.form, tc.traits)
			require.NoError(t, err)
			assert.JSONEq(t, string(tc.expect), string(got))
		})
	}
}
