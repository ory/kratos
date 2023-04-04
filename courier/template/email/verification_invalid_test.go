// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package email_test

import (
	"context"
	"testing"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/courier/template/testhelpers"
	"github.com/ory/kratos/internal"
)

func TestVerifyInvalid(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	t.Run("test=with courier templates directory", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		tpl := email.NewVerificationInvalid(reg, &email.VerificationInvalidModel{})

		testhelpers.TestRendered(t, ctx, tpl)
	})

	t.Run("test=with remote resources", func(t *testing.T) {
		t.Run("test=with remote resources", func(t *testing.T) {
			testhelpers.TestRemoteTemplates(t, "../courier/builtin/templates/verification/invalid", courier.TypeVerificationInvalid)
		})
	})
}
