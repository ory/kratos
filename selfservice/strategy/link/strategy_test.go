// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link_test

import (
	"context"
	"testing"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
)

func initViper(t *testing.T, c *config.Config) {
	ctx := context.Background()
	testhelpers.SetDefaultIdentitySchema(c, "file://./stub/default.schema.json")
	c.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
	c.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh"})
	c.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+identity.CredentialsTypePassword.String()+".enabled", true)
	c.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(recovery.RecoveryStrategyLink)+".enabled", true)
	c.MustSet(ctx, config.ViperKeySelfServiceRecoveryUse, "link")
	c.MustSet(ctx, config.ViperKeySelfServiceRecoveryEnabled, true)
	c.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
	c.MustSet(ctx, config.ViperKeySelfServiceVerificationUse, "link")
}
