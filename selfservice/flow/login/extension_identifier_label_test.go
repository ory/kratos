// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ory/kratos/text"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"

	"github.com/ory/kratos/schema"
)

func constructSchema(t *testing.T, ecModifier, ucModifier func(*schema.ExtensionConfig)) string {
	var emailConfig, usernameConfig schema.ExtensionConfig

	if ecModifier != nil {
		ecModifier(&emailConfig)
	}
	if ucModifier != nil {
		ucModifier(&usernameConfig)
	}

	ec, err := json.Marshal(&emailConfig)
	require.NoError(t, err)
	uc, err := json.Marshal(&usernameConfig)
	require.NoError(t, err)

	ec, err = sjson.DeleteBytes(ec, "verification")
	require.NoError(t, err)
	ec, err = sjson.DeleteBytes(ec, "recovery")
	require.NoError(t, err)
	ec, err = sjson.DeleteBytes(ec, "credentials.code.via")
	require.NoError(t, err)
	ec, err = sjson.DeleteBytes(ec, "organizations.matcher")
	require.NoError(t, err)

	uc, err = sjson.DeleteBytes(uc, "verification")
	require.NoError(t, err)
	uc, err = sjson.DeleteBytes(uc, "recovery")
	require.NoError(t, err)
	uc, err = sjson.DeleteBytes(uc, "credentials.code.via")
	require.NoError(t, err)
	uc, err = sjson.DeleteBytes(uc, "organizations.matcher")
	require.NoError(t, err)

	return "base64://" + base64.StdEncoding.EncodeToString(fmt.Appendf(nil, `
{
  "$id": "https://schemas.ory.sh/presets/kratos/identity.email.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "traits": {
	  "type": "object",
      "properties": {
		"email": {
		  "title": "Email",
		  "type": "string",
		  "ory.sh/kratos": %s
		},
		"username": {
		  "title": "Username",
		  "type": "string",
		  "ory.sh/kratos": %s
		}
	  }
	}
  }
}`, ec, uc))
}

func TestGetIdentifierLabelFromSchema(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		emailConfig, usernameConfig func(*schema.ExtensionConfig)
		expected                    *text.Message
	}{
		{
			name: "email for password",
			emailConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.Password.Identifier = true
			},
			expected: text.NewInfoNodeLabelGenerated("Email", "traits.email"),
		},
		{
			name: "email for webauthn",
			emailConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.WebAuthn.Identifier = true
			},
			expected: text.NewInfoNodeLabelGenerated("Email", "traits.email"),
		},
		{
			name: "email for totp",
			emailConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.TOTP.AccountName = true
			},
			expected: text.NewInfoNodeLabelGenerated("Email", "traits.email"),
		},
		{
			name: "email for code",
			emailConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.Code.Identifier = true
			},
			expected: text.NewInfoNodeLabelGenerated("Email", "traits.email"),
		},
		{
			name: "email for passkey",
			emailConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.Passkey.DisplayName = true
			},
			expected: text.NewInfoNodeLabelGenerated("Email", "traits.email"),
		},
		{
			name: "email for all",
			emailConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.Password.Identifier = true
				c.Credentials.WebAuthn.Identifier = true
				c.Credentials.TOTP.AccountName = true
				c.Credentials.Code.Identifier = true
			},
			expected: text.NewInfoNodeLabelGenerated("Email", "traits.email"),
		},
		{
			name: "username works as well",
			usernameConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.Password.Identifier = true
			},
			expected: text.NewInfoNodeLabelGenerated("Username", "traits.username"),
		},
		{
			name: "multiple identifiers",
			emailConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.Password.Identifier = true
			},
			usernameConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.Password.Identifier = true
			},
			expected: text.NewInfoNodeLabelID(),
		},
		{
			name:     "no identifiers",
			expected: text.NewInfoNodeLabelID(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			label, err := GetIdentifierLabelFromSchema(t.Context(), constructSchema(t, tc.emailConfig, tc.usernameConfig))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, label)
		})
	}
}
