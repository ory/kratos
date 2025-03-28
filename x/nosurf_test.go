// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"
	"github.com/ory/x/snapshotx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
	"github.com/ory/nosurf"
	"github.com/ory/x/randx"
)

func TestNosurfBaseCookieHandler(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	require.NoError(t, conf.Set(ctx, config.ViperKeyPublicBaseURL, "http://foo.com/bar"))

	cookie := x.NosurfBaseCookieHandler(reg)(httptest.NewRecorder(), httptest.NewRequest("GET", "https://foo/bar", nil))
	assert.EqualValues(t, "csrf_token_01c86631efd1537ee34a98e75884a6e21dd8e2d9e944934bca21204106bfd32f", cookie.Name, "base64 representation of http://foo.com/bar")
	assert.EqualValues(t, http.SameSiteLaxMode, cookie.SameSite, "is set to lax because https/secure is false - chrome rejects none samesite on non-https")
	assert.EqualValues(t, nosurf.MaxAge, cookie.MaxAge)
	assert.EqualValues(t, "/", cookie.Path, "cookie path is site root by default")
	assert.EqualValues(t, "", cookie.Domain, "domain for the cookie is set to empty per default")
	assert.False(t, cookie.Secure, "false because insecure dev mode")
	assert.True(t, cookie.HttpOnly)

	alNum := regexp.MustCompile("[a-zA-Z_0-9]+")
	for i := 0; i < 10; i++ {
		require.NoError(t, conf.Set(ctx, config.ViperKeyPublicBaseURL, randx.MustString(16, randx.AlphaNum)))
		cookie := x.NosurfBaseCookieHandler(reg)(httptest.NewRecorder(), httptest.NewRequest("GET", "https://foo/bar", nil))

		assert.NotEqual(t, "aHR0cDovL2Zvby5jb20vYmFy_csrf_token", cookie.Name, "should no longer be http://foo.com/bar")
		assert.True(t, alNum.MatchString(cookie.Name), "does not have any special chars")
	}

	require.NoError(t, conf.Set(ctx, config.ViperKeyCookieSameSite, "None"))
	require.NoError(t, conf.Set(ctx, "dev", false))
	cookie = x.NosurfBaseCookieHandler(reg)(httptest.NewRecorder(), httptest.NewRequest("GET", "https://foo/bar", nil))
	assert.EqualValues(t, http.SameSiteNoneMode, cookie.SameSite, "can be none because https/secure is true")
	assert.True(t, cookie.Secure, "true because secure mode")
	assert.True(t, cookie.HttpOnly)
}

func TestNosurfBaseCookieHandlerAliasing(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	require.NoError(t, conf.Set(ctx, config.ViperKeyPublicBaseURL, "http://foo.com/bar"))

	cookie := x.NosurfBaseCookieHandler(reg)(httptest.NewRecorder(), httptest.NewRequest("GET", "http://foo.com/bar", nil))
	assert.EqualValues(t, "", cookie.Domain, "remains unset")
	assert.EqualValues(t, "/", cookie.Path, "cookie path is site root by default")

	// Check root settings
	require.NoError(t, conf.Set(ctx, config.ViperKeyCookieDomain, "bar.com"))
	require.NoError(t, conf.Set(ctx, config.ViperKeyCookiePath, "/baz"))
	cookie = x.NosurfBaseCookieHandler(reg)(httptest.NewRecorder(), httptest.NewRequest("GET", "http://foo.com/bar", nil))
	assert.EqualValues(t, "bar.com", cookie.Domain, "domain doesn't change when request not from an alias but is overwritten by ViperKeyCookieDomain")
	assert.EqualValues(t, "/baz", cookie.Path, "cookie path is site root by default but is overwritten by ViperKeyCookiePath")
}

func TestNosurfBaseCookieErrorHandler(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)

	h := x.CSRFFailureHandler(reg)
	expectError := func(t *testing.T, err error, req *http.Request) {
		rec := httptest.NewRecorder()
		h(rec, req)
		assertx.EqualAsJSON(t, err, json.RawMessage(gjson.Get(rec.Body.String(), "error").Raw))
	}

	newAjaxRequest := func() *http.Request {
		req := httptest.NewRequest("GET", "https://foo/bar", nil)
		req.Header.Set("Sec-Fetch-Mode", "cors")
		return req
	}
	newBrowserRequest := func() *http.Request {
		return httptest.NewRequest("GET", "https://foo/bar", nil)
	}

	t.Run("case=without cookie", func(t *testing.T) {
		t.Run("source=ajax", func(t *testing.T) {
			expectError(t, x.ErrInvalidCSRFTokenAJAXNoCookies, newAjaxRequest())
		})

		t.Run("source=browser", func(t *testing.T) {
			expectError(t, x.ErrInvalidCSRFTokenServerNoCookies, newBrowserRequest())
		})
	})

	t.Run("case=with cookie header but without csrf cookie", func(t *testing.T) {
		test := func(t *testing.T, req *http.Request, err error) {
			req.Header.Set("Cookie", "foo=bar;")
			expectError(t, err, req)
		}

		t.Run("source=ajax", func(t *testing.T) {
			test(t, newAjaxRequest(), x.ErrInvalidCSRFTokenAJAXCookieMissing)
		})

		t.Run("source=browser", func(t *testing.T) {
			test(t, newBrowserRequest(), x.ErrInvalidCSRFTokenServerCookieMissing)
		})
	})

	t.Run("case=ajax with correct cookie but token was not sent in header", func(t *testing.T) {
		test := func(t *testing.T, req *http.Request, err error) {
			req.Header.Set("Cookie", x.CSRFCookieName(reg, req)+"=bar;")
			expectError(t, err, req)
		}

		t.Run("source=ajax", func(t *testing.T) {
			test(t, newAjaxRequest(), x.ErrInvalidCSRFTokenAJAXTokenNotSent)
		})

		t.Run("source=browser", func(t *testing.T) {
			test(t, newBrowserRequest(), x.ErrInvalidCSRFTokenServerTokenNotSent)
		})
	})

	t.Run("case=ajax with correct cookie and token in header but they do not match", func(t *testing.T) {
		test := func(t *testing.T, req *http.Request, err error) {
			req.Header.Set(nosurf.HeaderName, "bar")
			req.Header.Set("Cookie", x.CSRFCookieName(reg, req)+"=bar;")
			expectError(t, err, req)
		}

		t.Run("source=ajax", func(t *testing.T) {
			test(t, newAjaxRequest(), x.ErrInvalidCSRFTokenAJAXTokenMismatch)
		})

		t.Run("source=browser", func(t *testing.T) {
			test(t, newBrowserRequest(), x.ErrInvalidCSRFTokenServerTokenMismatch)
		})
	})

	t.Run("case=with correct cookie and token in body but they do not match", func(t *testing.T) {
		test := func(t *testing.T, req *http.Request, err error) {
			req.Header.Set("Accept", "application/x-www-form-urlencoded")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Cookie", x.CSRFCookieName(reg, req)+"=bar;")
			req.ParseForm()
			expectError(t, err, req)
		}

		t.Run("source=ajax", func(t *testing.T) {
			req := httptest.NewRequest("POST", "https://foo/bar", strings.NewReader(url.Values{nosurf.FormFieldName: {"bar"}}.Encode()))
			req.Header.Set("Sec-Fetch-Mode", "cors")
			test(t, req, x.ErrInvalidCSRFTokenAJAXTokenMismatch)
		})

		t.Run("source=browser", func(t *testing.T) {
			req := httptest.NewRequest("POST", "https://foo/bar", strings.NewReader(url.Values{nosurf.FormFieldName: {"bar"}}.Encode()))
			test(t, req, x.ErrInvalidCSRFTokenServerTokenMismatch)
		})
	})
}

func TestSnapshotCsrfErrors(t *testing.T) {
	type Snapshot = struct {
		Id  string
		Err error
	}

	errors := []Snapshot{
		{Id: "ErrInvalidCSRFToken", Err: x.ErrInvalidCSRFToken},
		{Id: "ErrInvalidCSRFTokenAJAX", Err: x.ErrInvalidCSRFTokenAJAX},
		{Id: "ErrInvalidCSRFTokenAJAXNoCookies", Err: x.ErrInvalidCSRFTokenAJAXNoCookies},
		{Id: "ErrInvalidCSRFTokenAJAXCookieMissing", Err: x.ErrInvalidCSRFTokenAJAXCookieMissing},
		{Id: "ErrInvalidCSRFTokenAJAXTokenMismatch", Err: x.ErrInvalidCSRFTokenAJAXTokenMismatch},
		{Id: "ErrInvalidCSRFTokenAJAXTokenNotSent", Err: x.ErrInvalidCSRFTokenAJAXTokenNotSent},
		{Id: "ErrInvalidCSRFTokenServer", Err: x.ErrInvalidCSRFTokenServer},
		{Id: "ErrInvalidCSRFTokenServerNoCookies", Err: x.ErrInvalidCSRFTokenServerNoCookies},
		{Id: "ErrInvalidCSRFTokenServerCookieMissing", Err: x.ErrInvalidCSRFTokenServerCookieMissing},
		{Id: "ErrInvalidCSRFTokenServerTokenMismatch", Err: x.ErrInvalidCSRFTokenServerTokenMismatch},
		{Id: "ErrInvalidCSRFTokenServerTokenNotSent", Err: x.ErrInvalidCSRFTokenServerTokenNotSent},
	}
	slices.SortFunc(errors, func(a Snapshot, b Snapshot) int {
		return strings.Compare(a.Id, b.Id)
	})
	snapshotx.SnapshotT(t, errors)
}
