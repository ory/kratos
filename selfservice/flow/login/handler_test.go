// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	stdtotp "github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/corpx"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/strategy/totp"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/assertx"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

func init() {
	corpx.RegisterFakes()
}

func TestFlowLifecycle(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	reg.SetHydra(hydra.NewFake())
	router := x.NewRouterPublic(reg)
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin(reg))
	loginTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)

	returnToTS := testhelpers.NewRedirTS(t, "return_to", conf)
	errorTS := testhelpers.NewErrorTestServer(t, reg)
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, returnToTS.URL)

	conf.MustSet(ctx, config.ViperKeyIdentitySchemas, config.Schemas{
		{ID: "default", URL: "file://./stub/password.schema.json"},
		{ID: "email", URL: "file://./stub/email.schema.json", SelfserviceSelectable: true},
		{ID: "phone", URL: "file://./stub/phone.schema.json", SelfserviceSelectable: true},
		{ID: "not-allowed", URL: "file://./stub/password.schema.json"},
	})
	conf.MustSet(ctx, config.ViperKeyDefaultIdentitySchemaID, "default")

	assertion := func(body []byte, isForced, isApi bool) {
		r := gjson.GetBytes(body, "refresh")
		assert.True(t, r.Exists(), "%s", body)
		assert.Equal(t, isForced, r.Bool(), "%s", body)
		if isApi {
			assert.Equal(t, "api", gjson.GetBytes(body, "type").String())
		} else {
			assert.Equal(t, "browser", gjson.GetBytes(body, "type").String())
		}
	}

	initAuthenticatedFlow := func(t *testing.T, extQuery url.Values, isAPI bool) (*http.Response, []byte) {
		route := login.RouteInitBrowserFlow
		if isAPI {
			route = login.RouteInitAPIFlow
		}
		req := testhelpers.NewTestHTTPRequest(t, "GET", ts.URL+route, nil)
		req.URL.RawQuery = extQuery.Encode()
		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router, req)
		if isAPI {
			assert.Len(t, res.Header.Get("Set-Cookie"), 0)
		}
		return res, body
	}

	initUnauthenticatedFlow := func(t *testing.T, extQuery url.Values, isAPI bool) (*http.Response, []byte) {
		route := login.RouteInitBrowserFlow
		if isAPI {
			route = login.RouteInitAPIFlow
		}
		client := ts.Client()
		req := testhelpers.NewTestHTTPRequest(t, "GET", ts.URL+route, nil)

		req.URL.RawQuery = extQuery.Encode()
		res, err := client.Do(req)
		require.NoError(t, errors.WithStack(err))

		body, err := io.ReadAll(res.Body)
		require.NoError(t, errors.WithStack(err))
		require.NoError(t, res.Body.Close())
		return res, body
	}

	initFlowWithAccept := func(t *testing.T, query url.Values, isAPI bool, accept string) (*http.Response, []byte) {
		route := login.RouteInitBrowserFlow
		if isAPI {
			route = login.RouteInitAPIFlow
		}
		c := ts.Client()
		req, err := http.NewRequest("GET", ts.URL+route+"?"+query.Encode(), nil)
		require.NoError(t, err)
		if accept != "" {
			req.Header.Set("Accept", accept)
		}

		res, err := c.Do(req)
		require.NoError(t, err)
		defer func() { _ = res.Body.Close() }()
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		return res, body
	}

	initFlow := func(t *testing.T, query url.Values, isAPI bool) (*http.Response, []byte) {
		return initFlowWithAccept(t, query, isAPI, "")
	}

	initSPAFlow := func(t *testing.T, query url.Values) (*http.Response, []byte) {
		return initFlowWithAccept(t, query, false, "application/json")
	}

	id1mail, id2mail := x.NewUUID().String(), x.NewUUID().String()
	identity1 := &identity.Identity{
		Credentials: map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypePassword: {
				Type:        identity.CredentialsTypePassword,
				Identifiers: []string{id1mail},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`), // foobar
			},
		},
		State:  identity.StateActive,
		Traits: identity.Traits(`{"username":"` + id1mail + `"}`),
	}
	identity2 := &identity.Identity{
		Credentials: map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypePassword: {
				Type:        identity.CredentialsTypePassword,
				Identifiers: []string{id2mail},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`), // foobar
			},
		},
		State:  identity.StateActive,
		Traits: identity.Traits(`{"username":"` + id2mail + `"}`),
	}

	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), identity1))
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), identity2))

	t.Run("lifecycle=submit", func(t *testing.T) {
		t.Run("interaction=unauthenticated", func(t *testing.T) {
			run := func(t *testing.T, tt flow.Type, aal string, values url.Values) (string, *http.Response) {
				f := login.Flow{
					Type: tt, ExpiresAt: time.Now().Add(time.Minute), IssuedAt: time.Now(),
					UI: container.New(""), Refresh: false, RequestedAAL: identity.AuthenticatorAssuranceLevel(aal),
				}
				require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), &f))

				res, err := http.PostForm(ts.URL+login.RouteSubmitFlow+"?flow="+f.ID.String(), values)
				require.NoError(t, err)
				body := x.MustReadAll(res.Body)
				require.NoError(t, res.Body.Close())
				return string(body), res
			}

			t.Run("case=ensure aal can only be upgraded with a session on submit", func(t *testing.T) {
				t.Run("type=api", func(t *testing.T) {
					body, res := run(t, flow.TypeAPI, "aal2", url.Values{"method": {"password"}})
					assert.Contains(t, res.Request.URL.String(), login.RouteSubmitFlow)
					assertx.EqualAsJSON(t, "You can not requested a higher AAL (AAL2/AAL3) without an active session.", gjson.Get(body, "error.reason").String())
				})

				t.Run("type=browser", func(t *testing.T) {
					body, res := run(t, flow.TypeBrowser, "aal2", url.Values{"method": {"password"}})
					assert.Contains(t, res.Request.URL.String(), errorTS.URL)
					assertx.EqualAsJSON(t, "You can not requested a higher AAL (AAL2/AAL3) without an active session.", gjson.Get(body, "reason").String())
				})
			})

			t.Run("case=end up with method missing when aal is ok", func(t *testing.T) {
				t.Run("type=api", func(t *testing.T) {
					body, res := run(t, flow.TypeAPI, "aal1", url.Values{"method": {"not-exist"}})
					assert.Contains(t, res.Request.URL.String(), login.RouteSubmitFlow)
					assertx.EqualAsJSON(t, text.NewErrorValidationLoginNoStrategyFound().Text, gjson.Get(body, "ui.messages.0.text").String(), body)
				})

				t.Run("type=browser", func(t *testing.T) {
					body, res := run(t, flow.TypeBrowser, "aal1", url.Values{"method": {"not-exist"}})
					assert.Contains(t, res.Request.URL.String(), loginTS.URL)
					assertx.EqualAsJSON(t, text.NewErrorValidationLoginNoStrategyFound().Text, gjson.Get(body, "ui.messages.0.text").String(), body)
				})
			})

			t.Run("case=reset the session when refresh is true but identity is different", func(t *testing.T) {
				testhelpers.NewRedirSessionEchoTS(t, reg)
				t.Cleanup(func() {
					conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, returnToTS.URL)
				})

				run := func(t *testing.T, tt flow.Type) (string, string) {
					f := login.Flow{Type: tt, ExpiresAt: time.Now().Add(time.Minute), IssuedAt: time.Now(), UI: container.New(""), Refresh: false, RequestedAAL: "aal1"}
					require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), &f))

					hc := testhelpers.NewClientWithCookies(t)
					res, err := hc.PostForm(ts.URL+login.RouteSubmitFlow+"?flow="+f.ID.String(), url.Values{"method": {"password"}, "password_identifier": {id1mail}, "password": {"foobar"}, "csrf_token": {nosurfx.FakeCSRFToken}})
					require.NoError(t, err)
					firstSession := x.MustReadAll(res.Body)
					require.NoError(t, res.Body.Close())

					f = login.Flow{Type: tt, ExpiresAt: time.Now().Add(time.Minute), IssuedAt: time.Now(), UI: container.New(""), Refresh: true, RequestedAAL: "aal1"}
					require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), &f))

					vv := testhelpers.EncodeFormAsJSON(t, tt == flow.TypeAPI, url.Values{"method": {"password"}, "password_identifier": {id2mail}, "password": {"foobar"}, "csrf_token": {nosurfx.FakeCSRFToken}})

					req, err := http.NewRequest("POST", ts.URL+login.RouteSubmitFlow+"?flow="+f.ID.String(), strings.NewReader(vv))
					require.NoError(t, err)

					if tt == flow.TypeAPI {
						req.Header.Set("Content-Type", "application/json")
						req.Header.Set("Authorization", "Bearer "+gjson.GetBytes(firstSession, "session_token").String())
					} else {
						req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					}

					res, err = hc.Do(req)
					require.NoError(t, err)
					secondSession := x.MustReadAll(res.Body)
					require.NoError(t, res.Body.Close())
					return string(firstSession), string(secondSession)
				}

				t.Run("type=browser", func(t *testing.T) {
					a, b := run(t, flow.TypeBrowser)

					assert.Equal(t, id1mail, gjson.Get(a, "identity.traits.username").String())
					assert.Equal(t, id2mail, gjson.Get(b, "identity.traits.username").String())

					assert.NotEmpty(t, gjson.Get(b, "id").String())
					assert.NotEqual(t, gjson.Get(b, "id").String(), gjson.Get(a, "id").String())
				})

				t.Run("type=api", func(t *testing.T) {
					a, b := run(t, flow.TypeAPI)

					assert.Equal(t, id1mail, gjson.Get(a, "session.identity.traits.username").String())
					assert.Equal(t, id2mail, gjson.Get(b, "session.identity.traits.username").String())

					assert.NotEmpty(t, gjson.Get(a, "session_token").String())
					assert.NotEqual(t, gjson.Get(a, "session_token").String(), gjson.Get(b, "session_token").String())

					assert.NotEmpty(t, gjson.Get(b, "session.id").String())
					assert.NotEqual(t, gjson.Get(b, "session.id").String(), gjson.Get(a, "id").String())
				})
			})

			t.Run("case=changed kratos session identifiers when refresh is true", func(t *testing.T) {
				t.Cleanup(func() {
					conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, returnToTS.URL)
				})

				t.Run("type=browser", func(t *testing.T) {
					// Setup flow
					f := login.Flow{Type: flow.TypeBrowser, ExpiresAt: time.Now().Add(time.Minute), IssuedAt: time.Now(), UI: container.New(""), Refresh: false, RequestedAAL: "aal1"}
					require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), &f))

					// Submit Login
					hc := testhelpers.NewClientWithCookies(t)
					res, err := hc.PostForm(ts.URL+login.RouteSubmitFlow+"?flow="+f.ID.String(), url.Values{"method": {"password"}, "password_identifier": {id1mail}, "password": {"foobar"}, "csrf_token": {nosurfx.FakeCSRFToken}})
					require.NoError(t, err)

					// Check response and session cookie presence
					assert.Equal(t, http.StatusOK, res.StatusCode)
					require.Len(t, hc.Jar.Cookies(urlx.ParseOrPanic(ts.URL+login.RouteGetFlow)), 1)
					require.Contains(t, fmt.Sprintf("%v", hc.Jar.Cookies(urlx.ParseOrPanic(ts.URL))), "ory_kratos_session")
					cookies1 := hc.Jar.Cookies(urlx.ParseOrPanic(ts.URL + login.RouteGetFlow))

					req, err := http.NewRequest("GET", ts.URL+"/sessions/whoami", nil)
					require.NoError(t, err)

					res, err = hc.Do(req)
					require.NoError(t, err)
					assert.Equal(t, http.StatusOK, res.StatusCode)
					firstSession := x.MustReadAll(res.Body)
					require.NoError(t, res.Body.Close())

					// Refresh
					f = login.Flow{Type: flow.TypeBrowser, ExpiresAt: time.Now().Add(time.Minute), IssuedAt: time.Now(), UI: container.New(""), Refresh: true, RequestedAAL: "aal1"}
					require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), &f))

					vv := testhelpers.EncodeFormAsJSON(t, false, url.Values{"method": {"password"}, "password_identifier": {id1mail}, "password": {"foobar"}, "csrf_token": {nosurfx.FakeCSRFToken}})

					req, err = http.NewRequest("POST", ts.URL+login.RouteSubmitFlow+"?flow="+f.ID.String(), strings.NewReader(vv))
					require.NoError(t, err)
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

					// Submit Login
					res, err = hc.Do(req)
					require.NoError(t, err)

					// Check response and session cookie presence
					assert.Equal(t, http.StatusOK, res.StatusCode)
					require.Len(t, hc.Jar.Cookies(urlx.ParseOrPanic(ts.URL+login.RouteGetFlow)), 1)
					require.Contains(t, fmt.Sprintf("%v", hc.Jar.Cookies(urlx.ParseOrPanic(ts.URL))), "ory_kratos_session")
					cookies2 := hc.Jar.Cookies(urlx.ParseOrPanic(ts.URL + login.RouteGetFlow))

					req, err = http.NewRequest("GET", ts.URL+"/sessions/whoami", nil)
					require.NoError(t, err)

					res, err = hc.Do(req)
					require.NoError(t, err)
					assert.Equal(t, http.StatusOK, res.StatusCode)
					secondSession := x.MustReadAll(res.Body)
					require.NoError(t, res.Body.Close())

					// Sessions should still be resolvable despite different kratos session identifier due to nonce
					assert.NotEqual(t, cookies1[0].String(), cookies2[0].String())
					assert.Equal(t, id1mail, gjson.Get(string(firstSession), "identity.traits.username").String())
					assert.Equal(t, id1mail, gjson.Get(string(secondSession), "identity.traits.username").String())
					assert.Equal(t, gjson.Get(string(secondSession), "id").String(), gjson.Get(string(firstSession), "id").String())
				})
			})
		})

		t.Run("case=ensure aal is checked for upgradeability on session", func(t *testing.T) {
			run := func(t *testing.T, tt flow.Type, values url.Values) (string, *http.Response) {
				f := login.Flow{
					Type: tt, ExpiresAt: time.Now().Add(time.Minute), IssuedAt: time.Now(),
					UI: container.New(""), Refresh: false, RequestedAAL: "aal1",
				}
				require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), &f))

				req, err := http.NewRequest("GET", ts.URL+login.RouteSubmitFlow+"?flow="+f.ID.String(), strings.NewReader(values.Encode()))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router, req)
				return string(body), res
			}

			t.Run("type=api", func(t *testing.T) {
				body, res := run(t, flow.TypeAPI, url.Values{"method": {"password"}})
				assert.Contains(t, res.Request.URL.String(), login.RouteSubmitFlow)
				assertx.EqualAsJSON(t, login.ErrAlreadyLoggedIn.Reason(), gjson.Get(body, "ui.messages.0.text").String(), body)
			})

			t.Run("type=browser", func(t *testing.T) {
				_, res := run(t, flow.TypeBrowser, url.Values{"method": {"password"}})
				assert.Contains(t, res.Request.URL.String(), returnToTS.URL)
			})
		})

		t.Run("case=should return an error because the request is expired", func(t *testing.T) {
			conf.MustSet(ctx, config.ViperKeySelfServiceLoginRequestLifespan, "50ms")
			t.Cleanup(func() {
				conf.MustSet(ctx, config.ViperKeySelfServiceLoginRequestLifespan, "10m")
			})

			expired := time.Now().Add(-time.Minute)
			run := func(t *testing.T, tt flow.Type, aal string, values string, isSPA bool) (string, *http.Response) {
				f := login.Flow{
					Type: tt, ExpiresAt: expired, IssuedAt: time.Now(),
					UI: container.New(""), Refresh: false, RequestedAAL: identity.AuthenticatorAssuranceLevel(aal),
				}
				require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), &f))

				req, err := http.NewRequest("POST", ts.URL+login.RouteSubmitFlow+"?flow="+f.ID.String(), strings.NewReader(values))
				require.NoError(t, err)

				if isSPA {
					req.Header.Set("Accept", "application/json")
					req.Header.Set("Content-Type", "application/json")
				} else {
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				}

				res, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				body := x.MustReadAll(res.Body)
				require.NoError(t, res.Body.Close())
				return string(body), res
			}

			t.Run("type=api", func(t *testing.T) {
				actual, res := run(t, flow.TypeAPI, "aal1", `{"method":"password"}`, false)
				assert.Contains(t, res.Request.URL.String(), login.RouteSubmitFlow)
				assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", gjson.Get(actual, "use_flow_id").String())
				assertx.EqualAsJSONExcept(t, flow.NewFlowExpiredError(expired), json.RawMessage(actual), []string{"use_flow_id", "since"}, "expired", "%s", actual)
			})

			t.Run("type=browser", func(t *testing.T) {
				body, res := run(t, flow.TypeBrowser, "aal1", url.Values{"method": {"password"}}.Encode(), false)
				assert.Contains(t, res.Request.URL.String(), loginTS.URL)
				assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "expired", "%s", body)
			})

			t.Run("type=SPA", func(t *testing.T) {
				actual, res := run(t, flow.TypeBrowser, "aal1", `{"method":"password"}`, true)
				assert.Contains(t, res.Request.URL.String(), login.RouteSubmitFlow)
				assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", gjson.Get(actual, "use_flow_id").String())
				assertx.EqualAsJSONExcept(t, flow.NewFlowExpiredError(expired), json.RawMessage(actual), []string{"use_flow_id", "since"}, "expired", "%s", actual)
			})
		})

		t.Run("case=should return to settings flow after successful mfa login after recovery", func(t *testing.T) {
			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
			conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
			testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeTOTP.String(), true)
			conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnToTS.URL})

			t.Cleanup(func() {
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, string(identity.AuthenticatorAssuranceLevel1))
				conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, string(identity.AuthenticatorAssuranceLevel1))
				testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeTOTP.String(), false)
			})

			key, err := totp.NewKey(context.Background(), "foo", reg)
			require.NoError(t, err)
			email := testhelpers.RandomEmail()
			id := &identity.Identity{
				Credentials: map[identity.CredentialsType]identity.Credentials{
					"password": {
						Type:        "password",
						Identifiers: []string{email},
						Config:      sqlxx.JSONRawMessage(`{"hashed_password": "$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`),
					},
				},
				Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, email)),
				SchemaID: config.DefaultIdentityTraitsSchemaID,
			}

			require.NoError(t, reg.IdentityManager().CreateIdentities(context.Background(), []*identity.Identity{id}, identity.ManagerAllowWriteProtectedTraits))

			id.SetCredentials(identity.CredentialsTypeTOTP, identity.Credentials{
				Type:        identity.CredentialsTypeTOTP,
				Identifiers: []string{id.ID.String()},
				Config:      sqlxx.JSONRawMessage(`{"totp_url":"` + string(key.URL()) + `"}`),
			})
			require.NoError(t, reg.IdentityManager().Update(context.Background(), id, identity.ManagerAllowWriteProtectedTraits))

			h := func(w http.ResponseWriter, r *http.Request) {
				sess, err := testhelpers.NewActiveSession(r, reg, id, time.Now().UTC(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
				require.NoError(t, err)
				sess.AuthenticatorAssuranceLevel = identity.AuthenticatorAssuranceLevel1
				require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), sess))
				require.NoError(t, reg.SessionManager().IssueCookie(context.Background(), w, r, sess))
				require.Equal(t, identity.AuthenticatorAssuranceLevel1, sess.AuthenticatorAssuranceLevel)
			}

			router.GET("/mock-session", h)

			client := testhelpers.NewClientWithCookies(t)

			testhelpers.MockHydrateCookieClient(t, client, ts.URL+"/mock-session")

			settingsURL := ts.URL + settings.RouteInitBrowserFlow + "?return_to=" + url.QueryEscape(returnToTS.URL)
			req, err := http.NewRequest("GET", settingsURL, nil)
			require.NoError(t, err)

			// we initialize the settings flow with a session that has AAL1 set
			resp, err := client.Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			// we expect the request to redirect to the login flow because the AAL1 session is not sufficient
			requestURL, err := url.Parse(resp.Request.Referer())
			require.NoError(t, err)
			require.Equal(t, login.RouteInitBrowserFlow, requestURL.Path)
			require.Equal(t, "aal2", requestURL.Query().Get("aal"))
			require.Equal(t, settingsURL, requestURL.Query().Get("return_to"))

			// we expect to be on the login page now
			respURL := resp.Request.URL
			require.NoError(t, err)
			require.Equal(t, "/login-ts", respURL.Path)
			flowID := respURL.Query().Get("flow")
			require.NotEmpty(t, flowID)

			code, err := stdtotp.GenerateCode(key.Secret(), time.Now())
			require.NoError(t, err)

			req, err = http.NewRequest("GET", ts.URL+login.RouteGetFlow+"?id="+flowID, nil)
			require.NoError(t, err)

			req.Header.Add("Content-Type", "application/json")

			resp, err = client.Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			body := string(x.MustReadAll(resp.Body))
			defer func() { _ = resp.Body.Close() }()

			totpNode := gjson.Get(body, "ui.nodes.#(attributes.name==totp_code)").String()
			require.NotEmpty(t, totpNode)
			require.NotEmpty(t, gjson.Get(body, "ui.action").String())

			csrfToken := gjson.Get(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()

			req, err = http.NewRequest("POST", ts.URL+login.RouteSubmitFlow+"?flow="+flowID, strings.NewReader(url.Values{
				"method":     {"totp"},
				"totp_code":  {code},
				"csrf_token": {csrfToken},
			}.Encode()))

			require.NoError(t, err)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}

			resp, err = client.Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusSeeOther, resp.StatusCode)

			location, err := resp.Location()
			require.NoError(t, err)
			require.Equal(t, settings.RouteInitBrowserFlow, location.Path)
		})
	})

	t.Run("lifecycle=init", func(t *testing.T) {
		t.Run("suite=identity schema in query", func(t *testing.T) {
			for _, tc := range []struct {
				name           string
				query          url.Values
				wantErr        bool
				wantIdentifier string
			}{{
				name:    "not-allowed",
				query:   url.Values{"identity_schema": {"not-allowed"}},
				wantErr: true,
			}, {
				name:    "not-found",
				query:   url.Values{"identity_schema": {"not-found"}},
				wantErr: true,
			}, {
				name:           "phone",
				query:          url.Values{"identity_schema": {"phone"}},
				wantIdentifier: "Phone Number",
			}, {
				name:           "email",
				query:          url.Values{"identity_schema": {"email"}},
				wantIdentifier: "E-Mail Address",
			}, {
				name:           "default",
				query:          url.Values{"identity_schema": {"default"}},
				wantIdentifier: "Username",
			}} {
				t.Run("case="+tc.name, func(t *testing.T) {
					t.Run("flow=api", func(t *testing.T) {
						res, body := initFlow(t, tc.query, true)
						if tc.wantErr {
							assert.Equal(t, http.StatusBadRequest, res.StatusCode)
							return
						}
						assert.Equalf(t, tc.wantIdentifier, gjson.GetBytes(body, "ui.nodes.#(attributes.name==identifier).meta.label.text").String(), "%s", body)
					})

					t.Run("flow=browser", func(t *testing.T) {
						res, body := initFlow(t, tc.query, false)
						if tc.wantErr {
							require.Contains(t, res.Request.URL.String(), errorTS.URL, "%s", body)
							assert.EqualValues(t, "Bad Request", gjson.GetBytes(body, "status").String(), "%s", body)
							return
						}
						assert.Equalf(t, tc.wantIdentifier, gjson.GetBytes(body, "ui.nodes.#(attributes.name==identifier).meta.label.context.title").String(), "%s", body)
					})
				})
			}
		})

		t.Run("flow=api", func(t *testing.T) {
			t.Run("case=does not set forced flag on unauthenticated request", func(t *testing.T) {
				res, body := initFlow(t, url.Values{}, true)
				assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
				assertion(body, false, true)
				assert.Empty(t, gjson.GetBytes(body, "session_token_exchange_code").String())
			})

			t.Run("case=returns session exchange code with any truthy value", func(t *testing.T) {
				conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnToTS.URL, "https://example.com"})
				parameters := []string{"true", "True", "1"}

				for _, param := range parameters {
					t.Run("return_session_token_exchange_code="+param, func(t *testing.T) {
						res, body := initFlow(t, url.Values{
							"return_session_token_exchange_code": {param},
							"return_to":                          {"https://example.com/redirect"},
						}, true)
						assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
						assert.NotEmpty(t, gjson.GetBytes(body, "session_token_exchange_code").String())
						assert.Equal(t, "https://example.com/redirect", gjson.GetBytes(body, "return_to").String())
					})
				}
			})

			t.Run("case=can not request refresh and aal at the same time on unauthenticated request", func(t *testing.T) {
				res, body := initFlow(t, url.Values{"refresh": {"true"}, "aal": {"aal2"}}, true)
				assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
				assertx.EqualAsJSON(t, "You can not requested a higher AAL (AAL2/AAL3) without an active session.", gjson.GetBytes(body, "error.reason").String())
			})

			t.Run("case=can request refresh and aal at the same time on authenticated request", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"refresh": {"true"}, "aal": {"aal2"}}, true)
				assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
				assertx.EqualAsJSON(t, "Please confirm this action by verifying that it is you.", gjson.GetBytes(body, "ui.messages.0.text").String(), "%s", body)
				assertx.EqualAsJSON(t, "Please complete the second authentication challenge.", gjson.GetBytes(body, "ui.messages.1.text").String(), "%s", body)
			})

			t.Run("case=can not request aal2 on unauthenticated request", func(t *testing.T) {
				res, body := initFlow(t, url.Values{"aal": {"aal2"}}, true)
				assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
				assertx.EqualAsJSON(t, "You can not requested a higher AAL (AAL2/AAL3) without an active session.", gjson.GetBytes(body, "error.reason").String())
			})

			t.Run("case=ignores aal1 if session has aal1 already", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"aal": {"aal1"}}, true)
				assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
				assertx.EqualAsJSON(t, "A valid session was detected and thus login is not possible. Did you forget to set `?refresh=true`?", gjson.GetBytes(body, "error.reason").String())
			})

			t.Run("case=aal0 is not a valid value", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"aal": {"aal0"}}, true)
				assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
				assertx.EqualAsJSON(t, "Unable to parse AuthenticationMethod Assurance Level (AAL): expected one of [aal1, aal2] but got aal0", gjson.GetBytes(body, "error.reason").String())
			})

			t.Run("case=indicates two factor auth", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"aal": {"aal2"}}, true)
				assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
				assert.Equal(t, gjson.GetBytes(body, "ui.messages.0.text").String(), text.NewInfoLoginMFA().Text)
			})

			t.Run("case=does not set forced flag on unauthenticated request with refresh=true", func(t *testing.T) {
				res, body := initFlow(t, url.Values{"refresh": {"true"}}, true)
				assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
				assertion(body, false, true)
			})

			t.Run("case=does not set forced flag on authenticated request without refresh=true", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{}, true)
				assert.Equal(t, http.StatusBadRequest, res.StatusCode)
				assertx.EqualAsJSON(t, login.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
			})

			t.Run("case=does not set forced flag on authenticated request with refresh=false", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"refresh": {"false"}}, true)
				assert.Equal(t, http.StatusBadRequest, res.StatusCode)
				assertx.EqualAsJSON(t, login.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
			})

			t.Run("case=does set forced flag on authenticated request with refresh=true", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"refresh": {"true"}}, true)
				assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
				assertion(body, true, true)
			})

			t.Run("case=check info message on authenticated request with refresh=true", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"refresh": {"true"}}, true)
				assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
				assertion(body, true, true)
				assert.Equal(t, gjson.GetBytes(body, "ui.messages.0.text").String(), text.NewInfoLoginReAuth().Text)
			})
		})

		t.Run("flow=browser", func(t *testing.T) {
			t.Run("case=does not set forced flag on unauthenticated request", func(t *testing.T) {
				res, body := initFlow(t, url.Values{}, false)
				assertion(body, false, false)
				assert.Contains(t, res.Request.URL.String(), loginTS.URL)
			})

			t.Run("case=never returns a session token exchange code", func(t *testing.T) {
				_, body := initFlow(t, urlx.ParseOrPanic("/?return_session_token_exchange_code=true").Query(), false)
				assertion(body, false, false)
				assert.Empty(t, gjson.GetBytes(body, "session_token_exchange_code").String())
			})

			t.Run("case=can not request refresh and aal at the same time on unauthenticated request", func(t *testing.T) {
				res, body := initFlow(t, url.Values{"refresh": {"true"}, "aal": {"aal2"}}, false)
				assert.Contains(t, res.Request.URL.String(), errorTS.URL)
				assertx.EqualAsJSON(t, "You can not requested a higher AAL (AAL2/AAL3) without an active session.", gjson.GetBytes(body, "reason").String(), "%s", body)
			})

			t.Run("case=can request refresh and aal at the same time on authenticated request", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"refresh": {"true"}, "aal": {"aal2"}}, false)
				assert.Contains(t, res.Request.URL.String(), loginTS.URL)
				assertx.EqualAsJSON(t, "Please confirm this action by verifying that it is you.", gjson.GetBytes(body, "ui.messages.0.text").String(), "%s", body)
				assertx.EqualAsJSON(t, "Please complete the second authentication challenge.", gjson.GetBytes(body, "ui.messages.1.text").String(), "%s", body)
			})

			t.Run("case=redirects if aal2 is requested and set up already without refresh", func(t *testing.T) {
				res, _ := initAuthenticatedFlow(t, url.Values{"aal": {"aal2"}, "set_aal": {"aal2"}}, false)
				assert.Contains(t, res.Request.URL.String(), returnToTS.URL)
			})

			t.Run("case=can not request aal2 on unauthenticated request", func(t *testing.T) {
				res, body := initFlow(t, url.Values{"aal": {"aal2"}}, false)
				assert.Contains(t, res.Request.URL.String(), errorTS.URL)
				assertx.EqualAsJSON(t, "You can not requested a higher AAL (AAL2/AAL3) without an active session.", gjson.GetBytes(body, "reason").String())
			})

			t.Run("case=ignores aal1 if session has aal1 already", func(t *testing.T) {
				res, _ := initAuthenticatedFlow(t, url.Values{"aal": {"aal1"}}, false)
				assert.Contains(t, res.Request.URL.String(), returnToTS.URL)
			})

			t.Run("case=aal0 is not a valid value", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"aal": {"aal0"}}, false)
				assert.Contains(t, res.Request.URL.String(), errorTS.URL)
				assertx.EqualAsJSON(t, "Unable to parse AuthenticationMethod Assurance Level (AAL): expected one of [aal1, aal2] but got aal0", gjson.GetBytes(body, "reason").String())
			})

			t.Run("case=indicates two factor auth", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"aal": {"aal2"}}, false)
				assert.Contains(t, res.Request.URL.String(), loginTS.URL)
				assert.Equal(t, gjson.GetBytes(body, "ui.messages.0.text").String(), text.NewInfoLoginMFA().Text)
			})

			t.Run("case=makes request with JSON", func(t *testing.T) {
				res, body := initSPAFlow(t, url.Values{})
				assertion(body, false, false)
				assert.NotContains(t, res.Request.URL.String(), loginTS.URL)
			})

			t.Run("case=does not set forced flag on unauthenticated request with refresh=true", func(t *testing.T) {
				res, body := initFlow(t, url.Values{"refresh": {"true"}}, false)
				assertion(body, false, false)
				assert.Contains(t, res.Request.URL.String(), loginTS.URL)
			})

			t.Run("case=does not set forced flag on authenticated request without refresh=true", func(t *testing.T) {
				res, _ := initAuthenticatedFlow(t, url.Values{}, false)
				assert.Contains(t, res.Request.URL.String(), returnToTS.URL)
			})

			t.Run("case=does not set forced flag on authenticated request with refresh=false", func(t *testing.T) {
				res, _ := initAuthenticatedFlow(t, url.Values{"refresh": {"false"}}, false)
				assert.Contains(t, res.Request.URL.String(), returnToTS.URL)
			})

			t.Run("case=does set forced flag on authenticated request with refresh=true", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"refresh": {"true"}}, false)
				assertion(body, true, false)
				assert.Contains(t, res.Request.URL.String(), loginTS.URL)
			})

			t.Run("case=redirects with 303", func(t *testing.T) {
				c := &http.Client{}
				// don't get the reference, instead copy the values, so we don't alter the client directly.
				*c = *ts.Client()
				// prevent the redirect
				c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
				req, err := http.NewRequest("GET", ts.URL+login.RouteInitBrowserFlow, nil)
				require.NoError(t, err)

				res, err := c.Do(req)
				require.NoError(t, err)
				defer func() { _ = res.Body.Close() }()
				// here we check that the redirect status is 303
				require.Equal(t, http.StatusSeeOther, res.StatusCode)
			})

			t.Run("case=refuses to parse oauth2 login challenge when Hydra is not configured", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"login_challenge": {hydra.FakeValidLoginChallenge}}, false)
				require.Contains(t, res.Request.URL.String(), errorTS.URL)
				require.Contains(t, string(body), "refusing to parse")
			})

			conf.MustSet(ctx, config.ViperKeyOAuth2ProviderURL, "https://fake-hydra")

			t.Run("case=oauth2 flow init should override return_to to the oauth2 request_url", func(t *testing.T) {
				conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh", "https://example.com"})
				conf.MustSet(ctx, config.ViperKeyOAuth2ProviderOverrideReturnTo, true)

				t.Cleanup(func() {
					conf.MustSet(ctx, config.ViperKeyOAuth2ProviderOverrideReturnTo, false)
				})

				res, _ := initUnauthenticatedFlow(t, url.Values{
					"return_to":       {"https://example.com"},
					"login_challenge": {hydra.FakeValidLoginChallenge},
				}, false)
				require.Equal(t, http.StatusOK, res.StatusCode)
				require.Contains(t, res.Request.URL.String(), loginTS.URL)

				c := ts.Client()
				req := testhelpers.NewTestHTTPRequest(t, "GET", ts.URL+login.RouteGetFlow, nil)
				req.URL.RawQuery = url.Values{"id": {res.Request.URL.Query().Get("flow")}}.Encode()

				res, err := c.Do(req)
				require.NoError(t, err)

				body, err := io.ReadAll(res.Body)
				require.NoError(t, errors.WithStack(err))

				require.NoError(t, res.Body.Close())

				assert.Equal(t, "https://www.ory.sh", gjson.GetBytes(body, "return_to").Value())
			})

			t.Run("case=invalid oauth2 login challenge returns 400 Bad Request", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"login_challenge": {hydra.FakeInvalidLoginChallenge}}, false)
				assert.Contains(t, res.Request.URL.String(), errorTS.URL)
				assert.Equal(t, int64(http.StatusBadRequest), gjson.GetBytes(body, "code").Int())
				assert.Contains(t, gjson.GetBytes(body, "reason").String(), "Unable to get OAuth 2.0 Login Challenge")
			})

			t.Run("case=oauth2 flow init succeeds", func(t *testing.T) {
				res, _ := initAuthenticatedFlow(t, url.Values{"login_challenge": {hydra.FakeValidLoginChallenge}}, false)
				require.Contains(t, res.Request.URL.String(), loginTS.URL)
			})

			t.Run("case=oauth2 flow init adds oauth2_login_request field", func(t *testing.T) {
				res, body := initSPAFlow(t, url.Values{"login_challenge": {hydra.FakeValidLoginChallenge}})
				assert.NotContains(t, res.Request.URL.String(), loginTS.URL)

				assert.NotEmpty(t, gjson.GetBytes(body, "oauth2_login_request").Value(), "%s", body)
			})
		})

		t.Run("case=relative redirect when self-service login ui is a relative URL", func(t *testing.T) {
			reg.Config().MustSet(ctx, config.ViperKeySelfServiceLoginUI, "/login-ts")
			assert.Regexp(
				t,
				"^/login-ts.*$",
				testhelpers.GetSelfServiceRedirectLocation(t, ts.URL+login.RouteInitBrowserFlow),
			)
		})
	})
}

func TestGetFlow(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	returnToTS := testhelpers.NewRedirTS(t, "", conf)

	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/password.schema.json")
	conf.MustSet(ctx, config.ViperKeyIdentitySchemas, config.Schemas{
		{ID: "default", URL: "file://./stub/password.schema.json"},
		{ID: "email", URL: "file://./stub/email.schema.json", SelfserviceSelectable: true},
		{ID: "phone", URL: "file://./stub/phone.schema.json", SelfserviceSelectable: true},
		{ID: "not-allowed", URL: "file://./stub/password.schema.json"},
	})

	setupLoginUI := func(t *testing.T, c *http.Client) *httptest.Server {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// It is important that we use a HTTP request to fetch the flow because that will show us if CSRF works or not
			_, err := w.Write(testhelpers.EasyGetBody(t, c, public.URL+login.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
			require.NoError(t, err)
		}))
		conf.MustSet(ctx, config.ViperKeySelfServiceLoginUI, ts.URL)
		conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, returnToTS.URL)
		t.Cleanup(ts.Close)
		return ts
	}

	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{
		"enabled": true,
	})

	t.Run("case=fetching successful", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupLoginUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+login.RouteInitBrowserFlow)

		assert.NotEmpty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), public.URL, "%s", body)
	})

	t.Run("case=csrf cookie missing", func(t *testing.T) {
		client := http.DefaultClient
		setupLoginUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+login.RouteInitBrowserFlow)

		assert.EqualValues(t, nosurfx.ErrInvalidCSRFToken.ReasonField, gjson.GetBytes(body, "error.reason").String(), "%s", body)
	})

	t.Run("case=expired", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupLoginUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+login.RouteInitBrowserFlow)

		// Expire the flow
		f, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), f))

		// Try the flow but it is expired
		res, body := testhelpers.EasyGet(t, client, public.URL+login.RouteGetFlow+"?id="+f.ID.String())
		assert.EqualValues(t, http.StatusGone, res.StatusCode)
		assert.Equal(t, public.URL+login.RouteInitBrowserFlow, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	})

	t.Run("case=expired with return_to and schema_id", func(t *testing.T) {
		returnTo := returnToTS.URL
		conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTo})

		client := testhelpers.NewClientWithCookies(t)
		setupLoginUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+login.RouteInitBrowserFlow+"?return_to="+returnTo+"&identity_schema=email")

		// Expire the flow
		f, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), f))

		// Retrieve the flow and verify that return_to is in the response
		getURL := fmt.Sprintf("%s%s?id=%s&return_to=%s", public.URL, login.RouteGetFlow, f.ID, returnTo)
		getBody := testhelpers.EasyGetBody(t, client, getURL)
		assert.Equal(t, gjson.GetBytes(getBody, "error.details.return_to").String(), returnTo)

		// submit the flow but it is expired
		u := public.URL + login.RouteSubmitFlow + "?flow=" + f.ID.String()
		res, err := client.PostForm(u, url.Values{"password_identifier": {"email@ory.sh"}, "csrf_token": {f.CSRFToken}, "password": {"password"}, "method": {"password"}})
		require.NoError(t, err)
		resBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		f, err = reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(resBody, "id").String()))
		require.NoError(t, err)
		assert.Equal(t, public.URL+login.RouteInitBrowserFlow+"?return_to="+returnTo+"&identity_schema=email", f.RequestURL)
	})

	t.Run("case=not found", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupLoginUI(t, client)

		res, _ := testhelpers.EasyGet(t, client, public.URL+login.RouteGetFlow+"?id="+x.NewUUID().String())
		assert.EqualValues(t, http.StatusNotFound, res.StatusCode)
	})
}
