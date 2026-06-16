// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sms_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template/sms"
	"github.com/ory/kratos/pkg"
)

func TestNewLoginCodeValid(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t)

	const (
		expectedPhone = "+12345678901"
		otp           = "012345"
	)

	tpl := sms.NewLoginCodeValid(reg, &sms.LoginCodeValidModel{To: expectedPhone, LoginCode: otp})

	expectedBody := "Your login code is: 012345\n\nIt expires in 0 minutes.\n"

	actualBody, err := tpl.SMSBody(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expectedBody, actualBody)

	actualPhone, err := tpl.PhoneNumber()
	require.NoError(t, err)
	assert.Equal(t, expectedPhone, actualPhone)
}
