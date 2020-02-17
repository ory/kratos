package verify_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/verify"
)

func TestManager(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/extension/schema.json")
	viper.Set(configuration.ViperKeyURLsSelfPublic, "https://www.ory.sh/")
	viper.Set(configuration.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")

	t.Run("method=SendCode", func(t *testing.T) {
		i := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)

		address, err := identity.NewVerifiableEmailAddress("tracked@ory.sh", i.ID, time.Minute)
		require.NoError(t, err)

		i.Addresses = []identity.VerifiableAddress{*address}
		i.Traits = identity.Traits("{}")
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

		address, err = reg.VerificationSender().SendCode(context.Background(), address.Via, address.Value)
		require.NoError(t, err)

		_, err = reg.VerificationSender().SendCode(context.Background(), address.Via, "not-tracked@ory.sh")
		require.EqualError(t, err, verify.ErrUnknownAddress.Error())

		messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
		require.NoError(t, err)
		require.Len(t, messages, 2)

		assert.EqualValues(t, address.Value, messages[0].Recipient)
		assert.Contains(t, messages[0].Subject, "Please verify")

		assert.Contains(t, messages[0].Body, address.Code)
		fromStore, err := reg.Persister().GetIdentity(context.Background(), i.ID)
		require.NoError(t, err)
		assert.Contains(t, messages[0].Body, fromStore.Addresses[0].Code)

		assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
		assert.Contains(t, messages[1].Subject, "tried to verify")
	})
}
