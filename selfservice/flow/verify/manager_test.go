package verify_test

import (
	"context"
	"testing"
	"time"

	"github.com/ory/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

	i := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
	require.NoError(t, reg.IdentityPool().CreateIdentity(context.Background(), i))

	t.Run("method=TrackAndSend", func(t *testing.T) {
		addresses := []verify.Address{
			*verify.MustNewEmailAddress("foo@ory.sh", i.ID, time.Minute),
			*verify.MustNewEmailAddress("bar@ory.sh", i.ID, time.Minute),
		}
		require.NoError(t, reg.VerificationManager().TrackAndSend(context.Background(), addresses))

		actual, err := reg.VerificationPersister().FindAddressByValue(context.Background(), verify.ViaEmail, "foo@ory.sh")
		require.NoError(t, err)
		assert.EqualValues(t, "foo@ory.sh", actual.Value)

		actual, err = reg.VerificationPersister().FindAddressByValue(context.Background(), verify.ViaEmail, "bar@ory.sh")
		require.NoError(t, err)
		assert.EqualValues(t, "bar@ory.sh", actual.Value)

		messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
		require.NoError(t, err)
		require.Len(t, messages, 2)

		assert.EqualValues(t, "foo@ory.sh", messages[0].Recipient)
		assert.Contains(t, messages[0].Subject, "Please verify")
		assert.EqualValues(t, "bar@ory.sh", messages[1].Recipient)
		assert.Contains(t, messages[1].Subject, "Please verify")
	})

	t.Run("method=SendCode", func(t *testing.T) {
		require.NoError(t, reg.VerificationManager().SendCode(context.Background(), verify.ViaEmail, "not-tracked@ory.sh"))

		messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
		require.NoError(t, err)

		require.Len(t, messages, 3)
		assert.EqualValues(t, "not-tracked@ory.sh", messages[2].Recipient)
		assert.Contains(t, messages[2].Subject, "tried to verify")
	})
}
