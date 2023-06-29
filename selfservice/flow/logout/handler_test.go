// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package logout_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"github.com/ory/kratos/session"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/x"
	"github.com/ory/x/urlx"
)

func TestLogout(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	errTS := testhelpers.NewErrorTestServer(t, reg)

	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
	public, _, publicRouter, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	publicRouter.GET("/session/browser/set", testhelpers.MockSetSession(t, reg, conf))
	publicRouter.GET("/session/browser/get", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		sess, err := reg.SessionManager().FetchFromRequest(r.Context(), r)
		if err != nil {
			reg.Writer().WriteError(w, r, err)
			return
		}
		reg.Writer().Write(w, r, sess)
	})
	publicRouter.POST("/csrf/check", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusNoContent)
	})
	conf.MustSet(ctx, config.ViperKeySelfServiceLogoutBrowserDefaultReturnTo, public.URL+"/session/browser/get")

	t.Run("case=successful logout for API clients", func(t *testing.T) {
		hc := testhelpers.NewDebugClient(t)

		sess := testhelpers.CreateSession(t, reg)
		_, res := testhelpers.HTTPRequestJSON(t, hc, "DELETE", public.URL+"/self-service/logout/api", json.RawMessage(`{"session_token": "`+sess.Token+`"}`))
		assert.Equal(t, http.StatusNoContent, res.StatusCode)

		actual, err := reg.SessionPersister().GetSession(ctx, sess.ID, session.ExpandNothing)
		require.NoError(t, err)
		assert.False(t, actual.IsActive())

		// Retrying logout should not fail:
		_, res = testhelpers.HTTPRequestJSON(t, hc, "DELETE", public.URL+"/self-service/logout/api", json.RawMessage(`{"session_token": "`+sess.Token+`"}`))
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
	})

	t.Run("case=unsuccessful logout for API clients because session token is invalid", func(t *testing.T) {
		hc := testhelpers.NewDebugClient(t)

		_, res := testhelpers.HTTPRequestJSON(t, hc, "DELETE", public.URL+"/self-service/logout/api", json.RawMessage(`{"session_token": "not a valid token"}`))
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
	})

	t.Run("case=unsuccessful logout for API clients because session token is missing", func(t *testing.T) {
		hc := testhelpers.NewDebugClient(t)

		_, res := testhelpers.HTTPRequestJSON(t, hc, "DELETE", public.URL+"/self-service/logout/api", json.RawMessage(`{}`))
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
	})

	makeBrowserLogout := func(t *testing.T, hc *http.Client, u string) ([]byte, *http.Response) {
		res, err := hc.Get(u)
		require.NoError(t, err)
		defer res.Body.Close()
		return x.MustReadAll(res.Body), res
	}

	getLogoutUrl := func(t *testing.T, params url.Values) (hc *http.Client, logoutUrl string) {
		hc = testhelpers.NewSessionClient(t, public.URL+"/session/browser/set")

		u, err := url.Parse(public.URL)
		require.NoError(t, err)
		u.Path = "/self-service/logout/browser"
		if params != nil {
			u.RawQuery = params.Encode()
		}

		body, res := testhelpers.HTTPRequestJSON(t, hc, "GET", u.String(), nil)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)

		logoutUrl = gjson.GetBytes(body, "logout_url").String()
		logoutToken := gjson.GetBytes(body, "logout_token").String()

		logoutUrlParsed, err := url.Parse(logoutUrl)
		require.NoError(t, err)

		assert.Equalf(t, logoutUrlParsed.Query().Get("token"), logoutToken, "%s", body)
		assert.Equalf(t, logoutUrlParsed.Path, "/self-service/logout", "%s", body)
		return
	}

	t.Run("case=successful logout for browser clients", func(t *testing.T) {
		ensurePrincipalChange := func(t *testing.T, originalCookies []*http.Cookie) {
			cj, err := cookiejar.New(nil)
			require.NoError(t, err)
			cj.SetCookies(urlx.ParseOrPanic(public.URL), originalCookies)
			res, err := (&http.Client{Jar: cj}).PostForm(public.URL+"/csrf/check", url.Values{})
			require.NoError(t, err)
			defer res.Body.Close()
			assert.EqualValues(t, http.StatusForbidden, res.StatusCode)
			body := x.MustReadAll(res.Body)
			assert.EqualValues(t, x.ErrInvalidCSRFToken.ReasonField, gjson.GetBytes(body, "error.reason").String(), "%s", body)
		}

		t.Run("type=browser", func(t *testing.T) {
			hc, logoutUrl := getLogoutUrl(t, nil)
			originalCookies := hc.Jar.Cookies(urlx.ParseOrPanic(public.URL))
			require.Contains(t, fmt.Sprintf("%v", hc.Jar.Cookies(urlx.ParseOrPanic(public.URL))), "ory_kratos_session")

			body, res := makeBrowserLogout(t, hc, logoutUrl)
			assert.EqualValues(t, public.URL+"/session/browser/get", res.Request.URL.String())
			require.NotContains(t, fmt.Sprintf("%v", hc.Jar.Cookies(urlx.ParseOrPanic(public.URL))), "ory_kratos_session")

			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode, "%s", body)
			assert.EqualValues(t, "No active session was found in this request.", gjson.GetBytes(body, "error.reason").String())

			// Logout means CSRF has also been changed.
			ensurePrincipalChange(t, originalCookies)
		})

		t.Run("type=ajax", func(t *testing.T) {
			hc, logoutUrl := getLogoutUrl(t, nil)
			originalCookies := hc.Jar.Cookies(urlx.ParseOrPanic(public.URL))
			require.Contains(t, fmt.Sprintf("%v", hc.Jar.Cookies(urlx.ParseOrPanic(public.URL))), "ory_kratos_session")

			body, res := testhelpers.HTTPRequestJSON(t, hc, "GET", logoutUrl, nil)
			assert.EqualValues(t, logoutUrl, res.Request.URL.String())

			assert.EqualValues(t, http.StatusNoContent, res.StatusCode, "%s", body)
			require.NotContains(t, fmt.Sprintf("%v", hc.Jar.Cookies(urlx.ParseOrPanic(public.URL))), "ory_kratos_session")

			// Logout means CSRF has also been changed.
			ensurePrincipalChange(t, originalCookies)
		})
	})

	t.Run("case=calling submission without token but with session", func(t *testing.T) {
		t.Run("type=browser", func(t *testing.T) {
			hc := testhelpers.NewSessionClient(t, public.URL+"/session/browser/set")
			body, res := makeBrowserLogout(t, hc, public.URL+"/self-service/logout")
			assert.Contains(t, res.Request.URL.String(), errTS.URL)
			assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
			assert.EqualValues(t, "Please include a token in the URL query.", gjson.GetBytes(body, "reason").String(), "%s", body)
		})

		t.Run("type=ajax", func(t *testing.T) {
			hc := testhelpers.NewSessionClient(t, public.URL+"/session/browser/set")
			body, res := testhelpers.HTTPRequestJSON(t, hc, "GET", public.URL+"/self-service/logout", nil)
			assert.EqualValues(t, public.URL+"/self-service/logout", res.Request.URL.String())
			assert.EqualValues(t, http.StatusBadRequest, res.StatusCode, "%s", body)
			assert.EqualValues(t, "Please include a token in the URL query.", gjson.GetBytes(body, "error.reason").String(), "%s", body)
		})
	})

	t.Run("case=calling submission with token but without session", func(t *testing.T) {
		t.Run("type=browser", func(t *testing.T) {
			_, logoutUrl := getLogoutUrl(t, nil)

			body, res := makeBrowserLogout(t, http.DefaultClient, logoutUrl)
			assert.Contains(t, res.Request.URL.String(), errTS.URL)
			assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
			assert.EqualValues(t, "No active session was found in this request.", gjson.GetBytes(body, "reason").String(), "%s", body)
		})

		t.Run("type=ajax", func(t *testing.T) {
			_, logoutUrl := getLogoutUrl(t, nil)

			body, res := testhelpers.HTTPRequestJSON(t, http.DefaultClient, "GET", logoutUrl, nil)
			assert.EqualValues(t, logoutUrl, res.Request.URL.String())
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode, "%s", body)
			assert.EqualValues(t, "No active session was found in this request.", gjson.GetBytes(body, "error.reason").String(), "%s", body)
		})
	})

	t.Run("case=calling submission with token but with session from another user", func(t *testing.T) {
		t.Run("type=browser", func(t *testing.T) {
			otherUser := testhelpers.NewSessionClient(t, public.URL+"/session/browser/set")
			_, logoutUrl := getLogoutUrl(t, nil)

			body, res := makeBrowserLogout(t, otherUser, logoutUrl)
			assert.Contains(t, res.Request.URL.String(), errTS.URL)
			assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
			assert.EqualValues(t, "Unable to log out because the logout token in the URL query does not match the session cookie.", gjson.GetBytes(body, "reason").String(), "%s", body)
		})

		t.Run("type=ajax", func(t *testing.T) {
			otherUser := testhelpers.NewSessionClient(t, public.URL+"/session/browser/set")
			_, logoutUrl := getLogoutUrl(t, nil)

			body, res := testhelpers.HTTPRequestJSON(t, otherUser, "GET", logoutUrl, nil)
			assert.EqualValues(t, logoutUrl, res.Request.URL.String())
			assert.EqualValues(t, http.StatusForbidden, res.StatusCode, "%s", body)
			assert.EqualValues(t, "Unable to log out because the logout token in the URL query does not match the session cookie.", gjson.GetBytes(body, "error.reason").String(), "%s", body)
		})
	})

	t.Run("case=calling submission with invalid token", func(t *testing.T) {
		t.Run("type=browser", func(t *testing.T) {
			hc, logoutUrl := getLogoutUrl(t, nil)

			body, res := makeBrowserLogout(t, hc, logoutUrl+"invalid")
			assert.Contains(t, res.Request.URL.String(), errTS.URL)
			assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
			assert.EqualValues(t, "Unable to log out because the logout token in the URL query does not match the session cookie.", gjson.GetBytes(body, "reason").String(), "%s", body)
		})

		t.Run("type=ajax", func(t *testing.T) {
			hc, logoutUrl := getLogoutUrl(t, nil)

			body, res := testhelpers.HTTPRequestJSON(t, hc, "GET", logoutUrl+"invalid", nil)
			assert.EqualValues(t, logoutUrl+"invalid", res.Request.URL.String())
			assert.EqualValues(t, http.StatusForbidden, res.StatusCode, "%s", body)
			assert.EqualValues(t, "Unable to log out because the logout token in the URL query does not match the session cookie.", gjson.GetBytes(body, "error.reason").String(), "%s", body)
		})
	})

	t.Run("case=calling browser init without session", func(t *testing.T) {
		body, res := testhelpers.HTTPRequestJSON(t, http.DefaultClient, "GET", public.URL+"/self-service/logout/browser", nil)
		assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
		assert.EqualValues(t, "No active session was found in this request.", gjson.GetBytes(body, "error.reason").String(), "%s", body)
	})

	t.Run("case=init logout through browser does 303 redirect", func(t *testing.T) {
		// init the logout
		hc, logoutUrl := getLogoutUrl(t, nil)
		// prevent the redirect, so we can get check the status code
		hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		// submit the login
		req, err := http.NewRequest("GET", logoutUrl, nil)
		require.NoError(t, err)

		res, err := hc.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		// here we check that the redirect status is 303
		require.Equal(t, http.StatusSeeOther, res.StatusCode)
	})

	t.Run("case=init logout with return_to should carry over return_to", func(t *testing.T) {
		reg.Config().MustSet(context.Background(), config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh"})

		hc, logoutUrl := getLogoutUrl(t, url.Values{"return_to": {"https://www.ory.sh"}})

		logoutUrlParsed, err := url.Parse(logoutUrl)
		require.NoError(t, err)

		assert.Equal(t, "https://www.ory.sh", logoutUrlParsed.Query().Get("return_to"))

		hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}

		body, res := makeBrowserLogout(t, hc, logoutUrl)
		assert.EqualValues(t, http.StatusSeeOther, res.StatusCode, "%s", body)
		assert.EqualValues(t, "https://www.ory.sh", res.Header.Get("Location"))
	})

	t.Run("case=init logout with return_to should not carry over return_to if not allowed", func(t *testing.T) {
		hc := testhelpers.NewSessionClient(t, public.URL+"/session/browser/set")

		logoutUrl := public.URL + "/self-service/logout/browser?return_to=https://www.ory.com"

		r, err := http.NewRequest("GET", logoutUrl, nil)
		require.NoError(t, err)
		r.Header.Set("Accept", "application/json")

		resp, err := hc.Do(r)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.EqualValues(t, "Requested return_to URL \"https://www.ory.com\" is not allowed.", gjson.GetBytes(body, "error.reason").String(), "%s", body)
		assert.EqualValues(t, http.StatusBadRequest, resp.StatusCode, "%s", body)
	})
}
