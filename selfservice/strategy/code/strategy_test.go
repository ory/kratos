// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	"testing"

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
