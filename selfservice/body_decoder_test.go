package selfservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewBodyDecoder(t *testing.T) {
	dec := NewBodyDecoder()

	t.Run("type=form", func(t *testing.T) {
		for k, tc := range []struct {
			d       string
			payload url.Values
			raw     string
			result  string
			opt     BodyDecoderOptions
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
				d: "should ignore unless prefix is set",
				payload: url.Values{
					"traits.foo": {"12345"},
					"password":   {"12345"},
				},
				opt:    BodyDecoderOptions{AssertTypesForPrefix: "traits."},
				result: `{"password":"12345","traits":{"foo":12345}}`,
			},
			{
				d:      "should work with true and false",
				raw:    "traits.consent.newsletter=false&traits.consent.newsletter=true&traits.consent.tos=false",
				result: `{"traits":{"consent":{"newsletter":true,"tos":false}}}`,
			},
		} {
			t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var result json.RawMessage
					require.NoError(t, dec.Decode(r, &result, tc.opt))
					require.JSONEq(t, tc.result, string(result), "%s", result)
				}))
				defer ts.Close()

				var res *http.Response
				var err error

				if tc.raw != "" {
					res, err = ts.Client().Post(ts.URL, "application/x-www-form-urlencoded", strings.NewReader(tc.raw))
				} else {
					res, err = ts.Client().PostForm(ts.URL, tc.payload)
				}

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
					require.NoError(t, dec.Decode(r, &result, BodyDecoderOptions{}))
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
