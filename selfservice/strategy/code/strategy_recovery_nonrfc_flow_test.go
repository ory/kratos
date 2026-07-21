// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	"maps"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/x/configx"
)

// TestNonRFCRecoveryFullFlow drives account recovery end to end through the
// real router for an identity whose email is a schema-approved non-RFC carrier
// address, covering submission decode, address lookup, and courier queue.
func TestNonRFCRecoveryFullFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cfg := map[string]any{
		config.ViperKeySelfServiceRecoveryEnabled: true,
		config.ViperKeySelfServiceRecoveryUse:     "code",
	}
	maps.Copy(cfg, testhelpers.DefaultIdentitySchemaConfig("file://./stub/recovery.nonrfc.schema.json"))
	maps.Copy(cfg, testhelpers.MethodEnableConfig(identity.CredentialsTypePassword, true))
	maps.Copy(cfg, testhelpers.MethodEnableConfig(identity.CredentialsTypeCodeAuth, true))

	conf, reg := pkg.NewFastRegistryWithMocks(t, configx.WithValues(cfg))

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	const email = "foo.@docomo.ne.jp"

	id := &identity.Identity{
		Traits:   identity.Traits(`{"email":"` + email + `"}`),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
		State:    identity.StateActive,
	}
	require.NoError(t, reg.IdentityManager().Create(ctx, id, identity.ManagerAllowWriteProtectedTraits))

	client := testhelpers.NewClientWithCookies(t)
	client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper

	testhelpers.SubmitRecoveryForm(t, false, false, client, public,
		func(v url.Values) {
			v.Set("email", email)
			v.Set("method", "code")
		},
		http.StatusOK,
		conf.SelfServiceFlowRecoveryUI(ctx).String(),
	)

	msg := testhelpers.CourierExpectMessage(ctx, t, reg, email, "recover access to your account")
	require.Equal(t, email, msg.Recipient)
}
