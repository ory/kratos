package settings_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	kratos "github.com/ory/kratos-client-go"

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
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePassword.String(), true)
	testhelpers.StrategyEnable(t, conf, settings.StrategyProfile, true)

	router := x.NewRouterPublic()
	admin := x.NewRouterAdmin()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, admin)

	_ = testhelpers.NewSettingsUITestServer(t, conf)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)

	conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")

	primaryIdentity := &identity.Identity{ID: x.NewUUID(), Traits: identity.Traits(`{}`)}
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), primaryIdentity))
	primaryUser := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, primaryIdentity)
	otherUser := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)

	newExpiredFlow := func() *settings.Flow {
		return settings.NewFlow(conf, -time.Minute,
			&http.Request{URL: urlx.ParseOrPanic(publicTS.URL + login.RouteInitBrowserFlow)},
			primaryIdentity, flow.TypeBrowser)
	}

	assertion := func(t *testing.T, body []byte, isApi bool) {
		if isApi {
			assert.Equal(t, "api", gjson.GetBytes(body, "type").String())
		} else {
			assert.Equal(t, "browser", gjson.GetBytes(body, "type").String())
		}
	}

	initAuthenticatedFlow := func(t *testing.T, hc *http.Client, isAPI bool, isSPA bool) (*http.Response, []byte) {
		route := settings.RouteInitBrowserFlow
		if isAPI {
			route = settings.RouteInitAPIFlow
		}
		req := x.NewTestHTTPRequest(t, "GET", publicTS.URL+route, nil)
		if isSPA {
			req.Header.Set("Accept", "application/json")
		}
		res, err := hc.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		if isAPI {
			assert.Len(t, res.Header.Get("Set-Cookie"), 0)
		}
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		return res, body
	}

	initFlow := func(t *testing.T, hc *http.Client, isAPI bool) (*http.Response, []byte) {
		return initAuthenticatedFlow(t, hc, isAPI, false)
	}

	initSPAFlow := func(t *testing.T, hc *http.Client) (*http.Response, []byte) {
		return initAuthenticatedFlow(t, hc, false, true)
	}

	t.Run("endpoint=init", func(t *testing.T) {
		t.Run("description=init a flow as API", func(t *testing.T) {
			user1 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
			res, body := initFlow(t, user1, true)
			assert.Contains(t, res.Request.URL.String(), settings.RouteInitAPIFlow)
			assertion(t, body, true)
		})

		t.Run("description=init a flow as browser", func(t *testing.T) {
			user1 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
			res, body := initFlow(t, user1, false)
			assert.Contains(t, res.Request.URL.String(), reg.Config(context.Background()).SelfServiceFlowSettingsUI().String())
			assertion(t, body, false)
		})

		t.Run("description=init a flow as SPA", func(t *testing.T) {
			user1 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
			res, body := initSPAFlow(t, user1)
			assert.Contains(t, res.Request.URL.String(), settings.RouteInitBrowserFlow)
			assertion(t, body, false)
		})
	})

	t.Run("endpoint=fetch", func(t *testing.T) {
		t.Run("description=fetching a non-existent flow should return a 403 error", func(t *testing.T) {
			_, _, err := testhelpers.NewSDKCustomClient(publicTS, otherUser).V0alpha1Api.GetSelfServiceSettingsFlow(context.Background()).Id("i-do-not-exist").Execute()
			require.Error(t, err)

			require.IsType(t, new(kratos.GenericOpenAPIError), err, "%T", err)
			assert.Equal(t, int64(http.StatusForbidden), gjson.GetBytes(err.(*kratos.GenericOpenAPIError).Body(), "error.code").Int())
		})

		t.Run("description=fetching an expired flow returns 410", func(t *testing.T) {
			pr := newExpiredFlow()
			require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(context.Background(), pr))

			_, _, err := testhelpers.NewSDKCustomClient(publicTS, primaryUser).V0alpha1Api.GetSelfServiceSettingsFlow(context.Background()).Id(pr.ID.String()).Execute()
			require.Error(t, err)

			require.IsType(t, new(kratos.GenericOpenAPIError), err, "%T", err)
			assert.Equal(t, int64(http.StatusGone), gjson.GetBytes(err.(*kratos.GenericOpenAPIError).Body(), "error.code").Int())
		})

		t.Run("case=expired with return_to", func(t *testing.T) {
			client := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
			body := x.EasyGetBody(t, client, publicTS.URL+settings.RouteInitBrowserFlow+"?return_to=https://www.ory.sh")

			// Expire the flow
			f, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
			require.NoError(t, err)
			f.ExpiresAt = time.Now().Add(-time.Second)
			require.NoError(t, reg.SettingsFlowPersister().UpdateSettingsFlow(context.Background(), f))

			// submit the flow but it is expired
			u := publicTS.URL + settings.RouteSubmitFlow + "?flow=" + f.ID.String()
			res, err := client.PostForm(u, url.Values{"method": {"password"}, "csrf_token": {"csrf"}, "password": {"password"}})
			resBody, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())

			f, err = reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(resBody, "id").String()))
			require.NoError(t, err)
			assert.Equal(t, publicTS.URL+settings.RouteInitBrowserFlow+"?return_to=https://www.ory.sh", f.RequestURL)
		})

		t.Run("description=should fail to fetch request if identity changed", func(t *testing.T) {
			t.Run("type=api", func(t *testing.T) {
				user1 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
				user2 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)

				t.Logf("%+v", user1.Jar)
				res, err := user1.Get(publicTS.URL + settings.RouteInitAPIFlow)
				require.NoError(t, err)
				defer res.Body.Close()
				t.Logf("%+v", user1.Jar)
				assert.Len(t, res.Header.Get("Set-Cookie"), 0)
				body := ioutilx.MustReadAll(res.Body)
				id := gjson.GetBytes(body, "id")
				require.NotEmpty(t, id)

				res, err = user2.Get(publicTS.URL + settings.RouteGetFlow)
				require.NoError(t, err)
				defer res.Body.Close()
				require.EqualValues(t, res.StatusCode, http.StatusForbidden)
				body = ioutilx.MustReadAll(res.Body)
				assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), "Access privileges are missing", "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				res, err := primaryUser.Get(publicTS.URL + settings.RouteInitBrowserFlow)
				require.NoError(t, err)

				rid := res.Request.URL.Query().Get("flow")
				require.NotEmpty(t, rid)

				testhelpers.NewSDKCustomClient(publicTS, primaryUser)

				_, _, err = testhelpers.NewSDKCustomClient(publicTS, otherUser).V0alpha1Api.GetSelfServiceSettingsFlow(context.Background()).Id(rid).Execute()
				require.Error(t, err)
				require.IsType(t, new(kratos.GenericOpenAPIError), err, "%T", err)
				assert.Equal(t, int64(http.StatusForbidden), gjson.GetBytes(err.(*kratos.GenericOpenAPIError).Body(), "error.code").Int(), "should return a 403 error because the identities from the cookies do not match")
			})
		})
	})

	t.Run("endpoint=submit", func(t *testing.T) {
		t.Run("description=fail to submit form as another user", func(t *testing.T) {
			t.Run("type=api", func(t *testing.T) {
				user1 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
				user2 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
				_, body := initFlow(t, user1, true)
				var f kratos.SelfServiceSettingsFlow
				require.NoError(t, json.Unmarshal(body, &f))

				actual, res := testhelpers.SettingsMakeRequest(t, true, false, &f, user2, `{"method":"not-exists"}`)
				assert.Equal(t, http.StatusBadRequest, res.StatusCode)
				assert.Equal(t, "You must restart the flow because the resumable session was initiated by another person.", gjson.Get(actual, "ui.messages.0.text").String(), actual)
			})

			t.Run("type=spa", func(t *testing.T) {
				user1 := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)
				user2 := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)
				_, body := initFlow(t, user1, true)
				var f kratos.SelfServiceSettingsFlow
				require.NoError(t, json.Unmarshal(body, &f))

				actual, res := testhelpers.SettingsMakeRequest(t, false, true, &f, user2, `{"method":"not-exists"}`)
				assert.Equal(t, http.StatusBadRequest, res.StatusCode)
				assert.Equal(t, "You must restart the flow because the resumable session was initiated by another person.", gjson.Get(actual, "ui.messages.0.text").String(), actual)
			})

			t.Run("type=browser", func(t *testing.T) {
				user1 := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)
				user2 := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)
				_, body := initFlow(t, user1, true)
				var f kratos.SelfServiceSettingsFlow
				require.NoError(t, json.Unmarshal(body, &f))

				actual, res := testhelpers.SettingsMakeRequest(t, false, false, &f, user2, `{"method":"not-exists"}`)
				assert.Equal(t, http.StatusBadRequest, res.StatusCode)
				assert.Equal(t, "You must restart the flow because the resumable session was initiated by another person.", gjson.Get(actual, "ui.messages.0.text").String(), actual)
			})
		})
	})
}
