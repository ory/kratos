// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCredentialsOIDC(t *testing.T) {
	_, err := NewCredentialsOIDC(new(CredentialsOIDCEncryptedTokens), "", "not-empty", "")
	require.Error(t, err)
	_, err = NewCredentialsOIDC(new(CredentialsOIDCEncryptedTokens), "not-empty", "", "")
	require.Error(t, err)
	_, err = NewCredentialsOIDC(new(CredentialsOIDCEncryptedTokens), "not-empty", "not-empty", "")
	require.NoError(t, err)
}

func TestGetProvider(t *testing.T) {
	c := CredentialsOIDC{
		Providers: []CredentialsOIDCProvider{
			{
				Subject:  "user-a",
				Provider: "google",
			},
			{
				Subject:  "user-a",
				Provider: "github",
			},
		},
	}

	k, found := c.GetProvider("github", "user-a")
	require.True(t, found)
	require.Equal(t, 1, k)

	k, found = c.GetProvider("not-found", "user-a")
	require.False(t, found)
	require.Equal(t, -1, k)
}
