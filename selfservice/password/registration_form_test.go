package password

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
)

func TestRegistrationFormDecoder(t *testing.T) {
	dec := NewRegistrationFormDecoder()
	writer := herodot.NewJSONWriter(logrus.New())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := dec.Decode(r)
		if err != nil {
			writer.WriteError(w, r, err)
			return
		}

		writer.Write(w, r, p)
	}))
	defer ts.Close()

	t.Run("type=form", func(t *testing.T) {
		for k, tc := range []struct {
			d       string
			payload url.Values
			code    int
			result  string
		}{
			{
				d:       "should fail because payload is empty",
				payload: url.Values{},
				code:    http.StatusInternalServerError,
				result:  `{"error":{"code":500,"message":"password is required"}}`,
			},
			{
				d:       "should fail because password is missing",
				payload: url.Values{"foo": {"bar"}},
				code:    http.StatusInternalServerError,
				result:  `{"error":{"code":500,"message":"password is required"}}`,
			},
			{
				d:       "should pass without traits",
				payload: url.Values{"request": {"bar"}, "password": {"bar"}},
				code:    http.StatusOK,
				result:  `{"password":"bar","traits":{}}`,
			},
			{
				d:       "should pass with traits",
				payload: url.Values{"traits[nested]": {"__object__"}, "request": {"bar"}, "password": {"bar"}, "traits[foo.bar]": {"baz"}, "traits[int]": {"1234"}, "traits[float]": {"1234.1234"}, "traits[boolt]": {"true"}, "traits[boolf]": {"false"}},
				code:    http.StatusOK,
				result:  `{"password":"bar","traits":{"nested":{},"boolt":true,"boolf":false,"int":1234,"foo":{"bar":"baz"},"float":1234.1234}}`,
			},
			{
					d:       "should not override existing nested objects",
				payload: url.Values{"traits[nested.inner]": {"foobar"}, "traits[nested]": {"__object__"}, "request": {"bar"}, "password": {"bar"}, "traits[foo.bar]": {"baz"}, "traits[int]": {"1234"}, "traits[float]": {"1234.1234"}, "traits[boolt]": {"true"}, "traits[boolf]": {"false"}},
				code:    http.StatusOK,
				result:  `{"password":"bar","traits":{"nested":{"inner":"foobar"},"boolt":true,"boolf":false,"int":1234,"foo":{"bar":"baz"},"float":1234.1234}}`,
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				res, err := ts.Client().PostForm(ts.URL, tc.payload)
				require.NoError(t, err)
				defer res.Body.Close()

				require.Equal(t, tc.code, res.StatusCode)
				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tc.result, strings.Replace(string(body), "\n", "", 1))
			})
		}
	})

	t.Run("type=json", func(t *testing.T) {
		for k, tc := range []struct {
			d       string
			payload string
			code    int
			result  string
		}{
			{
				d:       "should fail because password is missing",
				payload: `{}`,
				code:    http.StatusInternalServerError,
				result:  `{"error":{"code":500,"message":"password is required"}}`,
			},
			{
				d:       "should pass without traits",
				payload: `{ "password": "bar"}`,
				code:    http.StatusOK,
				result:  `{"password":"bar","traits":{}}`,
			},
			{
				d:       "should pass with traits",
				payload: `{"password":"bar","traits":{"foo":{"bar":"baz"}}}`,
				code:    http.StatusOK,
				result:  `{"password":"bar","traits":{"foo":{"bar":"baz"}}}`,
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				res, err := ts.Client().Post(ts.URL, "application/json", bytes.NewBufferString(tc.payload))
				require.NoError(t, err)
				defer res.Body.Close()

				require.Equal(t, tc.code, res.StatusCode)
				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				assert.EqualValues(t, tc.result, strings.Replace(string(body), "\n", "", 1))
			})
		}
	})
}
