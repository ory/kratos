// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

// Since:
// - This is a separate test package
// - Go creates one executable per test package
// - Go global variables are isolated per process (i.e. executable)
// - This test uses its own database
// , it means that this test is independent from the others and cannot interfere with them.
// We do not need to clear the global maps at the start of the test: they are still pristine (unset).

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	id "github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence/sql/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/x"
	"github.com/ory/x/dbal"
	"github.com/ory/x/sqlxx"
)

func TestNonStandardCredentialTypes(t *testing.T) {
	_, reg := pkg.NewRegistryDefaultWithDSN(t, dbal.NewSQLiteTestDatabase(t))
	_, p := testhelpers.NewNetwork(t, t.Context(), reg.Persister())

	// Replace the content of the table with custom UUIDs.
	{
		require.NoError(t, p.GetConnection(t.Context()).RawQuery(`DELETE FROM identity_credential_types`, x.NewUUID()).Exec())

		// Test the case of empty credentials in the database, so that we can exercise the case of loading credentials failing,
		// and then subsequent loads succeed, and the service does not get suck.
		_, err := identity.FindIdentityCredentialsTypeByName(p.GetConnection(t.Context()), id.CredentialsTypeOIDC)
		require.Error(t, err)

		// Now fill the table with non-standard UUIDs.
		q := strings.Builder{}
		q.WriteString(`INSERT INTO identity_credential_types (id, name) VALUES `)
		for i, ct := range id.AllCredentialTypes {
			fmt.Fprintf(&q, `('%s', '%s')`, x.NewUUID(), ct)
			if i < len(id.AllCredentialTypes)-1 {
				q.WriteRune(',')
			}
		}
		require.NoError(t, p.GetConnection(t.Context()).RawQuery(q.String()).Exec())
	}

	// Valid credential types are found.
	for _, ct := range id.AllCredentialTypes {
		t.Run("type="+ct.String(), func(t *testing.T) {
			id, err := identity.FindIdentityCredentialsTypeByName(p.GetConnection(t.Context()), ct)
			require.NoError(t, err)

			require.NotEqual(t, uuid.Nil, id)
			name, err := identity.FindIdentityCredentialsTypeByID(p.GetConnection(t.Context()), id)
			require.NoError(t, err)

			assert.Equal(t, ct, name)
		})
	}

	// Invalid credential types are not found.
	_, err := identity.FindIdentityCredentialsTypeByName(p.GetConnection(t.Context()), "unknown")
	require.Error(t, err)

	_, err = identity.FindIdentityCredentialsTypeByID(p.GetConnection(t.Context()), x.NewUUID())
	require.Error(t, err)

	// Create an identity and find it by credential identifier.
	ctx := testhelpers.WithDefaultIdentitySchemaFromRaw(t.Context(), []byte(`{"$id":"test","type":"object"}`))
	email := "test-" + x.NewUUID().String() + "@example.com"
	i := id.NewIdentity("")
	i.SetCredentials(id.CredentialsTypePassword, id.Credentials{
		Type:        id.CredentialsTypePassword,
		Identifiers: []string{email},
		Config:      sqlxx.JSONRawMessage(`{}`),
	})
	require.NoError(t, p.CreateIdentity(ctx, i))

	actualIdentity, actualCreds, err := p.FindByCredentialsIdentifier(ctx, id.CredentialsTypePassword, email)
	require.NoError(t, err)
	assert.Equal(t, i.ID, actualIdentity.ID)
	assert.Equal(t, []string{email}, actualCreds.Identifiers)
}
