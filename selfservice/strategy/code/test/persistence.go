// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"testing"
	"time"

	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/x/randx"

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

			newRecoveryCodeDTO := func(t *testing.T, email string) (*code.CreateRecoveryCodeParams, *recovery.Flow, *identity.RecoveryAddress) {
				var f recovery.Flow
				require.NoError(t, faker.FakeData(&f))
				require.NoError(t, p.CreateRecoveryFlow(ctx, &f))

				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))

				address := &identity.RecoveryAddress{Value: email, Via: identity.RecoveryAddressTypeEmail, IdentityID: i.ID}
				i.RecoveryAddresses = append(i.RecoveryAddresses, *address)

				require.NoError(t, p.CreateIdentity(ctx, &i))

				return &code.CreateRecoveryCodeParams{
					RawCode:         string(randx.MustString(8, randx.Numeric)),
					FlowID:          f.ID,
					RecoveryAddress: &i.RecoveryAddresses[0],
					ExpiresIn:       time.Minute,
					IdentityID:      i.ID,
				}, &f, &i.RecoveryAddresses[0]
			}

			t.Run("case=should error when the recovery token does not exist", func(t *testing.T) {
				_, err := p.UseRecoveryCode(ctx, x.NewUUID(), "i-do-not-exist")
				require.Error(t, err)
			})

			t.Run("case=should create a new recovery code", func(t *testing.T) {
				dto, f, a := newRecoveryCodeDTO(t, "foo-user@ory.sh")
				rCode, err := p.CreateRecoveryCode(ctx, dto)
				require.NoError(t, err)
				assert.Equal(t, f.ID, rCode.FlowID)
				assert.Equal(t, dto.IdentityID, rCode.IdentityID)
				require.True(t, rCode.RecoveryAddressID.Valid)
				assert.Equal(t, a.ID, rCode.RecoveryAddressID.UUID)
				assert.Equal(t, a.ID, rCode.RecoveryAddress.ID)
			})

			t.Run("case=should create a recovery code and use it", func(t *testing.T) {
				dto, f, _ := newRecoveryCodeDTO(t, "other-user@ory.sh")
				_, err := p.CreateRecoveryCode(ctx, dto)
				require.NoError(t, err)

				t.Run("not work on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.UseRecoveryCode(ctx, f.ID, dto.RawCode)
					require.ErrorIs(t, err, code.ErrCodeNotFound)
				})

				actual, err := p.UseRecoveryCode(ctx, f.ID, dto.RawCode)
				require.NoError(t, err)
				assert.Equal(t, nid, actual.NID)
				assert.Equal(t, dto.IdentityID, actual.IdentityID)
				assert.NotEqual(t, dto.RawCode, actual.CodeHMAC)
				assert.EqualValues(t, f.ID, actual.FlowID)

				_, err = p.UseRecoveryCode(ctx, f.ID, dto.RawCode)
				require.ErrorIs(t, err, code.ErrCodeAlreadyUsed)
			})

			t.Run("case=should not be able to use expired codes", func(t *testing.T) {
				dto, f, _ := newRecoveryCodeDTO(t, "expired-code@ory.sh")
				dto.ExpiresIn = -time.Hour
				_, err := p.CreateRecoveryCode(ctx, dto)
				require.NoError(t, err)

				_, err = p.UseRecoveryCode(ctx, f.ID, dto.RawCode)
				assert.Error(t, err)
			})

			t.Run("case=should increment flow submit count and fail after too many tries", func(t *testing.T) {
				dto, f, _ := newRecoveryCodeDTO(t, "submit-count@ory.sh")
				_, err := p.CreateRecoveryCode(ctx, dto)
				require.NoError(t, err)

				for i := 1; i <= 5; i++ {
					_, err = p.UseRecoveryCode(ctx, f.ID, "i-do-not-exist")
					require.Error(t, err)
				}

				_, err = p.UseRecoveryCode(ctx, f.ID, "i-do-not-exist")
				require.ErrorIs(t, err, code.ErrCodeSubmittedTooOften)

				// Submit again, just to be sure
				_, err = p.UseRecoveryCode(ctx, f.ID, "i-do-not-exist")
				require.ErrorIs(t, err, code.ErrCodeSubmittedTooOften)
			})

			t.Run("case=should delete codes of flow", func(t *testing.T) {
				dto, f, _ := newRecoveryCodeDTO(t, testhelpers.RandomEmail())
				for i := 0; i < 10; i++ {
					dto.RawCode = string(randx.MustString(8, randx.Numeric))
					_, err := p.CreateRecoveryCode(ctx, dto)
					require.NoError(t, err)
				}

				count, err := p.GetConnection(ctx).Where("selfservice_recovery_flow_id = ?", f.ID).Count(&code.RecoveryCode{})
				require.NoError(t, err)
				require.Equal(t, 10, count)

				err = p.DeleteRecoveryCodesOfFlow(ctx, f.ID)
				require.NoError(t, err)

				// Count again, should be 0
				count, err = p.GetConnection(ctx).Where("selfservice_recovery_flow_id = ?", f.ID).Count(&code.RecoveryCode{})
				require.NoError(t, err)
				require.Equal(t, 0, count)

			})
		})
	}
}
