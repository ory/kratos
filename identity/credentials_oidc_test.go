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
