package recoverytoken

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"
	"github.com/ory/x/assertx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/x"
)

func TestPersister(p interface {
	Persister
	recovery.RequestPersister
	identity.PrivilegedPool
}) func(t *testing.T) {
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")
	viper.Set(configuration.ViperKeySecretsDefault, []string{"secret-a", "secret-b"})
	return func(t *testing.T) {
		t.Run("case=should error when the recovery token does not exist", func(t *testing.T) {
			_, err := p.UseRecoveryToken(context.Background(), "i-do-not-exist")
			require.Error(t, err)
		})

		newRecoveryToken := func(t *testing.T, email string) *Token {
			var req recovery.Flow
			require.NoError(t, faker.FakeData(&req))
			require.NoError(t, p.CreateRecoveryRequest(context.Background(), &req))

			var i identity.Identity
			require.NoError(t, faker.FakeData(&i))

			address := &identity.RecoveryAddress{Value: email, Via: identity.RecoveryAddressTypeEmail}
			i.RecoveryAddresses = append(i.RecoveryAddresses, *address)

			require.NoError(t, p.CreateIdentity(context.Background(), &i))

			return &Token{Token: x.NewUUID().String(), Request: &req, RecoveryAddress: &i.RecoveryAddresses[0]}
		}

		t.Run("case=should error when the recovery token does not exist", func(t *testing.T) {
			_, err := p.UseRecoveryToken(context.Background(), "i-do-not-exist")
			require.Error(t, err)
		})

		t.Run("case=should create a new recovery token", func(t *testing.T) {
			token := newRecoveryToken(t, "foo-user@ory.sh")
			require.NoError(t, p.CreateRecoveryToken(context.Background(), token))
		})

		t.Run("case=should create a recovery token and use it", func(t *testing.T) {
			expected := newRecoveryToken(t, "other-user@ory.sh")
			require.NoError(t, p.CreateRecoveryToken(context.Background(), expected))
			actual, err := p.UseRecoveryToken(context.Background(), expected.Token)
			require.NoError(t, err)
			assertx.EqualAsJSON(t, expected.RecoveryAddress, actual.RecoveryAddress)
			assertx.EqualAsJSON(t, expected.RecoveryAddress, actual.RecoveryAddress)
			assert.Equal(t, expected.RecoveryAddress.IdentityID, actual.RecoveryAddress.IdentityID)
			assert.NotEqual(t, expected.Token, actual.Token)

			_, err = p.UseRecoveryToken(context.Background(), expected.Token)
			require.Error(t, err)
		})
	}
}
