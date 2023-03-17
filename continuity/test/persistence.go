// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/x"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
)

func TestPersister(ctx context.Context, p interface {
	persistence.Persister
	continuity.Persister
	identity.PrivilegedPool
}) func(t *testing.T) {
	var createIdentity = func(t *testing.T) *identity.Identity {
		id := identity.Identity{}
		require.NoError(t, p.CreateIdentity(ctx, &id))
		return &id
	}

	var createContainer = func(t *testing.T) continuity.Container {
		m := sqlxx.NullJSONRawMessage(`{"foo": "bar"}`)
		return continuity.Container{Name: "foo", IdentityID: x.PointToUUID(createIdentity(t).ID),
			ExpiresAt: time.Now().Add(time.Hour).UTC().Truncate(time.Second),
			Payload:   m,
		}
	}

	return func(t *testing.T) {
		nid, p := testhelpers.NewNetworkUnlessExisting(t, ctx, p)

		t.Run("case=not found", func(t *testing.T) {
			_, err := p.GetContinuitySession(ctx, x.NewUUID())
			require.EqualError(t, err, sqlcon.ErrNoRows.Error())
		})

		t.Run("case=save and find", func(t *testing.T) {
			expected := createContainer(t)
			require.NoError(t, p.SaveContinuitySession(ctx, &expected))

			actual, err := p.GetContinuitySession(ctx, expected.ID)
			require.NoError(t, err)
			actual.UpdatedAt, actual.CreatedAt, expected.UpdatedAt, expected.CreatedAt = time.Time{}, time.Time{}, time.Time{}, time.Time{}
			assert.EqualValues(t, expected.UTC(), actual.UTC())
		})

		t.Run("case=save and delete", func(t *testing.T) {
			expected := createContainer(t)

			require.NoError(t, p.SaveContinuitySession(ctx, &expected))
			require.NotEqual(t, uuid.Nil, expected.ID)
			require.NoError(t, p.DeleteContinuitySession(ctx, expected.ID))

			_, err := p.GetContinuitySession(ctx, expected.ID)
			require.EqualError(t, err, sqlcon.ErrNoRows.Error())
		})

		t.Run("case=network", func(t *testing.T) {
			id := x.NewUUID()

			t.Run("sets id on creation", func(t *testing.T) {
				expected := createContainer(t)
				expected.ID = id
				require.NoError(t, p.SaveContinuitySession(ctx, &expected))

				assert.EqualValues(t, id, expected.ID)
				assert.EqualValues(t, nid, expected.NID)

				actual, err := p.GetContinuitySession(ctx, id)
				require.NoError(t, err)
				assert.EqualValues(t, id, actual.ID)
				assert.EqualValues(t, nid, actual.NID)
			})

			t.Run("can not get on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, err := p.GetLoginFlow(ctx, id)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})

			t.Run("can not delete on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				err := p.DeleteContinuitySession(ctx, id)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})

		t.Run("case=cleanup", func(t *testing.T) {
			id := x.NewUUID()
			yesterday := time.Now().Add(-24 * time.Hour).UTC().Truncate(time.Second)
			m := sqlxx.NullJSONRawMessage(`{"foo": "bar"}`)
			expected := continuity.Container{Name: "foo", IdentityID: x.PointToUUID(createIdentity(t).ID),
				ExpiresAt: yesterday,
				Payload:   m,
			}
			expected.ID = id

			t.Run("can cleanup", func(t *testing.T) {
				require.NoError(t, p.SaveContinuitySession(ctx, &expected))

				assert.EqualValues(t, id, expected.ID)
				assert.EqualValues(t, nid, expected.NID)

				require.NoError(t, p.DeleteExpiredContinuitySessions(ctx, time.Now(), 5))

				_, err := p.GetContinuitySession(ctx, id)
				require.Error(t, err)
			})
		})
	}
}
