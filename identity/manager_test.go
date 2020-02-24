package identity_test

import (
	"context"
	"fmt"
	"testing"

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

	checkExtensionFields := func(t *testing.T, expected string, original *identity.Identity) {
		fromStore, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), original.ID)
		require.NoError(t, err)
		for k, i := range []identity.Identity{*original, *fromStore} {
			t.Run(fmt.Sprintf("identity=%d", k), func(t *testing.T) {
				require.Len(t, i.Addresses, 1)
				assert.EqualValues(t, expected, i.Addresses[0].Value)
				assert.EqualValues(t, identity.VerifiableAddressTypeEmail, i.Addresses[0].Via)

				require.NotNil(t, i.Credentials[identity.CredentialsTypePassword])
				assert.Equal(t, []string{expected}, i.Credentials[identity.CredentialsTypePassword].Identifiers)
			})
		}
	}

	t.Run("method=Create/case=should create identity and track extension fields", func(t *testing.T) {
		original := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
		original.Traits = identity.Traits(`{"email":"foo@ory.sh"}`)
		require.NoError(t, reg.IdentityManager().Create(context.Background(), original))
		checkExtensionFields(t, "foo@ory.sh", original)
	})

	t.Run("method=Update/case=should update identity and update extension fields", func(t *testing.T) {
		original := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
		original.Traits = identity.Traits(`{"email":"baz@ory.sh"}`)
		require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

		original.Traits = identity.Traits(`{"email":"bar@ory.sh"}`)
		require.NoError(t, reg.IdentityManager().Update(context.Background(), original))

		checkExtensionFields(t, "bar@ory.sh", original)
	})

	t.Run("method=Update/case=should update identity and update extension fields", func(t *testing.T) {
		original := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
		original.Traits = identity.Traits(`{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "foo"}`)
		require.NoError(t, reg.IdentityManager().Create(context.Background(), original))

		// These should all fail because they modify existing keys
		require.Error(t, reg.IdentityManager().UpdateTraits(context.Background(), original.ID, identity.Traits(`{"email":"not-baz@ory.sh","email_verify":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "foo"}`)))
		require.Error(t, reg.IdentityManager().UpdateTraits(context.Background(), original.ID, identity.Traits(`{"email":"baz@ory.sh","email_verify":"not-baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "foo"}`)))
		require.Error(t, reg.IdentityManager().UpdateTraits(context.Background(), original.ID, identity.Traits(`{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_creds":"not-baz@ory.sh","unprotected": "foo"}`)))

		require.NoError(t, reg.IdentityManager().UpdateTraits(context.Background(), original.ID, identity.Traits(`{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "bar"}`)))
		checkExtensionFields(t, "baz@ory.sh", original)

		actual, err := reg.IdentityPool().GetIdentity(context.Background(), original.ID)
		require.NoError(t, err)
		assert.JSONEq(t, `{"email":"baz@ory.sh","email_verify":"baz@ory.sh","email_creds":"baz@ory.sh","unprotected": "bar"}`, string(actual.Traits))
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
