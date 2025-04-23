package identity_test

import (
	"context"
	"github.com/ory/kratos/internal"
	"testing"
)

func TestImportCredentialsPassword(t *testing.T) {
	t.Parallel()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	ctx := context.Background()

	// Create the identity handler
	h := reg.IdentityHandler()

	testCases := []struct {
		name          string
		setupIdentity func() *identity.Identity
		credentials   *identity.AdminIdentityImportCredentialsPassword
		verify        func(t *testing.T, i *identity.Identity)
	}{
		{
			name: "import clear text password",
			setupIdentity: func() *identity.Identity {
				return identity.NewIdentity(conf.DefaultIdentityTraitsSchemaID())
			},
			credentials: &identity.AdminIdentityImportCredentialsPassword{
				Config: identity.AdminIdentityImportCredentialsPasswordConfig{
					Password: "password123",
				},
			},
			verify: func(t *testing.T, i *identity.Identity) {
				creds, ok := i.GetCredentials(identity.CredentialsTypePassword)
				require.True(t, ok, "password credentials should be set")

				var config identity.CredentialsPassword
				require.NoError(t, json.Unmarshal(creds.Config, &config))

				assert.NotEmpty(t, config.HashedPassword)
				assert.False(t, config.UsePasswordMigrationHook)
			},
		},
		{
			name: "import hashed password",
			setupIdentity: func() *identity.Identity {
				return identity.NewIdentity(conf.DefaultIdentityTraitsSchemaID())
			},
			credentials: &identity.AdminIdentityImportCredentialsPassword{
				Config: identity.AdminIdentityImportCredentialsPasswordConfig{
					HashedPassword: "$2a$10$JCU0ELjU1TCnFbV2jLi7huEQjVWZG0HzXQq/BWZyO30XR6DJxwN72",
				},
			},
			verify: func(t *testing.T, i *identity.Identity) {
				creds, ok := i.GetCredentials(identity.CredentialsTypePassword)
				require.True(t, ok, "password credentials should be set")

				var config identity.CredentialsPassword
				require.NoError(t, json.Unmarshal(creds.Config, &config))

				assert.Equal(t, "$2a$10$JCU0ELjU1TCnFbV2jLi7huEQjVWZG0HzXQq/BWZyO30XR6DJxwN72", config.HashedPassword)
				assert.False(t, config.UsePasswordMigrationHook)
			},
		},
		{
			name: "import with password migration hook",
			setupIdentity: func() *identity.Identity {
				return identity.NewIdentity(conf.DefaultIdentityTraitsSchemaID())
			},
			credentials: &identity.AdminIdentityImportCredentialsPassword{
				Config: identity.AdminIdentityImportCredentialsPasswordConfig{
					UsePasswordMigrationHook: true,
				},
			},
			verify: func(t *testing.T, i *identity.Identity) {
				creds, ok := i.GetCredentials(identity.CredentialsTypePassword)
				require.True(t, ok, "password credentials should be set")

				var config identity.CredentialsPassword
				require.NoError(t, json.Unmarshal(creds.Config, &config))

				assert.True(t, config.UsePasswordMigrationHook)
				assert.Empty(t, config.HashedPassword)
			},
		},
		{
			name: "update existing password credential",
			setupIdentity: func() *identity.Identity {
				i := identity.NewIdentity(conf.DefaultIdentityTraitsSchemaID())
				_ = i.SetCredentialsWithConfig(
					identity.CredentialsTypePassword,
					identity.Credentials{},
					identity.CredentialsPassword{
						HashedPassword: "old-hash",
					},
				)
				return i
			},
			credentials: &identity.AdminIdentityImportCredentialsPassword{
				Config: identity.AdminIdentityImportCredentialsPasswordConfig{
					Password: "new-password",
				},
			},
			verify: func(t *testing.T, i *identity.Identity) {
				creds, ok := i.GetCredentials(identity.CredentialsTypePassword)
				require.True(t, ok, "password credentials should be set")

				var config identity.CredentialsPassword
				require.NoError(t, json.Unmarshal(creds.Config, &config))

				assert.NotEmpty(t, config.HashedPassword)
				assert.NotEqual(t, "old-hash", config.HashedPassword)
				assert.False(t, config.UsePasswordMigrationHook)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up a fresh identity for each test
			i := tc.setupIdentity()

			// Perform the import
			err := h.ImportCredentialsPassword(ctx, i, tc.credentials)
			require.NoError(t, err)

			// Verify credential was set correctly
			tc.verify(t, i)

			// Take a snapshot of the credentials
			snapshotx.SnapshotT(t, i.Credentials)
		})
	}
}
