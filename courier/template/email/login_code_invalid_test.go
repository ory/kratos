// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package email_test

import (
	"context"
	"github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/courier/template/testhelpers"
	"testing"

	"github.com/ory/kratos/internal"
)

func TestNewLoginCodeInvalid(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	_, reg := internal.NewFastRegistryWithMocks(t)

	tpl := email.NewLoginCodeInvalid(reg, &email.LoginCodeInvalidModel{})

	testhelpers.TestRendered(t, ctx, tpl)
}
