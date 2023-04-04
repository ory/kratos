// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/x/assertx"
	"github.com/ory/x/logrusx"

	"github.com/stretchr/testify/assert"

	"github.com/ory/x/urlx"
)

func TestRequestURL(t *testing.T) {
	assert.EqualValues(t, RequestURL(&http.Request{
		URL: urlx.ParseOrPanic("/foo"), Host: "foobar", TLS: &tls.ConnectionState{},
	}).String(), "https://foobar/foo")
	assert.EqualValues(t, RequestURL(&http.Request{
		URL: urlx.ParseOrPanic("/foo"), Host: "foobar",
	}).String(), "http://foobar/foo")
	assert.EqualValues(t, RequestURL(&http.Request{
		URL: urlx.ParseOrPanic("/foo"), Host: "foobar", Header: http.Header{"X-Forwarded-Host": []string{"notfoobar"}, "X-Forwarded-Proto": {"https"}},
	}).String(), "https://notfoobar/foo")
}

func TestAcceptToRedirectOrJSON(t *testing.T) {
	wr := herodot.NewJSONWriter(logrusx.New("", ""))

	t.Run("case=browser", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Accept", "text/html")

		t.Run("regular payload", func(t *testing.T) {
			w := httptest.NewRecorder()
			AcceptToRedirectOrJSON(w, r, wr, json.RawMessage(`{"foo":"bar"}`), "https://www.ory.sh/redir")
			loc, err := w.Result().Location()
			require.NoError(t, err)
			assert.Equal(t, "https://www.ory.sh/redir", loc.String())
		})

		t.Run("error payload", func(t *testing.T) {
			w := httptest.NewRecorder()
			AcceptToRedirectOrJSON(w, r, wr, errors.New("foo"), "https://www.ory.sh/redir")
			loc, err := w.Result().Location()
			require.NoError(t, err)
			assert.Equal(t, "https://www.ory.sh/redir", loc.String())
		})
	})

	t.Run("case=json", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Accept", "application/json")

		t.Run("regular payload", func(t *testing.T) {
			msg := json.RawMessage(`{"foo":"bar"}`)
			w := httptest.NewRecorder()
			AcceptToRedirectOrJSON(w, r, wr, msg, "https://www.ory.sh/redir")
			_, err := w.Result().Location()
			require.ErrorIs(t, err, http.ErrNoLocation)

			body := MustReadAll(w.Result().Body)
			assertx.EqualAsJSON(t, msg, json.RawMessage(body))
		})

		t.Run("error payload", func(t *testing.T) {
			ee := errors.WithStack(herodot.ErrBadRequest)
			w := httptest.NewRecorder()
			AcceptToRedirectOrJSON(w, r, wr, ee, "https://www.ory.sh/redir")
			_, err := w.Result().Location()
			require.ErrorIs(t, err, http.ErrNoLocation)

			body := MustReadAll(w.Result().Body)
			assertx.EqualAsJSON(t, map[string]interface{}{"error": map[string]interface{}{"code": 400, "message": "The request was malformed or contained invalid parameters", "status": "Bad Request"}}, json.RawMessage(body))
			assert.EqualValues(t, http.StatusBadRequest, w.Result().StatusCode)
		})
	})
}
