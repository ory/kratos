package code

import (
	"context"
	"testing"
	"time"

	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/x/randx"
	"github.com/ory/x/sqlcon"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/x"
)

func TestPersister(ctx context.Context, conf *config.Config, p interface {
	persistence.Persister
}) func(t *testing.T) {
	return func(t *testing.T) {
		nid, p := testhelpers.NewNetworkUnlessExisting(t, ctx, p)

		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
		conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"secret-a", "secret-b"})

		t.Run("code=recovery", func(t *testing.T) {

			newRecoveryCode := func(t *testing.T, email string) (*code.RecoveryCode, *recovery.Flow) {
				var f recovery.Flow
				require.NoError(t, faker.FakeData(&f))
				require.NoError(t, p.CreateRecoveryFlow(ctx, &f))

				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))

				address := &identity.RecoveryAddress{Value: email, Via: identity.RecoveryAddressTypeEmail, IdentityID: i.ID}
				i.RecoveryAddresses = append(i.RecoveryAddresses, *address)

				require.NoError(t, p.CreateIdentity(ctx, &i))

				return &code.RecoveryCode{
					Code:            string(randx.MustString(8, randx.Numeric)),
					FlowID:          f.ID,
					RecoveryAddress: &i.RecoveryAddresses[0],
					ExpiresAt:       time.Now(),
					IssuedAt:        time.Now(),
					IdentityID:      i.ID,
				}, &f
			}

			t.Run("case=should error when the recovery token does not exist", func(t *testing.T) {
				_, err := p.UseRecoveryCode(ctx, x.NewUUID(), "i-do-not-exist")
				require.Error(t, err)
			})

			t.Run("case=should create a new recovery token", func(t *testing.T) {
				token, _ := newRecoveryCode(t, "foo-user@ory.sh")
				require.NoError(t, p.CreateRecoveryCode(ctx, token))
			})

			t.Run("case=should create a recovery token and use it", func(t *testing.T) {
				expected, f := newRecoveryCode(t, "other-user@ory.sh")
				require.NoError(t, p.CreateRecoveryCode(ctx, expected))

				t.Run("not work on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.UseRecoveryCode(ctx, f.ID, expected.Code)
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				actual, err := p.UseRecoveryCode(ctx, f.ID, expected.Code)
				require.NoError(t, err)
				assert.Equal(t, nid, actual.NID)
				assert.Equal(t, expected.IdentityID, actual.IdentityID)
				assert.NotEqual(t, expected.Code, actual.Code)
				assert.EqualValues(t, expected.FlowID, actual.FlowID)

				_, err = p.UseRecoveryCode(ctx, f.ID, expected.Code)
				require.Error(t, err)
			})

		})
	}
}
