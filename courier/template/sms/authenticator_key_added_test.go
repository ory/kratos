// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sms_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template/sms"
	"github.com/ory/kratos/pkg"
)

func TestAuthenticatorKeyAddedSMS(t *testing.T) {
	ctx := t.Context()
	_, reg := pkg.NewFastRegistryWithMocks(t)

	tpl := sms.NewAuthenticatorKeyAdded(reg, &sms.AuthenticatorKeyAddedModel{
		To:       "+15551234567",
		AddedAt:  "2026-04-21T12:00:00Z",
		Identity: map[string]any{"ID": "00000000-0000-0000-0000-000000000001"},
	})

	phone, err := tpl.PhoneNumber()
	require.NoError(t, err)
	assert.Equal(t, "+15551234567", phone)

	body, err := tpl.SMSBody(ctx)
	require.NoError(t, err)

	// Must stay under 160 chars so we do not fragment SMS billing.
	assert.LessOrEqual(t, len(body), 160, "SMS body too long: %q", body)
	assert.Contains(t, body, "sign-in method")
}
