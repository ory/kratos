// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package batch

import (
	"context"
	"errors"
	"testing"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	"github.com/ory/x/dbal"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
)

func TestPersister(ctx context.Context, tracer *otelx.Tracer, p persistence.Persister) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("method=batch.Create", func(t *testing.T) {

			ident1 := identity.NewIdentity("")
			ident1.NID = p.NetworkID(ctx)
			ident2 := identity.NewIdentity("")
			ident2.NID = p.NetworkID(ctx)

			// Create two identities
			_ = p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
				conn := &TracerConnection{
					Tracer:     tracer,
					Connection: tx,
				}

				err := Create(ctx, conn, []*identity.Identity{ident1, ident2})
				require.NoError(t, err)

				return nil
			})

			require.NotEqual(t, uuid.Nil, ident1.ID)
			require.NotEqual(t, uuid.Nil, ident2.ID)

			// Create conflicting verifiable addresses
			addresses := []*identity.VerifiableAddress{{
				Value:      "foo.1@bar.de",
				IdentityID: ident1.ID,
				NID:        ident1.NID,
			}, {
				Value:      "foo.2@bar.de",
				IdentityID: ident1.ID,
				NID:        ident1.NID,
			}, {
				Value:      "conflict@bar.de",
				IdentityID: ident1.ID,
				NID:        ident1.NID,
			}, {
				Value:      "foo.3@bar.de",
				IdentityID: ident1.ID,
				NID:        ident1.NID,
			}, {
				Value:      "conflict@bar.de",
				IdentityID: ident1.ID,
				NID:        ident1.NID,
			}, {
				Value:      "foo.4@bar.de",
				IdentityID: ident1.ID,
				NID:        ident1.NID,
			}}

			t.Run("case=fails all without partial inserts", func(t *testing.T) {
				_ = p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
					conn := &TracerConnection{
						Tracer:     tracer,
						Connection: tx,
					}
					err := Create(ctx, conn, addresses)
					assert.ErrorIs(t, err, sqlcon.ErrUniqueViolation)
					if partial := new(PartialConflictError[identity.VerifiableAddress]); errors.As(err, &partial) {
						require.NoError(t, partial, "expected no partial error")
					}
					return err
				})
			})

			t.Run("case=return partial error with partial inserts", func(t *testing.T) {
				_ = p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
					conn := &TracerConnection{
						Tracer:     tracer,
						Connection: tx,
					}

					err := Create(ctx, conn, addresses, WithPartialInserts)
					assert.ErrorIs(t, err, sqlcon.ErrUniqueViolation)

					if conn.Connection.Dialect.Name() != dbal.DriverMySQL {
						// MySQL does not support partial errors.
						partialErr := new(PartialConflictError[identity.VerifiableAddress])
						require.ErrorAs(t, err, &partialErr)
						assert.Len(t, partialErr.Failed, 1)
					}

					return nil
				})
			})
		})
	}
}
