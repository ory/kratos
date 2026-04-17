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

	run := func(t *testing.T, identifiers []string, config string, version int, credentialsType CredentialsType, expectedVersion int) {
		if identifiers == nil {
			identifiers = []string{"hi@example.org"}
		}
		i := &Identity{
			ID: uuid.FromStringOrNil("4d64fa08-20fc-450d-bebd-ebd7c7b6e249"),
			Credentials: map[CredentialsType]Credentials{
				credentialsType: {
					Identifiers: identifiers,
					Type:        credentialsType,
					Version:     version,
					Config:      []byte(config),
				},
			},
		}

		require.NoError(t, UpgradeCredentials(i))
		wc := WithCredentialsAndAdminMetadataInJSON(*i)
		snapshotx.SnapshotT(t, &wc)
		assert.Equal(t, expectedVersion, i.Credentials[credentialsType].Version)
	}

	t.Run("type=code", func(t *testing.T) {
		t.Run("from=v0 with email empty space value", func(t *testing.T) {
			t.Run("with one identifier", func(t *testing.T) {
				run(t, nil, `{"address_type": "email                               ", "used_at": {"Time": "0001-01-01T00:00:00Z", "Valid": false}}`, 0, CredentialsTypeCodeAuth, 1)
			})

			t.Run("with two identifiers", func(t *testing.T) {
				run(t, []string{"foo@example.org", "bar@example.org"}, `{"address_type": "email                               ", "used_at": {"Time": "0001-01-01T00:00:00Z", "Valid": false}}`, 0, CredentialsTypeCodeAuth, 1)
			})
		})

		t.Run("from=v0 with empty value", func(t *testing.T) {
			run(t, nil, `{"address_type": "", "used_at": {"Time": "0001-01-01T00:00:00Z", "Valid": false}}`, 0, CredentialsTypeCodeAuth, 1)
		})

		t.Run("from=v0 with correct value", func(t *testing.T) {
			run(t, nil, `{"address_type": "email", "used_at": {"Time": "0001-01-01T00:00:00Z", "Valid": false}}`, 0, CredentialsTypeCodeAuth, 1)
		})

		t.Run("from=v0 with unknown value", func(t *testing.T) {
			run(t, nil, `{"address_type": "other", "used_at": {"Time": "0001-01-01T00:00:00Z", "Valid": false}}`, 0, CredentialsTypeCodeAuth, 1)
		})

		t.Run("from=v2 with empty value", func(t *testing.T) {
			run(t, []string{"foo@example.org", "+12341234"}, `{"addresses": [{"address":"foo@example.org","channel":"email"},{"address":"+12341234","channel":"sms"}]}`, 1, CredentialsTypeCodeAuth, 1)
		})
	})

	t.Run("type=webauthn", func(t *testing.T) {
		t.Run("from=v0", func(t *testing.T) {
			run(t, []string{"4d64fa08-20fc-450d-bebd-ebd7c7b6e249"}, string(webAuthnV0), 0, CredentialsTypeWebAuthn, 1)
		})

		t.Run("from=v1", func(t *testing.T) {
			run(t, []string{}, string(webAuthnV1), 1, CredentialsTypeWebAuthn, 1)
		})
	})

	t.Run("type=password", func(t *testing.T) {
		t.Run("from=v0 with phone number", func(t *testing.T) {
			i := &Identity{
				ID: uuid.FromStringOrNil("4d64fa08-20fc-450d-bebd-ebd7c7b6e249"),
				Credentials: map[CredentialsType]Credentials{
					CredentialsTypePassword: {
						Identifiers: []string{"+49 176 671 11 638"},
						Type:        CredentialsTypePassword,
						Version:     0,
						Config:      []byte(`{}`),
					},
				},
			}

			require.NoError(t, UpgradeCredentials(i))
			wc := WithCredentialsAndAdminMetadataInJSON(*i)
			snapshotx.SnapshotT(t, &wc)
			assert.Equal(t, 1, i.Credentials[CredentialsTypePassword].Version)

			// Verify only normalized identifier is present
			identifiers := i.Credentials[CredentialsTypePassword].Identifiers
			assert.Contains(t, identifiers, "+4917667111638", "Should contain E.164 normalized phone number")
		})

		t.Run("from=v0 with email", func(t *testing.T) {
			i := &Identity{
				ID: uuid.FromStringOrNil("4d64fa08-20fc-450d-bebd-ebd7c7b6e249"),
				Credentials: map[CredentialsType]Credentials{
					CredentialsTypePassword: {
						Identifiers: []string{"  Test@Example.ORG  "},
						Type:        CredentialsTypePassword,
						Version:     0,
						Config:      []byte(`{}`),
					},
				},
			}

			require.NoError(t, UpgradeCredentials(i))
			wc := WithCredentialsAndAdminMetadataInJSON(*i)
			snapshotx.SnapshotT(t, &wc)
			assert.Equal(t, 1, i.Credentials[CredentialsTypePassword].Version)

			// Verify only normalized identifier is present
			identifiers := i.Credentials[CredentialsTypePassword].Identifiers
			assert.Contains(t, identifiers, "test@example.org", "Should contain normalized email")
		})

		t.Run("from=v0 with already normalized phone", func(t *testing.T) {
			i := &Identity{
				ID: uuid.FromStringOrNil("4d64fa08-20fc-450d-bebd-ebd7c7b6e249"),
				Credentials: map[CredentialsType]Credentials{
					CredentialsTypePassword: {
						Identifiers: []string{"+4917667111638"},
						Type:        CredentialsTypePassword,
						Version:     0,
						Config:      []byte(`{}`),
					},
				},
			}

			require.NoError(t, UpgradeCredentials(i))
			wc := WithCredentialsAndAdminMetadataInJSON(*i)
			snapshotx.SnapshotT(t, &wc)
			assert.Equal(t, 1, i.Credentials[CredentialsTypePassword].Version)

			// Verify no duplicate is created when already normalized
			identifiers := i.Credentials[CredentialsTypePassword].Identifiers
			assert.Len(t, identifiers, 1, "Should not create duplicate when already normalized")
			assert.Contains(t, identifiers, "+4917667111638", "Should contain the normalized phone number")
		})

		t.Run("from=v0 with mixed identifiers", func(t *testing.T) {
			i := &Identity{
				ID: uuid.FromStringOrNil("4d64fa08-20fc-450d-bebd-ebd7c7b6e249"),
				Credentials: map[CredentialsType]Credentials{
					CredentialsTypePassword: {
						Identifiers: []string{"+49 176 671 11 638", "  Test@Example.ORG  ", "username123"},
						Type:        CredentialsTypePassword,
						Version:     0,
						Config:      []byte(`{}`),
					},
				},
			}

			require.NoError(t, UpgradeCredentials(i))
			wc := WithCredentialsAndAdminMetadataInJSON(*i)
			snapshotx.SnapshotT(t, &wc)
			assert.Equal(t, 1, i.Credentials[CredentialsTypePassword].Version)

			// Verify only normalized identifiers are present
			identifiers := i.Credentials[CredentialsTypePassword].Identifiers
			assert.Contains(t, identifiers, "+4917667111638", "Should contain normalized phone")
			assert.Contains(t, identifiers, "test@example.org", "Should contain normalized email")
			assert.Contains(t, identifiers, "username123", "Should contain original username")
		})

		t.Run("from=v1", func(t *testing.T) {
			run(t, []string{"+4917667111638"}, `{}`, 1, CredentialsTypePassword, 1)
		})
	})
}
