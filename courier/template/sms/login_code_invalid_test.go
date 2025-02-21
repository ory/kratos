// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sms_test

import (
	"context"
	"github.com/ory/kratos/courier/template/testhelpers"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template/sms"
	"github.com/ory/kratos/internal"
)

func TestNewLoginCodeInvalid(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	_, reg := internal.NewFastRegistryWithMocks(t)

	const (
		expectedPhone = "+12345678901"
	)

	tpl := sms.NewLoginCodeInvalid(reg, &sms.LoginCodeInvalidModel{To: expectedPhone})

	testhelpers.TestSMSRendered(t, ctx, tpl)

	actualPhone, err := tpl.PhoneNumber()
	require.NoError(t, err)
	assert.Equal(t, expectedPhone, actualPhone)
}
