package oidc

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/selfservice"
)

func TestToFormValues(t *testing.T) {
	for k, tc := range []struct {
		json   json.RawMessage
		prefix string
		expect url.Values
	}{
		{
			json:   json.RawMessage(`{ "foo": [{ "bar": "baz" }] }`),
			expect: url.Values{"foo.0.bar": {"baz"}},
		},
		{
			json: json.RawMessage(`{ "foo": [{ "bar": "baz" },{ "bar": "baz" }], "baz": true, "baz1": 1234 }`),
			expect: url.Values{
				"foo.0.bar": {"baz"},
				"foo.1.bar": {"baz"},
				"baz":       {"true"},
				"baz1":      {"1234"},
			},
		},
		{
			json:   json.RawMessage(`[{ "bar": "baz" }]`),
			expect: url.Values{"0.bar": {"baz"}},
		},
		{
			json:   json.RawMessage(`{ "foo": [{ "bar": "baz" },{ "bar": "baz" }], "baz": true, "baz1": 1234 }`),
			prefix: "traits",
			expect: url.Values{
				"traits.foo.0.bar": {"baz"},
				"traits.foo.1.bar": {"baz"},
				"traits.baz":       {"true"},
				"traits.baz1":      {"1234"},
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			assert.EqualValues(t, tc.expect, toFormValues(tc.json, tc.prefix))
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
