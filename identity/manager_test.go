package identity_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
)

func TestManager(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/manager.schema.json")
	viper.Set(configuration.ViperKeyURLsSelfPublic, "https://www.ory.sh/")
	viper.Set(configuration.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")

	t.Run("case=should fail to create because validation fails", func(t *testing.T) {
		i := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
		i.Traits = identity.Traits("{}")
		require.Error(t, reg.IdentityManager().Create(context.Background(), i))
	})

	checkExtensionFields := func(i *identity.Identity, expected string) func(*testing.T) {
		return func(t *testing.T) {
			require.Len(t, i.Addresses, 1)
			assert.EqualValues(t, expected, i.Addresses[0].Value)
			assert.EqualValues(t, identity.VerifiableAddressTypeEmail, i.Addresses[0].Via)

			require.NotNil(t, i.Credentials[identity.CredentialsTypePassword])
			assert.Equal(t, []string{expected}, i.Credentials[identity.CredentialsTypePassword].Identifiers)
		}
	}

	checkExtensionFieldsForIdentities := func(t *testing.T, expected string, original *identity.Identity) {
		fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), original.ID)
		require.NoError(t, err)
		for k, i := range []identity.Identity{*original, *fromStore} {
			t.Run(fmt.Sprintf("identity=%d", k), checkExtensionFields(&i, expected))
		}
	}

	t.Run("method=Create", func(t *testing.T) {
		t.Run("case=should create identity and track extension fields", func(t *testing.T) {
			original := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"foo@ory.sh"}`)
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))
			checkExtensionFieldsForIdentities(t, "foo@ory.sh", original)
		})

		t.Run("case=should expose validation errors with option", func(t *testing.T) {
			original := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"not an email"}`)
			err := reg.IdentityManager().Create(context.Background(), original, identity.ManagerExposeValidationErrors)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "\"not an email\" is not valid \"email\"")
		})

		t.Run("case=should not expose validation errors without option", func(t *testing.T) {
			original := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"not an email"}`)
			err := reg.IdentityManager().Create(context.Background(), original)
			require.Error(t, err)
			assert.NotContains(t, err.Error(), "\"not an email\" is not valid \"email\"")
		})
	})

	t.Run("method=Update", func(t *testing.T) {
		t.Run("case=should update identity and update extension fields", func(t *testing.T) {
			original := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"baz@ory.sh"}`)
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

			original.Traits = identity.Traits(`{"email":"bar@ory.sh"}`)
			require.NoError(t, reg.IdentityManager().Update(context.Background(), original, identity.ManagerAllowWriteProtectedTraits))

			checkExtensionFieldsForIdentities(t, "bar@ory.sh", original)
		})

		t.Run("case=should update identity and update extension fields", func(t *testing.T) {
			original := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "foo"}`)
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

			// These should all fail because they modify existing keys
			require.Error(t, reg.IdentityManager().UpdateTraits(context.Background(), original.ID, identity.Traits(`{"email":"not-baz@ory.sh","email_verify":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "foo"}`)))
			require.Error(t, reg.IdentityManager().UpdateTraits(context.Background(), original.ID, identity.Traits(`{"email":"baz@ory.sh","email_verify":"not-baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "foo"}`)))
			require.Error(t, reg.IdentityManager().UpdateTraits(context.Background(), original.ID, identity.Traits(`{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_creds":"not-baz@ory.sh","unprotected": "foo"}`)))

			require.NoError(t, reg.IdentityManager().UpdateTraits(context.Background(), original.ID, identity.Traits(`{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "bar"}`)))
			checkExtensionFieldsForIdentities(t, "baz@ory.sh", original)

			actual, err := reg.IdentityPool().GetIdentity(context.Background(), original.ID)
			require.NoError(t, err)
			assert.JSONEq(t, `{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "bar"}`, string(actual.Traits))
		})

		t.Run("case=should not update protected traits without option", func(t *testing.T) {
			original := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"email-update-1@ory.sh"}`)
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

			original.Traits = identity.Traits(`{"email":"email-update-2@ory.sh"}`)
			err := reg.IdentityManager().Update(context.Background(), original)
			require.Error(t, err)
			assert.Equal(t, identity.ErrProtectedFieldModified, errors.Cause(err))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), original.ID)
			require.NoError(t, err)
			// As UpdateTraits takes only the ID as a parameter it cannot update the identity in place.
			// That is why we only check the identity in the store.
			checkExtensionFields(fromStore, "email-update-1@ory.sh")(t)
		})
	})

	t.Run("method=UpdateTraits", func(t *testing.T) {
		t.Run("case=should update protected traits with option", func(t *testing.T) {
			original := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"email-updatetraits-1@ory.sh"}`)
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

			require.NoError(t, reg.IdentityManager().UpdateTraits(
				context.Background(), original.ID, identity.Traits(`{"email":"email-updatetraits-2@ory.sh"}`),
				identity.ManagerAllowWriteProtectedTraits))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), original.ID)
			require.NoError(t, err)
			// As UpdateTraits takes only the ID as a parameter it cannot update the identity in place.
			// That is why we only check the identity in the store.
			checkExtensionFields(fromStore, "email-updatetraits-2@ory.sh")(t)
		})

		t.Run("case=should not update protected traits without option", func(t *testing.T) {
			original := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
			original.Traits = identity.Traits(`{"email":"email-updatetraits-1@ory.sh"}`)
			require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

			err := reg.IdentityManager().UpdateTraits(
				context.Background(), original.ID, identity.Traits(`{"email":"email-updatetraits-2@ory.sh"}`))
			require.Error(t, err)
			assert.Equal(t, identity.ErrProtectedFieldModified, errors.Cause(err))

			fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), original.ID)
			require.NoError(t, err)
			// As UpdateTraits takes only the ID as a parameter it cannot update the identity in place.
			// That is why we only check the identity in the store.
			checkExtensionFields(fromStore, "email-updatetraits-1@ory.sh")(t)
		})
	})

	t.Run("method=RefreshVerifyAddress", func(t *testing.T) {
		original := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
		original.Traits = identity.Traits(`{"email":"verifyme@ory.sh"}`)
		require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

		address, err := reg.IdentityPool().FindAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, "verifyme@ory.sh")
		require.NoError(t, err)

		pc := address.Code
		ea := address.ExpiresAt
		require.NoError(t, reg.IdentityManager().RefreshVerifyAddress(context.Background(), address))
		assert.NotEqual(t, pc, address.Code)
		assert.NotEqual(t, ea, address.ExpiresAt)

		fromStore, err := reg.IdentityPool().GetIdentity(context.Background(), original.ID)
		require.NoError(t, err)
		assert.NotEqual(t, pc, fromStore.Addresses[0].Code)
	})
}
