// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package profile_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/assertx"
	"github.com/ory/x/snapshotx"
)

func TestTwoStepRegistration(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")

	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePassword.String(), true)
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeOIDC.String(), true)
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePasskey.String(), true)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC)+".config", &oidc.ConfigurationCollection{Providers: []oidc.Configuration{
		{
			ID:           "google",
			Provider:     "google",
			ClientID:     "1234",
			ClientSecret: "1234",
		},
	}})

	conf.MustSet(ctx, config.ViperKeyWebAuthnPasswordless, true)
	conf.MustSet(ctx, config.ViperKeyPasskeyRPID, "localhost")
	conf.MustSet(ctx, config.ViperKeyPasskeyRPDisplayName, "localhost")
	conf.MustSet(ctx, config.ViperKeyWebAuthnRPID, "localhost")
	conf.MustSet(ctx, config.ViperKeyWebAuthnRPDisplayName, "localhost")
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.passwordless_enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth), true)

	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationFlowStyle, "profile_first")

	_ = testhelpers.NewErrorTestServer(t, reg)
	publicTS, _ := testhelpers.NewKratosServer(t, reg)
	_ = testhelpers.NewRedirSessionEchoTS(t, reg)
	ui := testhelpers.NewRegistrationUIFlowEchoServer(t, reg)

	t.Run("initial form is populated with identity traits", func(t *testing.T) {
		t.Run("type=browser", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)

			t.Run("empty_flow", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, publicTS, false, false, false)
				snapshotx.SnapshotT(t, f.Ui.Nodes, snapshotx.ExceptPaths(
					"1.attributes.value",
					"8.attributes.nonce",
					"8.attributes.src",
					"10.attributes.value",
				))
			})

			t.Run("select_credentials", func(t *testing.T) {
				res := testhelpers.SubmitRegistrationForm(t, false, client, publicTS, func(v url.Values) {
					v.Set("traits.email", "browser-1@example.org")
					v.Set("traits.booly", "true")
					v.Set("traits.numby", "1")
					v.Set("traits.stringy", "string")
					v.Set("traits.should_big_number", "1000000")
					v.Set("traits.should_long_string", "1111111111111111111111111111111111111111111111111111111111")

					v.Set("method", "profile")
				}, false, http.StatusOK, ui.URL)
				snapshotx.SnapshotT(t, json.RawMessage(gjson.Get(res, "ui.nodes").Raw), snapshotx.ExceptPaths(
					"1.attributes.value",
					"8.attributes.nonce",
					"8.attributes.src",
					"11.attributes.value",
				))
			})

			t.Run("return_to_profile", func(t *testing.T) {
				res := testhelpers.SubmitRegistrationForm(t, false, client, publicTS, func(v url.Values) {
					v.Set("traits.email", "browser-1-1@example.org")
					v.Set("traits.booly", "true")
					v.Set("traits.numby", "1")
					v.Set("traits.stringy", "string")
					v.Set("traits.should_big_number", "1000000")
					v.Set("traits.should_long_string", "1111111111111111111111111111111111111111111111111111111111")

					v.Set("screen", "previous")
				}, false, http.StatusOK, ui.URL)
				snapshotx.SnapshotT(t, json.RawMessage(gjson.Get(res, "ui.nodes").Raw), snapshotx.ExceptPaths(
					"1.attributes.value",
					"8.attributes.nonce",
					"8.attributes.src",
					"10.attributes.value",
				))
			})

			t.Run("select_credentials_again", func(t *testing.T) {
				res := testhelpers.SubmitRegistrationForm(t, false, client, publicTS, func(v url.Values) {
					v.Set("traits.email", "browser-1-1@example.org")
					v.Set("method", "profile")
				}, false, http.StatusOK, ui.URL)
				snapshotx.SnapshotT(t, json.RawMessage(gjson.Get(res, "ui.nodes").Raw), snapshotx.ExceptPaths(
					"1.attributes.value",
					"8.attributes.nonce",
					"8.attributes.src",
					"11.attributes.value",
				))
			})
		})
	})
}

func TestOneStepRegistration(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePassword.String(), true)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")
	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationFlowStyle, "unified")

	//ui := testhelpers.NewSettingsUIEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	//publicTS, _ := testhelpers.NewKratosServer(t, reg)

	t.Run("initial form is populated with identity traits", func(t *testing.T) {
		t.Run("type=browser", func(t *testing.T) {
		})
	})
}

func TestPopulateRegistrationMethod(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	ctx = testhelpers.WithDefaultIdentitySchema(ctx, "file://stub/identity.schema.json")

	s, err := reg.AllRegistrationStrategies().Strategy(identity.CredentialsTypeProfile)
	require.NoError(t, err)
	fh, ok := s.(registration.FormHydrator)
	require.True(t, ok)

	toSnapshot := func(t *testing.T, f node.Nodes) {
		t.Helper()
		// The CSRF token has a unique value that messes with the snapshot - ignore it.
		f.ResetNodes("csrf_token")
		f.ResetNodes("passkey_challenge")
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
