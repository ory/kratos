// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"context"
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

func TestWithBaseURL(t *testing.T) {
	ctx := WithBaseURL(context.Background(), urlx.ParseOrPanic("https://www.ory.com/"))
	assert.EqualValues(t, "https://www.ory.com/", BaseURLFromContext(ctx).String())
	assert.Nil(t, BaseURLFromContext(context.Background()))

	t.Run("preserves scheme", func(t *testing.T) {
		for _, raw := range []string{"http://localhost:4000", "https://example.com/"} {
			ctx := WithBaseURL(context.Background(), urlx.ParseOrPanic(raw))
			assert.EqualValuesf(t, raw, BaseURLFromContext(ctx).String(), "for input %q", raw)
		}
	})

	t.Run("BaseURLStringFromContext", func(t *testing.T) {
		assert.Equal(t, "", BaseURLStringFromContext(context.Background()))
		assert.Equal(t, "https://www.ory.com/",
			BaseURLStringFromContext(WithBaseURL(context.Background(), urlx.ParseOrPanic("https://www.ory.com/"))))
	})
}

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

func TestRequestBaseURL(t *testing.T) {
	t.Run("falls back to request scheme://host with no path", func(t *testing.T) {
		assert.Equal(t, "https://foobar", RequestBaseURL(&http.Request{
			URL: urlx.ParseOrPanic("/self-service/login?foo=bar"), Host: "foobar", TLS: &tls.ConnectionState{},
		}))
		assert.Equal(t, "http://foobar", RequestBaseURL(&http.Request{
			URL: urlx.ParseOrPanic("/foo"), Host: "foobar",
		}))
		assert.Equal(t, "https://notfoobar", RequestBaseURL(&http.Request{
			URL: urlx.ParseOrPanic("/foo"), Host: "foobar",
			Header: http.Header{"X-Forwarded-Host": []string{"notfoobar"}, "X-Forwarded-Proto": {"https"}},
		}))
	})

	t.Run("context-captured customer base URL wins, scheme preserved", func(t *testing.T) {
		// A proxy-aware middleware (e.g. the cloud courier middleware) validated
		// an Ory-Base-URL-Rewrite / X-Ory-Original-Host header and stashed the
		// real customer-facing base URL on the context. The OIDC/SAML state must
		// capture *that*, not the oryapis host this service was reached at.
		for _, captured := range []string{"http://localhost:4000", "https://login.customer.example.com"} {
			req := (&http.Request{URL: urlx.ParseOrPanic("/self-service/login"), Host: "slug.projects.oryapis.com", TLS: &tls.ConnectionState{}}).
				WithContext(WithBaseURL(context.Background(), urlx.ParseOrPanic(captured)))
			assert.Equalf(t, captured, RequestBaseURL(req), "captured %q must win over the oryapis host", captured)
		}
	})
}

func TestAcceptToRedirectOrJSON(t *testing.T) {
	wr := herodot.NewJSONWriter(logrusx.New("", ""))

	t.Run("case=browser", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Accept", "text/html")

		t.Run("regular payload", func(t *testing.T) {
			w := httptest.NewRecorder()
			SendFlowCompletedAsRedirectOrJSON(w, r, wr, json.RawMessage(`{"foo":"bar"}`), "https://www.ory.com/redir")
			loc, err := w.Result().Location()
			require.NoError(t, err)
			assert.Equal(t, "https://www.ory.com/redir", loc.String())
		})

		t.Run("error payload", func(t *testing.T) {
			w := httptest.NewRecorder()
			SendFlowCompletedAsRedirectOrJSON(w, r, wr, errors.New("foo"), "https://www.ory.com/redir")
			loc, err := w.Result().Location()
			require.NoError(t, err)
			assert.Equal(t, "https://www.ory.com/redir", loc.String())
		})
	})

	t.Run("case=json", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Accept", "application/json")

		t.Run("regular payload", func(t *testing.T) {
			msg := json.RawMessage(`{"foo":"bar"}`)
			w := httptest.NewRecorder()
			SendFlowCompletedAsRedirectOrJSON(w, r, wr, msg, "https://www.ory.com/redir")
			_, err := w.Result().Location()
			require.ErrorIs(t, err, http.ErrNoLocation)

			body := MustReadAll(w.Result().Body)
			assertx.EqualAsJSON(t, msg, json.RawMessage(body))
		})

		t.Run("error payload", func(t *testing.T) {
			ee := errors.WithStack(herodot.ErrBadRequest())
			w := httptest.NewRecorder()
			SendFlowCompletedAsRedirectOrJSON(w, r, wr, ee, "https://www.ory.com/redir")
			_, err := w.Result().Location()
			require.ErrorIs(t, err, http.ErrNoLocation)

			body := MustReadAll(w.Result().Body)
			assertx.EqualAsJSON(t, map[string]interface{}{"error": map[string]interface{}{"code": 400, "message": "The request was malformed or contained invalid parameters", "status": "Bad Request"}}, json.RawMessage(body))
			assert.EqualValues(t, http.StatusBadRequest, w.Result().StatusCode)
		})
	})
}
