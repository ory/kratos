// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	configtesthelpers "github.com/ory/kratos/driver/config/testhelpers"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/assertx"
	"github.com/ory/x/snapshotx"
)

func TestPopulateRegistrationMethod(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	ctx = testhelpers.WithDefaultIdentitySchema(ctx, "file://stub/registration.schema.json")
	ctx = configtesthelpers.WithConfigValue(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC)+".enabled", true)
	ctx = configtesthelpers.WithConfigValue(
		ctx,
		config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC)+".config",
		map[string]interface{}{
			"providers": []map[string]interface{}{
				{
					"provider":      "generic",
					"id":            "providerID",
					"client_id":     "invalid",
					"client_secret": "invalid",
					"issuer_url":    "https://foobar/",
					"mapper_url":    "file://./stub/oidc.facebook.jsonnet",
				},
			},
		},
	)

	s, err := reg.AllRegistrationStrategies().Strategy(identity.CredentialsTypeOIDC)
	require.NoError(t, err)

	fh, ok := s.(registration.FormHydrator)
	require.True(t, ok)

	toSnapshot := func(t *testing.T, f node.Nodes) {
		t.Helper()
		// The CSRF token has a unique value that messes with the snapshot - ignore it.
		f.ResetNodes("csrf_token")
		snapshotx.SnapshotT(t, f, snapshotx.ExceptNestedKeys("nonce", "src"))
	}

	newFlow := func(ctx context.Context, t *testing.T) (*http.Request, *registration.Flow) {
		r := httptest.NewRequest("GET", "/self-service/registration/browser", nil)
		r = r.WithContext(ctx)
		t.Helper()
		f, err := registration.NewFlow(conf, time.Minute, "csrf_token", r, flow.TypeBrowser)
		f.UI.Nodes = make(node.Nodes, 0)
		require.NoError(t, err)
		return r, f
	}

	t.Run("method=PopulateRegistrationMethod", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateRegistrationMethod(r, f))
		toSnapshot(t, f.UI.Nodes)
	})

	t.Run("method=PopulateRegistrationMethodProfile", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateRegistrationMethodProfile(r, f))
		toSnapshot(t, f.UI.Nodes)
	})

	t.Run("method=PopulateRegistrationMethodCredentials", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateRegistrationMethodCredentials(r, f))
		toSnapshot(t, f.UI.Nodes)
	})

	t.Run("method=idempotency", func(t *testing.T) {
		r, f := newFlow(ctx, t)

		var snapshots []node.Nodes

		t.Run("case=1", func(t *testing.T) {
			require.NoError(t, fh.PopulateRegistrationMethodProfile(r, f))
			snapshots = append(snapshots, f.UI.Nodes)
			toSnapshot(t, f.UI.Nodes)
		})

		t.Run("case=2", func(t *testing.T) {
			require.NoError(t, fh.PopulateRegistrationMethodCredentials(r, f))
			snapshots = append(snapshots, f.UI.Nodes)
			toSnapshot(t, f.UI.Nodes)
		})

		t.Run("case=3", func(t *testing.T) {
			require.NoError(t, fh.PopulateRegistrationMethodProfile(r, f))
			snapshots = append(snapshots, f.UI.Nodes)
			toSnapshot(t, f.UI.Nodes)
		})

		t.Run("case=4", func(t *testing.T) {
			require.NoError(t, fh.PopulateRegistrationMethodCredentials(r, f))
			snapshots = append(snapshots, f.UI.Nodes)
			toSnapshot(t, f.UI.Nodes)
		})

		t.Run("case=evaluate", func(t *testing.T) {
			assertx.EqualAsJSON(t, snapshots[0], snapshots[2])
			assertx.EqualAsJSON(t, snapshots[1], snapshots[3])
		})
	})
}
