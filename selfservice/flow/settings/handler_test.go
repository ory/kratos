package settings_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/pointerx"

	"github.com/ory/viper"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
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
	newExpiredRequest := func() *settings.Request {
		return settings.NewRequest(-time.Minute,
			&http.Request{URL: urlx.ParseOrPanic(publicTS.URL + login.BrowserInitPath)},
			&session.Session{Identity: primaryIdentity})
	}

	t.Run("daemon=admin", func(t *testing.T) {
		t.Run("description=fetching a non-existent request should return a 404 error", func(t *testing.T) {
			_, err := adminClient.Common.GetSelfServiceBrowserSettingsRequest(
				common.NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(otherUser).WithRequest("i-do-not-exist"),
			)
			require.Error(t, err)

			require.IsType(t, &common.GetSelfServiceBrowserSettingsRequestNotFound{}, err)
			assert.Equal(t, int64(http.StatusNotFound), err.(*common.GetSelfServiceBrowserSettingsRequestNotFound).Payload.Error.Code)
		})

		t.Run("description=fetching an expired request returns 410", func(t *testing.T) {
			pr := newExpiredRequest()
			require.NoError(t, reg.SettingsRequestPersister().CreateSettingsRequest(context.Background(), pr))

			_, err := adminClient.Common.GetSelfServiceBrowserSettingsRequest(
				common.NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(primaryUser).WithRequest(pr.ID.String()),
			)
			require.Error(t, err)

			require.IsType(t, &common.GetSelfServiceBrowserSettingsRequestGone{}, err, "%+v", err)
			assert.Equal(t, int64(http.StatusGone), err.(*common.GetSelfServiceBrowserSettingsRequestGone).Payload.Error.Code)
		})
	})

	t.Run("daemon=public", func(t *testing.T) {
		t.Run("description=fetching a non-existent request should return a 403 error", func(t *testing.T) {
			_, err := publicClient.Common.GetSelfServiceBrowserSettingsRequest(
				common.NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(otherUser).WithRequest("i-do-not-exist"),
			)
			require.Error(t, err)

			require.IsType(t, &common.GetSelfServiceBrowserSettingsRequestForbidden{}, err)
			assert.Equal(t, int64(http.StatusForbidden), err.(*common.GetSelfServiceBrowserSettingsRequestForbidden).Payload.Error.Code)
		})

		t.Run("description=fetching an expired request returns 410", func(t *testing.T) {
			pr := newExpiredRequest()
			require.NoError(t, reg.SettingsRequestPersister().CreateSettingsRequest(context.Background(), pr))

			_, err := publicClient.Common.GetSelfServiceBrowserSettingsRequest(
				common.NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(primaryUser).WithRequest(pr.ID.String()),
			)
			require.Error(t, err)

			require.IsType(t, &common.GetSelfServiceBrowserSettingsRequestGone{}, err)
			assert.Equal(t, int64(http.StatusGone), err.(*common.GetSelfServiceBrowserSettingsRequestGone).Payload.Error.Code)
		})

		t.Run("description=should fail to fetch request if identity changed", func(t *testing.T) {
			res, err := primaryUser.Get(publicTS.URL + settings.PublicPath)
			require.NoError(t, err)

			rid := res.Request.URL.Query().Get("request")
			require.NotEmpty(t, rid)

			_, err = publicClient.Common.GetSelfServiceBrowserSettingsRequest(
				common.NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(otherUser).WithRequest(rid),
			)
			require.Error(t, err)
			require.IsType(t, &common.GetSelfServiceBrowserSettingsRequestForbidden{}, err)
			assert.EqualValues(t, int64(http.StatusForbidden), err.(*common.GetSelfServiceBrowserSettingsRequestForbidden).Payload.Error.Code, "should return a 403 error because the identities from the cookies do not match")
		})

		t.Run("description=should fail to post data if CSRF is missing", func(t *testing.T) {
			f := testhelpers.GetSettingsMethodConfig(t, primaryUser, publicTS, settings.StrategyProfile)
			res, err := primaryUser.PostForm(pointerx.StringR(f.Action), url.Values{})
			require.NoError(t, err)
			assert.EqualValues(t, 400, res.StatusCode, "should return a 400 error because CSRF token is not set")
		})
	})
}
