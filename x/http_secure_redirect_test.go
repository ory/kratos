// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
)

func TestSecureContentNegotiationRedirection(t *testing.T) {
	conf, _ := internal.NewFastRegistryWithMocks(t)
	var jsonActual = json.RawMessage(`{"foo":"bar"}` + "\n")
	writer := herodot.NewJSONWriter(nil)

	router := httprouter.New()
	router.GET("/redir", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NoError(t, x.SecureContentNegotiationRedirection(w, r, jsonActual, x.RequestURL(r).String(), writer, conf))
	})
	router.GET("/default-return-to", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.GET("/return-to", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusNoContent)
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	ctx := context.Background()

	defaultReturnTo := ts.URL + "/default-return-to"
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, defaultReturnTo)
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, ts.URL)
	conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{ts.URL})

	run := func(t *testing.T, href string, contentType string) (*http.Response, string) {
		req, err := http.NewRequest("GET", href, nil)
		require.NoError(t, err)
		req.Header.Add("Accept", contentType)
		res, err := ts.Client().Do(req)
		require.NoError(t, err)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		return res, string(body)
	}

	t.Run("case=html browser causes redirect", func(t *testing.T) {
		res, _ := run(t, ts.URL+"/redir", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		assert.EqualValues(t, res.Request.URL.String(), defaultReturnTo)
	})

	t.Run("case=html browser causes redirect with redirect_to", func(t *testing.T) {
		res, _ := run(t, ts.URL+"/redir?return_to=/redirect-to", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		assert.EqualValues(t, res.Request.URL.String(), ts.URL+"/redirect-to")
	})

	t.Run("case=html browser causes redirect with redirect_to", func(t *testing.T) {
		res, body := run(t, ts.URL+"/redir?return_to=/redirect-to", "application/json")
		assert.EqualValues(t, res.Request.URL.String(), ts.URL+"/redir?return_to=/redirect-to")
		assert.EqualValues(t, body, jsonActual)
	})
}

func TestSecureRedirectToIsAllowedHost(t *testing.T) {
	type testCase struct {
		allowedURL  string
		redirectURL string
		valid       bool
	}
	tests := map[string]testCase{
		"case=Domain is allowed":              {allowedURL: "https://foo.bar", redirectURL: "https://foo.bar/redir", valid: true},
		"case=Domain prefix is allowed":       {allowedURL: "https://*.bar", redirectURL: "https://foo.bar/redir", valid: true},
		"case=Subdomain prefix is allowed":    {allowedURL: "https://*.foo.bar", redirectURL: "https://auth.foo.bar/redir", valid: true},
		"case=Domain is not allowed":          {allowedURL: "https://foo.baz", redirectURL: "https://foo.bar/redir", valid: false},
		"case=Domain wildcard is not allowed": {allowedURL: "https://*.foo.baz", redirectURL: "https://foo.bar/redir", valid: false},
		"case=Subdomain is not allowed":       {allowedURL: "https://*.foo.baz", redirectURL: "https://auth.foo.bar/redir", valid: false},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			allowedURL, err := url.Parse(tc.allowedURL)
			require.NoError(t, err)
			redirectURL, err := url.Parse(tc.redirectURL)
			require.NoError(t, err)
			assert.Equal(t, x.SecureRedirectToIsAllowedHost(redirectURL, *allowedURL), tc.valid)
		})
	}
}

func TestTakeOverReturnToParameter(t *testing.T) {
	type testCase struct {
		fromUrl           string
		toURL             string
		expectedOutputUrl string
	}
	tests := map[string]testCase{
		"case=return_to is taken over":                                             {fromUrl: "https://original.bar?return_to=https://allowed.domain", toURL: "https://output.bar", expectedOutputUrl: "https://output.bar?return_to=https%3A%2F%2Fallowed.domain"},
		"case=only return_to is taken over when multiple query parameters are set": {fromUrl: "https://original.bar?return_to=https://allowed.domain&flow=12312", toURL: "https://output.bar", expectedOutputUrl: "https://output.bar?return_to=https%3A%2F%2Fallowed.domain"},
		"case=output query parameters are preserved":                               {fromUrl: "https://original.bar?return_to=https://allowed.domain", toURL: "https://output.bar?flow=123321", expectedOutputUrl: "https://output.bar?flow=123321&return_to=https%3A%2F%2Fallowed.domain"},
		"case=when original return_to is empty do nothing":                         {fromUrl: "https://original.bar?return_to=", toURL: "https://output.bar?flow=123123", expectedOutputUrl: "https://output.bar?flow=123123"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := x.TakeOverReturnToParameter(tc.fromUrl, tc.toURL)
			require.NoError(t, err)
			assert.Equal(t, output, tc.expectedOutputUrl)
		})
	}
}

func TestSecureRedirectTo(t *testing.T) {
	newServer := func(t *testing.T, isTLS bool, isRelative bool, expectErr bool, opts func(ts *httptest.Server) []x.SecureRedirectOption) *httptest.Server {
		var ts *httptest.Server
		f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opts == nil {
				opts = func(ts *httptest.Server) []x.SecureRedirectOption {
					return nil
				}
			}
			defaultReturnTo := "/default-return-to"
			if !isRelative {
				defaultReturnTo = ts.URL + defaultReturnTo
			}
			ru, err := x.SecureRedirectTo(r, urlx.ParseOrPanic(defaultReturnTo), opts(ts)...)
			if expectErr {
				require.Error(t, err)
				_, _ = w.Write([]byte("error"))
				return
			}
			require.NoError(t, err)

			_, _ = w.Write([]byte(ru.String()))
		})

		if isTLS {
			ts = httptest.NewTLSServer(f)
		} else {
			ts = httptest.NewServer(f)
		}

		t.Cleanup(ts.Close)
		return ts
	}

	makeRequest := func(t *testing.T, ts *httptest.Server, path string) (*http.Response, string) {
		res, err := ts.Client().Get(ts.URL + "/" + path)
		require.NoError(t, err)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		return res, string(body)
	}

	t.Run("case=return to a relative path with anchor works", func(t *testing.T) {
		s := newServer(t, false, true, false, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{x.SecureRedirectAllowURLs([]url.URL{*urlx.ParseOrPanic("/foo")})}
		})
		_, body := makeRequest(t, s, "?return_to=/foo/kratos%23abcd")
		assert.Equal(t, body, "/foo/kratos#abcd")
	})

	t.Run("case=return to default URL if nothing is allowed", func(t *testing.T) {
		s := newServer(t, false, false, false, nil)
		_, body := makeRequest(t, s, "?return_to=/foo")
		assert.EqualValues(t, body, s.URL+"/default-return-to")
	})

	t.Run("case=return to foo with server baseURL if allowed", func(t *testing.T) {
		s := newServer(t, false, false, false, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{x.SecureRedirectAllowURLs([]url.URL{*urlx.ParseOrPanic(ts.URL)})}
		})
		_, body := makeRequest(t, s, "?return_to=/foo")
		assert.Equal(t, body, s.URL+"/foo")
	})

	t.Run("case=return to a relative path works", func(t *testing.T) {
		s := newServer(t, false, true, false, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{x.SecureRedirectAllowURLs([]url.URL{*urlx.ParseOrPanic("/foo")})}
		})
		_, body := makeRequest(t, s, "?return_to=/foo/kratos")
		assert.Equal(t, body, "/foo/kratos")
	})

	t.Run("case=return to a fully qualified domain is forbidden if allowlist is relative", func(t *testing.T) {
		s := newServer(t, false, true, true, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{x.SecureRedirectAllowURLs([]url.URL{*urlx.ParseOrPanic("/foo")})}
		})
		_, body := makeRequest(t, s, "?return_to=https://www.ory.sh/foo/kratos")
		assert.Equal(t, body, "error")
	})

	t.Run("case=return to another domain works", func(t *testing.T) {
		s := newServer(t, false, false, false, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{x.SecureRedirectAllowURLs([]url.URL{*urlx.ParseOrPanic("https://www.ory.sh/foo")})}
		})
		_, body := makeRequest(t, s, "?return_to=https://www.ory.sh/foo/kratos")
		assert.Equal(t, body, "https://www.ory.sh/foo/kratos")
	})

	t.Run("case=return to another domain fails if host mismatches", func(t *testing.T) {
		s := newServer(t, false, false, true, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{x.SecureRedirectAllowURLs([]url.URL{*urlx.ParseOrPanic("https://www.not-ory.sh/")})}
		})
		_, body := makeRequest(t, s, "?return_to=https://www.ory.sh/kratos")
		assert.Equal(t, body, "error")
	})

	t.Run("case=return to another domain fails if path mismatches", func(t *testing.T) {
		s := newServer(t, false, false, true, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{x.SecureRedirectAllowURLs([]url.URL{*urlx.ParseOrPanic("https://www.ory.sh/not-kratos")})}
		})
		_, body := makeRequest(t, s, "?return_to=https://www.ory.sh/kratos")
		assert.Equal(t, body, "error")
	})

	t.Run("case=return to another domain fails if scheme mismatches", func(t *testing.T) {
		s := newServer(t, false, false, true, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{x.SecureRedirectAllowURLs([]url.URL{*urlx.ParseOrPanic("http://www.ory.sh/")})}
		})
		_, body := makeRequest(t, s, "?return_to=https://www.ory.sh/kratos")
		assert.Equal(t, body, "error")
	})

	t.Run("case=should work with self-service modifier", func(t *testing.T) {
		s := newServer(t, false, false, false, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{x.SecureRedirectAllowSelfServiceURLs(urlx.ParseOrPanic(ts.URL))}
		})
		_, body := makeRequest(t, s, "?return_to=/self-service/foo")
		assert.Equal(t, body, s.URL+"/self-service/foo")
	})

	t.Run("case=should work with default return to", func(t *testing.T) {
		s := newServer(t, false, false, false, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{x.SecureRedirectOverrideDefaultReturnTo(urlx.ParseOrPanic(ts.URL + "/another-default"))}
		})
		_, body := makeRequest(t, s, "")
		assert.Equal(t, body, s.URL+"/another-default")
	})

	t.Run("case=should override return_to", func(t *testing.T) {
		s := newServer(t, false, false, false, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{
				x.SecureRedirectAllowURLs([]url.URL{*urlx.ParseOrPanic(ts.URL)}),
				x.SecureRedirectUseSourceURL("https://foo/bar?return_to=/override"),
			}
		})
		_, body := makeRequest(t, s, "?return_to=/original")
		assert.Equal(t, body, s.URL+"/override")
	})

	t.Run("case=should work with subdomain wildcard", func(t *testing.T) {
		s := newServer(t, false, false, false, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{x.SecureRedirectAllowURLs([]url.URL{*urlx.ParseOrPanic("https://*.ory.sh/")})}
		})
		_, body := makeRequest(t, s, "?return_to=https://www.ory.sh/kratos")
		assert.Equal(t, body, "https://www.ory.sh/kratos")
		_, body = makeRequest(t, s, "?return_to=https://even.deeper.nested.ory.sh/kratos")
		assert.Equal(t, body, "https://even.deeper.nested.ory.sh/kratos")
	})

	t.Run("case=should fallback to default return_to scheme", func(t *testing.T) {
		s := newServer(t, false, false, false, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{
				x.SecureRedirectAllowURLs([]url.URL{*urlx.ParseOrPanic("https://www.ory.sh")}),
				x.SecureRedirectOverrideDefaultReturnTo(urlx.ParseOrPanic("https://www.ory.sh/docs")),
			}
		})
		_, body := makeRequest(t, s, "?return_to=//www.ory.sh/kratos")
		assert.Equal(t, body, "https://www.ory.sh/kratos")
	})

	t.Run("case=should fallback to default return_to host", func(t *testing.T) {
		s := newServer(t, false, false, false, func(ts *httptest.Server) []x.SecureRedirectOption {
			return []x.SecureRedirectOption{
				x.SecureRedirectAllowURLs([]url.URL{
					*urlx.ParseOrPanic("https://www.ory.sh"),
					*urlx.ParseOrPanic("http://www.ory.sh"),
				}),
				x.SecureRedirectOverrideDefaultReturnTo(urlx.ParseOrPanic("https://www.ory.sh/docs")),
			}
		})
		_, body := makeRequest(t, s, "?return_to=http:///kratos")
		assert.Equal(t, body, "http://www.ory.sh/kratos")
	})
}
