// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {
	expected := &Message{ID: InfoSelfServiceSettingsUpdateSuccess, Text: "foo", Type: Info}

	v, err := expected.Value()
	require.NoError(t, err)

	var actual Message
	require.NoError(t, actual.Scan(v))

	assert.EqualValues(t, expected, &actual, v)
}

func TestNewErrorValidationDuplicateCredentialsWithHints(t *testing.T) {
	t.Run("title-cases real provider names", func(t *testing.T) {
		msg := NewErrorValidationDuplicateCredentialsWithHints(nil, []string{"google", "github"}, "user@example.com")
		assert.Contains(t, msg.Text, "Google, Github")
	})

	t.Run("preserves placeholder templates without title-casing", func(t *testing.T) {
		msg := NewErrorValidationDuplicateCredentialsWithHints(
			[]string{"{available_credential_types_list}"},
			[]string{"{available_oidc_providers_list}"},
			"{credential_identifier_hint}",
		)
		assert.Contains(t, msg.Text, "{available_oidc_providers_list}")
		assert.NotContains(t, msg.Text, "{Available_oidc_providers_list}")
	})
}

func TestMessages(t *testing.T) {
	expected := Messages{{ID: InfoSelfServiceSettingsUpdateSuccess, Text: "foo", Type: Info}}

	v, err := expected.Value()
	require.NoError(t, err)

	var actual Messages
	require.NoError(t, actual.Scan(v))

	assert.EqualValues(t, expected, actual, v)
}
