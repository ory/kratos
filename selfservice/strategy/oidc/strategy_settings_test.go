// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/corpx"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/snapshotx"
	"github.com/ory/x/sqlxx"
)

func init() {
	corpx.RegisterFakes()
}

func TestSettingsStrategy(t *testing.T) {
	t.Parallel()

	normalPrivilegedSessionFor := 5 * time.Minute
	conf, reg := internal.NewFastRegistryWithMocks(t,
		configx.WithValues(testhelpers.DefaultIdentitySchemaConfig("file://./stub/settings.schema.json")),
		configx.WithValues(map[string]any{
			config.ViperKeySelfServiceBrowserDefaultReturnTo:                "https://www.ory.sh/kratos",
			config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter: normalPrivilegedSessionFor,
		}),
	)

	var (
		subject string
		claims  idTokenClaims
		scope   []string
	)
	remoteAdmin, remotePublic, _ := newHydra(t, &subject, &claims, &scope)
	uiTS := newUI(t, reg)
	errTS := testhelpers.NewErrorTestServer(t, reg)
	publicTS, _ := testhelpers.NewKratosServer(t, reg)

	viperSetProviderConfig(
		t,
		conf,
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "ory", func(c *oidc.Configuration) {
			c.Label = "Ory"
		}),
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "ory-sso", func(c *oidc.Configuration) {
			c.OrganizationID = "org-1"
		}),
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "google"),
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "github"),
	)

	type userDataFunc = func() (string, *identity.Identity)
	passwordIdentity := func() (string, *identity.Identity) {
		e := testhelpers.RandomEmail()
		return e, &identity.Identity{
			Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, e)),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {
					Type:        "password",
					Identifiers: []string{e},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$argon2id$iammocked...."}`),
				},
			},
		}
	}
	singleOIDCIdentity := func() (string, *identity.Identity) {
		e := testhelpers.RandomEmail()
		return e, &identity.Identity{
			Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, e)),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypeOIDC: {
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"ory:" + e},
					Config:      sqlxx.JSONRawMessage(fmt.Sprintf(`{"providers":[{"provider":"ory","subject":"%s"}]}`, e)),
				},
			},
		}
	}
	multiOIDCIdentity := func() (string, *identity.Identity) {
		e := testhelpers.RandomEmail()
		return e, &identity.Identity{
			Traits: identity.Traits(fmt.Sprintf(`{"email":"%s"}`, e)),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypeOIDC: {
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"ory:" + e, "github:" + e},
					Config:      sqlxx.JSONRawMessage(fmt.Sprintf(`{"providers":[{"provider":"ory","subject":"%s"},{"provider":"github","subject":"%s"}]}`, e, e)),
				},
			},
			SchemaID: config.DefaultIdentityTraitsSchemaID,
		}
	}
	multiCredentialIdentity := func() (string, *identity.Identity) {
		e := testhelpers.RandomEmail()
		return e, &identity.Identity{
			Traits: identity.Traits(fmt.Sprintf(`{"email":"%s"}`, e)),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				"password": {
					Type:        "password",
					Identifiers: []string{e},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$argon2id$iammocked...."}`),
				},
				identity.CredentialsTypeOIDC: {
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"ory:" + e, "google:" + e},
					Config:      sqlxx.JSONRawMessage(fmt.Sprintf(`{"providers":[{"provider":"ory","subject":"%s"},{"provider":"google","subject":"%s"}]}`, e, e)),
				},
			},
			SchemaID: config.DefaultIdentityTraitsSchemaID,
		}
	}

	newProfileFlow := func(t *testing.T, client *http.Client, redirectTo string, exp time.Duration) *settings.Flow {
		req, err := reg.SettingsFlowPersister().GetSettingsFlow(t.Context(),
			x.ParseUUID(testhelpers.InitializeSettingsFlowViaBrowser(t, client, false, publicTS).Id))
		require.NoError(t, err)
		assert.Empty(t, req.Active)

		if redirectTo != "" {
			req.RequestURL = redirectTo
		}
		req.ExpiresAt = time.Now().Add(exp)
		require.NoError(t, reg.SettingsFlowPersister().UpdateSettingsFlow(t.Context(), req))

		// sanity check
		got, err := reg.SettingsFlowPersister().GetSettingsFlow(t.Context(), req.ID)
		require.NoError(t, err)
		require.Len(t, got.UI.Nodes, len(req.UI.Nodes))

		return req
	}

	// does the same as new profile request but uses the SDK
	nprSDK := func(t *testing.T, client *http.Client) *kratos.SettingsFlow {
		return testhelpers.InitializeSettingsFlowViaBrowser(t, client, false, publicTS)
	}

	_, sharedPasswordUser := passwordIdentity()
	sharedUserID, sharedUserClient := testhelpers.AddAndLoginIdentity(t, reg, sharedPasswordUser)

	t.Run("case=should not be able to continue a flow with a malformed ID", func(t *testing.T) {
		body, res := testhelpers.HTTPPostForm(t, sharedUserClient, publicTS.URL+settings.RouteSubmitFlow+"?flow=i-am-not-a-uuid", nil)
		AssertSystemError(t, errTS, res, body, 400, "malformed")
	})

	t.Run("case=should not be able to continue a flow without the flow query parameter", func(t *testing.T) {
		body, res := testhelpers.HTTPPostForm(t, sharedUserClient, publicTS.URL+settings.RouteSubmitFlow, nil)
		AssertSystemError(t, errTS, res, body, 400, "query parameter is missing")
	})

	t.Run("case=should not be able to continue a flow with a non-existing ID", func(t *testing.T) {
		body, res := testhelpers.HTTPPostForm(t, sharedUserClient, publicTS.URL+settings.RouteSubmitFlow+"?flow="+x.NewUUID().String(), nil)
		AssertSystemError(t, errTS, res, body, 404, "not be found")
	})

	t.Run("case=should not be able to continue a flow that is expired", func(t *testing.T) {
		req := newProfileFlow(t, sharedUserClient, "", -time.Hour)
		body, res := testhelpers.HTTPPostForm(t, sharedUserClient, publicTS.URL+settings.RouteSubmitFlow+"?flow="+req.ID.String(), nil)

		require.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow=")
		assert.NotContains(t, res.Request.URL.String(), req.ID.String(), "should initialize a new flow")
		assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(), "expired")
	})

	t.Run("case=should not be able to fetch another user's data", func(t *testing.T) {
		_, otherUserData := passwordIdentity()
		_, otherUserClient := testhelpers.AddAndLoginIdentity(t, reg, otherUserData)

		req := newProfileFlow(t, sharedUserClient, "", time.Hour)

		_, _, err := testhelpers.NewSDKCustomClient(publicTS, otherUserClient).FrontendAPI.GetSettingsFlow(t.Context()).Id(req.ID.String()).Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "403")
	})

	t.Run("case=should fetch the settings request and expect data to be set appropriately", func(t *testing.T) {
		req := newProfileFlow(t, sharedUserClient, "", time.Hour)

		rs, _, err := testhelpers.NewSDKCustomClient(publicTS, sharedUserClient).FrontendAPI.GetSettingsFlow(t.Context()).Id(req.ID.String()).Execute()
		require.NoError(t, err)

		// Check our sanity. Does the SDK relay the same info that we expect and got from the store?
		assert.Equal(t, publicTS.URL+"/self-service/settings/browser", req.RequestURL)
		assert.Empty(t, req.Active)
		assert.NotEmpty(t, req.IssuedAt)
		assert.EqualValues(t, sharedUserID, req.Identity.ID)
		assert.EqualValues(t, sharedPasswordUser.Traits, req.Identity.Traits)
		assert.EqualValues(t, sharedPasswordUser.SchemaID, req.Identity.SchemaID)

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
		_, pw := passwordIdentity()
		_, ory := singleOIDCIdentity()
		_, gh := multiOIDCIdentity()
		_, multi := multiCredentialIdentity()
		userData := map[string]*identity.Identity{
			"password":  pw,
			"oryer":     ory,
			"githuber":  gh,
			"multiuser": multi,
		}
		users := testhelpers.AddAndLoginIdentities(t, reg, userData)
		for name, user := range users {
			t.Run("agent="+name, func(t *testing.T) {
				rs := nprSDK(t, user.Client)
				snapshotx.SnapshotT(t, rs.Ui.Nodes, snapshotx.ExceptPaths("0.attributes.value", "1.attributes.value"))
			})
		}
	})

	action := func(req *kratos.SettingsFlow) string {
		return req.Ui.Action
	}

	checkCredentials := func(t *testing.T, shouldExist bool, iid uuid.UUID, provider, subject string, expectTokens bool) {
		actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(t.Context(), iid)
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

	t.Run("suite=unlink", func(t *testing.T) {
		unlink := func(t *testing.T, agent *http.Client, provider string) (body []byte, res *http.Response, req *kratos.SettingsFlow) {
			req = nprSDK(t, agent)
			body, res = testhelpers.HTTPPostForm(t, agent, action(req),
				&url.Values{"csrf_token": {nosurfx.FakeCSRFToken}, "unlink": {provider}})
			return
		}

		unlinkInvalid := func(userData userDataFunc, provider, errorMessage string) func(t *testing.T) {
			return func(t *testing.T) {
				_, identityData := userData()
				_, client := testhelpers.AddAndLoginIdentity(t, reg, identityData)

				body, res, req := unlink(t, client, provider)

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
			unlinkInvalid(singleOIDCIdentity, "ory", "can not unlink OpenID Connect connection because it is the last remaining first factor credential"))

		t.Run("case=should not be able to unlink an non-existing connection",
			unlinkInvalid(multiOIDCIdentity, "i-do-not-exist", "can not unlink non-existing OpenID Connect connection"))

		t.Run("case=should not be able to unlink a connection not yet linked",
			unlinkInvalid(multiOIDCIdentity, "google", "can not unlink non-existing OpenID Connect connection"))

		t.Run("case=should unlink a connection", func(t *testing.T) {
			email, userData := multiOIDCIdentity()
			// only keep the relevant identity
			userID, client := testhelpers.AddAndLoginIdentity(t, reg, userData)
			provider := "github"

			body, res, req := unlink(t, client, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)
			require.Equalf(t, "success", gjson.GetBytes(body, "state").String(), "%s", body)

			checkCredentials(t, false, userID, provider, email, false)
			checkCredentials(t, true, userID, "ory", email, false)
		})

		t.Run("case=should not be able to unlink a connection without a privileged session", func(t *testing.T) {
			userEmail, userData := multiOIDCIdentity()
			userID, client := testhelpers.AddAndLoginIdentity(t, reg, userData)
			provider := "github"

			runUnauthed := func(t *testing.T) *kratos.SettingsFlow {
				conf.MustSet(t.Context(), config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Millisecond)
				t.Cleanup(func() {
					conf.MustSet(t.Context(), config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, normalPrivilegedSessionFor)
				})
				_, res, req := unlink(t, client, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login")

				fa := testhelpers.NewSDKCustomClient(publicTS, client).FrontendAPI
				lf, _, err := fa.GetLoginFlow(t.Context()).Id(res.Request.URL.Query()["flow"][0]).Execute()
				require.NoError(t, err)

				for _, n := range lf.Ui.Nodes {
					if n.Group == "oidc" && n.Attributes.UiNodeInputAttributes.Name == "provider" {
						assert.Contains(t, []string{"ory", "github"}, n.Attributes.UiNodeInputAttributes.Value)
					}
				}

				rs, _, err := fa.GetSettingsFlow(t.Context()).Id(req.Id).Execute()
				require.NoError(t, err)
				require.EqualValues(t, flow.StateShowForm, rs.State)

				checkCredentials(t, true, userID, provider, userEmail, false)

				return req
			}

			t.Run("subcase=should not update without re-auth", func(t *testing.T) {
				_ = runUnauthed(t)
			})

			t.Run("subcase=should update after re-auth", func(t *testing.T) {
				req := runUnauthed(t)

				// fake login by allowing longer sessions...
				conf.MustSet(t.Context(), config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, normalPrivilegedSessionFor)

				body, res := testhelpers.HTTPPostForm(t, client, action(req),
					&url.Values{"csrf_token": {nosurfx.FakeCSRFToken}, "unlink": {provider}})
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)

				assert.Equal(t, "success", gjson.GetBytes(body, "state").String())

				checkCredentials(t, false, userID, provider, userEmail, false)
				checkCredentials(t, true, userID, "ory", userEmail, false)
			})
		})
	})

	t.Run("suite=link", func(t *testing.T) {
		userEmail, userData := multiOIDCIdentity()
		_, client := testhelpers.AddAndLoginIdentity(t, reg, userData)

		link := func(t *testing.T, client *http.Client, provider string) (body []byte, res *http.Response, req *kratos.SettingsFlow) {
			req = nprSDK(t, client)
			body, res = testhelpers.HTTPPostForm(t, client, action(req),
				&url.Values{"csrf_token": {nosurfx.FakeCSRFToken}, "link": {provider}})
			return
		}

		linkInvalid := func(userData userDataFunc, provider string) func(t *testing.T) {
			return func(t *testing.T) {
				_, identityData := userData()
				_, client := testhelpers.AddAndLoginIdentity(t, reg, identityData)

				body, res, req := link(t, client, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)

				// assert.EqualValues(t, identity.CredentialsTypeOIDC.String(), gjson.GetBytes(body, "active").String())
				assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), publicTS.URL+settings.RouteSubmitFlow+"?flow=")

				// The original options to link google and github are still there
				snapshotx.SnapshotT(t, json.RawMessage(gjson.GetBytes(body, "ui.nodes").Raw), snapshotx.ExceptPaths("0.attributes.value", "1.attributes.value"))

				assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(),
					"can not link unknown or already existing OpenID Connect connection")
			}
		}

		t.Run("case=should not be able to link an non-existing connection",
			linkInvalid(singleOIDCIdentity, "i-do-not-exist"))

		t.Run("case=should not be able to link a connection which already exists",
			linkInvalid(multiOIDCIdentity, "github"))

		t.Run("case=should not be able to link a connection already linked by another identity", func(t *testing.T) {
			// While this theoretically allows for account enumeration - because we see an error indicator if an
			// OIDC connection is being linked that exists already - it would require the attacker to already
			// have control over the social profile, in which case account enumeration is the least of our worries.
			// Instead of using the OIDC profile for enumeration, the attacker would use it for account takeover.
			_, otherUserData := singleOIDCIdentity()
			_, otherClient := testhelpers.AddAndLoginIdentity(t, reg, otherUserData)

			subject = userEmail
			scope = []string{"openid"}

			body, res, _ := link(t, otherClient, "github")

			assert.Contains(t, res.Request.URL.String(), uiTS.URL)
			assert.Containsf(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
		})

		t.Run("case=should not be able to link a connection which is missing the ID token", func(t *testing.T) {
			subject = "hackerman+scope-missing"
			scope = []string{}

			body, res, _ := link(t, client, "google")
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			assert.Containsf(t, gjson.GetBytes(body, `ui.messages.0.text`).String(),
				"no id_token was returned",
				"%s", body)
		})

		t.Run("case=should not be able to link a connection which is missing the ID token", func(t *testing.T) {
			subject = "hackerman+scope-missing"
			scope = []string{}

			body, res, _ := link(t, client, "google")
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			assert.Contains(t, gjson.GetBytes(body, `ui.messages.0.text`).String(),
				"no id_token was returned")
		})

		t.Run("case=should link a connection", func(t *testing.T) {
			subject = testhelpers.RandomEmail()
			scope = []string{"openid", "offline"}
			provider := "google"

			_, userData := multiOIDCIdentity()
			userID, client := testhelpers.AddAndLoginIdentity(t, reg, userData)

			updatedFlow, res, originalFlow := link(t, client, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			updatedFlowSDK, _, err := testhelpers.NewSDKCustomClient(publicTS, client).FrontendAPI.GetSettingsFlow(t.Context()).Id(originalFlow.Id).Execute()
			require.NoError(t, err)
			require.EqualValues(t, flow.StateSuccess, updatedFlowSDK.State)

			t.Run("flow=original", func(t *testing.T) {
				snapshotx.SnapshotT(t, originalFlow.Ui.Nodes, snapshotx.ExceptPaths("0.attributes.value", "1.attributes.value"))
			})
			t.Run("flow=response", func(t *testing.T) {
				snapshotx.SnapshotT(t, json.RawMessage(gjson.GetBytes(updatedFlow, "ui.nodes").Raw), snapshotx.ExceptPaths("0.attributes.value", "1.attributes.value"))
			})
			t.Run("flow=fetch", func(t *testing.T) {
				snapshotx.SnapshotT(t, updatedFlowSDK.Ui.Nodes, snapshotx.ExceptPaths("0.attributes.value", "1.attributes.value"))
			})

			checkCredentials(t, true, userID, provider, subject, true)
		})

		t.Run("case=should link a connection and add auth method to session", func(t *testing.T) {
			_, userData := multiOIDCIdentity()
			_, client := testhelpers.AddAndLoginIdentity(t, reg, userData)
			provider := "google"

			subject = testhelpers.RandomEmail()
			scope = []string{"openid", "offline"}

			_, res, _ := link(t, client, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			// Get the specific session for this agent using SDK
			sess, _, err := testhelpers.NewSDKCustomClient(publicTS, client).
				FrontendAPI.
				ToSession(t.Context()).
				Execute()
			require.NoError(t, err)
			require.NotNil(t, sess)

			// Check that the session has the expected auth method
			found := slices.ContainsFunc(sess.AuthenticationMethods, func(am kratos.SessionAuthenticationMethod) bool {
				return am.Method != nil &&
					am.Provider != nil &&
					*am.Method == string(identity.CredentialsTypeOIDC) &&
					*am.Provider == provider
			})

			require.True(t, found, "session should contain OIDC auth method for provider %s", provider)
		})

		t.Run("case=should link a connection even if user does not have oidc credentials yet", func(t *testing.T) {
			_, userData := passwordIdentity()
			userID, client := testhelpers.AddAndLoginIdentity(t, reg, userData)
			provider := "google"

			subject = testhelpers.RandomEmail()
			scope = []string{"openid", "offline"}

			_, res, req := link(t, client, provider)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)

			rs, _, err := testhelpers.NewSDKCustomClient(publicTS, client).FrontendAPI.GetSettingsFlow(t.Context()).Id(req.Id).Execute()
			require.NoError(t, err)
			require.EqualValues(t, flow.StateSuccess, rs.State)

			snapshotx.SnapshotT(t, rs.Ui.Nodes, snapshotx.ExceptPaths("0.attributes.value", "1.attributes.value"))

			checkCredentials(t, true, userID, provider, subject, true)
		})

		t.Run("case=upstream parameters", func(t *testing.T) {
			subject = ""
			scope = nil
			provider := "google"

			t.Run("case=should be able to pass upstream paramters when linking a connection", func(t *testing.T) {
				req := nprSDK(t, client)
				// copy over the client so we can disable redirects
				c := *client
				// We need to disable redirects because the upstream parameters are only passed on to the provider
				c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}

				values := &url.Values{}
				values.Set("csrf_token", nosurfx.FakeCSRFToken)
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
				req := nprSDK(t, client)
				// copy over the client so we can disable redirects
				c := *client
				// We need to disable redirects because the upstream parameters are only passed on to the provider
				c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}

				values := &url.Values{}
				values.Set("csrf_token", nosurfx.FakeCSRFToken)
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
			_, userData := singleOIDCIdentity()
			userID, client := testhelpers.AddAndLoginIdentity(t, reg, userData)
			provider := "google"

			subject = testhelpers.RandomEmail()
			scope = []string{"openid", "offline"}

			runUnauthed := func(t *testing.T) *kratos.SettingsFlow {
				conf.MustSet(t.Context(), config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Millisecond)
				t.Cleanup(func() {
					conf.MustSet(t.Context(), config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, normalPrivilegedSessionFor)
				})
				_, res, req := link(t, client, provider)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login")

				fa := testhelpers.NewSDKCustomClient(publicTS, client).FrontendAPI
				lf, _, err := fa.GetLoginFlow(t.Context()).Id(res.Request.URL.Query()["flow"][0]).Execute()
				require.NoError(t, err)

				for _, n := range lf.Ui.Nodes {
					if n.Group == "oidc" && n.Attributes.UiNodeInputAttributes.Name == "provider" {
						assert.Contains(t, []string{"ory", "github"}, n.Attributes.UiNodeInputAttributes.Value)
					}
				}

				rs, _, err := fa.GetSettingsFlow(t.Context()).Id(req.Id).Execute()
				require.NoError(t, err)
				require.EqualValues(t, flow.StateShowForm, rs.State)

				checkCredentials(t, false, userID, provider, subject, true)

				return req
			}

			t.Run("subcase=should not update without re-auth", func(t *testing.T) {
				_ = runUnauthed(t)
			})

			t.Run("subcase=should update after re-auth", func(t *testing.T) {
				req := runUnauthed(t)

				// fake login by allowing longer sessions...
				conf.MustSet(t.Context(), config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, normalPrivilegedSessionFor)

				body, res := testhelpers.HTTPPostForm(t, client, action(req),
					&url.Values{"csrf_token": {nosurfx.FakeCSRFToken}, "unlink": {provider}})
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/settings?flow="+req.Id)

				assert.Equal(t, "success", gjson.GetBytes(body, "state").String())

				checkCredentials(t, true, userID, provider, subject, true)
			})
		})
	})
}

func TestPopulateSettingsMethod(t *testing.T) {
	t.Parallel()
	nCtx := func(t *testing.T, conf *oidc.ConfigurationCollection) (*driver.RegistryDefault, context.Context) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		ctx := context.Background()
		ctx = testhelpers.WithDefaultIdentitySchema(ctx, "file://stub/registration.schema.json")
		ctx = contextx.WithConfigValue(ctx, config.ViperKeyPublicBaseURL, "https://www.ory.sh/")
		baseKey := fmt.Sprintf("%s.%s", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeOIDC)

		ctx = contextx.WithConfigValues(ctx, map[string]interface{}{
			baseKey + ".enabled": true,
			baseKey + ".config":  conf,
		})

		// Enabled per default:
		// 		conf.Set(ctx, configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": true})
		// viperSetProviderConfig(t, c, conf.Providers...)
		return reg, ctx
	}

	ns := func(t *testing.T, reg *driver.RegistryDefault, ctx context.Context) *oidc.Strategy {
		ss, err := reg.SettingsStrategies(ctx).Strategy(identity.CredentialsTypeOIDC.String())
		require.NoError(t, err)
		return ss.(*oidc.Strategy)
	}

	nr := func() *settings.Flow {
		return &settings.Flow{Type: flow.TypeBrowser, ID: x.NewUUID(), UI: container.New("")}
	}

	populate := func(t *testing.T, reg *driver.RegistryDefault, ctx context.Context, i *identity.Identity, f *settings.Flow) *container.Container {
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(ctx, i))
		req := new(http.Request)
		require.NoError(t, ns(t, reg, ctx).PopulateSettingsMethod(ctx, req, i, f))
		require.NotNil(t, f.UI)
		require.NotNil(t, f.UI.Nodes)
		assert.Equal(t, "POST", f.UI.Method)
		return f.UI
	}

	defaultConfig := []oidc.Configuration{
		{Provider: "generic", ID: "facebook"},
		{Provider: "generic", ID: "google"},
		{Provider: "generic", ID: "github"},
	}

	t.Run("case=should not populate non-browser flow", func(t *testing.T) {
		t.Parallel()
		reg, ctx := nCtx(t, &oidc.ConfigurationCollection{Providers: []oidc.Configuration{{Provider: "generic", ID: "github"}}})
		i := &identity.Identity{Traits: []byte(`{"subject":"foo@bar.com"}`)}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(ctx, i))
		f := &settings.Flow{Type: flow.TypeAPI, ID: x.NewUUID(), UI: container.New("")}
		req := new(http.Request)
		require.NoError(t, ns(t, reg, ctx).PopulateSettingsMethod(ctx, req, i, f))
		require.Empty(t, f.UI.Nodes)
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
				node.NewCSRFNode(nosurfx.FakeCSRFToken),
			},
		},
		{
			c: []oidc.Configuration{
				{Provider: "generic", ID: "github"},
			},
			e: node.Nodes{
				node.NewCSRFNode(nosurfx.FakeCSRFToken),
				oidc.NewLinkNode("github", "github"),
			},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(nosurfx.FakeCSRFToken),
				oidc.NewLinkNode("facebook", "facebook"),
				oidc.NewLinkNode("google", "google"),
				oidc.NewLinkNode("github", "github"),
			},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(nosurfx.FakeCSRFToken),
				oidc.NewLinkNode("facebook", "facebook"),
				oidc.NewLinkNode("google", "google"),
				oidc.NewLinkNode("github", "github"),
			},
			i: &identity.Credentials{Type: identity.CredentialsTypeOIDC, Identifiers: []string{}, Config: []byte(`{}`)},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(nosurfx.FakeCSRFToken),
				oidc.NewLinkNode("facebook", "facebook"),
				oidc.NewLinkNode("github", "github"),
			},
			i: &identity.Credentials{Type: identity.CredentialsTypeOIDC, Identifiers: []string{
				"google:1234",
			}, Config: []byte(`{"providers":[{"provider":"google","subject":"1234"}]}`)},
		},
		{
			c: defaultConfig,
			e: node.Nodes{
				node.NewCSRFNode(nosurfx.FakeCSRFToken),
				oidc.NewLinkNode("facebook", "facebook"),
				oidc.NewLinkNode("github", "github"),
				oidc.NewUnlinkNode("google", "google"),
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
				node.NewCSRFNode(nosurfx.FakeCSRFToken),
				oidc.NewLinkNode("github", "github"),
				oidc.NewUnlinkNode("google", "google"),
				oidc.NewUnlinkNode("facebook", "facebook"),
			},
			i: &identity.Credentials{
				Type: identity.CredentialsTypeOIDC, Identifiers: []string{
					"google:1234",
					"facebook:1234",
				},
				Config: []byte(`{"providers":[{"provider":"google","subject":"1234"},{"provider":"facebook","subject":"1234"}]}`),
			},
		},
		{
			c: []oidc.Configuration{
				{Provider: "generic", ID: "labeled", Label: "Labeled"},
			},
			e: node.Nodes{
				node.NewCSRFNode(nosurfx.FakeCSRFToken),
				oidc.NewLinkNode("labeled", "Labeled"),
			},
		},
		{
			c: []oidc.Configuration{
				{Provider: "generic", ID: "labeled", Label: "Labeled"},
				{Provider: "generic", ID: "facebook"},
			},
			e: node.Nodes{
				node.NewCSRFNode(nosurfx.FakeCSRFToken),
				oidc.NewUnlinkNode("labeled", "Labeled"),
				oidc.NewUnlinkNode("facebook", "facebook"),
			},
			i: &identity.Credentials{
				Type: identity.CredentialsTypeOIDC, Identifiers: []string{
					"labeled:1234",
					"facebook:1234",
				},
				Config: []byte(`{"providers":[{"provider":"labeled","subject":"1234"},{"provider":"facebook","subject":"1234"}]}`),
			},
		},
	} {
		t.Run("iteration="+strconv.Itoa(k), func(t *testing.T) {
			t.Parallel()
			reg, ctx := nCtx(t, &oidc.ConfigurationCollection{Providers: tc.c})
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
			actual := populate(t, reg, ctx, i, nr())
			assert.EqualValues(t, tc.e, actual.Nodes)
		})
	}
}
