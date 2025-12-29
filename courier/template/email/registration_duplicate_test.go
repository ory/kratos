// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package email_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/internal"
)

func TestNewRegistrationDuplicate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	_, reg := internal.NewFastRegistryWithMocks(t)

	model := &email.RegistrationDuplicateModel{
		To:         "test@example.com",
		RequestURL: "https://www.ory.sh/verify",
		TransientPayload: map[string]interface{}{
			"foo": "bar",
		},
	}

	tpl := email.NewRegistrationDuplicate(reg, model)

	t.Run("case=renders subject", func(t *testing.T) {
		subject, err := tpl.EmailSubject(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, subject)
		assert.Contains(t, subject, "Account registration attempt")
	})

	t.Run("case=renders body html", func(t *testing.T) {
		body, err := tpl.EmailBody(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, body)
		assert.Contains(t, body, "Someone tried to create an account")
	})

	t.Run("case=renders body plaintext", func(t *testing.T) {
		body, err := tpl.EmailBodyPlaintext(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, body)
		assert.Contains(t, body, "Someone tried to create an account")
	})

	t.Run("case=email recipient", func(t *testing.T) {
		recipient, err := tpl.EmailRecipient()
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", recipient)
	})
}
