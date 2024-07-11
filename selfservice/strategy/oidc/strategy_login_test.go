// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ory/kratos/selfservice/strategy/idfirst"

	configtesthelpers "github.com/ory/kratos/driver/config/testhelpers"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
	"github.com/ory/x/snapshotx"
)

func createIdentity(t *testing.T, ctx context.Context, reg driver.Registry, id uuid.UUID, provider string) *identity.Identity {
	creds, err := identity.NewCredentialsOIDC(new(identity.CredentialsOIDCEncryptedTokens), provider, id.String(), "")
	require.NoError(t, err)

	i := identity.NewIdentity("default")
	i.SetCredentials(identity.CredentialsTypeOIDC, *creds)

	require.NoError(t, reg.IdentityManager().Create(ctx, i))
	return i
}

func TestFormHydration(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	providerID := "test-provider"

	ctx = configtesthelpers.WithConfigValue(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC)+".enabled", true)
	ctx = configtesthelpers.WithConfigValue(
		ctx,
		config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC)+".config",
		map[string]interface{}{
			"providers": []map[string]interface{}{
				{
					"provider":      "generic",
					"id":            providerID,
					"client_id":     "invalid",
					"client_secret": "invalid",
					"issuer_url":    "https://foobar/",
					"mapper_url":    "file://./stub/oidc.facebook.jsonnet",
				},
			},
		},
	)
	ctx = testhelpers.WithDefaultIdentitySchema(ctx, "file://stub/stub.schema.json")

	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypeOIDC)
	require.NoError(t, err)
	fh, ok := s.(login.FormHydrator)
	require.True(t, ok)

	toSnapshot := func(t *testing.T, f *login.Flow) {
		t.Helper()
		// The CSRF token has a unique value that messes with the snapshot - ignore it.
		f.UI.Nodes.ResetNodes("csrf_token")
		snapshotx.SnapshotT(t, f.UI.Nodes)
	}
	newFlow := func(ctx context.Context, t *testing.T) (*http.Request, *login.Flow) {
		r := httptest.NewRequest("GET", "/self-service/login/browser", nil)
		r = r.WithContext(ctx)
		t.Helper()
		f, err := login.NewFlow(conf, time.Minute, "csrf_token", r, flow.TypeBrowser)
		require.NoError(t, err)
		return r, f
	}

	t.Run("method=PopulateLoginMethodSecondFactor", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		f.RequestedAAL = identity.AuthenticatorAssuranceLevel2
		require.NoError(t, fh.PopulateLoginMethodSecondFactor(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodFirstFactor", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodFirstFactor(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodFirstFactorRefresh", func(t *testing.T) {
		r, f := newFlow(ctx, t)

		id := createIdentity(t, ctx, reg, x.NewUUID(), providerID)
		r.Header = testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, id).Transport.(*testhelpers.TransportWithHeader).GetHeader()
		f.Refresh = true

		require.NoError(t, fh.PopulateLoginMethodFirstFactorRefresh(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodSecondFactorRefresh", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodFirstFactorRefresh(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodIdentifierFirstCredentials", func(t *testing.T) {
		t.Run("case=no options", func(t *testing.T) {
			r, f := newFlow(ctx, t)
			require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f), idfirst.ErrNoCredentialsFound)
			toSnapshot(t, f)
		})

		t.Run("case=WithIdentifier", func(t *testing.T) {
			r, f := newFlow(ctx, t)
			require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentifier("foo@bar.com")), idfirst.ErrNoCredentialsFound)
			toSnapshot(t, f)
		})

		t.Run("case=WithIdentityHint", func(t *testing.T) {
			t.Run("case=account enumeration mitigation enabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, true)

				id := identity.NewIdentity(providerID)
				r, f := newFlow(ctx, t)
				require.NoError(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)))
				toSnapshot(t, f)
			})

			t.Run("case=account enumeration mitigation disabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, false)

				t.Run("case=identity has oidc", func(t *testing.T) {
					identifier := x.NewUUID()
					id := createIdentity(t, ctx, reg, identifier, providerID)

					r, f := newFlow(ctx, t)
					require.NoError(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)))
					toSnapshot(t, f)
				})

				t.Run("case=identity does not have a oidc", func(t *testing.T) {
					id := identity.NewIdentity("default")
					r, f := newFlow(ctx, t)
					require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)), idfirst.ErrNoCredentialsFound)
					toSnapshot(t, f)
				})
			})
		})
	})

	t.Run("method=PopulateLoginMethodIdentifierFirstIdentification", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodIdentifierFirstIdentification(r, f))
		toSnapshot(t, f)
	})
}
