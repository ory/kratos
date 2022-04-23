package credentialmigrate

import (
	_ "embed"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/x/snapshotx"
)

//go:embed stub/webauthn/v0.json
var webAuthnV0 []byte

//go:embed stub/webauthn/v1.json
var webAuthnV1 []byte

func TestUpgradeCredentials(t *testing.T) {
	t.Run("empty credentials", func(t *testing.T) {
		i := &identity.Identity{}

		err := UpgradeCredentials(i)
		require.NoError(t, err)
		wc := identity.WithCredentialsAndAdminMetadataInJSON(*i)
		snapshotx.SnapshotTExcept(t, &wc, nil)
	})

	identityID := uuid.FromStringOrNil("4d64fa08-20fc-450d-bebd-ebd7c7b6e249")
	t.Run("type=webauthn", func(t *testing.T) {
		t.Run("from=v0", func(t *testing.T) {
			i := &identity.Identity{
				ID: identityID,
				Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypeWebAuthn: {
						Identifiers: []string{"4d64fa08-20fc-450d-bebd-ebd7c7b6e249"},
						Type:        identity.CredentialsTypeWebAuthn,
						Version:     0,
						Config:      webAuthnV0,
					}},
			}

			require.NoError(t, UpgradeCredentials(i))
			wc := identity.WithCredentialsAndAdminMetadataInJSON(*i)
			snapshotx.SnapshotTExcept(t, &wc, nil)

			assert.Equal(t, 1, i.Credentials[identity.CredentialsTypeWebAuthn].Version)
		})

		t.Run("from=v1", func(t *testing.T) {
			i := &identity.Identity{
				ID: identityID,
				Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypeWebAuthn: {
						Type:    identity.CredentialsTypeWebAuthn,
						Version: 1,
						Config:  webAuthnV1,
					}},
			}

			require.NoError(t, UpgradeCredentials(i))
			wc := identity.WithCredentialsAndAdminMetadataInJSON(*i)
			snapshotx.SnapshotTExcept(t, &wc, nil)

			assert.Equal(t, 1, i.Credentials[identity.CredentialsTypeWebAuthn].Version)
		})
	})
}
