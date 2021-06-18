package test

import (
	"context"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/randx"
	"github.com/ory/x/sqlcon"
)

func TestPersister(ctx context.Context, conf *config.Config, p interface {
	persistence.Persister
}) func(t *testing.T) {
	return func(t *testing.T) {
		_, p := testhelpers.NewNetworkUnlessExisting(t, ctx, p)

		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

		t.Run("case=not found", func(t *testing.T) {
			_, err := p.GetSession(ctx, x.NewUUID())
			require.Error(t, err)
		})

		t.Run("case=create session", func(t *testing.T) {
			var expected session.Session
			require.NoError(t, faker.FakeData(&expected))
			expected.Active = true
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))

			assert.Equal(t, uuid.Nil, expected.ID)
			require.NoError(t, p.CreateSession(ctx, &expected))
			assert.NotEqual(t, uuid.Nil, expected.ID)

			check := func(actual *session.Session, err error) {
				require.NoError(t, err)
				assert.Equal(t, expected.Identity.ID, actual.Identity.ID)
				assert.NotEmpty(t, actual.Identity.SchemaURL)
				assert.NotEmpty(t, actual.Identity.SchemaID)
				assert.Equal(t, expected.ID, actual.ID)
				assert.Equal(t, expected.Active, actual.Active)
				assert.Equal(t, expected.Token, actual.Token)
				assert.EqualValues(t, expected.ExpiresAt.Unix(), actual.ExpiresAt.Unix())
				assert.Equal(t, expected.AuthenticatedAt.Unix(), actual.AuthenticatedAt.Unix())
				assert.Equal(t, expected.IssuedAt.Unix(), actual.IssuedAt.Unix())
			}

			t.Run("method=get by id", func(t *testing.T) {
				check(p.GetSession(ctx, expected.ID))

				t.Run("on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.GetSession(ctx, expected.ID)
					assert.ErrorIs(t, err, sqlcon.ErrNoRows)
				})
			})

			t.Run("method=get by token", func(t *testing.T) {
				check(p.GetSessionByToken(ctx, expected.Token))

				t.Run("on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.GetSessionByToken(ctx, expected.Token)
					assert.ErrorIs(t, err, sqlcon.ErrNoRows)
				})
			})
		})

		t.Run("case=delete session", func(t *testing.T) {
			var expected session.Session
			require.NoError(t, faker.FakeData(&expected))
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.CreateSession(ctx, &expected))

			t.Run("on another network", func(t *testing.T) {
				_, other := testhelpers.NewNetwork(t, ctx, p)
				err := other.DeleteSession(ctx, expected.ID)
				assert.ErrorIs(t, err, sqlcon.ErrNoRows)

				_, err = p.GetSession(ctx, expected.ID)
				assert.NoError(t, err)
			})

			require.NoError(t, p.DeleteSession(ctx, expected.ID))
			_, err := p.GetSession(ctx, expected.ID)
			assert.ErrorIs(t, err, sqlcon.ErrNoRows)
		})

		t.Run("case=delete session by token", func(t *testing.T) {
			var expected session.Session
			require.NoError(t, faker.FakeData(&expected))
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.CreateSession(ctx, &expected))

			t.Run("on another network", func(t *testing.T) {
				_, other := testhelpers.NewNetwork(t, ctx, p)
				err := other.DeleteSessionByToken(ctx, expected.Token)
				assert.ErrorIs(t, err, sqlcon.ErrNoRows)

				_, err = p.GetSessionByToken(ctx, expected.Token)
				assert.NoError(t, err)
			})

			require.NoError(t, p.DeleteSessionByToken(ctx, expected.Token))
			_, err := p.GetSession(ctx, expected.ID)
			require.Error(t, err)
		})

		t.Run("case=revoke session by token", func(t *testing.T) {
			var expected session.Session
			require.NoError(t, faker.FakeData(&expected))
			expected.Active = true
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.CreateSession(ctx, &expected))

			actual, err := p.GetSession(ctx, expected.ID)
			require.NoError(t, err)
			assert.True(t, actual.Active)

			t.Run("on another network", func(t *testing.T) {
				_, other := testhelpers.NewNetwork(t, ctx, p)
				err := other.RevokeSessionByToken(ctx, expected.Token)
				assert.ErrorIs(t, err, sqlcon.ErrNoRows)

				actual, err = p.GetSession(ctx, expected.ID)
				require.NoError(t, err)
				assert.True(t, actual.Active)
			})

			require.NoError(t, p.RevokeSessionByToken(ctx, expected.Token))

			actual, err = p.GetSession(ctx, expected.ID)
			require.NoError(t, err)
			assert.False(t, actual.Active)
		})

		t.Run("case=delete session for", func(t *testing.T) {
			var expected1 session.Session
			var expected2 session.Session
			require.NoError(t, faker.FakeData(&expected1))
			require.NoError(t, p.CreateIdentity(ctx, expected1.Identity))

			require.NoError(t, p.CreateSession(ctx, &expected1))

			require.NoError(t, faker.FakeData(&expected2))
			expected2.Identity = expected1.Identity
			expected2.IdentityID = expected1.IdentityID
			require.NoError(t, p.CreateSession(ctx, &expected2))

			t.Run("on another network", func(t *testing.T) {
				_, other := testhelpers.NewNetwork(t, ctx, p)
				err := other.DeleteSessionsByIdentity(ctx, expected2.IdentityID)
				assert.ErrorIs(t, err, sqlcon.ErrNoRows)

				_, err = p.GetSession(ctx, expected1.ID)
				require.NoError(t, err)
			})

			require.NoError(t, p.DeleteSessionsByIdentity(ctx, expected2.IdentityID))
			_, err := p.GetSession(ctx, expected1.ID)
			require.Error(t, err)
			_, err = p.GetSession(ctx, expected2.ID)
			require.Error(t, err)
		})

		t.Run("network isolation", func(t *testing.T) {
			nid1, p := testhelpers.NewNetwork(t, ctx, p)
			nid2, _ := testhelpers.NewNetwork(t, ctx, p)

			iid1, iid2 := x.NewUUID(), x.NewUUID()
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at) VALUES (?, ?, 'default', '{}', ?, ?)", iid1, nid1, time.Now(), time.Now()).Exec())
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at) VALUES (?, ?, 'default', '{}', ?, ?)", iid2, nid2, time.Now(), time.Now()).Exec())

			t1, t2 := randx.MustString(32, randx.AlphaNum), randx.MustString(32, randx.AlphaNum)
			sid1, sid2 := x.NewUUID(), x.NewUUID()
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO sessions (id, nid, identity_id, token, expires_at,authenticated_at, created_at, updated_at, logout_token) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", sid1, nid1, iid1, t1, time.Now().Add(time.Hour), time.Now(), time.Now(), time.Now(), randx.MustString(32, randx.AlphaNum)).Exec())
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO sessions (id, nid, identity_id, token, expires_at,authenticated_at, created_at, updated_at, logout_token) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", sid2, nid2, iid2, t2, time.Now().Add(time.Hour), time.Now(), time.Now(), time.Now(), randx.MustString(32, randx.AlphaNum)).Exec())

			_, err := p.GetSession(ctx, sid1)
			require.NoError(t, err)
			_, err = p.GetSession(ctx, sid2)
			require.ErrorIs(t, err, sqlcon.ErrNoRows)

			_, err = p.GetSessionByToken(ctx, t1)
			require.NoError(t, err)
			_, err = p.GetSessionByToken(ctx, t2)
			require.ErrorIs(t, err, sqlcon.ErrNoRows)
		})
	}
}
