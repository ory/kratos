// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/events"
	"github.com/ory/pop/v6"
	"github.com/ory/x/dbal"
	"github.com/ory/x/otelx"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/popx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/stringsx"
)

var _ session.Persister = new(Persister)

const (
	SessionDeviceUserAgentMaxLength = 512
	SessionDeviceLocationMaxLength  = 512
	paginationMaxItemsSize          = 1000
	paginationDefaultItemsSize      = 250
)

func (p *Persister) GetSession(ctx context.Context, sid uuid.UUID, expandables session.Expandables) (_ *session.Session, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetSession")
	defer otelx.End(span, &err)

	var s session.Session
	s.Devices = make([]session.Device, 0)
	nid := p.NetworkID(ctx)

	q := p.GetConnection(ctx).Q()
	// if len(expandables) > 0 {
	if expandables.Has(session.ExpandSessionDevices) {
		q = q.Eager(expandables.ToEager()...)
	}

	if err := q.Where("id = ? AND nid = ?", sid, nid).First(&s); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if expandables.Has(session.ExpandSessionIdentity) {
		// This is needed because of how identities are fetched from the store (if we use eager not all fields are
		// available!).
		i, err := p.GetIdentity(ctx, s.IdentityID, identity.ExpandDefault)
		if err != nil {
			return nil, err
		}
		s.Identity = i
	}

	s.Active = s.IsActive()
	return &s, nil
}

func (p *Persister) ListSessions(ctx context.Context, active *bool, paginatorOpts []keysetpagination.Option, expandables session.Expandables) (_ []session.Session, _ *keysetpagination.Paginator, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListSessions")
	defer otelx.End(span, &err)

	var s []session.Session
	nid := p.NetworkID(ctx)

	paginatorOpts = append(paginatorOpts, keysetpagination.WithDefaultSize(paginationDefaultItemsSize))
	paginatorOpts = append(paginatorOpts, keysetpagination.WithMaxSize(paginationMaxItemsSize))
	paginatorOpts = append(paginatorOpts, keysetpagination.WithDefaultToken(new(session.Session).DefaultPageToken()))
	paginatorOpts = append(paginatorOpts, keysetpagination.WithColumn("created_at", "DESC"))
	paginator := keysetpagination.GetPaginator(paginatorOpts...)

	if _, err := uuid.FromString(paginator.Token().Parse("id")["id"]); err != nil {
		return nil, nil, errors.WithStack(x.PageTokenInvalid)
	}

	if err := p.listWithinReadCommittedReadOnlyTx(ctx, func(ctx context.Context, c *pop.Connection) error {
		s = make([]session.Session, 0)

		q := c.Where("nid = ?", nid)
		if active != nil {
			if *active {
				q.Where("active = ? AND expires_at >= ?", *active, time.Now().UTC())
			} else {
				q.Where("(active = ? OR expires_at < ?)", *active, time.Now().UTC())
			}
		}

		if len(expandables) > 0 {
			q = q.EagerPreload(expandables.ToEager()...)
		}

		if err := q.Scope(keysetpagination.Paginate[session.Session](paginator)).All(&s); err != nil {
			return sqlcon.HandleError(err)
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	for k := range s {
		if s[k].Identity == nil {
			continue
		}
		if err := p.InjectTraitsSchemaURL(ctx, s[k].Identity); err != nil {
			return nil, nil, err
		}
	}

	s, nextPage := keysetpagination.Result(s, paginator)
	return s, nextPage, nil
}

// ListSessionsByIdentity retrieves sessions for an identity from the store.
func (p *Persister) ListSessionsByIdentity(
	ctx context.Context,
	iID uuid.UUID,
	active *bool,
	page, perPage int,
	except uuid.UUID,
	expandables session.Expandables,
) (_ []session.Session, _ int64, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListSessionsByIdentity")
	defer otelx.End(span, &err)

	var s []session.Session
	t := int64(0)
	nid := p.NetworkID(ctx)

	if err := p.listWithinReadCommittedReadOnlyTx(ctx, func(ctx context.Context, c *pop.Connection) error {
		s = make([]session.Session, 0)

		q := c.Where("identity_id = ? AND nid = ?", iID, nid)
		if except != uuid.Nil {
			q = q.Where("id != ?", except)
		}
		if active != nil {
			if *active {
				q.Where("active = ? AND expires_at >= ?", *active, time.Now().UTC())
			} else {
				q.Where("(active = ? OR expires_at < ?)", *active, time.Now().UTC())
			}
		}

		if len(expandables) > 0 {
			q = q.EagerPreload(expandables.ToEager()...)
		}

		total, err := q.Count(new(session.Session))
		if err != nil {
			return sqlcon.HandleError(err)
		}
		t = int64(total)

		q.Order("created_at DESC")

		if err := q.Paginate(page, perPage).All(&s); err != nil {
			return sqlcon.HandleError(err)
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	return s, t, nil
}

// ExtendSession updates the expiry of a session.
func (p *Persister) ExtendSession(ctx context.Context, sessionID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ExtendSession")
	defer otelx.End(span, &err)

	nid := p.NetworkID(ctx)
	s := new(session.Session)
	var didRefresh bool
	if err := errors.WithStack(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) (err error) {
		lockBehavior := ""
		if tx.Dialect.Name() == dbal.DriverCockroachDB {
			// SKIP LOCKED returns no rows if the row is locked by another transaction.
			lockBehavior = "FOR UPDATE SKIP LOCKED"
		}

		if err := tx.
			Where(
				// We make use of the fact that CRDB supports FOR UPDATE as part of the WHERE clause.
				fmt.Sprintf("id = ? AND nid = ? %s", lockBehavior),
				sessionID, nid,
			).First(s); err != nil {

			// This is a special case for CockroachDB. If the row is locked, we do not see the session. Therefor we return
			// a 404 not found error indicating to the user that the session might already be updated by someone else.
			if errors.Is(err, sqlcon.ErrNoRows()) && tx.Dialect.Name() == dbal.DriverCockroachDB {
				return errors.WithStack(herodot.ErrNotFound().WithReason("The session you are trying to extend is already being extended by another request or does not exist."))
			}

			return sqlcon.HandleError(err)
		}

		if !s.CanBeRefreshed(ctx, p.r.Config()) {
			// This prevents excessive writes to the database.
			return nil
		}

		didRefresh = true
		s = s.Refresh(ctx, p.r.Config())

		if _, err := tx.Where("id = ? AND nid = ?", sessionID, nid).UpdateQuery(s, "expires_at"); err != nil {
			return sqlcon.HandleError(err)
		}

		return nil
	})); err != nil {
		return err
	}

	if didRefresh {
		trace.SpanFromContext(ctx).AddEvent(events.NewSessionLifespanExtended(ctx, s.ID, s.IdentityID, s.ExpiresAt))
	}

	return nil
}

// UpsertSession creates a session if not found else updates.
// This operation also inserts Session device records when a session is being created.
// The update operation skips updating Session device records since only one record would need to be updated in this case.
func (p *Persister) UpsertSession(ctx context.Context, s *session.Session) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpsertSession")
	defer otelx.End(span, &err)

	s.NID = p.NetworkID(ctx)
	if s.Identity != nil {
		s.IdentityID = s.Identity.ID
	} else if s.IdentityID.IsNil() {
		return errors.WithStack(herodot.ErrInternalServerError().WithReasonf("cannot upsert session without an identity or identity ID set"))
	}

	var updated bool
	defer func() {
		if err != nil {
			return
		}
		if updated {
			trace.SpanFromContext(ctx).AddEvent(events.NewSessionChanged(ctx, string(s.AuthenticatorAssuranceLevel), s.ID, s.IdentityID))
		} else {
			trace.SpanFromContext(ctx).AddEvent(events.NewSessionIssued(ctx, string(s.AuthenticatorAssuranceLevel), s.ID, s.IdentityID))
		}
	}()

	return errors.WithStack(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) (err error) {
		updated = false
		exists := false
		if !s.ID.IsNil() {
			exists, err = tx.Where("id = ? AND nid = ?", s.ID, s.NID).Exists(new(session.Session))
			if err != nil {
				return sqlcon.HandleError(err)
			}
		}

		if exists {
			// This must not be eager or identities will be created / updated
			// Only update session and not corresponding session device records
			if err := tx.Update(s, "issued_at", "identity_id", "nid"); err != nil {
				return sqlcon.HandleError(err)
			}
			updated = true
			return nil
		}

		// This must not be eager or identities will be created / updated
		if err := sqlcon.HandleError(tx.Create(s)); err != nil {
			return err
		}

		for i := range s.Devices {
			device := &(s.Devices[i])
			device.SessionID = s.ID
			device.NID = s.NID
			device.IdentityID = new(s.IdentityID)

			if device.Location != nil {
				device.Location = new(stringsx.TruncateByteLen(*device.Location, SessionDeviceLocationMaxLength))
			}
			if device.UserAgent != nil {
				device.UserAgent = new(stringsx.TruncateByteLen(*device.UserAgent, SessionDeviceUserAgentMaxLength))
			}

			if err := p.CreateDevice(ctx, device); err != nil {
				return err
			}
		}

		return nil
	}))
}

// DeleteSession permanently deletes a single session. Returns
// sqlcon.ErrNoRows() when no matching session is found in the caller's
// network.
func (p *Persister) DeleteSession(ctx context.Context, sid uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteSession")
	defer otelx.End(span, &err)

	n, err := p.DeleteSessionsByIDs(ctx, []uuid.UUID{sid})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.WithStack(sqlcon.ErrNoRows())
	}
	return nil
}

func (p *Persister) DeleteSessionsByIdentity(ctx context.Context, identityID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteSessionsByIdentity")
	defer otelx.End(span, &err)

	//#nosec G201 -- TableName is static
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE identity_id = ? AND nid = ?",
		session.Session{}.TableName(),
	),
		identityID,
		p.NetworkID(ctx),
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows())
	}
	return nil
}

func (p *Persister) GetSessionByToken(ctx context.Context, token string, expand session.Expandables, identityExpand identity.Expandables) (res *session.Session, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetSessionByToken")
	defer otelx.End(span, &err)

	var s session.Session
	s.Devices = make([]session.Device, 0)
	nid := p.NetworkID(ctx)

	con := p.GetConnection(ctx)
	if err := con.Where("token = ? AND nid = ?", token, nid).First(&s); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	var (
		i  *identity.Identity
		sd []session.Device
	)

	eg, ctx := errgroup.WithContext(ctx)
	if expand.Has(session.ExpandSessionDevices) {
		eg.Go(func() error {
			return sqlcon.HandleError(con.WithContext(ctx).
				Where("session_id = ? AND nid = ?", s.ID, nid).All(&sd))
		})
	}

	// This is needed because of how identities are fetched from the store (if we use eager not all fields are
	// available!).
	if expand.Has(session.ExpandSessionIdentity) {
		eg.Go(func() (err error) {
			i, err = p.GetIdentity(ctx, s.IdentityID, identityExpand)
			return err
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	s.Identity = i
	s.Devices = sd

	return &s, nil
}

func (p *Persister) DeleteSessionByToken(ctx context.Context, token string) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteSessionByToken")
	defer otelx.End(span, &err)

	//#nosec G201 -- TableName is static
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE token = ? AND nid = ?",
		session.Session{}.TableName(),
	),
		token,
		p.NetworkID(ctx),
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows())
	}
	return nil
}

// RevokeSessionByToken marks the session with the given token inactive and
// returns its IDs so the caller can attach them to observability events without
// a separate GetSessionByToken round trip.
// Returns sqlcon.ErrNoRows() when no matching session exists in the caller's
// network; the returned RevokedSession is the zero value in that case.
func (p *Persister) RevokeSessionByToken(ctx context.Context, token string) (revoked session.RevokedSession, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSessionByToken")
	defer otelx.End(span, &err)

	con := p.GetConnection(ctx)
	nid := p.NetworkID(ctx)
	var dst struct {
		ID         uuid.UUID `db:"id"`
		IdentityID uuid.UUID `db:"identity_id"`
	}

	switch con.Dialect.Name() {
	case dbal.DriverCockroachDB, dbal.DriverPostgreSQL:
		// CTE: identify the row by (token, nid), conditionally flip active=false
		// only when currently true, and return the matched row's identifiers in
		// one round trip. See revokeMatchingSessions for the contention rationale.
		const query = `WITH found AS (SELECT id, identity_id FROM sessions WHERE token = ? AND nid = ?),
     upd AS (UPDATE sessions SET active = false FROM found WHERE sessions.id = found.id AND sessions.active = true RETURNING 1)
SELECT id, identity_id FROM found`

		err = p.runInReadCommittedOnCRDB(ctx, func(c *pop.Connection) error {
			return c.RawQuery(query, token, nid).First(&dst)
		})
	default:
		// SQLite and MySQL: data-modifying CTEs are not portable here, so issue
		// a separate SELECT followed by the legacy UPDATE. Same two-statement
		// shape as today's caller (GetSessionByToken + RevokeSessionByToken),
		// so no regression on these dialects.
		err = con.RawQuery("SELECT id, identity_id FROM sessions WHERE token = ? AND nid = ?", token, nid).First(&dst)
		if err == nil {
			err = con.RawQuery("UPDATE sessions SET active = false WHERE token = ? AND nid = ?", token, nid).Exec()
		}
	}
	if err != nil {
		return session.RevokedSession{}, sqlcon.HandleError(err)
	}
	return session.RevokedSession{ID: dst.ID, IdentityID: dst.IdentityID}, nil
}

// RevokeSessionById revokes a given session. Returns sqlcon.ErrNoRows() when
// no matching session is found in the caller's network.
func (p *Persister) RevokeSessionById(ctx context.Context, sID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSessionById")
	defer otelx.End(span, &err)

	count, err := p.revokeMatchingSessions(ctx, "id = ? AND nid = ?", sID, p.NetworkID(ctx))
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows())
	}
	return nil
}

// RevokeSession revokes a given session. If the session does not exist or was not modified,
// it effectively has been revoked already, and therefore that case does not return an error.
func (p *Persister) RevokeSession(ctx context.Context, iID, sID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSession")
	defer otelx.End(span, &err)

	return p.runInReadCommittedOnCRDB(ctx, func(c *pop.Connection) error {
		return c.RawQuery(
			"UPDATE sessions SET active = false WHERE id = ? AND identity_id = ? AND nid = ? AND active = true",
			sID, iID, p.NetworkID(ctx),
		).Exec()
	})
}

// RevokeSessionsIdentityExcept marks all except the given session of an identity inactive.
func (p *Persister) RevokeSessionsIdentityExcept(ctx context.Context, iID, sID uuid.UUID) (res int, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSessionsIdentityExcept")
	defer otelx.End(span, &err)

	return p.revokeMatchingSessions(ctx, "identity_id = ? AND id != ? AND nid = ?", iID, sID, p.NetworkID(ctx))
}

// RevokeSessionsByIdentities marks all currently active sessions inactive for the given identity IDs.
func (p *Persister) RevokeSessionsByIdentities(ctx context.Context, identityIDs []uuid.UUID) (count int, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSessionsByIdentities")
	defer otelx.End(span, &err)

	if len(identityIDs) == 0 {
		return 0, nil
	}

	err = p.runInReadCommittedOnCRDB(ctx, func(c *pop.Connection) error {
		var inner error
		count, inner = c.RawQuery(
			"UPDATE sessions SET active = false WHERE identity_id IN (?) AND active = true AND nid = ?",
			identityIDs, p.NetworkID(ctx),
		).ExecWithCount()
		return inner
	})
	if err != nil {
		return 0, sqlcon.HandleError(err)
	}
	return count, nil
}

// RevokeSessionsByIDs marks the listed sessions inactive (only ones currently active).
func (p *Persister) RevokeSessionsByIDs(ctx context.Context, sessionIDs []uuid.UUID) (count int, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSessionsByIDs")
	defer otelx.End(span, &err)

	if len(sessionIDs) == 0 {
		return 0, nil
	}

	err = p.runInReadCommittedOnCRDB(ctx, func(c *pop.Connection) error {
		var inner error
		count, inner = c.RawQuery(
			"UPDATE sessions SET active = false WHERE id IN (?) AND active = true AND nid = ?",
			sessionIDs, p.NetworkID(ctx),
		).ExecWithCount()
		return inner
	})
	if err != nil {
		return 0, sqlcon.HandleError(err)
	}
	return count, nil
}

// RevokeAllSessions deactivates up to `limit` currently-active sessions in
// the caller's network in a single SQL statement and returns the number of
// rows actually updated (in the range [0, limit]).
func (p *Persister) RevokeAllSessions(ctx context.Context, limit int) (count int, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeAllSessions")
	defer otelx.End(span, &err)

	err = p.runInReadCommittedOnCRDB(ctx, func(c *pop.Connection) error {
		var inner error
		count, inner = c.RawQuery(
			"UPDATE sessions SET active = false WHERE id IN (SELECT id FROM (SELECT id FROM sessions c WHERE active = true AND nid = ? LIMIT ?) AS s)",
			p.NetworkID(ctx), limit,
		).ExecWithCount()
		return inner
	})
	if err != nil {
		return 0, sqlcon.HandleError(err)
	}
	return count, nil
}

// revokeMatchingSessions flips active=false on every sessions row matching the
// given WHERE clause and returns the number of rows that matched the clause
// before the update (irrespective of whether they were already inactive). Used
// by the revoke entry points whose caller relies on the matched-count
// distinction — either to map count==0 to sqlcon.ErrNoRows() or to surface
// the count as an API response field.
//
// On CockroachDB and PostgreSQL, the update is expressed as a CTE so that
// only rows whose current active=true get rewritten. CockroachDB's sessions
// table is LOCALITY GLOBAL: every write commits at a future timestamp and
// any concurrent write to the same key inside that closed-timestamp window
// raises WriteTooOldError + a serializable refresh. Suppressing the
// redundant rewrite of already-inactive rows eliminates the dominant
// same-fingerprint contention pattern observed in production (concurrent
// revokes for the same token from logout retries / multiple clients).
// SQLite and MySQL retain the single-statement UPDATE; neither is subject
// to the GLOBAL-table closed-timestamp contention.
func (p *Persister) revokeMatchingSessions(ctx context.Context, predicate string, args ...any) (int, error) {
	con := p.GetConnection(ctx)
	var (
		count int
		err   error
	)

	switch con.Dialect.Name() {
	case dbal.DriverCockroachDB, dbal.DriverPostgreSQL:
		//#nosec G201 -- predicate is a static persister-internal constant, not user input
		query := fmt.Sprintf(`WITH found AS (SELECT id FROM sessions WHERE %s),
     upd AS (UPDATE sessions SET active = false FROM found WHERE sessions.id = found.id AND sessions.active = true RETURNING 1)
SELECT count(*) FROM found`, predicate)

		err = p.runInReadCommittedOnCRDB(ctx, func(c *pop.Connection) error {
			return c.RawQuery(query, args...).First(&count)
		})
	default:
		//#nosec G201 -- predicate is a static persister-internal constant, not user input
		count, err = con.RawQuery(
			fmt.Sprintf("UPDATE sessions SET active = false WHERE %s", predicate),
			args...,
		).ExecWithCount()
	}
	if err != nil {
		return 0, sqlcon.HandleError(err)
	}
	return count, nil
}

// runInReadCommittedOnCRDB invokes fn against a connection bound to a READ
// COMMITTED transaction on CockroachDB, and against the bare persister
// connection on every other dialect. CRDB needs the explicit isolation
// downgrade so the cluster transparently advances the statement read
// timestamp on WriteTooOldError instead of surfacing the serializable
// retry. Postgres already runs single statements at READ COMMITTED by
// default; SQLite and MySQL are not subject to the GLOBAL-table
// closed-timestamp contention.
func (p *Persister) runInReadCommittedOnCRDB(ctx context.Context, fn func(*pop.Connection) error) error {
	con := p.GetConnection(ctx)
	if con.Dialect.Name() != dbal.DriverCockroachDB {
		return fn(con)
	}
	return popx.TransactionWithOptions(ctx, con,
		&sql.TxOptions{Isolation: sql.LevelReadCommitted},
		func(ctx context.Context, tx *pop.Connection) error { return fn(tx) },
	)
}

// listWithinReadCommittedReadOnlyTx invokes fn, which must only read, inside a READ
// COMMITTED read-only transaction on CockroachDB and PostgreSQL, and inside a
// regular transaction on every other dialect. The paginated session list
// queries run more than one statement per transaction (count + page, or page
// + eager preloads); under SERIALIZABLE a concurrent write to the scanned
// rows — a login upsert, extension, or revocation — between those statements
// invalidates the transaction's read timestamp and surfaces as a client-side
// RETRY_SERIALIZABLE retry. READ COMMITTED gives each statement its own read
// timestamp so CockroachDB absorbs those conflicts server-side. READ
// COMMITTED is already the PostgreSQL default; there the options make the
// read-only intent explicit and let the server skip write-path bookkeeping.
// SQLite and MySQL keep the plain transaction.
func (p *Persister) listWithinReadCommittedReadOnlyTx(ctx context.Context, fn func(ctx context.Context, c *pop.Connection) error) error {
	con := p.GetConnection(ctx)
	switch con.Dialect.Name() {
	case dbal.DriverCockroachDB, dbal.DriverPostgreSQL:
		return popx.TransactionWithOptions(ctx, con,
			&sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true},
			fn,
		)
	default:
		return p.Transaction(ctx, fn)
	}
}

// DeleteSessionsByIdentities permanently deletes all sessions belonging to the given identity IDs.
func (p *Persister) DeleteSessionsByIdentities(ctx context.Context, identityIDs []uuid.UUID) (count int, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteSessionsByIdentities")
	defer otelx.End(span, &err)

	if len(identityIDs) == 0 {
		return 0, nil
	}

	//#nosec G201 -- TableName is static.
	count, err = p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE identity_id IN (?) AND nid = ?",
		session.Session{}.TableName(),
	), identityIDs, p.NetworkID(ctx)).ExecWithCount()
	if err != nil {
		return 0, sqlcon.HandleError(err)
	}
	return count, nil
}

// DeleteSessionsByIDs permanently deletes the listed sessions.
func (p *Persister) DeleteSessionsByIDs(ctx context.Context, sessionIDs []uuid.UUID) (count int, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteSessionsByIDs")
	defer otelx.End(span, &err)

	if len(sessionIDs) == 0 {
		return 0, nil
	}

	//#nosec G201 -- TableName is static.
	count, err = p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE id IN (?) AND nid = ?",
		session.Session{}.TableName(),
	), sessionIDs, p.NetworkID(ctx)).ExecWithCount()
	if err != nil {
		return 0, sqlcon.HandleError(err)
	}
	return count, nil
}

// DeleteAllSessions permanently deletes up to `limit` sessions in the
// caller's network in a single SQL statement and returns the number of rows
// actually deleted (in the range [0, limit]).
func (p *Persister) DeleteAllSessions(ctx context.Context, limit int) (count int, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteAllSessions")
	defer otelx.End(span, &err)

	//#nosec G201 -- TableName is static
	count, err = p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %[1]s WHERE id IN (SELECT id FROM (SELECT id FROM %[1]s c WHERE nid = ? LIMIT ?) AS s)",
		session.Session{}.TableName(),
	), p.NetworkID(ctx), limit).ExecWithCount()
	if err != nil {
		return 0, sqlcon.HandleError(err)
	}
	return count, nil
}

func (p *Persister) DeleteExpiredSessions(ctx context.Context, expiresAt time.Time, limit int) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteExpiredSessions")
	defer otelx.End(span, &err)

	//#nosec G201 -- TableName is static
	err = p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %[1]s WHERE id in (SELECT id FROM (SELECT id FROM %[1]s c WHERE expires_at <= ? and nid = ? ORDER BY expires_at ASC LIMIT ?) AS s)",
		session.Session{}.TableName(),
	),
		expiresAt,
		p.NetworkID(ctx),
		limit,
	).Exec()

	return sqlcon.HandleError(err)
}
