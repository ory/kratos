package login_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/ui/container"

	"github.com/ory/kratos/text"

	"github.com/gobuffalo/httptest"
	"github.com/gofrs/uuid"

	"github.com/ory/kratos/corpx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

func init() {
	corpx.RegisterFakes()
}

func TestFlowLifecycle(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	router := x.NewRouterPublic()
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())
	loginTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)

	errorTS := testhelpers.NewErrorTestServer(t, reg)
	conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/password.schema.json")

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
		req := x.NewTestHTTPRequest(t, "GET", ts.URL+route, nil)
		req.URL.RawQuery = extQuery.Encode()
		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router.Router, req)
		if isAPI {
			assert.Len(t, res.Header.Get("Set-Cookie"), 0)
		}
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
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
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
			}},
		State:  identity.StateActive,
		Traits: identity.Traits(`{"username":"` + id1mail + `"}`),
	}
	identity2 := &identity.Identity{
		Credentials: map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypePassword: {
				Type:        identity.CredentialsTypePassword,
				Identifiers: []string{id2mail},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`), // foobar
			}},
		State:  identity.StateActive,
		Traits: identity.Traits(`{"username":"` + id2mail + `"}`),
	}

	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), identity1))
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), identity2))

	t.Run("lifecycle=submit", func(t *testing.T) {
		t.Run("interaction=unauthenticated", func(t *testing.T) {
			run := func(t *testing.T, tt flow.Type, aal string, values url.Values) (string, *http.Response) {
				f := login.Flow{Type: tt, ExpiresAt: time.Now().Add(time.Minute), IssuedAt: time.Now(),
					UI: container.New(""), Refresh: false, RequestedAAL: identity.AuthenticatorAssuranceLevel(aal)}
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
					conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
				})

				run := func(t *testing.T, tt flow.Type) (string, string) {
					f := login.Flow{Type: tt, ExpiresAt: time.Now().Add(time.Minute), IssuedAt: time.Now(), UI: container.New(""), Refresh: false, RequestedAAL: "aal1"}
					require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), &f))

					hc := testhelpers.NewClientWithCookies(t)
					res, err := hc.PostForm(ts.URL+login.RouteSubmitFlow+"?flow="+f.ID.String(), url.Values{"method": {"password"}, "password_identifier": {id1mail}, "password": {"foobar"}, "csrf_token": {x.FakeCSRFToken}})
					require.NoError(t, err)
					firstSession := x.MustReadAll(res.Body)
					require.NoError(t, res.Body.Close())

					f = login.Flow{Type: tt, ExpiresAt: time.Now().Add(time.Minute), IssuedAt: time.Now(), UI: container.New(""), Refresh: true, RequestedAAL: "aal1"}
					require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), &f))

					vv := testhelpers.EncodeFormAsJSON(t, tt == flow.TypeAPI, url.Values{"method": {"password"}, "password_identifier": {id2mail}, "password": {"foobar"}, "csrf_token": {x.FakeCSRFToken}})

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
		})

		t.Run("case=ensure aal is checked for upgradeability on session", func(t *testing.T) {
			run := func(t *testing.T, tt flow.Type, values url.Values) (string, *http.Response) {
				f := login.Flow{Type: tt, ExpiresAt: time.Now().Add(time.Minute), IssuedAt: time.Now(),
					UI: container.New(""), Refresh: false, RequestedAAL: "aal1"}
				require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), &f))

				req, err := http.NewRequest("GET", ts.URL+login.RouteSubmitFlow+"?flow="+f.ID.String(), strings.NewReader(values.Encode()))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router.Router, req)
				return string(body), res
			}

			t.Run("type=api", func(t *testing.T) {
				body, res := run(t, flow.TypeAPI, url.Values{"method": {"password"}})
				assert.Contains(t, res.Request.URL.String(), login.RouteSubmitFlow)
				assertx.EqualAsJSON(t, login.ErrAlreadyLoggedIn.Reason(), gjson.Get(body, "ui.messages.0.text").String(), body)
			})

			t.Run("type=browser", func(t *testing.T) {
				_, res := run(t, flow.TypeBrowser, url.Values{"method": {"password"}})
				assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
			})
		})

		t.Run("case=should return an error because the request is expired", func(t *testing.T) {
			conf.MustSet(config.ViperKeySelfServiceLoginRequestLifespan, "50ms")
			t.Cleanup(func() {
				conf.MustSet(config.ViperKeySelfServiceLoginRequestLifespan, "10m")
			})

			expired := time.Now().Add(-time.Minute)
			run := func(t *testing.T, tt flow.Type, aal string, values string, isSPA bool) (string, *http.Response) {
				f := login.Flow{Type: tt, ExpiresAt: expired, IssuedAt: time.Now(),
					UI: container.New(""), Refresh: false, RequestedAAL: identity.AuthenticatorAssuranceLevel(aal)}
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
	})

	t.Run("lifecycle=init", func(t *testing.T) {
		t.Run("flow=api", func(t *testing.T) {
			t.Run("case=does not set forced flag on unauthenticated request", func(t *testing.T) {
				res, body := initFlow(t, url.Values{}, true)
				assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
				assertion(body, false, true)
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
				assertion(body, true, true)
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
				assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
			})

			t.Run("case=can not request aal2 on unauthenticated request", func(t *testing.T) {
				res, body := initFlow(t, url.Values{"aal": {"aal2"}}, false)
				assert.Contains(t, res.Request.URL.String(), errorTS.URL)
				assertx.EqualAsJSON(t, "You can not requested a higher AAL (AAL2/AAL3) without an active session.", gjson.GetBytes(body, "reason").String())
			})

			t.Run("case=ignores aal1 if session has aal1 already", func(t *testing.T) {
				res, _ := initAuthenticatedFlow(t, url.Values{"aal": {"aal1"}}, false)
				assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
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
				assertion(body, true, false)
				assert.Contains(t, res.Request.URL.String(), loginTS.URL)
			})

			t.Run("case=does not set forced flag on authenticated request without refresh=true", func(t *testing.T) {
				res, _ := initAuthenticatedFlow(t, url.Values{}, false)
				assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
			})

			t.Run("case=does not set forced flag on authenticated request with refresh=false", func(t *testing.T) {
				res, _ := initAuthenticatedFlow(t, url.Values{"refresh": {"false"}}, false)
				assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
			})

			t.Run("case=does set forced flag on authenticated request with refresh=true", func(t *testing.T) {
				res, body := initAuthenticatedFlow(t, url.Values{"refresh": {"true"}}, false)
				assertion(body, true, false)
				assert.Contains(t, res.Request.URL.String(), loginTS.URL)
			})
		})
		t.Run("case=relative redirect when self-service login ui is a relative URL", func(t *testing.T) {
			reg.Config(context.Background()).MustSet(config.ViperKeySelfServiceLoginUI, "/login-ts")
			assert.Regexp(
				t,
				"^/login-ts.*$",
				testhelpers.GetSelfServiceRedirectLocation(t, ts.URL+login.RouteInitBrowserFlow),
			)
		})
	})
}

func TestGetFlow(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "", conf)

	setupLoginUI := func(t *testing.T, c *http.Client) *httptest.Server {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// It is important that we use a HTTP request to fetch the flow because that will show us if CSRF works or not
			_, err := w.Write(x.EasyGetBody(t, c, public.URL+login.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
			require.NoError(t, err)
		}))
		conf.MustSet(config.ViperKeySelfServiceLoginUI, ts.URL)
		conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
		t.Cleanup(ts.Close)
		return ts
	}

	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{
		"enabled": true})

	t.Run("case=fetching successful", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupLoginUI(t, client)
		body := x.EasyGetBody(t, client, public.URL+login.RouteInitBrowserFlow)

		assert.NotEmpty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), public.URL, "%s", body)
	})

	t.Run("case=csrf cookie missing", func(t *testing.T) {
		client := http.DefaultClient
		setupLoginUI(t, client)
		body := x.EasyGetBody(t, client, public.URL+login.RouteInitBrowserFlow)

		assert.EqualValues(t, x.ErrInvalidCSRFToken.ReasonField, gjson.GetBytes(body, "error.reason").String(), "%s", body)
	})

	t.Run("case=expired", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupLoginUI(t, client)
		body := x.EasyGetBody(t, client, public.URL+login.RouteInitBrowserFlow)

		// Expire the flow
		f, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), f))

		// Try the flow but it is expired
		res, body := x.EasyGet(t, client, public.URL+login.RouteGetFlow+"?id="+f.ID.String())
		assert.EqualValues(t, http.StatusGone, res.StatusCode)
		assert.Equal(t, public.URL+login.RouteInitBrowserFlow, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	})

	t.Run("case=expired with return_to", func(t *testing.T) {
		conf.MustSet(config.ViperKeyURLsWhitelistedReturnToDomains, []string{"https://www.ory.sh/"})

		client := testhelpers.NewClientWithCookies(t)
		setupLoginUI(t, client)
		body := x.EasyGetBody(t, client, public.URL+login.RouteInitBrowserFlow+"?return_to=https://www.ory.sh")

		// Expire the flow
		f, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), f))

		// submit the flow but it is expired
		u := public.URL + login.RouteSubmitFlow + "?flow=" + f.ID.String()
		res, err := client.PostForm(u, url.Values{"password_identifier": {"email@ory.sh"}, "csrf_token": {f.CSRFToken}, "password": {"password"}, "method": {"password"}})
		resBody, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		f, err = reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(resBody, "id").String()))
		require.NoError(t, err)
		assert.Equal(t, public.URL+login.RouteInitBrowserFlow+"?return_to=https://www.ory.sh", f.RequestURL)
	})

	t.Run("case=not found", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupLoginUI(t, client)

		res, _ := x.EasyGet(t, client, public.URL+login.RouteGetFlow+"?id="+x.NewUUID().String())
		assert.EqualValues(t, http.StatusNotFound, res.StatusCode)
	})
}
