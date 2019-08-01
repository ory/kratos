package selfservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewBodyDecoder(t *testing.T) {
	dec := NewBodyDecoder()

	t.Run("type=form", func(t *testing.T) {
		for k, tc := range []struct {
			d       string
			payload url.Values
			result  string
		}{
			{
				d: "should work with nested keys",
				payload: url.Values{
					"traits.foo": {"bar"},
					"request":    {"bar"},
				},
				result: `{"request":"bar","traits":{"foo":"bar"}}`,
			},
			{
				d:       "should work with __object__ special key",
				payload: url.Values{"traits.nested": {"__object__"}, "request": {"bar"}, "password": {"bar"}, "traits.foo.bar": {"baz"}, "traits.int": {"1234"}, "traits.float": {"1234.1234"}, "traits.boolt": {"true"}, "traits.boolf": {"false"}},
				result:  `{"password":"bar","request":"bar","traits":{"nested":{},"boolt":true,"boolf":false,"int":1234,"foo":{"bar":"baz"},"float":1234.1234}}`,
			},
			{
				d:       "should not override existing object when __object__ is being used",
				payload: url.Values{"traits.nested.inner": {"foobar"}, "traits.nested": {"__object__"}, "request": {"bar"}, "password": {"bar"}, "traits.foo.bar": {"baz"}, "traits.int": {"1234"}, "traits.float": {"1234.1234"}, "traits.boolt": {"true"}, "traits.boolf": {"false"}},
				result:  `{"password":"bar","request":"bar","traits":{"nested":{"inner":"foobar"},"boolt":true,"boolf":false,"int":1234,"foo":{"bar":"baz"},"float":1234.1234}}`,
			},
		} {
			t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var result json.RawMessage
					require.NoError(t, dec.Decode(r, &result))
					require.JSONEq(t, tc.result, string(result), "%s", result)
				}))
				defer ts.Close()

				res, err := ts.Client().PostForm(ts.URL, tc.payload)
				require.NoError(t, err)
				require.NoError(t, res.Body.Close())
			})
		}
	})

	t.Run("type=json", func(t *testing.T) {
		for k, tc := range []struct {
			d       string
			payload string
			result  string
		}{
			{
				d:       "should work with nested keys",
				payload: `{"request":"bar","traits":{"foo":"bar"}}`,
				result:  `{"request":"bar","traits":{"foo":"bar"}}`,
			},
		} {
			t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var result json.RawMessage
					require.NoError(t, dec.Decode(r, &result))
					require.JSONEq(t, tc.result, string(result), "%s", result)
				}))
				defer ts.Close()

				res, err := ts.Client().Post(ts.URL, "application/json", bytes.NewBufferString(tc.payload))
				require.NoError(t, err)
				require.NoError(t, res.Body.Close())
			})
		}

	})
}
