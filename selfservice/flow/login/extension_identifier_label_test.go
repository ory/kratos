// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"
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
	uc, err = sjson.DeleteBytes(uc, "verification")
	require.NoError(t, err)
	uc, err = sjson.DeleteBytes(uc, "recovery")
	require.NoError(t, err)
	uc, err = sjson.DeleteBytes(uc, "credentials.code.via")
	require.NoError(t, err)

	return "base64://" + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`
{
  "properties": {
    "traits": {
      "properties": {
		"email": {
		  "title": "Email",
		  "ory.sh/kratos": %s
		},
		"username": {
		  "title": "Username",
		  "ory.sh/kratos": %s
		}
	  }
	}
  }
}`, ec, uc)))
}

func TestGetIdentifierLabelFromSchema(t *testing.T) {
	ctx := context.Background()

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
			expected: text.NewInfoNodeLabelGenerated("Email"),
		},
		{
			name: "email for webauthn",
			emailConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.WebAuthn.Identifier = true
			},
			expected: text.NewInfoNodeLabelGenerated("Email"),
		},
		{
			name: "email for totp",
			emailConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.TOTP.AccountName = true
			},
			expected: text.NewInfoNodeLabelGenerated("Email"),
		},
		{
			name: "email for code",
			emailConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.Code.Identifier = true
			},
			expected: text.NewInfoNodeLabelGenerated("Email"),
		},
		{
			name: "email for passkey",
			emailConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.Passkey.DisplayName = true
			},
			expected: text.NewInfoNodeLabelGenerated("Email"),
		},
		{
			name: "email for all",
			emailConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.Password.Identifier = true
				c.Credentials.WebAuthn.Identifier = true
				c.Credentials.TOTP.AccountName = true
				c.Credentials.Code.Identifier = true
			},
			expected: text.NewInfoNodeLabelGenerated("Email"),
		},
		{
			name: "username works as well",
			usernameConfig: func(c *schema.ExtensionConfig) {
				c.Credentials.Password.Identifier = true
			},
			expected: text.NewInfoNodeLabelGenerated("Username"),
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
			label, err := GetIdentifierLabelFromSchema(ctx, constructSchema(t, tc.emailConfig, tc.usernameConfig))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, label)
		})
	}
}
