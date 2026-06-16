// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sms_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template/sms"
	"github.com/ory/kratos/pkg"
)

// TestOriginBoundSMSBodies asserts that the login, registration, and
// verification code SMS bodies include the Web OTP origin-bound last line
// `@<domain> #<code>`, mirroring the recovery code template.
func TestOriginBoundSMSBodies(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t)

	t.Run("template=login_code", func(t *testing.T) {
		body, err := sms.NewLoginCodeValid(reg, &sms.LoginCodeValidModel{LoginCode: "012345", RequestURLDomain: "auth.example.com"}).SMSBody(t.Context())
		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(strings.TrimRight(body, "\n"), "@auth.example.com #012345"),
			"body must end with the origin-bound line, got: %q", body)
	})

	t.Run("template=registration_code", func(t *testing.T) {
		body, err := sms.NewRegistrationCodeValid(reg, &sms.RegistrationCodeValidModel{RegistrationCode: "012345", RequestURLDomain: "auth.example.com"}).SMSBody(t.Context())
		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(strings.TrimRight(body, "\n"), "@auth.example.com #012345"),
			"body must end with the origin-bound line, got: %q", body)
	})

	t.Run("template=verification_code", func(t *testing.T) {
		body, err := sms.NewVerificationCodeValid(reg, &sms.VerificationCodeValidModel{VerificationCode: "012345", RequestURLDomain: "auth.example.com"}).SMSBody(t.Context())
		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(strings.TrimRight(body, "\n"), "@auth.example.com #012345"),
			"body must end with the origin-bound line, got: %q", body)
	})
}
