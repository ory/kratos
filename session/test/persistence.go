package test

import (
	"context"
	"testing"
	"time"

	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/identity"

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
			expected.AMR = session.AuthenticationMethods{
				{Method: identity.CredentialsTypePassword, CompletedAt: time.Now().UTC().Round(time.Second)},
				{Method: identity.CredentialsTypeOIDC, CompletedAt: time.Now().UTC().Round(time.Second)},
			}
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))

			assert.Equal(t, uuid.Nil, expected.ID)
			require.NoError(t, p.UpsertSession(ctx, &expected))
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
				assert.Equal(t, expected.AuthenticatorAssuranceLevel, actual.AuthenticatorAssuranceLevel)
				assert.Equal(t, expected.AMR, actual.AMR)
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

			t.Run("method=list by identity", func(t *testing.T) {
				i := identity.NewIdentity("")
				require.NoError(t, p.CreateIdentity(ctx, i))
				sess := make([]session.Session, 4)
				for j := range sess {
					require.NoError(t, faker.FakeData(&sess[j]))
					sess[j].Identity = i
					sess[j].Active = j%2 == 0
					require.NoError(t, p.UpsertSession(ctx, &sess[j]))
				}

				for _, tc := range []struct {
					desc     string
					except   uuid.UUID
					expected []session.Session
					active   *bool
				}{
					{
						desc:     "all",
						expected: sess,
					},
					{
						desc:   "except one",
						except: sess[0].ID,
						expected: []session.Session{
							sess[1],
							sess[2],
							sess[3],
						},
					},
					{
						desc:   "active only",
						active: pointerx.Bool(true),
						expected: []session.Session{
							sess[0],
							sess[2],
						},
					},
					{
						desc:   "active only and except",
						active: pointerx.Bool(true),
						except: sess[0].ID,
						expected: []session.Session{
							sess[2],
						},
					},
					{
						desc:   "inactive only",
						active: pointerx.Bool(false),
						expected: []session.Session{
							sess[1],
							sess[3],
						},
					},
					{
						desc:   "inactive only and except",
						active: pointerx.Bool(false),
						except: sess[3].ID,
						expected: []session.Session{
							sess[1],
						},
					},
				} {
					t.Run("case="+tc.desc, func(t *testing.T) {
						actual, err := p.ListSessionsByIdentity(ctx, i.ID, tc.active, 1, 10, tc.except)
						require.NoError(t, err)

						require.Equal(t, len(tc.expected), len(actual))
						for _, es := range tc.expected {
							found := false
							for _, as := range actual {
								if as.ID == es.ID {
									found = true
								}
							}
							assert.True(t, found)
						}
					})
				}

				t.Run("other network", func(t *testing.T) {
					_, other := testhelpers.NewNetwork(t, ctx, p)
					actual, err := other.ListSessionsByIdentity(ctx, i.ID, nil, 1, 10, uuid.Nil)
					require.NoError(t, err)
					assert.Len(t, actual, 0)
				})
			})

			t.Run("case=update session", func(t *testing.T) {
				expected.AuthenticatorAssuranceLevel = identity.AuthenticatorAssuranceLevel3
				require.NoError(t, p.UpsertSession(ctx, &expected))

				actual, err := p.GetSessionByToken(ctx, expected.Token)
				check(actual, err)
				assert.Equal(t, identity.AuthenticatorAssuranceLevel3, actual.AuthenticatorAssuranceLevel)
			})

			t.Run("case=remove amr and update", func(t *testing.T) {
				expected.AMR = nil
				require.NoError(t, p.UpsertSession(ctx, &expected))

				actual, err := p.GetSessionByToken(ctx, expected.Token)
				check(actual, err)
				assert.Empty(t, actual.AMR)
			})
		})

		t.Run("case=delete session", func(t *testing.T) {
			var expected session.Session
			require.NoError(t, faker.FakeData(&expected))
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.UpsertSession(ctx, &expected))

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
			require.NoError(t, p.UpsertSession(ctx, &expected))

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
			require.NoError(t, p.UpsertSession(ctx, &expected))

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

		t.Run("method=revoke other sessions for identity", func(t *testing.T) {
			// here we set up 2 identities with each having 2 sessions
			sessions := make([]session.Session, 4)
			for i := range sessions {
				require.NoError(t, faker.FakeData(&sessions[i]))
			}
			require.NoError(t, p.CreateIdentity(ctx, sessions[0].Identity))
			require.NoError(t, p.CreateIdentity(ctx, sessions[2].Identity))
			sessions[1].IdentityID, sessions[1].Identity = sessions[0].IdentityID, sessions[0].Identity
			sessions[3].IdentityID, sessions[3].Identity = sessions[2].IdentityID, sessions[2].Identity
			for i := range sessions {
				sessions[i].Active = true
				require.NoError(t, p.UpsertSession(ctx, &sessions[i]))
			}

			t.Run("on another network", func(t *testing.T) {
				_, other := testhelpers.NewNetwork(t, ctx, p)
				n, err := other.RevokeSessionsIdentityExcept(ctx, sessions[0].IdentityID, sessions[0].ID)
				require.NoError(t, err)
				assert.Equal(t, 0, n)

				for _, s := range sessions {
					actual, err := p.GetSession(ctx, s.ID)
					require.NoError(t, err)
					assert.True(t, actual.Active)
				}
			})

			n, err := p.RevokeSessionsIdentityExcept(ctx, sessions[0].IdentityID, sessions[0].ID)
			require.NoError(t, err)
			assert.Equal(t, 1, n)

			actual, err := p.ListSessionsByIdentity(ctx, sessions[0].IdentityID, nil, 1, 10, uuid.Nil)
			require.NoError(t, err)
			require.Len(t, actual, 2)

			if actual[0].ID == sessions[0].ID {
				assert.True(t, actual[0].Active)
				assert.False(t, actual[1].Active)
			} else {
				assert.Equal(t, actual[0].ID, sessions[1].ID)
				assert.True(t, actual[1].Active)
				assert.False(t, actual[0].Active)
			}

			otherIdentitiesSessions, err := p.ListSessionsByIdentity(ctx, sessions[2].IdentityID, nil, 1, 10, uuid.Nil)
			require.NoError(t, err)
			require.Len(t, actual, 2)

			for _, s := range otherIdentitiesSessions {
				assert.True(t, s.Active)
			}
		})

		t.Run("method=revoke specific session for identity", func(t *testing.T) {
			sessions := make([]session.Session, 2)
			for i := range sessions {
				require.NoError(t, faker.FakeData(&sessions[i]))
			}
			require.NoError(t, p.CreateIdentity(ctx, sessions[0].Identity))
			sessions[1].IdentityID, sessions[1].Identity = sessions[0].IdentityID, sessions[0].Identity
			for i := range sessions {
				sessions[i].Active = true
				require.NoError(t, p.UpsertSession(ctx, &sessions[i]))
			}

			t.Run("on another network", func(t *testing.T) {
				_, other := testhelpers.NewNetwork(t, ctx, p)
				require.NoError(t, other.RevokeSession(ctx, sessions[0].IdentityID, sessions[0].ID))

				for _, s := range sessions {
					actual, err := p.GetSession(ctx, s.ID)
					require.NoError(t, err)
					assert.True(t, actual.Active)
				}
			})

			require.NoError(t, p.RevokeSession(ctx, sessions[0].IdentityID, sessions[0].ID))

			actual, err := p.ListSessionsByIdentity(ctx, sessions[0].IdentityID, nil, 1, 10, uuid.Nil)
			require.NoError(t, err)
			require.Len(t, actual, 2)

			if actual[0].ID == sessions[0].ID {
				assert.False(t, actual[0].Active)
				assert.True(t, actual[1].Active)
			} else {
				assert.Equal(t, actual[0].ID, sessions[1].ID)
				assert.False(t, actual[1].Active)
				assert.True(t, actual[0].Active)
			}
		})

		t.Run("case=delete session for", func(t *testing.T) {
			var expected1 session.Session
			var expected2 session.Session
			require.NoError(t, faker.FakeData(&expected1))
			require.NoError(t, p.CreateIdentity(ctx, expected1.Identity))

			require.NoError(t, p.UpsertSession(ctx, &expected1))

			require.NoError(t, faker.FakeData(&expected2))
			expected2.Identity = expected1.Identity
			expected2.IdentityID = expected1.IdentityID
			require.NoError(t, p.UpsertSession(ctx, &expected2))

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
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO sessions (id, nid, identity_id, token, expires_at,authenticated_at, created_at, updated_at, logout_token, authentication_methods) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", sid1, nid1, iid1, t1, time.Now().Add(time.Hour), time.Now(), time.Now(), time.Now(), randx.MustString(32, randx.AlphaNum), "[]").Exec())
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO sessions (id, nid, identity_id, token, expires_at,authenticated_at, created_at, updated_at, logout_token, authentication_methods) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", sid2, nid2, iid2, t2, time.Now().Add(time.Hour), time.Now(), time.Now(), time.Now(), randx.MustString(32, randx.AlphaNum), "[]").Exec())

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
