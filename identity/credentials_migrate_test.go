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
				}},
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
}
