package oidc_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ory/x/sqlxx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"
	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

var (
	csrfField = &models.FormField{Name: pointerx.String("csrf_token"), Value: "nosurf",
		Required: true, Type: pointerx.String("hidden")}
)

func TestSettingsStrategy(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var (
		_, reg  = internal.NewFastRegistryWithMocks(t)
		subject string
		scope   []string
	)

	remoteAdmin, remotePublic, _ := newHydra(t, &subject, &scope)
	uiTS := newUI(t, reg)
	errTS := testhelpers.NewErrorTestServer(t, reg)
	publicTS, adminTS := testhelpers.NewKratosServers(t)
	public := testhelpers.NewSDKClient(publicTS)
	admin := testhelpers.NewSDKClient(adminTS)

	viperSetProviderConfig(
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "ory", "ory"),
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "google", "google"),
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "github", "github"),
	)
	testhelpers.InitKratosServers(t, reg, publicTS, adminTS)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/settings.schema.json")
	viper.Set(configuration.ViperKeyURLsDefaultReturnTo, "https://www.ory.sh/kratos")

	// Make test data for this test run unique
	testID := x.NewUUID().String()
	users := map[string]*identity.Identity{
		"password": {ID: x.NewUUID(), Traits: identity.Traits(`{"email":"john` + testID + `@doe.com"}`),
			TraitsSchemaID: configuration.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {Type: "password",
					Identifiers: []string{"john+" + testID + "@doe.com"},
					Config: sqlxx.JSONRawMessage(`{"hashed_password":"$argon2id$iammocked...."}`)}},
		},
		"oryer": {ID: x.NewUUID(), Traits: identity.Traits(`{"email":"hackerman+` + testID + `@ory.sh"}`),
			TraitsSchemaID: configuration.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypeOIDC: {Type: identity.CredentialsTypeOIDC,
					Identifiers: []string{"ory:hackerman+" + testID},
					Config: sqlxx.JSONRawMessage(`{"providers":[{"provider":"ory","subject":"hackerman+` + testID + `"}]}`)}},
		},
		"githuber": {ID: x.NewUUID(), Traits: identity.Traits(`{"email":"hackerman+github+` + testID + `@ory.sh"}`),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypeOIDC: {Type: identity.CredentialsTypeOIDC,
					Identifiers: []string{"ory:hackerman+github+" + testID, "github:hackerman+github+" + testID},
					Config: sqlxx.JSONRawMessage(`{"providers":[{"provider":"ory","subject":"hackerman+github+` + testID + `"},{"provider":"github","subject":"hackerman+github+` + testID + `"}]}`)}},
			TraitsSchemaID: configuration.DefaultIdentityTraitsSchemaID,
		},
		"multiuser": {ID: x.NewUUID(), Traits: identity.Traits(`{"email":"hackerman+multiuser+` + testID + `@ory.sh"}`),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {Type: "password",
					Identifiers: []string{"hackerman+multiuser+" + testID + "@ory.sh"},
					Config: sqlxx.JSONRawMessage(`{"hashed_password":"$argon2id$iammocked...."}`)},
				identity.CredentialsTypeOIDC: {Type: identity.CredentialsTypeOIDC,
					Identifiers: []string{"ory:hackerman+multiuser+" + testID, "google:hackerman+multiuser+" + testID},
					Config: sqlxx.JSONRawMessage(`{"providers":[{"provider":"ory","subject":"hackerman+multiuser+` + testID + `"},{"provider":"google","subject":"hackerman+multiuser+` + testID + `"}]}`)}},
			TraitsSchemaID: configuration.DefaultIdentityTraitsSchemaID,
		},
	}
	agents := testhelpers.AddAndLoginIdentities(t, reg, publicTS, users)

	// new profile request
	var npr = func(t *testing.T, client *http.Client, redirectTo string, exp time.Duration) *settings.Request {
		req, err := reg.SettingsRequestPersister().GetSettingsRequest(context.Background(),
			x.ParseUUID(string(testhelpers.GetSettingsRequest(t, client, publicTS).Payload.ID)))
		require.NoError(t, err)
		assert.Empty(t, req.Active)

		if redirectTo != "" {
			req.RequestURL = redirectTo
		}
		req.ExpiresAt = time.Now().Add(exp)
		require.NoError(t, reg.SettingsRequestPersister().UpdateSettingsRequest(context.Background(), req))

		// sanity check
		got, err := reg.SettingsRequestPersister().GetSettingsRequest(context.Background(), req.ID)
		require.NoError(t, err)
		require.Len(t, got.Methods, len(req.Methods))

		return req
	}

	// does the same as new profile request but uses the SDK
	var nprSDK = func(t *testing.T, client *http.Client, redirectTo string, exp time.Duration) *models.SettingsRequest {
		req := npr(t, client, redirectTo, exp)
		rs, err := admin.Common.GetSelfServiceBrowserSettingsRequest(common.
			NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(client).
			WithRequest(req.ID.String()))
		require.NoError(t, err)
		return rs.Payload
	}

	t.Run("case=should not be able to continue a flow with a malformed ID", func(t *testing.T) {
		body, res := testhelpers.HTTPPostForm(t, agents["password"], publicTS.URL+oidc.SettingsPath+"?request=i-am-not-a-uuid", nil)
		AssertSystemError(t, errTS, res, body, 400, "malformed")
	})

	t.Run("case=should not be able to continue a flow without the request query parameter", func(t *testing.T) {
		body, res := testhelpers.HTTPPostForm(t, agents["password"], publicTS.URL+oidc.SettingsPath, nil)
		AssertSystemError(t, errTS, res, body, 400, "query parameter is missing")
	})

	t.Run("case=should not be able to continue a flow with a non-existing ID", func(t *testing.T) {
		body, res := testhelpers.HTTPPostForm(t, agents["password"], publicTS.URL+oidc.SettingsPath+"?request="+x.NewUUID().String(), nil)
		AssertSystemError(t, errTS, res, body, 404, "not be found")
	})

	t.Run("case=should not be able to continue a flow that is expired", func(t *testing.T) {
		req := npr(t, agents["password"], "", -time.Hour)
		body, res := testhelpers.HTTPPostForm(t, agents["password"], publicTS.URL+oidc.SettingsPath+"?request="+req.ID.String(), nil)
		AssertSystemError(t, errTS, res, body, 400, "expired")
	})

	t.Run("case=should not be able to fetch another user's data", func(t *testing.T) {
		req := npr(t, agents["password"], "", time.Hour)

		_, err := public.Common.GetSelfServiceBrowserSettingsRequest(common.
			NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(agents["oryer"]).
			WithRequest(req.ID.String()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "403")
	})

	t.Run("case=should fetch the settings request and expect data to be set appropriately", func(t *testing.T) {
		req := npr(t, agents["password"], "", time.Hour)

		rs, err := admin.Common.GetSelfServiceBrowserSettingsRequest(common.
			NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(agents["password"]).
			WithRequest(req.ID.String()))
		require.NoError(t, err)

		// Check our sanity. Does the SDK relay the same info that we expect and got from the store?
		assert.Equal(t, publicTS.URL+"/self-service/browser/flows/settings", req.RequestURL)
		assert.Empty(t, req.Active)
		assert.NotEmpty(t, req.IssuedAt)
		assert.EqualValues(t, users["password"].ID, req.Identity.ID)
		assert.EqualValues(t, users["password"].Traits, req.Identity.Traits)
		assert.EqualValues(t, users["password"].TraitsSchemaID, req.Identity.TraitsSchemaID)

		assert.EqualValues(t, req.ID.String(), rs.Payload.ID)
		assert.EqualValues(t, req.RequestURL, *rs.Payload.RequestURL)
		assert.EqualValues(t, req.Identity.ID.String(), rs.Payload.Identity.ID)
		assert.EqualValues(t, req.IssuedAt, time.Time(*rs.Payload.IssuedAt))

		require.NotNil(t, identity.CredentialsTypeOIDC.String(), rs.Payload.Methods[identity.CredentialsTypeOIDC.String()])
		require.EqualValues(t, identity.CredentialsTypeOIDC.String(), rs.Payload.Methods[identity.CredentialsTypeOIDC.String()].Method)
		require.EqualValues(t, publicTS.URL+oidc.SettingsPath+"?request="+req.ID.String(),
			*rs.Payload.Methods[identity.CredentialsTypeOIDC.String()].Config.Action)
	})

	expectedOryerFields := models.FormFields{
		{Type: pointerx.String("submit"), Name: pointerx.String("link"), Value: "google"},
		{Type: pointerx.String("submit"), Name: pointerx.String("link"), Value: "github"}}
	expectedGithuberFields := models.FormFields{
		{Type: pointerx.String("submit"), Name: pointerx.String("link"), Value: "google"},
		{Type: pointerx.String("submit"), Name: pointerx.String("unlink"), Value: "ory"},
		{Type: pointerx.String("submit"), Name: pointerx.String("unlink"), Value: "github"}}
	t.Run("case=should adjust linkable providers based on linked credentials", func(t *testing.T) {
		for _, tc := range []struct {
			agent    string
			expected models.FormFields
		}{
			{agent: "password", expected: models.FormFields{
				{Type: pointerx.String("submit"), Name: pointerx.String("link"), Value: "ory"},
				{Type: pointerx.String("submit"), Name: pointerx.String("link"), Value: "google"},
				{Type: pointerx.String("submit"), Name: pointerx.String("link"), Value: "github"}}},
			{agent: "oryer", expected: expectedOryerFields},
			{agent: "githuber", expected: expectedGithuberFields},
			{agent: "multiuser", expected: models.FormFields{
				{Type: pointerx.String("submit"), Name: pointerx.String("link"), Value: "github"},
				{Type: pointerx.String("submit"), Name: pointerx.String("unlink"), Value: "ory"},
				{Type: pointerx.String("submit"), Name: pointerx.String("unlink"), Value: "google"}}},
		} {
			t.Run("agent="+tc.agent, func(t *testing.T) {
				rs := nprSDK(t, agents[tc.agent], "", time.Hour)
				assert.EqualValues(t, append(models.FormFields{csrfField}, tc.expected...),
					rs.Methods[identity.CredentialsTypeOIDC.String()].Config.Fields)
			})
		}
	})

	var action = func(req *models.SettingsRequest) string {
		return *req.Methods[identity.CredentialsTypeOIDC.String()].Config.Action
	}

	var checkCredentials = func(t *testing.T, shouldExist bool, iid uuid.UUID, provider, subject string) {
		actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), iid)
		require.NoError(t, err)

		var cc oidc.CredentialsConfig
		creds, err := actual.ParseCredentials(identity.CredentialsTypeOIDC, &cc)
		require.NoError(t, err)

		if shouldExist {
			assert.Contains(t, creds.Identifiers, provider+":"+subject)
		} else {
			assert.NotContains(t, creds.Identifiers, provider+":"+subject)
		}

		var found bool
		for _, p := range cc.Providers {
			if p.Provider == provider && p.Subject == subject {
				found = true
				break
			}
		}

		require.EqualValues(t, shouldExist, found)
	}

	var reset = func(t *testing.T) func() {
		return func() {
			viper.Set(configuration.ViperKeySelfServicePrivilegedAuthenticationAfter, time.Minute*5)
			agents = testhelpers.AddAndLoginIdentities(t, reg, publicTS, users)
		}
	}

	t.Run("suite=unlink", func(t *testing.T) {
		var unlink = func(t *testing.T, agent, provider string) (body []byte, res *http.Response, req *models.SettingsRequest) {
			req = nprSDK(t, agents[agent], "", time.Hour)
			body, res = testhelpers.HTTPPostForm(t, agents[agent], action(req),
				&url.Values{"csrf_token": {"nosurf"}, "unlink": {provider}})
			return
		}

		var unlinkInvalid = func(agent, provider string, expectedFields models.FormFields) func(t *testing.T) {
			return func(t *testing.T) {
				body, res, req := unlink(t, agent, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?request="+string(req.ID))

				assert.EqualValues(t, identity.CredentialsTypeOIDC.String(), gjson.GetBytes(body, "active").String())
				assert.Contains(t, gjson.GetBytes(body, "methods.oidc.config.action").String(), publicTS.URL+oidc.SettingsPath+"?request=")

				// The original options to link google and github are still there
				testhelpers.JSONEq(t, append(models.FormFields{csrfField}, expectedFields...),
					json.RawMessage(gjson.GetBytes(body, `methods.oidc.config.fields`).Raw))

				assert.Contains(t, gjson.GetBytes(body, `methods.oidc.config.messages.0.text`).String(),
					"can not unlink non-existing OpenID Connect")
			}
		}

		t.Run("case=should not be able to unlink the last remaining connection",
			unlinkInvalid("oryer", "ory", expectedOryerFields))

		t.Run("case=should not be able to unlink an non-existing connection",
			unlinkInvalid("oryer", "i-do-not-exist", expectedOryerFields))

		t.Run("case=should not be able to unlink a connection not yet linked",
			unlinkInvalid("githuber", "google", expectedGithuberFields))

		t.Run("case=should unlink a connection", func(t *testing.T) {
			agent, provider := "githuber", "github"
			t.Cleanup(reset(t))

			body, res, req := unlink(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?request="+string(req.ID))
			require.True(t, gjson.GetBytes(body, "update_successful").Bool())

			checkCredentials(t, false, users[agent].ID, provider, "hackerman+github+"+testID)
		})

		t.Run("case=should not be able to unlink a connection without a privileged session", func(t *testing.T) {
			agent, provider := "githuber", "github"

			var runUnauthed = func(t *testing.T) *models.SettingsRequest {
				viper.Set(configuration.ViperKeySelfServicePrivilegedAuthenticationAfter, time.Millisecond)
				time.Sleep(time.Millisecond)
				t.Cleanup(reset(t))
				_, res, req := unlink(t, agent, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login")

				rs, err := admin.Common.GetSelfServiceBrowserSettingsRequest(common.
					NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(agents[agent]).
					WithRequest(string(req.ID)))
				require.NoError(t, err)
				require.EqualValues(t, false, *rs.Payload.UpdateSuccessful)

				checkCredentials(t, true, users[agent].ID, provider, "hackerman+github+"+testID)

				return req
			}

			t.Run("subcase=should not update without re-auth", func(t *testing.T) {
				_ = runUnauthed(t)
			})

			t.Run("subcase=should update after re-auth", func(t *testing.T) {
				req := runUnauthed(t)

				// fake login by allowing longer sessions...
				viper.Set(configuration.ViperKeySelfServicePrivilegedAuthenticationAfter, time.Minute*5)

				body, res := testhelpers.HTTPPostForm(t, agents[agent], action(req),
					&url.Values{"csrf_token": {"nosurf"}, "unlink": {provider}})
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?request="+string(req.ID))

				assert.True(t, gjson.GetBytes(body, "update_successful").Bool())

				checkCredentials(t, false, users[agent].ID, provider, "hackerman+github+"+testID)
			})
		})
	})

	t.Run("suite=link", func(t *testing.T) {
		var link = func(t *testing.T, agent, provider string) (body []byte, res *http.Response, req *models.SettingsRequest) {
			req = nprSDK(t, agents[agent], "", time.Hour)
			body, res = testhelpers.HTTPPostForm(t, agents[agent], action(req),
				&url.Values{"csrf_token": {"nosurf"}, "link": {provider}})
			return
		}

		var linkInvalid = func(agent, provider string, expectedFields models.FormFields) func(t *testing.T) {
			return func(t *testing.T) {
				body, res, req := link(t, agent, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?request="+string(req.ID))

				assert.EqualValues(t, identity.CredentialsTypeOIDC.String(), gjson.GetBytes(body, "active").String())
				assert.Contains(t, gjson.GetBytes(body, "methods.oidc.config.action").String(), publicTS.URL+oidc.SettingsPath+"?request=")

				// The original options to link google and github are still there
				testhelpers.JSONEq(t, append(models.FormFields{csrfField}, expectedFields...),
					json.RawMessage(gjson.GetBytes(body, `methods.oidc.config.fields`).Raw))

				assert.Contains(t, gjson.GetBytes(body, `methods.oidc.config.messages.0.text`).String(),
					"can not link unknown or already existing OpenID Connect connection")
			}
		}

		t.Run("case=should not be able to link an non-existing connection",
			linkInvalid("oryer", "i-do-not-exist", expectedOryerFields))

		t.Run("case=should not be able to link a connection which already exists",
			linkInvalid("githuber", "github", expectedGithuberFields))

		t.Run("case=should not be able to link a connection already linked by another identity", func(t *testing.T) {
			// While this theoretically allows for account enumeration - because we see an error indicator if an
			// oidc connection is being linked that exists already - it would require the attacker to already
			// have control over the social profile, in which case account enumeration is the least of our worries.
			// Instead of using the oidc profile for enumeration, the attacker would use it for account takeover.

			// This is the multiuser login id for google
			subject = "hackerman+multiuser+" + testID
			scope = []string{"openid"}

			agent, provider := "githuber", "google"
			body, res, _ := link(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), errTS.URL)

			t.Logf("%s", body)
			assert.EqualValues(t, 409, gjson.GetBytes(body, `0.code`).Int())
			assert.Contains(t, gjson.GetBytes(body, `0.message`).String(), "insert or update resource because a resource")
		})

		t.Run("case=should not be able to link a connection which is missing the ID token", func(t *testing.T) {
			t.Cleanup(reset(t))

			subject = "hackerman+scope-missing+" + testID
			scope = []string{}

			agent, provider := "githuber", "google"
			body, res, _ := link(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			t.Logf("%s", body)
			assert.Contains(t, gjson.GetBytes(body, `methods.oidc.config.messages.0.text`).String(),
				"no id_token was returned")
		})

		t.Run("case=should not be able to link a connection which is missing the ID token", func(t *testing.T) {
			t.Cleanup(reset(t))

			subject = "hackerman+scope-missing+" + testID
			scope = []string{}

			agent, provider := "githuber", "google"
			body, res, _ := link(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			assert.Contains(t, gjson.GetBytes(body, `methods.oidc.config.messages.0.text`).String(),
				"no id_token was returned")
		})

		t.Run("case=should link a connection", func(t *testing.T) {
			t.Cleanup(reset(t))

			subject = "hackerman+new-connection+" + testID
			scope = []string{"openid"}

			agent, provider := "githuber", "google"
			_, res, req := link(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			rs, err := admin.Common.GetSelfServiceBrowserSettingsRequest(common.
				NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(agents[agent]).
				WithRequest(string(req.ID)))
			require.NoError(t, err)
			require.EqualValues(t, true, *rs.Payload.UpdateSuccessful)

			testhelpers.JSONEq(t, append(models.FormFields{csrfField}, models.FormFields{
				{Type: pointerx.String("submit"), Name: pointerx.String("unlink"), Value: "ory"},
				{Type: pointerx.String("submit"), Name: pointerx.String("unlink"), Value: "github"},
				{Type: pointerx.String("submit"), Name: pointerx.String("unlink"), Value: "google"},
			}...), rs.Payload.Methods[identity.CredentialsTypeOIDC.String()].Config.Fields)

			checkCredentials(t, true, users[agent].ID, provider, subject)
		})

		t.Run("case=should not be able to link a connection without a privileged session", func(t *testing.T) {
			agent, provider := "githuber", "google"
			subject = "hackerman+new+google+" + testID

			var runUnauthed = func(t *testing.T) *models.SettingsRequest {
				viper.Set(configuration.ViperKeySelfServicePrivilegedAuthenticationAfter, time.Millisecond)
				time.Sleep(time.Millisecond)
				t.Cleanup(reset(t))
				_, res, req := link(t, agent, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login")

				rs, err := admin.Common.GetSelfServiceBrowserSettingsRequest(common.
					NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(agents[agent]).
					WithRequest(string(req.ID)))
				require.NoError(t, err)
				require.EqualValues(t, false, *rs.Payload.UpdateSuccessful)

				checkCredentials(t, false, users[agent].ID, provider, subject)

				return req
			}

			t.Run("subcase=should not update without re-auth", func(t *testing.T) {
				_ = runUnauthed(t)
			})

			t.Run("subcase=should update after re-auth", func(t *testing.T) {
				req := runUnauthed(t)

				// fake login by allowing longer sessions...
				viper.Set(configuration.ViperKeySelfServicePrivilegedAuthenticationAfter, time.Minute*5)

				body, res := testhelpers.HTTPPostForm(t, agents[agent], action(req),
					&url.Values{"csrf_token": {"nosurf"}, "unlink": {provider}})
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?request="+string(req.ID))

				assert.True(t, gjson.GetBytes(body, "update_successful").Bool())

				checkCredentials(t, true, users[agent].ID, provider, subject)
			})
		})
	})
}

func TestPopulateSettingsMethod(t *testing.T) {
	nreg := func(t *testing.T, conf *oidc.ConfigurationCollection) *driver.RegistryDefault {
		_, reg := internal.NewFastRegistryWithMocks(t)

		viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://stub/registration.schema.json")
		viper.Set(configuration.ViperKeyURLsSelfPublic, "https://www.ory.sh/")

		viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{
			"enabled": true})
		viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC), map[string]interface{}{
			"enabled": true,
			"config":  conf})
		return reg
	}

	ns := func(t *testing.T, reg *driver.RegistryDefault) *oidc.Strategy {
		ss, err := reg.SettingsStrategies().Strategy(identity.CredentialsTypeOIDC.String())
		require.NoError(t, err)
		return ss.(*oidc.Strategy)
	}

	nr := func() *settings.Request {
		return &settings.Request{ID: x.NewUUID(), Methods: map[string]*settings.RequestMethod{}}
	}

	populate := func(t *testing.T, reg *driver.RegistryDefault, i *identity.Identity, req *settings.Request) *form.HTMLForm {
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
		require.NoError(t, ns(t, reg).PopulateSettingsMethod(new(http.Request), &session.Session{Identity: i, IdentityID: i.ID}, req))
		require.NotNil(t, req.Methods[identity.CredentialsTypeOIDC.String()])
		require.NotNil(t, req.Methods[identity.CredentialsTypeOIDC.String()].Config)
		require.NotNil(t, req.Methods[identity.CredentialsTypeOIDC.String()].Config.RequestMethodConfigurator)
		require.Equal(t, identity.CredentialsTypeOIDC.String(), req.Methods[identity.CredentialsTypeOIDC.String()].Method)
		f := req.Methods[identity.CredentialsTypeOIDC.String()].Config.RequestMethodConfigurator.(*oidc.RequestMethod).HTMLForm
		assert.Equal(t, "https://www.ory.sh"+oidc.SettingsPath+"?request="+req.ID.String(), f.Action)
		assert.Equal(t, "POST", f.Method)
		return f
	}

	defaultConfig := []oidc.Configuration{
		{Provider: "generic", ID: "facebook"},
		{Provider: "generic", ID: "google"},
		{Provider: "generic", ID: "github"},
	}

	for k, tc := range []struct {
		c      []oidc.Configuration
		i      identity.Credentials
		e      form.Fields
		withpw bool
	}{
		{
			c: []oidc.Configuration{},
			e: form.Fields{
				{Name: "csrf_token", Type: "hidden", Required: true, Value: "nosurf"},
			},
		},
		{
			c: []oidc.Configuration{
				{Provider: "generic", ID: "github"},
			},
			e: form.Fields{
				{Name: "csrf_token", Type: "hidden", Required: true, Value: "nosurf"},
				{Name: "link", Type: "submit", Value: "github"},
			},
		},
		{
			c: defaultConfig,
			e: form.Fields{
				{Name: "csrf_token", Type: "hidden", Required: true, Value: "nosurf"},
				{Name: "link", Type: "submit", Value: "facebook"},
				{Name: "link", Type: "submit", Value: "google"},
				{Name: "link", Type: "submit", Value: "github"},
			},
		},
		{
			c: defaultConfig,
			e: form.Fields{
				{Name: "csrf_token", Type: "hidden", Required: true, Value: "nosurf"},
				{Name: "link", Type: "submit", Value: "facebook"},
				{Name: "link", Type: "submit", Value: "google"},
				{Name: "link", Type: "submit", Value: "github"},
			},
			i: identity.Credentials{Type: identity.CredentialsTypeOIDC, Identifiers: []string{}, Config: []byte(`{}`)},
		},
		{
			c: defaultConfig,
			e: form.Fields{
				{Name: "csrf_token", Type: "hidden", Required: true, Value: "nosurf"},
				{Name: "link", Type: "submit", Value: "facebook"},
				{Name: "link", Type: "submit", Value: "github"},
			},
			i: identity.Credentials{Type: identity.CredentialsTypeOIDC, Identifiers: []string{
				"google:1234",
			}, Config: []byte(`{"providers":[{"provider":"google","subject":"1234"}]}`)},
		},
		{
			c: defaultConfig,
			e: form.Fields{
				{Name: "csrf_token", Type: "hidden", Required: true, Value: "nosurf"},
				{Name: "link", Type: "submit", Value: "facebook"},
				{Name: "link", Type: "submit", Value: "github"},
				{Name: "unlink", Type: "submit", Value: "google"},
			},
			withpw: true,
			i: identity.Credentials{Type: identity.CredentialsTypeOIDC, Identifiers: []string{
				"google:1234",
			},
				Config: []byte(`{"providers":[{"provider":"google","subject":"1234"}]}`)},
		},
		{
			c: defaultConfig,
			e: form.Fields{
				{Name: "csrf_token", Type: "hidden", Required: true, Value: "nosurf"},
				{Name: "link", Type: "submit", Value: "github"},
				{Name: "unlink", Type: "submit", Value: "google"},
				{Name: "unlink", Type: "submit", Value: "facebook"},
			},
			i: identity.Credentials{Type: identity.CredentialsTypeOIDC, Identifiers: []string{
				"google:1234",
				"facebook:1234",
			},
				Config: []byte(`{"providers":[{"provider":"google","subject":"1234"},{"provider":"facebook","subject":"1234"}]}`)},
		},
	} {
		t.Run("iteration="+strconv.Itoa(k), func(t *testing.T) {
			reg := nreg(t, &oidc.ConfigurationCollection{Providers: tc.c})
			i := &identity.Identity{Traits: []byte(`{"subject":"foo@bar.com"}`),
				Credentials: map[identity.CredentialsType]identity.Credentials{identity.CredentialsTypeOIDC: tc.i}}
			if tc.withpw {
				i.Credentials[identity.CredentialsTypePassword] = identity.Credentials{
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{"foo@bar.com"},
					Config:      []byte(`{"hashed_password":"$argon2id$..."}`),
				}
			}
			actual := populate(t, reg, i, nr())
			assert.EqualValues(t, tc.e, actual.Fields)
		})
	}
}
