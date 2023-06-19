// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos/text"

	"github.com/ory/x/assertx"

	"github.com/ory/kratos/session"

	"github.com/gofrs/uuid"

	kratos "github.com/ory/kratos/internal/httpclient"

	"github.com/ory/kratos/corpx"

	"github.com/ory/x/ioutilx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/x"
)

func init() {
	corpx.RegisterFakes()
}

func TestHandler(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePassword.String(), true)
	testhelpers.StrategyEnable(t, conf, settings.StrategyProfile, true)

	router := x.NewRouterPublic()
	admin := x.NewRouterAdmin()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, admin)

	_ = testhelpers.NewSettingsUITestServer(t, conf)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)

	conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")

	primaryIdentity := &identity.Identity{ID: x.NewUUID(), Traits: identity.Traits(`{}`)}
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), primaryIdentity))
	primaryUser := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, primaryIdentity)
	otherUser := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)

	newExpiredFlow := func() *settings.Flow {
		f, err := settings.NewFlow(conf, -time.Minute,
			&http.Request{URL: urlx.ParseOrPanic(publicTS.URL + login.RouteInitBrowserFlow)},
			primaryIdentity, flow.TypeBrowser)
		require.NoError(t, err)
		return f
	}

	assertion := func(t *testing.T, body []byte, isApi bool) {
		if isApi {
			assert.Equal(t, "api", gjson.GetBytes(body, "type").String())
		} else {
			assert.Equal(t, "browser", gjson.GetBytes(body, "type").String())
		}
	}

	type options struct {
		query url.Values
	}
	type initAuthOptions func(*options)

	WithInitAuthQuery := func(q url.Values) initAuthOptions {
		return func(o *options) {
			o.query = q
		}
	}

	initAuthenticatedFlow := func(t *testing.T, hc *http.Client, isAPI bool, isSPA bool, opts ...initAuthOptions) (*http.Response, []byte) {
		op := new(options)
		for _, o := range opts {
			o(op)
		}
		route := settings.RouteInitBrowserFlow
		if isAPI {
			route = settings.RouteInitAPIFlow
		}
		reqURL, err := url.Parse(publicTS.URL + route)
		require.NoError(t, err)
		reqURL.RawQuery = op.query.Encode()

		req := x.NewTestHTTPRequest(t, "GET", reqURL.String(), nil)
		if isSPA || isAPI {
			req.Header.Set("Accept", "application/json")
		}
		res, err := hc.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		if isAPI {
			assert.Len(t, res.Header.Get("Set-Cookie"), 0)
		}
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		return res, body
	}

	initFlow := func(t *testing.T, hc *http.Client, isAPI bool, opts ...initAuthOptions) (*http.Response, []byte) {
		return initAuthenticatedFlow(t, hc, isAPI, false, opts...)
	}

	initSPAFlow := func(t *testing.T, hc *http.Client, opts ...initAuthOptions) (*http.Response, []byte) {
		return initAuthenticatedFlow(t, hc, false, true, opts...)
	}

	aal2Identity := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, &identity.Identity{
		State:  identity.StateActive,
		Traits: []byte(`{"email":"foo@bar"}`),
		Credentials: map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword, Config: []byte(`{"hashed_password":"$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`), Identifiers: []string{"foo@bar"}},
			identity.CredentialsTypeWebAuthn: {Type: identity.CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":false}]}`), Identifiers: []string{"foo@bar"}},
		},
	})

	t.Run("endpoint=init", func(t *testing.T) {
		t.Run("description=init a flow as API", func(t *testing.T) {
			t.Run("description=without privileges", func(t *testing.T) {
				res, body := initFlow(t, new(http.Client), true)
				assert.Equalf(t, http.StatusUnauthorized, res.StatusCode, "%s", body)
				assert.Equalf(t, text.ErrNoActiveSession, gjson.GetBytes(body, "error.id").String(), "%s", body)
			})

			t.Run("description=success", func(t *testing.T) {
				user1 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
				res, body := initFlow(t, user1, true)
				assert.Contains(t, res.Request.URL.String(), settings.RouteInitAPIFlow)
				assertion(t, body, true)
			})

			t.Run("description=can not init if identity has aal2 but session has aal1", func(t *testing.T) {
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
				res, body := initFlow(t, aal2Identity, true)
				assert.Equalf(t, http.StatusForbidden, res.StatusCode, "%s", body)
				assertx.EqualAsJSON(t, session.NewErrAALNotSatisfied(publicTS.URL+"/self-service/login/browser?aal=aal2"), json.RawMessage(body))
			})
		})

		t.Run("description=init a flow as browser", func(t *testing.T) {
			t.Run("case=unauthorized users are redirected to login preserving redirect_to param", func(t *testing.T) {
				c := testhelpers.NewClientWithCookies(t)
				// prevent the redirect
				c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
				returnTo := "?return_to=validRedirect"
				req, err := http.NewRequest("GET", publicTS.URL+settings.RouteInitBrowserFlow+returnTo, nil)
				require.NoError(t, err)

				res, err := c.Do(req)
				require.NoError(t, err)
				defer res.Body.Close()
				// here we check that the redirect status is 303
				require.Equal(t, http.StatusSeeOther, res.StatusCode)
				location, err := res.Location()
				require.NoError(t, err)
				require.Equal(t, publicTS.URL+login.RouteInitBrowserFlow+returnTo, location.String())
			})

			t.Run("case=unauthorized users are redirected to login", func(t *testing.T) {
				c := testhelpers.NewClientWithCookies(t)
				// prevent the redirect
				c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
				req, err := http.NewRequest("GET", publicTS.URL+settings.RouteInitBrowserFlow, nil)
				require.NoError(t, err)

				res, err := c.Do(req)
				require.NoError(t, err)
				defer res.Body.Close()
				// here we check that the redirect status is 303
				require.Equal(t, http.StatusSeeOther, res.StatusCode)
				location, err := res.Location()
				require.NoError(t, err)
				require.Equal(t, publicTS.URL+login.RouteInitBrowserFlow, location.String())
			})

			t.Run("description=success", func(t *testing.T) {
				user1 := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)
				res, body := initFlow(t, user1, false)
				assert.Contains(t, res.Request.URL.String(), reg.Config().SelfServiceFlowSettingsUI(ctx).String())
				assertion(t, body, false)
			})

			t.Run("description=can not init if identity has aal2 but session has aal1", func(t *testing.T) {
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
				res, body := initFlow(t, aal2Identity, false)
				assert.Contains(t, res.Request.URL.String(), reg.Config().SelfServiceFlowLoginUI(ctx).String())
				assert.EqualValues(t, "Please complete the second authentication challenge.", gjson.GetBytes(body, "ui.messages.0.text").String(), "%s", body)
			})

			t.Run("description=settings return_to is persisted through aal2 flow", func(t *testing.T) {
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
				res, _ := initFlow(t, aal2Identity, false, WithInitAuthQuery(url.Values{"return_to": {"https://www.ory.sh"}}))
				assert.Contains(t, res.Request.URL.String(), reg.Config().SelfServiceFlowLoginUI(ctx).String())

				lf, err := reg.LoginFlowPersister().GetLoginFlow(ctx, uuid.FromStringOrNil(res.Request.URL.Query().Get("flow")))
				require.NoError(t, err)

				reqURL, err := url.Parse(lf.RequestURL)
				require.NoError(t, err)

				// the return_to of login will be set to the full URL of the settings URL with its query parameters
				settingsURL, err := url.Parse(publicTS.URL + settings.RouteInitBrowserFlow)
				require.NoError(t, err)
				q := settingsURL.Query()
				q.Set("return_to", "https://www.ory.sh")
				settingsURL.RawQuery = q.Encode()

				assert.Equal(t, settingsURL.String(), reqURL.Query().Get("return_to"))
			})

			t.Run("case=redirects with 303", func(t *testing.T) {
				c := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)
				// prevent the redirect
				c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
				req, err := http.NewRequest("GET", publicTS.URL+settings.RouteInitBrowserFlow, nil)
				require.NoError(t, err)

				res, err := c.Do(req)
				require.NoError(t, err)
				defer res.Body.Close()
				// here we check that the redirect status is 303
				require.Equal(t, http.StatusSeeOther, res.StatusCode)
				location, err := res.Location()
				require.NoError(t, err)
				require.Contains(t, location.String(), conf.SelfServiceFlowSettingsUI(ctx).String())
			})
		})

		t.Run("description=init a flow as SPA", func(t *testing.T) {
			t.Run("description=without privileges", func(t *testing.T) {
				res, body := initSPAFlow(t, new(http.Client))
				assert.Equalf(t, http.StatusUnauthorized, res.StatusCode, "%s", body)
				assert.Equalf(t, text.ErrNoActiveSession, gjson.GetBytes(body, "error.id").String(), "%s", body)
			})

			t.Run("description=success", func(t *testing.T) {
				user1 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
				res, body := initSPAFlow(t, user1)
				assert.Contains(t, res.Request.URL.String(), settings.RouteInitBrowserFlow)
				assertion(t, body, false)
			})

			t.Run("description=can not init if identity has aal2 but session has aal1", func(t *testing.T) {
				email := testhelpers.RandomEmail()
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
				user1 := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, &identity.Identity{
					State:  identity.StateActive,
					Traits: []byte(`{"email":"` + email + `"}`),
					Credentials: map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword, Config: []byte(`{"hashed_password":"$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`), Identifiers: []string{email}},
						identity.CredentialsTypeWebAuthn: {Type: identity.CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":false}]}`), Identifiers: []string{email}},
					},
				})
				res, body := initSPAFlow(t, user1)
				assert.Equal(t, http.StatusForbidden, res.StatusCode)

				url := url.URL{
					Scheme: conf.SelfPublicURL(ctx).Scheme,
					Host:   conf.SelfPublicURL(ctx).Host,
					Path:   login.RouteInitBrowserFlow,
				}
				q := url.Query()
				q.Add("aal", "aal2")
				url.RawQuery = q.Encode()

				assertx.EqualAsJSON(t, session.NewErrAALNotSatisfied(url.String()), json.RawMessage(body))
			})

			t.Run("description=settings return_to should persist through mfa flows", func(t *testing.T) {
				email := testhelpers.RandomEmail()
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
				user1 := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, &identity.Identity{
					State:  identity.StateActive,
					Traits: []byte(`{"email":"` + email + `"}`),
					Credentials: map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword, Config: []byte(`{"hashed_password":"$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`), Identifiers: []string{email}},
						identity.CredentialsTypeWebAuthn: {Type: identity.CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":false}]}`), Identifiers: []string{email}},
					},
				})
				res, body := initSPAFlow(t, user1, WithInitAuthQuery(url.Values{"return_to": {"https://www.ory.sh/"}}))
				assert.Equal(t, http.StatusForbidden, res.StatusCode)

				returnToURL := gjson.GetBytes(body, "redirect_browser_to").String()
				require.NotEmpty(t, returnToURL)

				reqURL, err := url.Parse(gjson.GetBytes(body, "redirect_browser_to").String())
				require.NoError(t, err)

				settingsURL, err := url.Parse(publicTS.URL + settings.RouteInitBrowserFlow)
				require.NoError(t, err)

				q := settingsURL.Query()
				q.Add("return_to", "https://www.ory.sh/")
				settingsURL.RawQuery = q.Encode()

				assert.Equal(t, settingsURL.String(), reqURL.Query().Get("return_to"))
			})
		})
	})

	t.Run("endpoint=fetch", func(t *testing.T) {
		t.Run("description=fetching a non-existent flow should return a 404 error", func(t *testing.T) {
			_, _, err := testhelpers.NewSDKCustomClient(publicTS, otherUser).FrontendApi.GetSettingsFlow(context.Background()).Id("i-do-not-exist").Execute()
			require.Error(t, err)

			require.IsTypef(t, new(kratos.GenericOpenAPIError), err, "%T", err)
			assert.Equal(t, int64(http.StatusNotFound), gjson.GetBytes(err.(*kratos.GenericOpenAPIError).Body(), "error.code").Int())
		})

		t.Run("description=fetching an expired flow returns 410", func(t *testing.T) {
			pr := newExpiredFlow()
			require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(context.Background(), pr))

			_, _, err := testhelpers.NewSDKCustomClient(publicTS, primaryUser).FrontendApi.GetSettingsFlow(context.Background()).Id(pr.ID.String()).Execute()
			require.Error(t, err)

			require.IsTypef(t, new(kratos.GenericOpenAPIError), err, "%T", err)
			assert.Equal(t, int64(http.StatusGone), gjson.GetBytes(err.(*kratos.GenericOpenAPIError).Body(), "error.code").Int())
		})

		t.Run("case=expired with return_to", func(t *testing.T) {
			returnTo := "https://www.ory.sh"
			conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTo})

			client := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
			body := x.EasyGetBody(t, client, publicTS.URL+settings.RouteInitBrowserFlow+"?return_to="+returnTo)

			// Expire the flow
			f, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
			require.NoError(t, err)
			f.ExpiresAt = time.Now().Add(-time.Second)
			require.NoError(t, reg.SettingsFlowPersister().UpdateSettingsFlow(context.Background(), f))

			// Retrieve the flow and verify that return_to is in the response
			getURL := fmt.Sprintf("%s%s?id=%s&return_to=%s", publicTS.URL, settings.RouteGetFlow, f.ID, returnTo)
			getBody := x.EasyGetBody(t, client, getURL)
			assert.Equal(t, gjson.GetBytes(getBody, "error.details.return_to").String(), returnTo)

			// submit the flow but it is expired
			u := publicTS.URL + settings.RouteSubmitFlow + "?flow=" + f.ID.String()
			res, err := client.PostForm(u, url.Values{"method": {"password"}, "csrf_token": {"csrf"}, "password": {"password"}})
			require.NoError(t, err)
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())

			f, err = reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(resBody, "id").String()))
			require.NoError(t, err)
			assert.Equal(t, publicTS.URL+settings.RouteInitBrowserFlow+"?return_to="+returnTo, f.RequestURL)
		})

		t.Run("description=should fail to fetch request if identity changed", func(t *testing.T) {
			t.Run("type=api", func(t *testing.T) {
				user1 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
				user2 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)

				res, err := user1.Get(publicTS.URL + settings.RouteInitAPIFlow)
				require.NoError(t, err)
				defer res.Body.Close()

				assert.Len(t, res.Header.Get("Set-Cookie"), 0)
				body := ioutilx.MustReadAll(res.Body)
				id := gjson.GetBytes(body, "id")
				require.NotEmpty(t, id)

				res, err = user2.Get(publicTS.URL + settings.RouteGetFlow + "?id=" + id.String())
				require.NoError(t, err)
				defer res.Body.Close()

				require.EqualValues(t, res.StatusCode, http.StatusForbidden)
				body = ioutilx.MustReadAll(res.Body)
				assert.Containsf(t, gjson.GetBytes(body, "error.reason").String(), "The request was made for another identity and has been blocked for security reasons", "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				res, err := primaryUser.Get(publicTS.URL + settings.RouteInitBrowserFlow)
				require.NoError(t, err)

				rid := res.Request.URL.Query().Get("flow")
				require.NotEmpty(t, rid)

				_, _, err = testhelpers.NewSDKCustomClient(publicTS, otherUser).FrontendApi.GetSettingsFlow(context.Background()).Id(rid).Execute()
				require.Error(t, err)
				require.IsTypef(t, new(kratos.GenericOpenAPIError), err, "%T", err)
				assert.Equal(t, int64(http.StatusForbidden), gjson.GetBytes(err.(*kratos.GenericOpenAPIError).Body(), "error.code").Int(), "should return a 403 error because the identities from the cookies do not match")
			})
		})

		t.Run("description=can not fetch if identity has aal2 but session has aal1", func(t *testing.T) {
			t.Cleanup(func() {
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
			})
			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, "aal1")

			res, body := initFlow(t, aal2Identity, false)
			require.Equal(t, http.StatusOK, res.StatusCode)

			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
			flowID := gjson.GetBytes(body, "id").String()
			settingsURL := publicTS.URL + settings.RouteGetFlow + "?id=" + flowID
			res, err := aal2Identity.Get(settingsURL)
			require.NoError(t, err)
			body = ioutilx.MustReadAll(res.Body)
			require.NoError(t, res.Body.Close())

			url := url.URL{
				Scheme: conf.SelfPublicURL(ctx).Scheme,
				Host:   conf.SelfPublicURL(ctx).Host,
				Path:   login.RouteInitBrowserFlow,
			}

			returnTo := conf.SelfServiceFlowSettingsUI(context.Background())
			rq := returnTo.Query()
			// the flow ID should be perserved within the return_to query parameter of the login flow
			rq.Set("flow", flowID)
			returnTo.RawQuery = rq.Encode()

			q := url.Query()
			q.Set("aal", "aal2")
			q.Set("return_to", returnTo.String())
			url.RawQuery = q.Encode()

			require.EqualValues(t, http.StatusForbidden, res.StatusCode)
			assertx.EqualAsJSON(t, session.NewErrAALNotSatisfied(url.String()), json.RawMessage(body))
		})
	})

	t.Run("endpoint=submit", func(t *testing.T) {
		t.Run("description=can not submit if identity has aal2 but session has aal1", func(t *testing.T) {
			t.Run("type=browser", func(t *testing.T) {
				t.Cleanup(func() {
					conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
				})
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, "aal1")

				res, body := initFlow(t, aal2Identity, false)
				require.Equal(t, http.StatusOK, res.StatusCode)
				var f kratos.SettingsFlow
				require.NoError(t, json.Unmarshal(body, &f))

				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
				actual, res := testhelpers.SettingsMakeRequest(t, false, false, &f, aal2Identity, `{"method":"not-exists"}`)
				assert.Equal(t, http.StatusOK, res.StatusCode)
				assert.Equal(t, "Please complete the second authentication challenge.", gjson.Get(actual, "ui.messages.0.text").String(), actual)
			})

			t.Run("type=spa", func(t *testing.T) {
				t.Cleanup(func() {
					conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
				})
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, "aal1")

				res, body := initFlow(t, aal2Identity, false)
				require.Equal(t, http.StatusOK, res.StatusCode)
				var f kratos.SettingsFlow
				require.NoError(t, json.Unmarshal(body, &f))

				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
				actual, res := testhelpers.SettingsMakeRequest(t, false, true, &f, aal2Identity, `{"method":"not-exists"}`)
				assert.Equal(t, http.StatusForbidden, res.StatusCode)

				url := url.URL{
					Scheme: conf.SelfPublicURL(ctx).Scheme,
					Host:   conf.SelfPublicURL(ctx).Host,
					Path:   login.RouteInitBrowserFlow,
				}
				q := url.Query()
				q.Set("aal", "aal2")
				q.Set("return_to", publicTS.URL+"/self-service/settings?flow="+f.GetId())
				url.RawQuery = q.Encode()

				assert.Equal(t, url.String(), gjson.Get(actual, "redirect_browser_to").String(), actual)
			})

			t.Run("type=api", func(t *testing.T) {
				t.Cleanup(func() {
					conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
				})
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, "aal1")

				res, body := initFlow(t, aal2Identity, true)
				require.Equal(t, http.StatusOK, res.StatusCode)
				var f kratos.SettingsFlow
				require.NoError(t, json.Unmarshal(body, &f))

				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
				actual, res := testhelpers.SettingsMakeRequest(t, true, false, &f, aal2Identity, `{"method":"not-exists"}`)
				assert.Equal(t, http.StatusForbidden, res.StatusCode)

				url := url.URL{
					Scheme: conf.SelfPublicURL(ctx).Scheme,
					Host:   conf.SelfPublicURL(ctx).Host,
					Path:   login.RouteInitBrowserFlow,
				}

				q := url.Query()
				q.Set("aal", "aal2")
				q.Set("return_to", publicTS.URL+"/self-service/settings?flow="+f.GetId())
				url.RawQuery = q.Encode()
				assert.Equal(t, url.String(), gjson.Get(actual, "redirect_browser_to").String(), actual)
			})
		})

		t.Run("description=fail to submit form as another user", func(t *testing.T) {
			t.Run("type=api", func(t *testing.T) {
				user1 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
				user2 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
				_, body := initFlow(t, user1, true)
				var f kratos.SettingsFlow
				require.NoError(t, json.Unmarshal(body, &f))

				actual, res := testhelpers.SettingsMakeRequest(t, true, false, &f, user2, `{"method":"not-exists"}`)
				assert.Equal(t, http.StatusBadRequest, res.StatusCode)
				assert.Equal(t, "You must restart the flow because the resumable session was initiated by another person.", gjson.Get(actual, "ui.messages.0.text").String(), actual)
			})

			t.Run("type=spa", func(t *testing.T) {
				user1 := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)
				user2 := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)
				_, body := initFlow(t, user1, true)
				var f kratos.SettingsFlow
				require.NoError(t, json.Unmarshal(body, &f))

				actual, res := testhelpers.SettingsMakeRequest(t, false, true, &f, user2, `{"method":"not-exists"}`)
				assert.Equal(t, http.StatusBadRequest, res.StatusCode)
				assert.Equal(t, "You must restart the flow because the resumable session was initiated by another person.", gjson.Get(actual, "ui.messages.0.text").String(), actual)
			})

			t.Run("type=browser", func(t *testing.T) {
				user1 := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)
				user2 := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)
				_, body := initFlow(t, user1, true)
				var f kratos.SettingsFlow
				require.NoError(t, json.Unmarshal(body, &f))

				actual, res := testhelpers.SettingsMakeRequest(t, false, false, &f, user2, `{"method":"not-exists"}`)
				assert.Equal(t, http.StatusBadRequest, res.StatusCode)
				assert.Equal(t, "You must restart the flow because the resumable session was initiated by another person.", gjson.Get(actual, "ui.messages.0.text").String(), actual)
			})
		})

		t.Run("description=submit - kratos session cookie issued", func(t *testing.T) {
			t.Run("type=spa", func(t *testing.T) {
				_, body := initFlow(t, primaryUser, false)
				var f kratos.SettingsFlow
				require.NoError(t, json.Unmarshal(body, &f))

				actual, res := testhelpers.SettingsMakeRequest(t, false, true, &f, primaryUser, fmt.Sprintf(`{"method":"profile", "numby": 15, "csrf_token": "%s"}`, x.FakeCSRFToken))
				assert.Equal(t, http.StatusOK, res.StatusCode)
				require.Len(t, primaryUser.Jar.Cookies(urlx.ParseOrPanic(publicTS.URL+login.RouteGetFlow)), 1)
				require.Contains(t, fmt.Sprintf("%v", primaryUser.Jar.Cookies(urlx.ParseOrPanic(publicTS.URL))), "ory_kratos_session")
				assert.Equal(t, "Your changes have been saved!", gjson.Get(actual, "ui.messages.0.text").String(), actual)
			})

			t.Run("type=browser", func(t *testing.T) {
				_, body := initFlow(t, primaryUser, false)
				var f kratos.SettingsFlow
				require.NoError(t, json.Unmarshal(body, &f))

				actual, res := testhelpers.SettingsMakeRequest(t, false, false, &f, primaryUser, `method=profile&traits.numby=15&csrf_token=`+x.FakeCSRFToken)
				assert.Equal(t, http.StatusOK, res.StatusCode)
				require.Len(t, primaryUser.Jar.Cookies(urlx.ParseOrPanic(publicTS.URL+login.RouteGetFlow)), 1)
				require.Contains(t, fmt.Sprintf("%v", primaryUser.Jar.Cookies(urlx.ParseOrPanic(publicTS.URL))), "ory_kratos_session")
				assert.Equal(t, "Your changes have been saved!", gjson.Get(actual, "ui.messages.0.text").String(), actual)
			})
		})
	})

	t.Run("case=relative redirect when self-service settings ui is a relative url", func(t *testing.T) {
		reg.Config().MustSet(ctx, config.ViperKeySelfServiceSettingsURL, "/settings-ts")
		user1 := testhelpers.NewNoRedirectHTTPClientWithArbitrarySessionCookie(t, reg)
		res, _ := initFlow(t, user1, false)
		assert.Regexp(
			t,
			"^/settings-ts.*$",
			res.Header.Get("Location"),
		)
	})
}
