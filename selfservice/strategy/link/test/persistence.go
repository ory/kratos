package link

import (
	"context"
	"testing"
	"time"

	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/x/sqlcon"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/assertx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/x"
)

func TestPersister(ctx context.Context, conf *config.Config, p interface {
	persistence.Persister
}) func(t *testing.T) {
	return func(t *testing.T) {
		nid, p := testhelpers.NewNetworkUnlessExisting(t, ctx, p)

		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")
		conf.MustSet(config.ViperKeySecretsDefault, []string{"secret-a", "secret-b"})

		t.Run("token=recovery", func(t *testing.T) {
			t.Run("case=should error when the recovery token does not exist", func(t *testing.T) {
				_, err := p.UseRecoveryToken(ctx, "i-do-not-exist")
				require.Error(t, err)
			})

			newRecoveryToken := func(t *testing.T, email string) *link.RecoveryToken {
				var req recovery.Flow
				require.NoError(t, faker.FakeData(&req))
				require.NoError(t, p.CreateRecoveryFlow(ctx, &req))

				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))

				address := &identity.RecoveryAddress{Value: email, Via: identity.RecoveryAddressTypeEmail}
				i.RecoveryAddresses = append(i.RecoveryAddresses, *address)

				require.NoError(t, p.CreateIdentity(ctx, &i))

				return &link.RecoveryToken{Token: x.NewUUID().String(), FlowID: uuid.NullUUID{UUID: req.ID, Valid: true},
					RecoveryAddress: &i.RecoveryAddresses[0],
					ExpiresAt:       time.Now(),
					IssuedAt:        time.Now(),
				}
			}

			t.Run("case=should error when the recovery token does not exist", func(t *testing.T) {
				_, err := p.UseRecoveryToken(ctx, "i-do-not-exist")
				require.Error(t, err)
			})

			t.Run("case=should create a new recovery token", func(t *testing.T) {
				token := newRecoveryToken(t, "foo-user@ory.sh")
				require.NoError(t, p.CreateRecoveryToken(ctx, token))
			})

			t.Run("case=should create a recovery token and use it", func(t *testing.T) {
				expected := newRecoveryToken(t, "other-user@ory.sh")
				require.NoError(t, p.CreateRecoveryToken(ctx, expected))

				t.Run("not work on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.UseRecoveryToken(ctx, expected.Token)
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				actual, err := p.UseRecoveryToken(ctx, expected.Token)
				require.NoError(t, err)
				assertx.EqualAsJSONExcept(t, expected.RecoveryAddress, actual.RecoveryAddress, []string{"created_at", "updated_at"})
				assert.Equal(t, nid, actual.NID)
				assert.Equal(t, expected.RecoveryAddress.IdentityID, actual.RecoveryAddress.IdentityID)
				assert.NotEqual(t, expected.Token, actual.Token)
				assert.EqualValues(t, expected.FlowID, actual.FlowID)

				_, err = p.UseRecoveryToken(ctx, expected.Token)
				require.Error(t, err)
			})

		})

		t.Run("token=verification", func(t *testing.T) {
			t.Run("case=should error when the verification token does not exist", func(t *testing.T) {
				_, err := p.UseVerificationToken(ctx, "i-do-not-exist")
				require.Error(t, err)
			})

			newVerificationToken := func(t *testing.T, email string) *link.VerificationToken {
				var req verification.Flow
				require.NoError(t, faker.FakeData(&req))
				require.NoError(t, p.CreateVerificationFlow(ctx, &req))

				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))

				address := &identity.VerifiableAddress{Value: email, Via: identity.VerifiableAddressTypeEmail}
				i.VerifiableAddresses = append(i.VerifiableAddresses, *address)

				require.NoError(t, p.CreateIdentity(ctx, &i))
				return &link.VerificationToken{
					Token:             x.NewUUID().String(),
					FlowID:            uuid.NullUUID{UUID: req.ID, Valid: true},
					VerifiableAddress: &i.VerifiableAddresses[0],
					ExpiresAt:         time.Now(),
					IssuedAt:          time.Now(),
				}
			}

			t.Run("case=should error when the verification token does not exist", func(t *testing.T) {
				_, err := p.UseVerificationToken(ctx, "i-do-not-exist")
				require.Error(t, err)
			})

			t.Run("case=should create a new verification token", func(t *testing.T) {
				token := newVerificationToken(t, "foo-user@ory.sh")
				require.NoError(t, p.CreateVerificationToken(ctx, token))
			})

			t.Run("case=should create a verification token and use it", func(t *testing.T) {
				expected := newVerificationToken(t, "other-user@ory.sh")
				require.NoError(t, p.CreateVerificationToken(ctx, expected))

				t.Run("not work on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.UseVerificationToken(ctx, expected.Token)
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				actual, err := p.UseVerificationToken(ctx, expected.Token)
				require.NoError(t, err)
				assertx.EqualAsJSONExcept(t, expected.VerifiableAddress, actual.VerifiableAddress, []string{"created_at", "updated_at"})
				assert.Equal(t, nid, actual.NID)
				assert.Equal(t, expected.VerifiableAddress.IdentityID, actual.VerifiableAddress.IdentityID)
				assert.NotEqual(t, expected.Token, actual.Token)
				assert.EqualValues(t, expected.FlowID, actual.FlowID)

				_, err = p.UseVerificationToken(ctx, expected.Token)
				require.Error(t, err)
			})
		})
	}
}
