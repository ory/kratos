package otp_test

import (
	"context"
	"testing"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/recovery"
)

func initConfigEmail(ctx context.Context, t *testing.T, c *config.Config) {
	testhelpers.SetDefaultIdentitySchema(c, "file://./stub/email.schema.json")
	c.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
	c.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh"})

	c.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+identity.CredentialsTypePassword.String()+".enabled", true)
	c.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+recovery.StrategyRecoveryOTPName+".enabled", true)

	c.MustSet(ctx, config.ViperKeySelfServiceRecoveryEnabled, true)
	c.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
}

func initConfigPhone(ctx context.Context, t *testing.T, c *config.Config) {
	testhelpers.SetDefaultIdentitySchema(c, "file://./stub/phone.schema.json")
	c.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
	c.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh"})

	c.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+identity.CredentialsTypePassword.String()+".enabled", true)
	c.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+recovery.StrategyRecoveryOTPName+".enabled", true)

	c.MustSet(ctx, config.ViperKeySelfServiceRecoveryEnabled, true)
	c.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
}
