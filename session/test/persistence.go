// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/contextx"
	"github.com/ory/x/dbal"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/randx"
	"github.com/ory/x/sqlcon"
)

func TestPersister(ctx context.Context, conf *config.Config, p interface {
	persistence.Persister
},
) func(t *testing.T) {
	return func(t *testing.T) {
		_, p := testhelpers.NewNetworkUnlessExisting(t, ctx, p)

		t.Run("case=not found", func(t *testing.T) {
			_, err := p.GetSession(ctx, x.NewUUID(), session.ExpandNothing)
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

			var expectedSessionDevice session.Device
			require.NoError(t, faker.FakeData(&expectedSessionDevice))
			expected.Devices = []session.Device{
				expectedSessionDevice,
			}

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

			checkDevices := func(actual []session.Device, err error) {
				require.NoError(t, err)
				assert.Equal(t, len(expected.Devices), len(actual))

				for i, d := range actual {
					assert.Equal(t, expected.Devices[i].SessionID, d.SessionID)
					assert.Equal(t, expected.Devices[i].NID, d.NID)
					assert.Equal(t, *expected.Devices[i].IPAddress, *d.IPAddress)
					assert.Equal(t, expected.Devices[i].UserAgent, d.UserAgent)
					assert.Equal(t, *expected.Devices[i].Location, *d.Location)
					assert.Equal(t, *expected.Devices[i].IdentityID, *d.IdentityID)
				}
			}

			t.Run("method=get by id", func(t *testing.T) {
				sess, err := p.GetSession(ctx, expected.ID, session.ExpandEverything)
				check(sess, err)
				checkDevices(sess.Devices, err)

				t.Run("on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.GetSession(ctx, expected.ID, session.ExpandEverything)
					assert.ErrorIs(t, err, sqlcon.ErrNoRows())
				})
			})

			t.Run("method=get by token", func(t *testing.T) {
				sess, err := p.GetSessionByToken(ctx, expected.Token, session.ExpandEverything, identity.ExpandDefault)
				check(sess, err)
				checkDevices(sess.Devices, err)

				t.Run("on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.GetSessionByToken(ctx, expected.Token, session.ExpandNothing, identity.ExpandDefault)
					assert.ErrorIs(t, err, sqlcon.ErrNoRows())
				})
			})

			t.Run("case=update session", func(t *testing.T) {
				expected.AuthenticatorAssuranceLevel = identity.AuthenticatorAssuranceLevel1
				require.NoError(t, p.UpsertSession(ctx, &expected))

				actual, err := p.GetSessionByToken(ctx, expected.Token, session.ExpandDefault, identity.ExpandDefault)
				check(actual, err)
				assert.Equal(t, identity.AuthenticatorAssuranceLevel1, actual.AuthenticatorAssuranceLevel)
			})

			t.Run("case=remove amr and update", func(t *testing.T) {
				expected.AMR = nil
				require.NoError(t, p.UpsertSession(ctx, &expected))

				actual, err := p.GetSessionByToken(ctx, expected.Token, session.ExpandDefault, identity.ExpandDefault)
				check(actual, err)
				assert.Empty(t, actual.AMR)
			})
		})

		t.Run("case=list sessions", func(t *testing.T) {
			var identity1 identity.Identity
			require.NoError(t, faker.FakeData(&identity1))

			// Second identity to test listing by identity isolation
			var identity2 identity.Identity
			var identity2Session session.Session
			require.NoError(t, faker.FakeData(&identity2))
			require.NoError(t, faker.FakeData(&identity2Session))

			// Create seed identities
			_, l := testhelpers.NewNetwork(t, ctx, p)
			require.NoError(t, l.CreateIdentity(ctx, &identity1))
			require.NoError(t, l.CreateIdentity(ctx, &identity2))

			seedSessionIDs := make([]uuid.UUID, 5)
			seedSessionsList := make([]session.Session, 5)
			now := time.Now().UTC()
			for j := range seedSessionsList {
				require.NoError(t, faker.FakeData(&seedSessionsList[j]))
				seedSessionsList[j].Identity = &identity1
				seedSessionsList[j].Active = j%2 == 0

				if seedSessionsList[j].Active {
					seedSessionsList[j].ExpiresAt = time.Now().UTC().Add(time.Hour)
				} else {
					seedSessionsList[j].ExpiresAt = time.Now().UTC().Add(-time.Hour)
				}

				var device session.Device
				require.NoError(t, faker.FakeData(&device))
				seedSessionsList[j].Devices = []session.Device{
					device,
				}
				require.NoError(t, l.UpsertSession(ctx, &seedSessionsList[j]))
				require.NoError(t, p.GetConnection(ctx).
					RawQuery(
						"UPDATE sessions SET created_at = ?, updated_at = ? WHERE id = ?",
						now.Add(time.Duration(j)*time.Minute).Round(time.Second),
						now.Add(time.Duration(j)*time.Minute).Round(time.Second),
						seedSessionsList[j].ID).
					Exec())
				seedSessionIDs[j] = seedSessionsList[j].ID
			}

			identity2Session.Identity = &identity2
			identity2Session.Active = true
			identity2Session.ExpiresAt = time.Now().UTC().Add(time.Hour)
			require.NoError(t, l.UpsertSession(ctx, &identity2Session))

			for _, tc := range []struct {
				desc               string
				except             uuid.UUID
				expectedSessionIds []uuid.UUID
				active             *bool
			}{
				{
					desc:               "all",
					expectedSessionIds: seedSessionIDs,
				},
				{
					desc:   "except one",
					except: seedSessionsList[0].ID,
					expectedSessionIds: []uuid.UUID{
						seedSessionIDs[1],
						seedSessionIDs[2],
						seedSessionIDs[3],
						seedSessionIDs[4],
					},
				},
				{
					desc:   "active only",
					active: new(true),
					expectedSessionIds: []uuid.UUID{
						seedSessionIDs[0],
						seedSessionIDs[2],
						seedSessionIDs[4],
					},
				},
				{
					desc:   "active only and except",
					active: new(true),
					except: seedSessionsList[0].ID,
					expectedSessionIds: []uuid.UUID{
						seedSessionIDs[2],
						seedSessionIDs[4],
					},
				},
				{
					desc:   "inactive only",
					active: new(false),
					expectedSessionIds: []uuid.UUID{
						seedSessionIDs[1],
						seedSessionIDs[3],
					},
				},
				{
					desc:   "inactive only and except",
					active: new(false),
					except: seedSessionsList[3].ID,
					expectedSessionIds: []uuid.UUID{
						seedSessionIDs[1],
					},
				},
			} {
				t.Run("case=by Identity "+tc.desc, func(t *testing.T) {
					actual, total, err := l.ListSessionsByIdentity(ctx, identity1.ID, tc.active, 1, 10, tc.except, session.ExpandEverything)
					require.NoError(t, err)

					actualSessionIds := make([]uuid.UUID, 0)
					for _, s := range actual {
						actualSessionIds = append(actualSessionIds, s.ID)
					}

					assert.Equal(t, int64(len(tc.expectedSessionIds)), total)
					assert.ElementsMatch(t, tc.expectedSessionIds, actualSessionIds)
				})
			}

			t.Run("case=by Identity on other network", func(t *testing.T) {
				_, other := testhelpers.NewNetwork(t, ctx, p)
				actual, total, err := other.ListSessionsByIdentity(ctx, identity1.ID, nil, 1, 10, uuid.Nil, session.ExpandNothing)
				require.NoError(t, err)
				require.Equal(t, int64(0), total)
				assert.Len(t, actual, 0)
			})

			for _, tc := range []struct {
				desc     string
				except   uuid.UUID
				expected []session.Session
				active   *bool
			}{
				{
					desc:     "all",
					expected: append(seedSessionsList, identity2Session),
				},
				{
					desc:   "active only",
					active: new(true),
					expected: []session.Session{
						seedSessionsList[0],
						seedSessionsList[2],
						seedSessionsList[4],
						identity2Session,
					},
				},
				{
					desc:   "inactive only",
					active: new(false),
					expected: []session.Session{
						seedSessionsList[1],
						seedSessionsList[3],
					},
				},
			} {
				t.Run("case=all "+tc.desc, func(t *testing.T) {
					paginatorOpts := make([]keysetpagination.Option, 0)
					actual, nextPage, err := l.ListSessions(ctx, tc.active, paginatorOpts, session.ExpandEverything)
					require.NoError(t, err, "%+v", err)

					require.Equal(t, len(tc.expected), len(actual))
					assert.Equal(t, true, nextPage.IsLast())

					mapPageToken := nextPage.Token().Parse("")
					assert.Equal(t, uuid.Nil.String(), mapPageToken["id"])

					assert.Equal(t, 250, nextPage.Size())
					for _, es := range tc.expected {
						found := false
						for _, as := range actual {
							if as.ID == es.ID {
								found = true
								assert.Equal(t, len(es.Devices), len(as.Devices))
								assert.Equal(t, es.Identity.ID.String(), as.Identity.ID.String())
							}
						}
						assert.True(t, found)
					}
				})
			}

			t.Run("case=all sessions pagination only one page", func(t *testing.T) {
				paginatorOpts := make([]keysetpagination.Option, 0)
				actual, page, err := l.ListSessions(ctx, nil, paginatorOpts, session.ExpandEverything)
				require.NoError(t, err)

				require.Equal(t, 6, len(actual))
				assert.Equal(t, true, page.IsLast())
				mapPageToken := page.Token().Parse("")
				assert.Equal(t, uuid.Nil.String(), mapPageToken["id"])
				assert.Equal(t, 250, page.Size())
			})

			t.Run("case=all sessions pagination multiple pages", func(t *testing.T) {
				paginatorOpts := make([]keysetpagination.Option, 0)
				paginatorOpts = append(paginatorOpts, keysetpagination.WithSize(3))
				firstPageItems, page1, err := l.ListSessions(ctx, nil, paginatorOpts, session.ExpandEverything)
				require.NoError(t, err)
				assert.Len(t, firstPageItems, 3)

				assert.Equal(t, false, page1.IsLast())
				mapPageToken := page1.Token().Parse("")
				assert.Equal(t, firstPageItems[len(firstPageItems)-1].ID.String(), mapPageToken["id"])
				assert.Equal(t, 3, page1.Size())

				// Validate secondPageItems page
				secondPageItems, page2, err := l.ListSessions(ctx, nil, page1.ToOptions(), session.ExpandEverything)
				require.NoError(t, err)
				assert.Len(t, secondPageItems, 3)

				acutalIDs := make([]uuid.UUID, 0)
				for _, s := range append(firstPageItems, secondPageItems...) {
					acutalIDs = append(acutalIDs, s.ID)
				}
				expect := append(seedSessionIDs, identity2Session.ID)
				require.Len(t, acutalIDs, len(expect))
				assert.ElementsMatch(t, expect, acutalIDs)

				assert.True(t, page2.IsLast())
				assert.Equal(t, 3, page2.Size())
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
				assert.ErrorIs(t, err, sqlcon.ErrNoRows())

				_, err = p.GetSession(ctx, expected.ID, session.ExpandNothing)
				assert.NoError(t, err)
			})

			require.NoError(t, p.DeleteSession(ctx, expected.ID))
			_, err := p.GetSession(ctx, expected.ID, session.ExpandNothing)
			assert.ErrorIs(t, err, sqlcon.ErrNoRows())
		})

		t.Run("case=delete session by token", func(t *testing.T) {
			var expected session.Session
			require.NoError(t, faker.FakeData(&expected))
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.UpsertSession(ctx, &expected))

			t.Run("on another network", func(t *testing.T) {
				_, other := testhelpers.NewNetwork(t, ctx, p)
				err := other.DeleteSessionByToken(ctx, expected.Token)
				assert.ErrorIs(t, err, sqlcon.ErrNoRows())

				_, err = p.GetSessionByToken(ctx, expected.Token, session.ExpandNothing, identity.ExpandDefault)
				assert.NoError(t, err)
			})

			require.NoError(t, p.DeleteSessionByToken(ctx, expected.Token))
			_, err := p.GetSession(ctx, expected.ID, session.ExpandNothing)
			require.Error(t, err)
		})

		t.Run("case=revoke session by token", func(t *testing.T) {
			var expected session.Session
			require.NoError(t, faker.FakeData(&expected))
			expected.Active = true
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.UpsertSession(ctx, &expected))

			actual, err := p.GetSession(ctx, expected.ID, session.ExpandNothing)
			require.NoError(t, err)
			assert.True(t, actual.Active)

			t.Run("on another network", func(t *testing.T) {
				_, other := testhelpers.NewNetwork(t, ctx, p)
				revoked, err := other.RevokeSessionByToken(ctx, expected.Token)
				assert.ErrorIs(t, err, sqlcon.ErrNoRows())
				assert.Equal(t, uuid.Nil, revoked.ID, "session id must be nil when not found")
				assert.Equal(t, uuid.Nil, revoked.IdentityID, "identity id must be nil when not found")

				actual, err = p.GetSession(ctx, expected.ID, session.ExpandNothing)
				require.NoError(t, err)
				assert.True(t, actual.Active)
			})

			revoked, err := p.RevokeSessionByToken(ctx, expected.Token)
			require.NoError(t, err)
			assert.Equal(t, expected.ID, revoked.ID, "must return the session id")
			assert.Equal(t, expected.IdentityID, revoked.IdentityID, "must return the identity id")

			actual, err = p.GetSession(ctx, expected.ID, session.ExpandNothing)
			require.NoError(t, err)
			assert.False(t, actual.Active)

			// Re-revoking an already-inactive session is silently
			// idempotent — the CTE-based implementation returns the matched
			// row, not the flipped one, so the IDs come back regardless.
			revoked, err = p.RevokeSessionByToken(ctx, expected.Token)
			require.NoError(t, err,
				"re-revoking an already-inactive session must succeed silently on %s",
				p.GetConnection(ctx).Dialect.Name())
			assert.Equal(t, expected.ID, revoked.ID, "matched id must be returned on re-revoke")
			assert.Equal(t, expected.IdentityID, revoked.IdentityID, "matched identity id must be returned on re-revoke")

			actual, err = p.GetSession(ctx, expected.ID, session.ExpandNothing)
			require.NoError(t, err, "session must still exist after re-revoke")
			assert.False(t, actual.Active, "session must remain inactive after re-revoke")

			t.Run("returns ErrNoRows for an unknown token", func(t *testing.T) {
				revoked, err := p.RevokeSessionByToken(ctx, "this-token-does-not-exist-"+x.NewUUID().String())
				assert.ErrorIs(t, err, sqlcon.ErrNoRows())
				assert.Equal(t, uuid.Nil, revoked.ID)
				assert.Equal(t, uuid.Nil, revoked.IdentityID)
			})
		})

		t.Run("case=revoke session by id", func(t *testing.T) {
			var expected session.Session
			require.NoError(t, faker.FakeData(&expected))
			expected.Active = true
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.UpsertSession(ctx, &expected))

			actual, err := p.GetSession(ctx, expected.ID, session.ExpandNothing)
			require.NoError(t, err)
			assert.True(t, actual.Active)

			t.Run("on another network", func(t *testing.T) {
				_, other := testhelpers.NewNetwork(t, ctx, p)
				err := other.RevokeSessionById(ctx, expected.ID)
				assert.ErrorIs(t, err, sqlcon.ErrNoRows())

				actual, err = p.GetSession(ctx, expected.ID, session.ExpandNothing)
				require.NoError(t, err)
				assert.True(t, actual.Active)
			})

			require.NoError(t, p.RevokeSessionById(ctx, expected.ID))

			actual, err = p.GetSession(ctx, expected.ID, session.ExpandNothing)
			require.NoError(t, err)
			assert.False(t, actual.Active)

			// Re-revoking an already-inactive session is silently
			// idempotent — the admin disableSession handler relies on it.
			require.NoError(t, p.RevokeSessionById(ctx, expected.ID),
				"re-revoking an already-inactive session must succeed silently on %s",
				p.GetConnection(ctx).Dialect.Name())

			actual, err = p.GetSession(ctx, expected.ID, session.ExpandNothing)
			require.NoError(t, err, "session must still exist after re-revoke")
			assert.False(t, actual.Active, "session must remain inactive after re-revoke")

			t.Run("returns ErrNoRows for an unknown id", func(t *testing.T) {
				err := p.RevokeSessionById(ctx, x.NewUUID())
				assert.ErrorIs(t, err, sqlcon.ErrNoRows())
			})
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
					actual, err := p.GetSession(ctx, s.ID, session.ExpandNothing)
					require.NoError(t, err)
					assert.True(t, actual.Active)
				}
			})

			n, err := p.RevokeSessionsIdentityExcept(ctx, sessions[0].IdentityID, sessions[0].ID)
			require.NoError(t, err)
			assert.Equal(t, 1, n)

			actual, total, err := p.ListSessionsByIdentity(ctx, sessions[0].IdentityID, nil, 1, 10, uuid.Nil, session.ExpandNothing)
			require.NoError(t, err)
			require.Len(t, actual, 2)
			require.Equal(t, int64(2), total)

			if actual[0].ID == sessions[0].ID {
				assert.True(t, actual[0].Active)
				assert.False(t, actual[1].Active)
			} else {
				assert.Equal(t, actual[0].ID, sessions[1].ID)
				assert.True(t, actual[1].Active)
				assert.False(t, actual[0].Active)
			}

			otherIdentitiesSessions, total, err := p.ListSessionsByIdentity(ctx, sessions[2].IdentityID, nil, 1, 10, uuid.Nil, session.ExpandNothing)
			require.NoError(t, err)
			require.Len(t, actual, 2)
			require.Equal(t, int64(2), total)

			for _, s := range otherIdentitiesSessions {
				assert.True(t, s.Active)
			}

			t.Run("returns matched count, including already-inactive sessions", func(t *testing.T) {
				// One identity, three sessions: the keeper, one already-
				// inactive other, one active other. Returned count must be 2
				// (both "other" sessions matched the predicate) — not 1
				// (only the active "other" was actually flipped). The CTE on
				// CRDB/Postgres must preserve matched-count, not flipped-count.
				var keeper session.Session
				require.NoError(t, faker.FakeData(&keeper))
				keeper.Active = true
				require.NoError(t, p.CreateIdentity(ctx, keeper.Identity))
				require.NoError(t, p.UpsertSession(ctx, &keeper))

				makeSibling := func() session.Session {
					var s session.Session
					require.NoError(t, faker.FakeData(&s))
					s.Identity = keeper.Identity
					s.IdentityID = keeper.IdentityID
					s.Active = true
					require.NoError(t, p.UpsertSession(ctx, &s))
					return s
				}
				alreadyInactive := makeSibling()
				require.NoError(t, p.RevokeSessionById(ctx, alreadyInactive.ID))
				_ = makeSibling() // stillActive — will be flipped by the call below

				n, err := p.RevokeSessionsIdentityExcept(ctx, keeper.IdentityID, keeper.ID)
				require.NoError(t, err)
				assert.Equal(t, 2, n,
					"matched count must include already-inactive sessions on %s",
					p.GetConnection(ctx).Dialect.Name())
			})
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
					actual, err := p.GetSession(ctx, s.ID, session.ExpandNothing)
					require.NoError(t, err)
					assert.True(t, actual.Active)
				}
			})

			require.NoError(t, p.RevokeSession(ctx, sessions[0].IdentityID, sessions[0].ID))

			actual, total, err := p.ListSessionsByIdentity(ctx, sessions[0].IdentityID, nil, 1, 10, uuid.Nil, session.ExpandNothing)
			require.NoError(t, err)
			require.Len(t, actual, 2)
			require.Equal(t, int64(2), total)

			if actual[0].ID == sessions[0].ID {
				assert.False(t, actual[0].Active)
				assert.True(t, actual[1].Active)
			} else {
				assert.Equal(t, actual[0].ID, sessions[1].ID)
				assert.False(t, actual[1].Active)
				assert.True(t, actual[0].Active)
			}

			// Re-revoking an already-inactive session must succeed silently —
			// the `AND active = true` predicate added on the CRDB/Postgres path
			// must not change the function's nil-on-no-rows contract.
			require.NoError(t, p.RevokeSession(ctx, sessions[0].IdentityID, sessions[0].ID),
				"re-revoking an already-inactive session must succeed silently")

			// Wrong identity_id for an existing session id is a silent no-op,
			// not an error.
			require.NoError(t, p.RevokeSession(ctx, x.NewUUID(), sessions[1].ID))
			actual1, err := p.GetSession(ctx, sessions[1].ID, session.ExpandNothing)
			require.NoError(t, err)
			assert.True(t, actual1.Active, "session must stay active when identity_id mismatches")
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
				assert.ErrorIs(t, err, sqlcon.ErrNoRows())

				_, err = p.GetSession(ctx, expected1.ID, session.ExpandNothing)
				require.NoError(t, err)
			})

			require.NoError(t, p.DeleteSessionsByIdentity(ctx, expected2.IdentityID))
			_, err := p.GetSession(ctx, expected1.ID, session.ExpandNothing)
			require.Error(t, err)
			_, err = p.GetSession(ctx, expected2.ID, session.ExpandNothing)
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

			_, err := p.GetSession(ctx, sid1, session.ExpandEverything)
			require.NoError(t, err)
			_, err = p.GetSession(ctx, sid2, session.ExpandNothing)
			require.ErrorIs(t, err, sqlcon.ErrNoRows())

			_, err = p.GetSessionByToken(ctx, t1, session.ExpandNothing, identity.ExpandDefault)
			require.NoError(t, err)
			_, err = p.GetSessionByToken(ctx, t2, session.ExpandNothing, identity.ExpandDefault)
			require.ErrorIs(t, err, sqlcon.ErrNoRows())
		})

		t.Run("extend session lifespan but min time is not yet reached", func(t *testing.T) {
			ctx := contextx.WithConfigValues(ctx, map[string]any{config.ViperKeySessionRefreshMinTimeLeft: 2 * time.Hour})

			var expected session.Session
			require.NoError(t, faker.FakeData(&expected))
			expected.ExpiresAt = time.Now().Add(time.Hour * 10).Round(time.Second).UTC()
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.UpsertSession(ctx, &expected))

			require.NoError(t, p.ExtendSession(ctx, expected.ID))
			actual, err := p.GetSession(ctx, expected.ID, session.ExpandNothing)
			require.NoError(t, err)
			assert.Equal(t, expected.ExpiresAt, actual.ExpiresAt)
		})

		t.Run("extend session lifespan", func(t *testing.T) {
			ctx := contextx.WithConfigValues(ctx, map[string]any{config.ViperKeySessionRefreshMinTimeLeft: 2 * time.Hour})

			var expected session.Session
			require.NoError(t, faker.FakeData(&expected))
			expected.ExpiresAt = time.Now().Add(time.Hour).UTC()
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.UpsertSession(ctx, &expected))

			expectedExpiry := expected.Refresh(ctx, conf).ExpiresAt
			require.NoError(t, p.ExtendSession(ctx, expected.ID))
			actual, err := p.GetSession(ctx, expected.ID, session.ExpandNothing)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, 10*time.Second, expectedExpiry.Sub(actual.ExpiresAt).Abs())
		})

		t.Run("extend session lifespan on CockroachDB", func(t *testing.T) {
			if p.GetConnection(ctx).Dialect.Name() != dbal.DriverCockroachDB {
				t.Skip("Skipping test because driver is not CockroachDB")
			}

			ctx := contextx.WithConfigValue(ctx, config.ViperKeySessionRefreshMinTimeLeft, 2*time.Hour)

			var expected session.Session
			require.NoError(t, faker.FakeData(&expected))
			expected.ExpiresAt = time.Now().Add(time.Hour).UTC()
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.UpsertSession(ctx, &expected))

			expectedExpiry := expected.Refresh(ctx, conf).ExpiresAt

			foundExpectedCockroachError := atomic.Bool{}
			g := errgroup.Group{}
			for range 10 {
				g.Go(func() error {
					err := p.ExtendSession(ctx, expected.ID)
					if errors.Is(err, sqlcon.ErrNoRows()) {
						foundExpectedCockroachError.Store(true)
						return nil
					}
					return err
				})
			}
			require.NoError(t, g.Wait())

			actual, err := p.GetSession(ctx, expected.ID, session.ExpandNothing)
			require.NoError(t, err)
			assert.LessOrEqual(t, expectedExpiry.Sub(actual.ExpiresAt).Abs(), 10*time.Second)
			assert.True(t, foundExpectedCockroachError.Load(), "We expect to find a not found error caused by ... FOR UPDATE SKIP LOCKED")
		})

		// Helpers for the manage-sessions persister batch tests below. Each
		// helper creates an active session belonging to a fresh identity and
		// returns both, so callers can mutate state independently across
		// subtests without sharing fixtures.
		newActiveSession := func(t *testing.T) *session.Session {
			t.Helper()
			var s session.Session
			require.NoError(t, faker.FakeData(&s))
			s.Active = true
			require.NoError(t, p.CreateIdentity(ctx, s.Identity))
			require.NoError(t, p.UpsertSession(ctx, &s))
			return &s
		}
		isActive := func(t *testing.T, sID uuid.UUID) bool {
			t.Helper()
			s, err := p.GetSession(ctx, sID, session.ExpandNothing)
			require.NoError(t, err)
			return s.Active
		}
		exists := func(t *testing.T, sID uuid.UUID) bool {
			t.Helper()
			_, err := p.GetSession(ctx, sID, session.ExpandNothing)
			if errors.Is(err, sqlcon.ErrNoRows()) {
				return false
			}
			require.NoError(t, err)
			return true
		}

		t.Run("case=RevokeSessionsByIdentities updates only matching active rows", func(t *testing.T) {
			s1 := newActiveSession(t)
			s2 := newActiveSession(t)
			other := newActiveSession(t)

			n, err := p.RevokeSessionsByIdentities(ctx, []uuid.UUID{s1.IdentityID, s2.IdentityID})
			require.NoError(t, err)
			assert.Equal(t, 2, n)
			assert.False(t, isActive(t, s1.ID))
			assert.False(t, isActive(t, s2.ID))
			assert.True(t, isActive(t, other.ID), "untargeted session must stay active")

			n, err = p.RevokeSessionsByIdentities(ctx, []uuid.UUID{s1.IdentityID})
			require.NoError(t, err)
			assert.Equal(t, 0, n, "already-inactive sessions must not be re-touched")
		})

		t.Run("case=RevokeSessionsByIdentities is network-isolated", func(t *testing.T) {
			s := newActiveSession(t)
			_, other := testhelpers.NewNetwork(t, ctx, p)
			n, err := other.RevokeSessionsByIdentities(ctx, []uuid.UUID{s.IdentityID})
			require.NoError(t, err)
			assert.Equal(t, 0, n)
			assert.True(t, isActive(t, s.ID))
		})

		t.Run("case=RevokeSessionsByIdentities with empty slice is no-op", func(t *testing.T) {
			n, err := p.RevokeSessionsByIdentities(ctx, nil)
			require.NoError(t, err)
			assert.Equal(t, 0, n)
		})

		t.Run("case=RevokeSessionsByIDs updates only matching active rows", func(t *testing.T) {
			s1 := newActiveSession(t)
			s2 := newActiveSession(t)
			untouched := newActiveSession(t)

			n, err := p.RevokeSessionsByIDs(ctx, []uuid.UUID{s1.ID, s2.ID})
			require.NoError(t, err)
			assert.Equal(t, 2, n)
			assert.False(t, isActive(t, s1.ID))
			assert.False(t, isActive(t, s2.ID))
			assert.True(t, isActive(t, untouched.ID))

			n, err = p.RevokeSessionsByIDs(ctx, []uuid.UUID{s1.ID})
			require.NoError(t, err)
			assert.Equal(t, 0, n, "already-inactive sessions must not be re-touched")
		})

		t.Run("case=RevokeSessionsByIDs is network-isolated", func(t *testing.T) {
			s := newActiveSession(t)
			_, other := testhelpers.NewNetwork(t, ctx, p)
			n, err := other.RevokeSessionsByIDs(ctx, []uuid.UUID{s.ID})
			require.NoError(t, err)
			assert.Equal(t, 0, n)
			assert.True(t, isActive(t, s.ID))
		})

		t.Run("case=RevokeSessionsByIDs with empty slice is no-op", func(t *testing.T) {
			n, err := p.RevokeSessionsByIDs(ctx, nil)
			require.NoError(t, err)
			assert.Equal(t, 0, n)
		})

		t.Run("case=RevokeAllSessions deactivates only currently-active rows in this network", func(t *testing.T) {
			_, scoped := testhelpers.NewNetwork(t, ctx, p)
			active1 := newActiveSession(t)
			active2 := newActiveSession(t)

			// Create one session that's already inactive, in the original network.
			alreadyInactive := newActiveSession(t)
			require.NoError(t, p.RevokeSessionById(ctx, alreadyInactive.ID))

			// Create a session in the other network — must remain untouched.
			var foreign session.Session
			require.NoError(t, faker.FakeData(&foreign))
			foreign.Active = true
			require.NoError(t, scoped.CreateIdentity(ctx, foreign.Identity))
			require.NoError(t, scoped.UpsertSession(ctx, &foreign))

			n, err := p.RevokeAllSessions(ctx, 100)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, n, 2, "must have flipped active1 and active2 at least")

			assert.False(t, isActive(t, active1.ID))
			assert.False(t, isActive(t, active2.ID))
			// The foreign session must still be active when read via its own network.
			fs, err := scoped.GetSession(ctx, foreign.ID, session.ExpandNothing)
			require.NoError(t, err)
			assert.True(t, fs.Active, "foreign-network session must not be touched")
		})

		t.Run("case=RevokeAllSessions caps the work at limit", func(t *testing.T) {
			// Isolated network so previous tests can't influence the count.
			_, scoped := testhelpers.NewNetwork(t, ctx, p)
			ids := make([]uuid.UUID, 3)
			for i := range ids {
				var s session.Session
				require.NoError(t, faker.FakeData(&s))
				s.Active = true
				require.NoError(t, scoped.CreateIdentity(ctx, s.Identity))
				require.NoError(t, scoped.UpsertSession(ctx, &s))
				ids[i] = s.ID
			}

			n, err := scoped.RevokeAllSessions(ctx, 2)
			require.NoError(t, err)
			assert.Equal(t, 2, n, "must flip exactly limit rows")

			activeCount := 0
			for _, id := range ids {
				s, err := scoped.GetSession(ctx, id, session.ExpandNothing)
				require.NoError(t, err)
				if s.Active {
					activeCount++
				}
			}
			assert.Equal(t, 1, activeCount, "limit must leave one session active")
		})

		t.Run("case=DeleteSessionsByIdentities removes matching rows", func(t *testing.T) {
			s1 := newActiveSession(t)
			s2 := newActiveSession(t)
			survivor := newActiveSession(t)

			n, err := p.DeleteSessionsByIdentities(ctx, []uuid.UUID{s1.IdentityID, s2.IdentityID})
			require.NoError(t, err)
			assert.Equal(t, 2, n)
			assert.False(t, exists(t, s1.ID))
			assert.False(t, exists(t, s2.ID))
			assert.True(t, exists(t, survivor.ID))
		})

		t.Run("case=DeleteSessionsByIdentities is network-isolated", func(t *testing.T) {
			s := newActiveSession(t)
			_, other := testhelpers.NewNetwork(t, ctx, p)
			n, err := other.DeleteSessionsByIdentities(ctx, []uuid.UUID{s.IdentityID})
			require.NoError(t, err)
			assert.Equal(t, 0, n)
			assert.True(t, exists(t, s.ID))
		})

		t.Run("case=DeleteSessionsByIDs removes matching rows", func(t *testing.T) {
			s1 := newActiveSession(t)
			s2 := newActiveSession(t)
			survivor := newActiveSession(t)

			n, err := p.DeleteSessionsByIDs(ctx, []uuid.UUID{s1.ID, s2.ID})
			require.NoError(t, err)
			assert.Equal(t, 2, n)
			assert.False(t, exists(t, s1.ID))
			assert.False(t, exists(t, s2.ID))
			assert.True(t, exists(t, survivor.ID))
		})

		t.Run("case=DeleteSessionsByIDs is network-isolated", func(t *testing.T) {
			s := newActiveSession(t)
			_, other := testhelpers.NewNetwork(t, ctx, p)
			n, err := other.DeleteSessionsByIDs(ctx, []uuid.UUID{s.ID})
			require.NoError(t, err)
			assert.Equal(t, 0, n)
			assert.True(t, exists(t, s.ID))
		})

		t.Run("case=DeleteAllSessions removes only this network's rows", func(t *testing.T) {
			_, scoped := testhelpers.NewNetwork(t, ctx, p)
			s1 := newActiveSession(t)
			s2 := newActiveSession(t)

			var foreign session.Session
			require.NoError(t, faker.FakeData(&foreign))
			foreign.Active = true
			require.NoError(t, scoped.CreateIdentity(ctx, foreign.Identity))
			require.NoError(t, scoped.UpsertSession(ctx, &foreign))

			n, err := p.DeleteAllSessions(ctx, 100)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, n, 2)

			assert.False(t, exists(t, s1.ID))
			assert.False(t, exists(t, s2.ID))

			fs, err := scoped.GetSession(ctx, foreign.ID, session.ExpandNothing)
			require.NoError(t, err)
			assert.NotNil(t, fs, "foreign-network session must survive")
		})
	}
}
