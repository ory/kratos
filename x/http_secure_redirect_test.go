package x_test

import (
	"encoding/json"
	"io/ioutil"
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

	defaultReturnTo := ts.URL + "/default-return-to"
	conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, defaultReturnTo)
	conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)
	conf.MustSet(config.ViperKeyURLsWhitelistedReturnToDomains, []string{ts.URL})

	run := func(t *testing.T, href string, contentType string) (*http.Response, string) {
		req, err := http.NewRequest("GET", href, nil)
		require.NoError(t, err)
		req.Header.Add("Accept", contentType)
		res, err := ts.Client().Do(req)
		require.NoError(t, err)
		body, err := ioutil.ReadAll(res.Body)
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

func TestSecureRedirectTo(t *testing.T) {

	var newServer = func(t *testing.T, isTLS bool, isRelative bool, expectErr bool, opts func(ts *httptest.Server) []x.SecureRedirectOption) *httptest.Server {
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

	var makeRequest = func(t *testing.T, ts *httptest.Server, path string) (*http.Response, string) {
		res, err := ts.Client().Get(ts.URL + "/" + path)
		require.NoError(t, err)

		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		return res, string(body)
	}

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

	t.Run("case=return to a fully qualified domain is forbidden if whitelist is relative", func(t *testing.T) {
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
}
