// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/ory/x/snapshotx"

	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"

	"github.com/ory/kratos/corpx"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"

	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
)

func init() {
	corpx.RegisterFakes()
}

func TestSettingsStrategy(t *testing.T) {
	ctx := context.Background()
	if testing.Short() {
		t.Skip()
	}

	var (
		conf, reg = internal.NewFastRegistryWithMocks(t)
		subject   string
		claims    idTokenClaims
		scope     []string
	)

	remoteAdmin, remotePublic, _ := newHydra(t, &subject, &claims, &scope)
	uiTS := newUI(t, reg)
	errTS := testhelpers.NewErrorTestServer(t, reg)
	publicTS, adminTS := testhelpers.NewKratosServers(t)

	viperSetProviderConfig(
		t,
		conf,
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "ory"),
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "google"),
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "github"),
	)
	testhelpers.InitKratosServers(t, reg, publicTS, adminTS)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/settings.schema.json")
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/kratos")

	// Make test data for this test run unique
	testID := x.NewUUID().String()
	users := map[string]*identity.Identity{
		"password": {
			ID: x.NewUUID(), Traits: identity.Traits(`{"email":"john` + testID + `@doe.com"}`),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {
					Type:        "password",
					Identifiers: []string{"john+" + testID + "@doe.com"},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$argon2id$iammocked...."}`),
				},
			},
		},
		"oryer": {
			ID: x.NewUUID(), Traits: identity.Traits(`{"email":"hackerman+` + testID + `@ory.sh"}`),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypeOIDC: {
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"ory:hackerman+" + testID},
					Config:      sqlxx.JSONRawMessage(`{"providers":[{"provider":"ory","subject":"hackerman+` + testID + `"}]}`),
				},
			},
		},
		"githuber": {
			ID: x.NewUUID(), Traits: identity.Traits(`{"email":"hackerman+github+` + testID + `@ory.sh"}`),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypeOIDC: {
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"ory:hackerman+github+" + testID, "github:hackerman+github+" + testID},
					Config:      sqlxx.JSONRawMessage(`{"providers":[{"provider":"ory","subject":"hackerman+github+` + testID + `"},{"provider":"github","subject":"hackerman+github+` + testID + `"}]}`),
				},
			},
			SchemaID: config.DefaultIdentityTraitsSchemaID,
		},
		"multiuser": {
			ID: x.NewUUID(), Traits: identity.Traits(`{"email":"hackerman+multiuser+` + testID + `@ory.sh"}`),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {
					Type:        "password",
					Identifiers: []string{"hackerman+multiuser+" + testID + "@ory.sh"},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$argon2id$iammocked...."}`),
				},
				identity.CredentialsTypeOIDC: {
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"ory:hackerman+multiuser+" + testID, "google:hackerman+multiuser+" + testID},
					Config:      sqlxx.JSONRawMessage(`{"providers":[{"provider":"ory","subject":"hackerman+multiuser+` + testID + `"},{"provider":"google","subject":"hackerman+multiuser+` + testID + `"}]}`),
				},
			},
			SchemaID: config.DefaultIdentityTraitsSchemaID,
		},
	}
	agents := testhelpers.AddAndLoginIdentities(t, reg, publicTS, users)

	newProfileFlow := func(t *testing.T, client *http.Client, redirectTo string, exp time.Duration) *settings.Flow {
		req, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(),
			x.ParseUUID(string(testhelpers.InitializeSettingsFlowViaBrowser(t, client, false, publicTS).Id)))
		require.NoError(t, err)
		assert.Empty(t, req.Active)

		if redirectTo != "" {
			req.RequestURL = redirectTo
		}
		req.ExpiresAt = time.Now().Add(exp)
		require.NoError(t, reg.SettingsFlowPersister().UpdateSettingsFlow(context.Background(), req))

		// sanity check
		got, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), req.ID)
		require.NoError(t, err)
		require.Len(t, got.UI.Nodes, len(req.UI.Nodes))

		return req
	}

	// does the same as new profile request but uses the SDK
	nprSDK := func(t *testing.T, client *http.Client, redirectTo string, exp time.Duration) *kratos.SettingsFlow {
		return testhelpers.InitializeSettingsFlowViaBrowser(t, client, false, publicTS)
	}

	t.Run("case=should not be able to continue a flow with a malformed ID", func(t *testing.T) {
		body, res := testhelpers.HTTPPostForm(t, agents["password"], publicTS.URL+settings.RouteSubmitFlow+"?flow=i-am-not-a-uuid", nil)
		AssertSystemError(t, errTS, res, body, 400, "malformed")
	})

	t.Run("case=should not be able to continue a flow without the flow query parameter", func(t *testing.T) {
		body, res := testhelpers.HTTPPostForm(t, agents["password"], publicTS.URL+settings.RouteSubmitFlow, nil)
		AssertSystemError(t, errTS, res, body, 400, "query parameter is missing")
	})

	t.Run("case=should not be able to continue a flow with a non-existing ID", func(t *testing.T) {
		body, res := testhelpers.HTTPPostForm(t, agents["password"], publicTS.URL+settings.RouteSubmitFlow+"?flow="+x.NewUUID().String(), nil)
		AssertSystemError(t, errTS, res, body, 404, "not be found")
	})

	t.Run("case=should not be able to continue a flow that is expired", func(t *testing.T) {
		req := newProfileFlow(t, agents["password"], "", -time.Hour)
		body, res := testhelpers.HTTPPostForm(t, agents["password"], publicTS.URL+settings.RouteSubmitFlow+"?flow="+req.ID.String(), nil)

		require.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow=")
		assert.NotContains(t, res.Request.URL.String(), req.ID.String(), "should initialize a new flow")
		assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(), "expired")
	})

	t.Run("case=should not be able to fetch another user's data", func(t *testing.T) {
		req := newProfileFlow(t, agents["password"], "", time.Hour)

		_, _, err := testhelpers.NewSDKCustomClient(publicTS, agents["oryer"]).FrontendApi.GetSettingsFlow(context.Background()).Id(req.ID.String()).Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "403")
	})

	t.Run("case=should fetch the settings request and expect data to be set appropriately", func(t *testing.T) {
		req := newProfileFlow(t, agents["password"], "", time.Hour)

		rs, _, err := testhelpers.NewSDKCustomClient(publicTS, agents["password"]).FrontendApi.GetSettingsFlow(context.Background()).Id(req.ID.String()).Execute()
		require.NoError(t, err)

		// Check our sanity. Does the SDK relay the same info that we expect and got from the store?
		assert.Equal(t, publicTS.URL+"/self-service/settings/browser", req.RequestURL)
		assert.Empty(t, req.Active)
		assert.NotEmpty(t, req.IssuedAt)
		assert.EqualValues(t, users["password"].ID, req.Identity.ID)
		assert.EqualValues(t, users["password"].Traits, req.Identity.Traits)
		assert.EqualValues(t, users["password"].SchemaID, req.Identity.SchemaID)

		assert.EqualValues(t, req.ID.String(), rs.Id)
		assert.EqualValues(t, req.RequestURL, rs.RequestUrl)
		assert.EqualValues(t, req.Identity.ID.String(), rs.Identity.Id)
		assert.EqualValues(t, req.IssuedAt, rs.IssuedAt)

		require.NotNil(t, identity.CredentialsTypeOIDC.String(), rs.Ui)
		require.EqualValues(t, "POST", rs.Ui.Method)
		require.EqualValues(t, publicTS.URL+settings.RouteSubmitFlow+"?flow="+req.ID.String(),
			rs.Ui.Action)
	})

	t.Run("case=should adjust linkable providers based on linked credentials", func(t *testing.T) {
		for _, tc := range []struct {
			agent string
		}{
			{agent: "password"},
			{agent: "oryer"},
			{agent: "githuber"},
			{agent: "multiuser"},
		} {
			t.Run("agent="+tc.agent, func(t *testing.T) {
				rs := nprSDK(t, agents[tc.agent], "", time.Hour)
				snapshotx.SnapshotTExcept(t, rs.Ui.Nodes, []string{"0.attributes.value", "1.attributes.value"})
			})
		}
	})

	action := func(req *kratos.SettingsFlow) string {
		return req.Ui.Action
	}

	checkCredentials := func(t *testing.T, shouldExist bool, iid uuid.UUID, provider, subject string, expectTokens bool) {
		actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), iid)
		require.NoError(t, err)

		var cc identity.CredentialsOIDC
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
				if expectTokens {
					assert.NotEmpty(t, p.InitialIDToken)
					assert.NotEmpty(t, p.InitialAccessToken)
					assert.NotEmpty(t, p.InitialRefreshToken)
				}
				break
			}
		}

		require.EqualValues(t, shouldExist, found)
	}

	reset := func(t *testing.T) func() {
		return func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Minute*5)
			agents = testhelpers.AddAndLoginIdentities(t, reg, publicTS, users)
		}
	}

	t.Run("suite=unlink", func(t *testing.T) {
		unlink := func(t *testing.T, agent, provider string) (body []byte, res *http.Response, req *kratos.SettingsFlow) {
			req = nprSDK(t, agents[agent], "", time.Hour)
			body, res = testhelpers.HTTPPostForm(t, agents[agent], action(req),
				&url.Values{"csrf_token": {x.FakeCSRFToken}, "unlink": {provider}})
			return
		}

		unlinkInvalid := func(agent, provider, errorMessage string) func(t *testing.T) {
			return func(t *testing.T) {
				body, res, req := unlink(t, agent, provider)

				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)

				// assert.EqualValues(t, identity.CredentialsTypeOIDC.String(), gjson.GetBytes(body, "active").String())

				// The original options to link google and github are still there
				t.Run("flow=fetch", func(t *testing.T) {
					snapshotx.SnapshotT(t, req.Ui.Nodes, snapshotx.ExceptPaths("0.attributes.value", "1.attributes.value"))
				})

				t.Run("flow=json", func(t *testing.T) {
					snapshotx.SnapshotT(t, json.RawMessage(gjson.GetBytes(body, `ui.nodes`).Raw), snapshotx.ExceptPaths("0.attributes.value", "1.attributes.value"))
				})

				assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), publicTS.URL+settings.RouteSubmitFlow+"?flow=")
				assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(), errorMessage)
			}
		}

		t.Run("case=should not be able to unlink the last remaining connection",
			unlinkInvalid("oryer", "ory", "can not unlink OpenID Connect connection because it is the last remaining first factor credential"))

		t.Run("case=should not be able to unlink an non-existing connection",
			unlinkInvalid("githuber", "i-do-not-exist", "can not unlink non-existing OpenID Connect connection"))

		t.Run("case=should not be able to unlink a connection not yet linked",
			unlinkInvalid("githuber", "google", "can not unlink non-existing OpenID Connect connection"))

		t.Run("case=should unlink a connection", func(t *testing.T) {
			agent, provider := "githuber", "github"
			t.Cleanup(reset(t))

			body, res, req := unlink(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)
			require.Equal(t, "success", gjson.GetBytes(body, "state").String(), "%s", body)

			checkCredentials(t, false, users[agent].ID, provider, "hackerman+github+"+testID, false)
		})

		t.Run("case=should not be able to unlink a connection without a privileged session", func(t *testing.T) {
			agent, provider := "githuber", "github"

			runUnauthed := func(t *testing.T) *kratos.SettingsFlow {
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Millisecond)
				time.Sleep(time.Millisecond)
				t.Cleanup(reset(t))
				_, res, req := unlink(t, agent, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login")

				rs, _, err := testhelpers.NewSDKCustomClient(publicTS, agents[agent]).FrontendApi.GetSettingsFlow(context.Background()).Id(req.Id).Execute()
				require.NoError(t, err)
				require.EqualValues(t, flow.StateShowForm, rs.State)

				checkCredentials(t, true, users[agent].ID, provider, "hackerman+github+"+testID, false)

				return req
			}

			t.Run("subcase=should not update without re-auth", func(t *testing.T) {
				_ = runUnauthed(t)
			})

			t.Run("subcase=should update after re-auth", func(t *testing.T) {
				req := runUnauthed(t)

				// fake login by allowing longer sessions...
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Minute*5)

				body, res := testhelpers.HTTPPostForm(t, agents[agent], action(req),
					&url.Values{"csrf_token": {x.FakeCSRFToken}, "unlink": {provider}})
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)

				assert.Equal(t, "success", gjson.GetBytes(body, "state").String())

				checkCredentials(t, false, users[agent].ID, provider, "hackerman+github+"+testID, false)
			})
		})
	})

	t.Run("suite=link", func(t *testing.T) {
		link := func(t *testing.T, agent, provider string) (body []byte, res *http.Response, req *kratos.SettingsFlow) {
			req = nprSDK(t, agents[agent], "", time.Hour)
			body, res = testhelpers.HTTPPostForm(t, agents[agent], action(req),
				&url.Values{"csrf_token": {x.FakeCSRFToken}, "link": {provider}})
			return
		}

		linkInvalid := func(agent, provider string) func(t *testing.T) {
			return func(t *testing.T) {
				body, res, req := link(t, agent, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)

				// assert.EqualValues(t, identity.CredentialsTypeOIDC.String(), gjson.GetBytes(body, "active").String())
				assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), publicTS.URL+settings.RouteSubmitFlow+"?flow=")

				// The original options to link google and github are still there
				snapshotx.SnapshotTExcept(t, json.RawMessage(gjson.GetBytes(body, `ui.nodes`).Raw), []string{"0.attributes.value", "1.attributes.value"})

				assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(),
					"can not link unknown or already existing OpenID Connect connection")
			}
		}

		t.Run("case=should not be able to link an non-existing connection",
			linkInvalid("oryer", "i-do-not-exist"))

		t.Run("case=should not be able to link a connection which already exists",
			linkInvalid("githuber", "github"))

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

			assert.Contains(t, res.Request.URL.String(), uiTS.URL)
			assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
		})

		t.Run("case=should not be able to link a connection which is missing the ID token", func(t *testing.T) {
			t.Cleanup(reset(t))

			subject = "hackerman+scope-missing+" + testID
			scope = []string{}

			agent, provider := "githuber", "google"
			body, res, _ := link(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			t.Logf("%s", body)
			assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(),
				"no id_token was returned")
		})

		t.Run("case=should not be able to link a connection which is missing the ID token", func(t *testing.T) {
			t.Cleanup(reset(t))

			subject = "hackerman+scope-missing+" + testID
			scope = []string{}

			agent, provider := "githuber", "google"
			body, res, _ := link(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(),
				"no id_token was returned")
		})

		t.Run("case=should link a connection", func(t *testing.T) {
			t.Cleanup(reset(t))

			subject = "hackerman+new-connection+" + testID
			scope = []string{"openid", "offline"}

			agent, provider := "githuber", "google"
			updatedFlow, res, originalFlow := link(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			updatedFlowSDK, _, err := testhelpers.NewSDKCustomClient(publicTS, agents[agent]).FrontendApi.GetSettingsFlow(context.Background()).Id(originalFlow.Id).Execute()
			require.NoError(t, err)
			require.EqualValues(t, flow.StateSuccess, updatedFlowSDK.State)

			t.Run("flow=original", func(t *testing.T) {
				snapshotx.SnapshotTExcept(t, originalFlow.Ui.Nodes, []string{"0.attributes.value", "1.attributes.value"})
			})
			t.Run("flow=response", func(t *testing.T) {
				snapshotx.SnapshotTExcept(t, json.RawMessage(gjson.GetBytes(updatedFlow, "ui.nodes").Raw), []string{"0.attributes.value", "1.attributes.value"})
			})
			t.Run("flow=fetch", func(t *testing.T) {
				snapshotx.SnapshotTExcept(t, updatedFlowSDK.Ui.Nodes, []string{"0.attributes.value", "1.attributes.value"})
			})

			checkCredentials(t, true, users[agent].ID, provider, subject, true)
		})

		t.Run("case=should link a connection even if user does not have oidc credentials yet", func(t *testing.T) {
			t.Cleanup(reset(t))

			subject = "hackerman+new-connection-new-oidc+" + testID
			scope = []string{"openid", "offline"}

			agent, provider := "password", "google"
			_, res, req := link(t, agent, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			rs, _, err := testhelpers.NewSDKCustomClient(publicTS, agents[agent]).FrontendApi.GetSettingsFlow(context.Background()).Id(req.Id).Execute()
			require.NoError(t, err)
			require.EqualValues(t, flow.StateSuccess, rs.State)

			snapshotx.SnapshotTExcept(t, rs.Ui.Nodes, []string{"0.attributes.value", "1.attributes.value"})

			checkCredentials(t, true, users[agent].ID, provider, subject, true)
		})

		t.Run("case=upstream parameters", func(t *testing.T) {
			t.Cleanup(reset(t))

			subject = "hackerman+new-connection-new-oidc-with-parameters+" + testID
			scope = []string{"openid", "offline"}
			agent, provider := "password", "google"

			a := agents[agent]

			t.Run("case=should be able to pass upstream paramters when linking a connection", func(t *testing.T) {
				c := &http.Client{}
				req := nprSDK(t, a, "", time.Hour)
				// copy over the client so we can disable redirects
				*c = *a
				// We need to disable redirects because the upstream parameters are only passed on to the provider
				c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}

				values := &url.Values{}
				values.Set("csrf_token", x.FakeCSRFToken)
				values.Set("link", provider)
				values.Set("upstream_parameters.login_hint", "foo@bar.com")
				values.Set("upstream_parameters.hd", "bar.com")
				values.Set("upstream_parameters.prompt", "consent")

				resp, err := c.PostForm(action(req), *values)
				require.NoError(t, err)
				require.Equal(t, http.StatusSeeOther, resp.StatusCode)

				loc, err := resp.Location()
				require.NoError(t, err)

				require.EqualValues(t, "foo@bar.com", loc.Query().Get("login_hint"))
				require.EqualValues(t, "bar.com", loc.Query().Get("hd"))
				require.EqualValues(t, "consent", loc.Query().Get("prompt"))
			})

			t.Run("case=invalid query parameters should be ignored", func(t *testing.T) {
				c := &http.Client{}
				req := nprSDK(t, a, "", time.Hour)
				// copy over the client so we can disable redirects
				*c = *a
				// We need to disable redirects because the upstream parameters are only passed on to the provider
				c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}

				values := &url.Values{}
				values.Set("csrf_token", x.FakeCSRFToken)
				values.Set("link", provider)
				values.Set("upstream_parameters.lol", "invalid")

				resp, err := c.PostForm(action(req), *values)
				require.NoError(t, err)
				require.Equal(t, http.StatusSeeOther, resp.StatusCode)

				loc, err := resp.Location()
				require.NoError(t, err)

				require.Empty(t, loc.Query().Get("lol"))
			})
		})

		t.Run("case=should not be able to link a connection without a privileged session", func(t *testing.T) {
			agent, provider := "githuber", "google"
			subject = "hackerman+new+google+" + testID

			runUnauthed := func(t *testing.T) *kratos.SettingsFlow {
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Millisecond)
				time.Sleep(time.Millisecond)
				t.Cleanup(reset(t))
				_, res, req := link(t, agent, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login")

				rs, _, err := testhelpers.NewSDKCustomClient(publicTS, agents[agent]).FrontendApi.GetSettingsFlow(context.Background()).Id(req.Id).Execute()
				require.NoError(t, err)
				require.EqualValues(t, flow.StateShowForm, rs.State)

				checkCredentials(t, false, users[agent].ID, provider, subject, true)

				return req
			}

			t.Run("subcase=should not update without re-auth", func(t *testing.T) {
				_ = runUnauthed(t)
			})

			t.Run("subcase=should update after re-auth", func(t *testing.T) {
				req := runUnauthed(t)

				// fake login by allowing longer sessions...
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Minute*5)

				body, res := testhelpers.HTTPPostForm(t, agents[agent], action(req),
					&url.Values{"csrf_token": {x.FakeCSRFToken}, "unlink": {provider}})
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)

				assert.Equal(t, "success", gjson.GetBytes(body, "state").String())

				checkCredentials(t, true, users[agent].ID, provider, subject, true)
			})
		})
	})
}

func TestPopulateSettingsMethod(t *testing.T) {
	ctx := context.Background()
	nreg := func(t *testing.T, conf *oidc.ConfigurationCollection) *driver.RegistryDefault {
		c, reg := internal.NewFastRegistryWithMocks(t)

		testhelpers.SetDefaultIdentitySchema(c, "file://stub/registration.schema.json")
		c.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://www.ory.sh/")

		// Enabled per default:
		// 		conf.Set(ctx, configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": true})
		viperSetProviderConfig(t, c, conf.Providers...)
		return reg
	}

	ns := func(t *testing.T, reg *driver.RegistryDefault) *oidc.Strategy {
		ss, err := reg.SettingsStrategies(context.Background()).Strategy(identity.CredentialsTypeOIDC.String())
		require.NoError(t, err)
		return ss.(*oidc.Strategy)
	}

	nr := func() *settings.Flow {
		return &settings.Flow{Type: flow.TypeBrowser, ID: x.NewUUID(), UI: container.New("")}
	}

	populate := func(t *testing.T, reg *driver.RegistryDefault, i *identity.Identity, req *settings.Flow) *container.Container {
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
		require.NoError(t, ns(t, reg).PopulateSettingsMethod(new(http.Request), i, req))
		require.NotNil(t, req.UI)
		require.NotNil(t, req.UI.Nodes)
		assert.Equal(t, "POST", req.UI.Method)
		return req.UI
	}

	defaultConfig := []oidc.Configuration{
		{Provider: "generic", ID: "facebook"},
		{Provider: "generic", ID: "google"},
		{Provider: "generic", ID: "github"},
	}

	t.Run("case=should not populate non-browser flow", func(t *testing.T) {
		reg := nreg(t, &oidc.ConfigurationCollection{Providers: []oidc.Configuration{{Provider: "generic", ID: "github"}}})
		i := &identity.Identity{Traits: []byte(`{"subject":"foo@bar.com"}`)}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
		req := &settings.Flow{Type: flow.TypeAPI, ID: x.NewUUID(), UI: container.New("")}
		require.NoError(t, ns(t, reg).PopulateSettingsMethod(new(http.Request), i, req))
		require.Empty(t, req.UI.Nodes)
	})

	for k, tc := range []struct {
		c      []oidc.Configuration
		i      *identity.Credentials
		e      node.Nodes
		withpw bool
	}{
		{
			c: []oidc.Configuration{},
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
			},
		},
		{
			c: []oidc.Configuration{
				{Provider: "generic", ID: "github"},
			},
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				oidc.NewLinkNode("github"),
			},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				oidc.NewLinkNode("facebook"),
				oidc.NewLinkNode("google"),
				oidc.NewLinkNode("github"),
			},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				oidc.NewLinkNode("facebook"),
				oidc.NewLinkNode("google"),
				oidc.NewLinkNode("github"),
			},
			i: &identity.Credentials{Type: identity.CredentialsTypeOIDC, Identifiers: []string{}, Config: []byte(`{}`)},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				oidc.NewLinkNode("facebook"),
				oidc.NewLinkNode("github"),
			},
			i: &identity.Credentials{Type: identity.CredentialsTypeOIDC, Identifiers: []string{
				"google:1234",
			}, Config: []byte(`{"providers":[{"provider":"google","subject":"1234"}]}`)},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				oidc.NewLinkNode("facebook"),
				oidc.NewLinkNode("github"),
				oidc.NewUnlinkNode("google"),
			},
			withpw: true,
			i: &identity.Credentials{
				Type: identity.CredentialsTypeOIDC, Identifiers: []string{
					"google:1234",
				},
				Config: []byte(`{"providers":[{"provider":"google","subject":"1234"}]}`),
			},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				oidc.NewLinkNode("github"),
				oidc.NewUnlinkNode("google"),
				oidc.NewUnlinkNode("facebook"),
			},
			i: &identity.Credentials{
				Type: identity.CredentialsTypeOIDC, Identifiers: []string{
					"google:1234",
					"facebook:1234",
				},
				Config: []byte(`{"providers":[{"provider":"google","subject":"1234"},{"provider":"facebook","subject":"1234"}]}`),
			},
		},
	} {
		t.Run("iteration="+strconv.Itoa(k), func(t *testing.T) {
			reg := nreg(t, &oidc.ConfigurationCollection{Providers: tc.c})
			i := &identity.Identity{
				Traits:      []byte(`{"subject":"foo@bar.com"}`),
				Credentials: make(map[identity.CredentialsType]identity.Credentials, 2),
			}
			if tc.i != nil {
				i.Credentials[identity.CredentialsTypeOIDC] = *tc.i
			}
			if tc.withpw {
				i.Credentials[identity.CredentialsTypePassword] = identity.Credentials{
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{"foo@bar.com"},
					Config:      []byte(`{"hashed_password":"$argon2id$..."}`),
				}
			}
			actual := populate(t, reg, i, nr())
			assert.EqualValues(t, tc.e, actual.Nodes)
		})
	}
}
