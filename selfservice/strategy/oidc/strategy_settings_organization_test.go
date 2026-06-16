// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/configx"
	"github.com/ory/x/sqlxx"
)

// TestSettingsOrganizationFlow exercises the org-scoped settings flow end to
// end at the HTTP layer in the OSS package. Coverage:
//
//   - case=init flow with ?organization= persists OrganizationID on the flow.
//   - case=session identity org id overrides ?organization= (session wins).
//   - case=link redirects through RouteOrganizationCallback for the flow's org.
//   - case=unlink under org-scoped flow is rejected (focused integration check
//     complementing the existing TestSettingsStrategy/suite=unlink case).
//
// The full OIDC callback dance — completing the upstream auth and persisting
// identity.OrganizationID via linkCredentials — is exercised end-to-end in the
// Playwright e2e suite (test/e2epw). It cannot run inside this OSS package
// because RouteOrganizationCallback is registered only by the cloud
// organization.Strategy in kratos/kratos/internal/organization/strategy_login.go.
//
// The URL-path mismatch case (Task 11) is covered by
// kratos/kratos/internal/organization/hook_test.go (TestExecuteSettingsPostPersistHook)
// because it lives in the cloud org hook.
func TestSettingsOrganizationFlow(t *testing.T) {
	t.Parallel()

	orgA := uuid.Must(uuid.NewV4())
	orgB := uuid.Must(uuid.NewV4())

	// PrepareOrganizations only honors the ?organization= query parameter when
	// one of the identity's email addresses is under the organization's
	// configured domains, so the orgs claim a domain and the test identities
	// carry a matching email verifiable address.
	const orgADomain = "org-a.example"

	privilegedSession := 5 * time.Minute
	conf, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(testhelpers.DefaultIdentitySchemaConfig("file://./stub/settings.schema.json")),
		configx.WithValues(map[string]any{
			config.ViperKeySelfServiceBrowserDefaultReturnTo:                "https://www.ory.com/kratos",
			config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter: privilegedSession,
			config.ViperKeyOrganizations: []map[string]any{
				{"id": orgA.String(), "domains": []string{orgADomain}},
				{"id": orgB.String(), "domains": []string{"org-b.example"}},
			},
		}),
	)

	hydraAdmin, hydraPublic := newHydra(t)

	_ = newUI(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	publicTS, _ := testhelpers.NewKratosServer(t, reg)

	// newOrgOIDCProvider mirrors newOIDCProvider but additionally registers the
	// org-scoped callback URL (RouteOrganizationCallback) with the upstream
	// hydra client. The non-helper provider only registers /callback/<id>; an
	// org-scoped flow uses /organization/<orgID>/callback/<id>.
	newOrgOIDCProvider := func(id string, organizationID uuid.UUID, opts ...func(*oidc.Configuration)) oidc.Configuration {
		clientID, secret := createClient(t, hydraAdmin, []string{
			publicTS.URL + oidc.RouteBase + "/callback/" + id,
			publicTS.URL + oidc.RouteBase + "/organization/" + organizationID.String() + "/callback/" + id,
			publicTS.URL + oidc.RouteCallbackGeneric,
		})
		cfg := oidc.Configuration{
			Provider:       "generic",
			ID:             id,
			ClientID:       clientID,
			ClientSecret:   secret,
			IssuerURL:      hydraPublic + "/",
			Mapper:         "file://./stub/oidc.hydra.jsonnet",
			OrganizationID: organizationID.String(),
		}
		for _, opt := range opts {
			opt(&cfg)
		}
		return cfg
	}

	setProviderConfig(
		t,
		conf,
		newOIDCProvider(t, publicTS, hydraPublic, hydraAdmin, "ory"),
		newOrgOIDCProvider("orgA-google", orgA, func(c *oidc.Configuration) {
			c.Label = "Org A Google"
		}),
		newOrgOIDCProvider("orgB-google", orgB),
	)

	// Identity with a password credential and no OrganizationID — the typical
	// case for an initial org binding through the settings flow. The email
	// lives under orgA's claimed domain and is carried as a verifiable
	// address so PrepareOrganizations honors ?organization=<orgA>.
	newPasswordIdentity := func(t *testing.T) *identity.Identity {
		t.Helper()
		email := "user-" + x.NewUUID().String() + "@" + orgADomain
		return &identity.Identity{
			Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, email)),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			VerifiableAddresses: []identity.VerifiableAddress{
				*identity.NewVerifiableEmailAddress(email, uuid.Nil),
			},
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{email},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$argon2id$iammocked...."}`),
				},
			},
		}
	}

	// initBrowserSettingsFlow creates a settings flow via the public browser
	// endpoint with an optional ?organization= query parameter, then loads it
	// back from the persister.
	initBrowserSettingsFlow := func(t *testing.T, client *http.Client, orgQuery string) *settings.Flow {
		t.Helper()
		u := publicTS.URL + settings.RouteInitBrowserFlow
		if orgQuery != "" {
			u += "?organization=" + orgQuery
		}
		req, err := http.NewRequest("GET", u, nil)
		require.NoError(t, err)

		res, err := client.Do(req)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		flowID := res.Request.URL.Query().Get("flow")
		require.NotEmpty(t, flowID, "browser-init redirected without a flow id (response URL: %s)", res.Request.URL.String())

		f, err := reg.SettingsFlowPersister().GetSettingsFlow(t.Context(), x.ParseUUID(flowID))
		require.NoError(t, err)
		return f
	}

	t.Run("case=init flow with organization query persists OrganizationID on the flow", func(t *testing.T) {
		_, client := testhelpers.AddAndLoginIdentity(t, reg, newPasswordIdentity(t))

		f := initBrowserSettingsFlow(t, client, orgA.String())

		require.True(t, f.OrganizationID.Valid, "flow should be org-scoped")
		assert.Equal(t, orgA, f.OrganizationID.UUID)
	})

	t.Run("case=init flow with organization query but non-matching email stays unscoped", func(t *testing.T) {
		// The ?organization= parameter is free-form caller input. When none of
		// the identity's email addresses are under the org's configured
		// domains, the flow must stay unscoped — otherwise any authenticated
		// identity could enumerate the org's SSO providers.
		email := "user-" + x.NewUUID().String() + "@nomatch.example"
		outsideIdentity := &identity.Identity{
			Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, email)),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			VerifiableAddresses: []identity.VerifiableAddress{
				*identity.NewVerifiableEmailAddress(email, uuid.Nil),
			},
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{email},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$argon2id$iammocked...."}`),
				},
			},
		}
		_, client := testhelpers.AddAndLoginIdentity(t, reg, outsideIdentity)

		f := initBrowserSettingsFlow(t, client, orgA.String())

		assert.False(t, f.OrganizationID.Valid,
			"flow must not be org-scoped when the identity's email is not under the organization's domains")
	})

	t.Run("case=session identity org id overrides organization query", func(t *testing.T) {
		// An identity already bound to an organization cannot be silently
		// re-bound through this path (PrepareOrganizations: session wins).
		boundIdentity := newPasswordIdentity(t)
		boundIdentity.OrganizationID = uuid.NullUUID{UUID: orgA, Valid: true}
		_, client := testhelpers.AddAndLoginIdentity(t, reg, boundIdentity)

		f := initBrowserSettingsFlow(t, client, orgB.String())

		require.True(t, f.OrganizationID.Valid)
		assert.Equal(t, orgA, f.OrganizationID.UUID,
			"session identity organization must override the organization query parameter")
	})

	t.Run("case=link redirects through RouteOrganizationCallback for the flow's org", func(t *testing.T) {
		_, client := testhelpers.AddAndLoginIdentity(t, reg, newPasswordIdentity(t))

		// Init via HTTP so PrepareOrganizations populates OrganizationID.
		f := initBrowserSettingsFlow(t, client, orgA.String())
		require.True(t, f.OrganizationID.Valid)
		require.Equal(t, orgA, f.OrganizationID.UUID)

		// Disable redirects so we can inspect the location header that points
		// at the upstream provider's auth URL.
		c := *client
		c.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}

		values := url.Values{
			"csrf_token": {nosurfx.FakeCSRFToken},
			"link":       {"orgA-google"},
		}
		res, err := c.PostForm(publicTS.URL+settings.RouteSubmitFlow+"?flow="+f.ID.String(), values)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusSeeOther, res.StatusCode)

		loc, err := res.Location()
		require.NoError(t, err)

		// The upstream auth URL embeds the redirect_uri pointing back at our
		// org-scoped callback path.
		redirectURI := loc.Query().Get("redirect_uri")
		require.NotEmpty(t, redirectURI, "expected redirect_uri on upstream auth URL: %s", loc.String())
		assert.Contains(t, redirectURI, "/self-service/methods/oidc/organization/"+orgA.String()+"/callback/orgA-google",
			"upstream redirect_uri should target RouteOrganizationCallback for orgA")
	})

	t.Run("case=unlink rejection under org-scoped flow", func(t *testing.T) {
		// Identity that already has the orgA-google credential linked, so an
		// unlink request would otherwise be a candidate for processing.
		email := "user-" + x.NewUUID().String() + "@" + orgADomain
		linkedIdentity := &identity.Identity{
			Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, email)),
			SchemaID: config.DefaultIdentityTraitsSchemaID,
			VerifiableAddresses: []identity.VerifiableAddress{
				*identity.NewVerifiableEmailAddress(email, uuid.Nil),
			},
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{email},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$argon2id$iammocked...."}`),
				},
				identity.CredentialsTypeOIDC: {
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"orgA-google:" + email},
					Config: sqlxx.JSONRawMessage(fmt.Sprintf(
						`{"providers":[{"provider":"orgA-google","subject":"%s","organization":"%s"}]}`,
						email, orgA.String())),
				},
			},
		}
		_, client := testhelpers.AddAndLoginIdentity(t, reg, linkedIdentity)

		f := initBrowserSettingsFlow(t, client, orgA.String())
		require.True(t, f.OrganizationID.Valid)

		body, res := testhelpers.HTTPPostForm(t, client, publicTS.URL+settings.RouteSubmitFlow+"?flow="+f.ID.String(),
			&url.Values{"csrf_token": {nosurfx.FakeCSRFToken}, "unlink": {"orgA-google"}})

		// The handler bubbles the unlink rejection through the settings UI;
		// the response payload carries the rejection message rendered into the
		// flow.
		assert.Equal(t, http.StatusOK, res.StatusCode, "%s", body)
		assert.Contains(t,
			gjson.GetBytes(body, "ui.messages.0.text").String(),
			"Cannot unlink",
			"%s", body)
	})

	t.Run("case=link rejection when provider belongs to a different organization", func(t *testing.T) {
		// A flow scoped to orgA must not accept a provider that belongs to
		// orgB: otherwise the user is redirected to orgB's IdP and bound to
		// orgB through an orgA-scoped flow.
		_, client := testhelpers.AddAndLoginIdentity(t, reg, newPasswordIdentity(t))

		f := initBrowserSettingsFlow(t, client, orgA.String())
		require.True(t, f.OrganizationID.Valid)
		require.Equal(t, orgA, f.OrganizationID.UUID)

		body, res := testhelpers.HTTPPostForm(t, client, publicTS.URL+settings.RouteSubmitFlow+"?flow="+f.ID.String(),
			&url.Values{"csrf_token": {nosurfx.FakeCSRFToken}, "link": {"orgB-google"}})

		assert.Equal(t, http.StatusOK, res.StatusCode, "%s", body)
		assert.Contains(t,
			gjson.GetBytes(body, "ui.messages.0.text").String(),
			"can not link unknown or already existing OpenID Connect connection",
			"%s", body)
	})

	t.Run("case=link rejection when org-scoped flow links a non-org provider", func(t *testing.T) {
		// An org-scoped flow must also reject a provider that has no
		// organization at all (the "ory" provider). Only providers of the
		// flow's own organization are linkable.
		_, client := testhelpers.AddAndLoginIdentity(t, reg, newPasswordIdentity(t))

		f := initBrowserSettingsFlow(t, client, orgA.String())
		require.True(t, f.OrganizationID.Valid)
		require.Equal(t, orgA, f.OrganizationID.UUID)

		body, res := testhelpers.HTTPPostForm(t, client, publicTS.URL+settings.RouteSubmitFlow+"?flow="+f.ID.String(),
			&url.Values{"csrf_token": {nosurfx.FakeCSRFToken}, "link": {"ory"}})

		assert.Equal(t, http.StatusOK, res.StatusCode, "%s", body)
		assert.Contains(t,
			gjson.GetBytes(body, "ui.messages.0.text").String(),
			"can not link unknown or already existing OpenID Connect connection",
			"%s", body)
	})

	t.Run("case=link rejection when non-org flow links an org provider", func(t *testing.T) {
		// The symmetric direction: a non-org settings flow must not accept an
		// org-scoped provider on the submit path. The UI already hides it; the
		// server must enforce it too.
		_, client := testhelpers.AddAndLoginIdentity(t, reg, newPasswordIdentity(t))

		f := initBrowserSettingsFlow(t, client, "")
		require.False(t, f.OrganizationID.Valid, "flow should not be org-scoped")

		body, res := testhelpers.HTTPPostForm(t, client, publicTS.URL+settings.RouteSubmitFlow+"?flow="+f.ID.String(),
			&url.Values{"csrf_token": {nosurfx.FakeCSRFToken}, "link": {"orgA-google"}})

		assert.Equal(t, http.StatusOK, res.StatusCode, "%s", body)
		assert.Contains(t,
			gjson.GetBytes(body, "ui.messages.0.text").String(),
			"can not link unknown or already existing OpenID Connect connection",
			"%s", body)
	})
}
