package x_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
	"github.com/ory/nosurf"
	"github.com/ory/x/randx"
)

func TestNosurfBaseCookieHandler(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	require.NoError(t, conf.Source().Set(config.ViperKeyPublicBaseURL, "http://foo.com/bar"))

	cookie := x.NosurfBaseCookieHandler(reg)(httptest.NewRecorder(), httptest.NewRequest("GET", "https://foo/bar", nil))
	assert.EqualValues(t, "aHR0cDovL2Zvby5jb20vYmFy_csrf_token", cookie.Name, "base64 representation of http://foo.com/bar")
	assert.EqualValues(t, http.SameSiteLaxMode, cookie.SameSite, "is set to lax because https/secure is false - chrome rejects none samesite on non-https")
	assert.EqualValues(t, nosurf.MaxAge, cookie.MaxAge)
	assert.False(t, cookie.Secure, "false because insecure dev mode")
	assert.True(t, cookie.HttpOnly)

	for i := 0; i < 10; i++ {
		require.NoError(t, conf.Source().Set(config.ViperKeyPublicBaseURL, randx.MustString(16, randx.AlphaNum)))
		cookie := x.NosurfBaseCookieHandler(reg)(httptest.NewRecorder(), httptest.NewRequest("GET", "https://foo/bar", nil))

		assert.NotEqual(t, "aHR0cDovL2Zvby5jb20vYmFy_csrf_token", cookie.Name, "should no longer be http://foo.com/bar")

		matches, err := regexp.MatchString("[a-zA-Z_0-9]+", cookie.Name)
		require.NoError(t, err)
		assert.True(t, matches, "does not have any special chars")
	}

	require.NoError(t, conf.Source().Set("dev", false))
	cookie = x.NosurfBaseCookieHandler(reg)(httptest.NewRecorder(), httptest.NewRequest("GET", "https://foo/bar", nil))
	assert.EqualValues(t, http.SameSiteNoneMode, cookie.SameSite, "can be none because https/secure is true")
	assert.True(t, cookie.Secure, "true because secure mode")
	assert.True(t, cookie.HttpOnly)
}

func TestNosurfBaseCookieErrorHandler(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)

	h := x.CSRFFailureHandler(reg)
	expectError := func(t *testing.T, err error, req *http.Request) {
		rec := httptest.NewRecorder()
		h(rec, req)
		assertx.EqualAsJSON(t, x.ErrInvalidCSRFTokenAJAXNoCookies, json.RawMessage(gjson.Get(rec.Body.String(), "error").Raw))
	}

	newAjaxRequest := func() *http.Request {
		req := httptest.NewRequest("GET", "https://foo/bar", nil)
		req.Header.Set("Origin", "foo")
		return req
	}
	newBrowserRequest := func() *http.Request {
		return httptest.NewRequest("GET", "https://foo/bar", nil)
	}

	t.Run("case=without cookie", func(t *testing.T) {
		t.Run("source=ajax", func(t *testing.T) {
			expectError(t, x.ErrInvalidCSRFTokenAJAXNoCookies, newAjaxRequest())
		})

		t.Run("source=ajax", func(t *testing.T) {
			expectError(t, x.ErrInvalidCSRFTokenAJAXNoCookies, newBrowserRequest())
		})
	})

	t.Run("case=ajax with cookie but without csrf cookie", func(t *testing.T) {
		test := func(t *testing.T, req *http.Request) {
			req.Header.Set("Cookie", "foo=bar;")
			expectError(t, x.ErrInvalidCSRFTokenAJAXNoCookies, req)
		}

		t.Run("source=ajax", func(t *testing.T) {
			test(t, newAjaxRequest())
		})

		t.Run("source=ajax", func(t *testing.T) {
			test(t, newBrowserRequest())
		})
	})

	t.Run("case=ajax with correct cookie but token was not sent in header", func(t *testing.T) {
		test := func(t *testing.T, req *http.Request) {
			req.Header.Set("Cookie", x.CSRFCookieName(reg, req)+"=bar;")
			expectError(t, x.ErrInvalidCSRFTokenAJAXTokenNotSent, req)
		}

		t.Run("source=ajax", func(t *testing.T) {
			test(t, newAjaxRequest())
		})

		t.Run("source=ajax", func(t *testing.T) {
			test(t, newBrowserRequest())
		})
	})

	t.Run("case=ajax with correct cookie and token in header but they do not match", func(t *testing.T) {
		test := func(t *testing.T, req *http.Request) {
			req.Header.Set(nosurf.HeaderName, "bar")
			req.Header.Set("Cookie", x.CSRFCookieName(reg, req)+"=bar;")
			expectError(t, x.ErrInvalidCSRFTokenAJAXTokenMismatch, req)
		}

		t.Run("source=ajax", func(t *testing.T) {
			test(t, newAjaxRequest())
		})

		t.Run("source=ajax", func(t *testing.T) {
			test(t, newBrowserRequest())
		})
	})

	t.Run("case=ajax with correct cookie and token in body but they do not match", func(t *testing.T) {
		test := func(t *testing.T, req *http.Request) {
			req.Header.Set("Accept", "application/x-www-form-urlencoded")
			req.Header.Set("Cookie", x.CSRFCookieName(reg, req)+"=bar;")
			expectError(t, x.ErrInvalidCSRFTokenAJAXTokenMismatch, req)
		}

		t.Run("source=ajax", func(t *testing.T) {
			req := httptest.NewRequest("POST", "https://foo/bar", strings.NewReader(url.Values{nosurf.CookieName: {"bar"}}.Encode()))
			req.Header.Set("Origin", "foo")
			test(t, req)
		})

		t.Run("source=ajax", func(t *testing.T) {
			req := httptest.NewRequest("POST", "https://foo/bar", strings.NewReader(url.Values{nosurf.CookieName: {"bar"}}.Encode()))
			test(t, req)
		})
	})
}
