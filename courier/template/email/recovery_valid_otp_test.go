package email_test

import (
	"context"
	"testing"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/courier/template/testhelpers"
	"github.com/ory/kratos/internal"
)

func TestRecoverValidOTP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	t.Run("test=with courier templates directory", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		tpl := email.NewRecoveryValidOTP(reg, &email.RecoveryValidOTPModel{})

		testhelpers.TestRendered(t, ctx, tpl)
	})

	t.Run("test=with remote resources", func(t *testing.T) {
		testhelpers.TestRemoteTemplates(t, "../courier/builtin/templates/otp/recovery/valid", courier.TypeRecoveryValid)
	})
}
