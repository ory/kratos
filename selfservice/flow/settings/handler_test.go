package settings_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/pointerx"

	"github.com/ory/viper"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	sdkp "github.com/ory/kratos/internal/httpclient/client/public"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

func TestHandler(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")
	testhelpers.StrategyEnable(identity.CredentialsTypePassword.String(), true)
	testhelpers.StrategyEnable(settings.StrategyProfile, true)

	_ = testhelpers.NewSettingsUITestServer(t)
	_ = testhelpers.NewErrorTestServer(t, reg)

	viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
	primaryIdentity := &identity.Identity{ID: x.NewUUID(), Traits: identity.Traits(`{}`)}
	publicTS, adminTS, clients := testhelpers.NewSettingsAPIServer(t, reg, map[string]*identity.Identity{
		"primary":   primaryIdentity,
		"secondary": {ID: x.NewUUID(), Traits: identity.Traits(`{}`)}})

	primaryUser, otherUser := clients["primary"], clients["secondary"]
	publicClient, adminClient := testhelpers.NewSDKClient(publicTS), testhelpers.NewSDKClient(adminTS)
	newExpiredFlow := func() *settings.Flow {
		return settings.NewFlow(-time.Minute,
			&http.Request{URL: urlx.ParseOrPanic(publicTS.URL + login.RouteInitBrowserFlow)},
			primaryIdentity, flow.TypeBrowser)
	}

	t.Run("daemon=admin", func(t *testing.T) {
		t.Run("description=fetching a non-existent flow should return a 404 error", func(t *testing.T) {
			_, err := adminClient.Public.GetSelfServiceSettingsFlow(
				sdkp.NewGetSelfServiceSettingsFlowParams().WithHTTPClient(otherUser).WithID("i-do-not-exist"), nil)
			require.Error(t, err)

			require.IsType(t, &sdkp.GetSelfServiceSettingsFlowNotFound{}, err)
			assert.Equal(t, int64(http.StatusNotFound), err.(*sdkp.GetSelfServiceSettingsFlowNotFound).Payload.Error.Code)
		})

		t.Run("description=fetching an expired flow returns 410", func(t *testing.T) {
			pr := newExpiredFlow()
			require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(context.Background(), pr))

			_, err := adminClient.Public.GetSelfServiceSettingsFlow(
				sdkp.NewGetSelfServiceSettingsFlowParams().WithHTTPClient(primaryUser).WithID(pr.ID.String()), nil,
			)
			require.Error(t, err)

			require.IsType(t, &sdkp.GetSelfServiceSettingsFlowGone{}, err, "%+v", err)
			assert.Equal(t, int64(http.StatusGone), err.(*sdkp.GetSelfServiceSettingsFlowGone).Payload.Error.Code)
		})
	})

	t.Run("daemon=public", func(t *testing.T) {
		t.Run("description=fetching a non-existent flow should return a 403 error", func(t *testing.T) {
			_, err := publicClient.Public.GetSelfServiceSettingsFlow(
				sdkp.NewGetSelfServiceSettingsFlowParams().WithHTTPClient(otherUser).WithID("i-do-not-exist"), nil,
			)
			require.Error(t, err)

			require.IsType(t, &sdkp.GetSelfServiceSettingsFlowForbidden{}, err)
			assert.Equal(t, int64(http.StatusForbidden), err.(*sdkp.GetSelfServiceSettingsFlowForbidden).Payload.Error.Code)
		})

		t.Run("description=fetching an expired flow returns 410", func(t *testing.T) {
			pr := newExpiredFlow()
			require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(context.Background(), pr))

			_, err := publicClient.Public.GetSelfServiceSettingsFlow(
				sdkp.NewGetSelfServiceSettingsFlowParams().WithHTTPClient(primaryUser).WithID(pr.ID.String()), nil,
			)
			require.Error(t, err)

			require.IsType(t, &sdkp.GetSelfServiceSettingsFlowGone{}, err)
			assert.Equal(t, int64(http.StatusGone), err.(*sdkp.GetSelfServiceSettingsFlowGone).Payload.Error.Code)
		})

		t.Run("description=should fail to fetch request if identity changed", func(t *testing.T) {
			t.Run("type=api", func(t *testing.T) {
				user1 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
				user2 := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)

				res, err := user1.Get(publicTS.URL + settings.RouteInitAPIFlow)
				require.NoError(t, err)
				defer res.Body.Close()
				assert.Len(t, res.Header.Get("Set-Cookie"), 0)
				body := x.MustReadAll(res.Body)
				id := gjson.GetBytes(body, "id")
				require.NotEmpty(t, id)

				res, err = user2.Get(publicTS.URL + settings.RouteGetFlow)
				require.NoError(t, err)
				defer res.Body.Close()
				require.EqualValues(t, res.StatusCode, http.StatusForbidden)
				body = x.MustReadAll(res.Body)
				assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), "Access privileges are missing", "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				res, err := primaryUser.Get(publicTS.URL + settings.RouteInitBrowserFlow)
				require.NoError(t, err)

				rid := res.Request.URL.Query().Get("flow")
				require.NotEmpty(t, rid)

				_, err = publicClient.Public.GetSelfServiceSettingsFlow(
					sdkp.NewGetSelfServiceSettingsFlowParams().WithHTTPClient(otherUser).WithID(rid), nil,
				)
				require.Error(t, err)
				require.IsType(t, &sdkp.GetSelfServiceSettingsFlowForbidden{}, err)
				assert.EqualValues(t, int64(http.StatusForbidden), err.(*sdkp.GetSelfServiceSettingsFlowForbidden).Payload.Error.Code, "should return a 403 error because the identities from the cookies do not match")
			})
		})

		t.Run("description=should fail to post data if CSRF is missing", func(t *testing.T) {
			f := testhelpers.GetSettingsFlowMethodConfigDeprecated(t, primaryUser, publicTS, settings.StrategyProfile)
			res, err := primaryUser.PostForm(pointerx.StringR(f.Action), url.Values{})
			require.NoError(t, err)
			assert.EqualValues(t, 400, res.StatusCode, "should return a 400 error because CSRF token is not set")
		})
	})
}
