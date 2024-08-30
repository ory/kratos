// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	confighelpers "github.com/ory/kratos/driver/config/testhelpers"
	"github.com/ory/kratos/internal"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/x/stringslice"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/code"
)

func initViper(t *testing.T, ctx context.Context, c *config.Config) {
	testhelpers.SetDefaultIdentitySchema(c, "file://./stub/default.schema.json")
	c.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
	c.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh"})
	c.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+identity.CredentialsTypePassword.String()+".enabled", true)
	c.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(recovery.RecoveryStrategyCode)+".enabled", true)
	c.MustSet(ctx, config.ViperKeySelfServiceRecoveryEnabled, true)
	c.MustSet(ctx, config.ViperKeySelfServiceRecoveryUse, "code")
	c.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
	c.MustSet(ctx, config.ViperKeySelfServiceVerificationUse, "code")
}

func TestGenerateCode(t *testing.T) {
	codes := make([]string, 100)
	for k := range codes {
		codes[k] = code.GenerateCode()
	}

	assert.Len(t, stringslice.Unique(codes), len(codes))
}

func TestMaskAddress(t *testing.T) {
	for _, tc := range []struct {
		address  string
		expected string
	}{
		{
			address:  "a",
			expected: "a",
		},
		{
			address:  "ab@cd",
			expected: "ab****@cd",
		},
		{
			address:  "fixed@ory.sh",
			expected: "fi****@ory.sh",
		},
		{
			address:  "f@ory.sh",
			expected: "f@ory.sh",
		},
		{
			address:  "+12345678910",
			expected: "+12****10",
		},
		{
			address:  "+123456",
			expected: "+12****56",
		},
	} {
		t.Run("case="+tc.address, func(t *testing.T) {
			assert.Equal(t, tc.expected, code.MaskAddress(tc.address))
		})
	}
}

func TestCountActiveCredentials(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	strategy := code.NewStrategy(reg)
	ctx := context.Background()

	t.Run("first factor", func(t *testing.T) {
		for k, tc := range []struct {
			in                  map[identity.CredentialsType]identity.Credentials
			expected            int
			passwordlessEnabled bool
			enabled             bool
		}{
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:   strategy.ID(),
					Config: []byte{},
				}},
				passwordlessEnabled: false,
				enabled:             true,
				expected:            0,
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:   strategy.ID(),
					Config: []byte{},
				}},
				passwordlessEnabled: true,
				enabled:             false,
				expected:            1,
			},
			{
				in:                  map[identity.CredentialsType]identity.Credentials{},
				passwordlessEnabled: true,
				enabled:             true,
				expected:            1,
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:   strategy.ID(),
					Config: []byte(`{}`),
				}},
				passwordlessEnabled: true,
				enabled:             true,
				expected:            1,
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				ctx := confighelpers.WithConfigValue(ctx, "selfservice.methods.code.passwordless_enabled", tc.passwordlessEnabled)
				ctx = confighelpers.WithConfigValue(ctx, "selfservice.methods.code.enabled", tc.enabled)

				cc := map[identity.CredentialsType]identity.Credentials{}
				for _, c := range tc.in {
					cc[c.Type] = c
				}

				actual, err := strategy.CountActiveFirstFactorCredentials(ctx, cc)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			})
		}
	})

	t.Run("second factor", func(t *testing.T) {
		for k, tc := range []struct {
			in         map[identity.CredentialsType]identity.Credentials
			expected   int
			mfaEnabled bool
			enabled    bool
		}{
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:   strategy.ID(),
					Config: []byte{},
				}},
				mfaEnabled: false,
				enabled:    true,
				expected:   0,
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:   strategy.ID(),
					Config: []byte{},
				}},
				mfaEnabled: true,
				enabled:    false,
				expected:   1,
			},
			{
				in:         map[identity.CredentialsType]identity.Credentials{},
				mfaEnabled: true,
				enabled:    true,
				expected:   1,
			},
			{
				in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
					Type:   strategy.ID(),
					Config: []byte(`{}`),
				}},
				mfaEnabled: true,
				enabled:    true,
				expected:   1,
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				ctx := confighelpers.WithConfigValue(ctx, "selfservice.methods.code.mfa_enabled", tc.mfaEnabled)
				ctx = confighelpers.WithConfigValue(ctx, "selfservice.methods.code.enabled", tc.enabled)

				cc := map[identity.CredentialsType]identity.Credentials{}
				for _, c := range tc.in {
					cc[c.Type] = c
				}

				actual, err := strategy.CountActiveMultiFactorCredentials(ctx, cc)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			})
		}
	})
}
