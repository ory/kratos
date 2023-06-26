// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x/events"
	"github.com/ory/x/otelx"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/stringsx"
)

var _ session.Persister = new(Persister)

const SessionDeviceUserAgentMaxLength = 512
const SessionDeviceLocationMaxLength = 512
const paginationMaxItemsSize = 1000
const paginationDefaultItemsSize = 250

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
		i, err := p.PrivilegedPool.GetIdentity(ctx, s.IdentityID, identity.ExpandDefault)
		if err != nil {
			return nil, err
		}
		s.Identity = i
	}

	s.Active = s.IsActive()
	return &s, nil
}

func (p *Persister) ListSessions(ctx context.Context, active *bool, paginatorOpts []keysetpagination.Option, expandables session.Expandables) (_ []session.Session, _ int64, _ *keysetpagination.Paginator, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListSessions")
	defer otelx.End(span, &err)

	s := make([]session.Session, 0)
	t := int64(0)
	nid := p.NetworkID(ctx)

	paginatorOpts = append(paginatorOpts, keysetpagination.WithDefaultSize(paginationDefaultItemsSize))
	paginatorOpts = append(paginatorOpts, keysetpagination.WithMaxSize(paginationMaxItemsSize))
	paginatorOpts = append(paginatorOpts, keysetpagination.WithDefaultToken(new(session.Session).DefaultPageToken()))
	paginator := keysetpagination.GetPaginator(paginatorOpts...)

	if err := p.Transaction(ctx, func(ctx context.Context, c *pop.Connection) error {
		q := c.Where("nid = ?", nid)
		if active != nil {
			if *active {
				q.Where("active = ? AND expires_at >= ?", *active, time.Now().UTC())
			} else {
				q.Where("(active = ? OR expires_at < ?)", *active, time.Now().UTC())
			}
		}

		// Get the total count of matching items
		total, err := q.Count(new(session.Session))
		if err != nil {
			return sqlcon.HandleError(err)
		}
		t = int64(total)

		if len(expandables) > 0 {
			q = q.EagerPreload(expandables.ToEager()...)
		}

		// Get the paginated list of matching items
		if err := q.Scope(keysetpagination.Paginate[session.Session](paginator)).All(&s); err != nil {
			return sqlcon.HandleError(err)
		}

		return nil
	}); err != nil {
		return nil, 0, nil, err
	}

	for k := range s {
		if s[k].Identity == nil {
			continue
		}
		if err := p.InjectTraitsSchemaURL(ctx, s[k].Identity); err != nil {
			return nil, 0, nil, err
		}
	}

	s, nextPage := keysetpagination.Result(s, paginator)
	return s, t, nextPage, nil
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

	s := make([]session.Session, 0)
	t := int64(0)
	nid := p.NetworkID(ctx)

	if err := p.Transaction(ctx, func(ctx context.Context, c *pop.Connection) error {
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

		// Get the total count of matching items
		total, err := q.Count(new(session.Session))
		if err != nil {
			return sqlcon.HandleError(err)
		}
		t = int64(total)

		q.Order("authenticated_at DESC")

		// Get the paginated list of matching items
		if err := q.Paginate(page, perPage).All(&s); err != nil {
			return sqlcon.HandleError(err)
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	return s, t, nil
}

// UpsertSession creates a session if not found else updates.
// This operation also inserts Session device records when a session is being created.
// The update operation skips updating Session device records since only one record would need to be updated in this case.
func (p *Persister) UpsertSession(ctx context.Context, s *session.Session) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpsertSession")
	defer otelx.End(span, &err)

	s.NID = p.NetworkID(ctx)

	return errors.WithStack(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		exists := false
		if !s.ID.IsNil() {
			var err error
			exists, err = tx.Where("id = ? AND nid = ?", s.ID, s.NID).Exists(new(session.Session))
			if err != nil {
				return sqlcon.HandleError(err)
			}
		}

		if exists {
			// This must not be eager or identities will be created / updated
			// Only update session and not corresponding session device records
			if err := tx.Update(s); err != nil {
				return sqlcon.HandleError(err)
			}
			trace.SpanFromContext(ctx).AddEvent(events.NewSessionChanged(ctx, string(s.AuthenticatorAssuranceLevel), s.ID, s.IdentityID))
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

			if device.Location != nil {
				device.Location = stringsx.GetPointer(stringsx.TruncateByteLen(*device.Location, SessionDeviceLocationMaxLength))
			}
			if device.UserAgent != nil {
				device.UserAgent = stringsx.GetPointer(stringsx.TruncateByteLen(*device.UserAgent, SessionDeviceUserAgentMaxLength))
			}

			if err := p.DevicePersister.CreateDevice(ctx, device); err != nil {
				return err
			}
		}

		trace.SpanFromContext(ctx).AddEvent(events.NewSessionIssued(ctx, string(s.AuthenticatorAssuranceLevel), s.ID, s.IdentityID))
		return nil
	}))
}

func (p *Persister) DeleteSession(ctx context.Context, sid uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteSession")
	defer otelx.End(span, &err)

	nid := p.NetworkID(ctx)
	//#nosec G201 -- TableName is static
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE id = ? AND nid = ?", new(session.Session).TableName(ctx)),
		sid,
		nid,
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}
	return nil
}

func (p *Persister) DeleteSessionsByIdentity(ctx context.Context, identityID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteSessionsByIdentity")
	defer otelx.End(span, &err)

	//#nosec G201 -- TableName is static
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE identity_id = ? AND nid = ?",
		new(session.Session).TableName(ctx),
	),
		identityID,
		p.NetworkID(ctx),
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
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
			i, err = p.PrivilegedPool.GetIdentity(ctx, s.IdentityID, identityExpand)
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
		new(session.Session).TableName(ctx),
	),
		token,
		p.NetworkID(ctx),
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}
	return nil
}

func (p *Persister) RevokeSessionByToken(ctx context.Context, token string) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSessionByToken")
	defer otelx.End(span, &err)

	//#nosec G201 -- TableName is static
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"UPDATE %s SET active = false WHERE token = ? AND nid = ?",
		new(session.Session).TableName(ctx),
	),
		token,
		p.NetworkID(ctx),
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}
	return nil
}

// RevokeSessionById revokes a given session
func (p *Persister) RevokeSessionById(ctx context.Context, sID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSessionById")
	defer otelx.End(span, &err)

	//#nosec G201 -- TableName is static
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"UPDATE %s SET active = false WHERE id = ? AND nid = ?",
		new(session.Session).TableName(ctx),
	),
		sID,
		p.NetworkID(ctx),
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}
	return nil
}

// RevokeSession revokes a given session. If the session does not exist or was not modified,
// it effectively has been revoked already, and therefore that case does not return an error.
func (p *Persister) RevokeSession(ctx context.Context, iID, sID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSession")
	defer otelx.End(span, &err)

	//#nosec G201 -- TableName is static
	err = p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"UPDATE %s SET active = false WHERE id = ? AND identity_id = ? AND nid = ?",
		new(session.Session).TableName(ctx),
	),
		sID,
		iID,
		p.NetworkID(ctx),
	).Exec()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	return nil
}

// RevokeSessionsIdentityExcept marks all except the given session of an identity inactive.
func (p *Persister) RevokeSessionsIdentityExcept(ctx context.Context, iID, sID uuid.UUID) (res int, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSessionsIdentityExcept")
	defer otelx.End(span, &err)

	//#nosec G201 -- TableName is static
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"UPDATE %s SET active = false WHERE identity_id = ? AND id != ? AND nid = ?",
		new(session.Session).TableName(ctx),
	),
		iID,
		sID,
		p.NetworkID(ctx),
	).ExecWithCount()
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
		"DELETE FROM %s WHERE id in (SELECT id FROM (SELECT id FROM %s c WHERE expires_at <= ? and nid = ? ORDER BY expires_at ASC LIMIT %d ) AS s )",
		new(session.Session).TableName(ctx),
		new(session.Session).TableName(ctx),
		limit,
	),
		expiresAt,
		p.NetworkID(ctx),
	).Exec()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	return nil
}
