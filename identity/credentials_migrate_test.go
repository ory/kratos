// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	_ "embed"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/snapshotx"
)

//go:embed stub/webauthn/v0.json
var webAuthnV0 []byte

//go:embed stub/webauthn/v1.json
var webAuthnV1 []byte

func TestUpgradeCredentials(t *testing.T) {
	t.Run("empty credentials", func(t *testing.T) {
		i := &Identity{}

		err := UpgradeCredentials(i)
		require.NoError(t, err)
		wc := WithCredentialsAndAdminMetadataInJSON(*i)
		snapshotx.SnapshotTExcept(t, &wc, nil)
	})

	identityID := uuid.FromStringOrNil("4d64fa08-20fc-450d-bebd-ebd7c7b6e249")
	t.Run("type=webauthn", func(t *testing.T) {
		t.Run("from=v0", func(t *testing.T) {
			i := &Identity{
				ID: identityID,
				Credentials: map[CredentialsType]Credentials{
					CredentialsTypeWebAuthn: {
						Identifiers: []string{"4d64fa08-20fc-450d-bebd-ebd7c7b6e249"},
						Type:        CredentialsTypeWebAuthn,
						Version:     0,
						Config:      webAuthnV0,
					}},
			}

			require.NoError(t, UpgradeCredentials(i))
			wc := WithCredentialsAndAdminMetadataInJSON(*i)
			snapshotx.SnapshotTExcept(t, &wc, nil)

			assert.Equal(t, 1, i.Credentials[CredentialsTypeWebAuthn].Version)
		})

		t.Run("from=v1", func(t *testing.T) {
			i := &Identity{
				ID: identityID,
				Credentials: map[CredentialsType]Credentials{
					CredentialsTypeWebAuthn: {
						Type:    CredentialsTypeWebAuthn,
						Version: 1,
						Config:  webAuthnV1,
					}},
			}

			require.NoError(t, UpgradeCredentials(i))
			wc := WithCredentialsAndAdminMetadataInJSON(*i)
			snapshotx.SnapshotTExcept(t, &wc, nil)

			assert.Equal(t, 1, i.Credentials[CredentialsTypeWebAuthn].Version)
		})
	})
}
