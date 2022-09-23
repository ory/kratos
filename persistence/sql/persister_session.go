package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/ory/x/stringsx"

	"github.com/gobuffalo/pop/v6"

	"github.com/pkg/errors"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/session"
)

var _ session.Persister = new(Persister)

const SessionDeviceUserAgentMaxLength = 512
const SessionDeviceLocationMaxLength = 512

func (p *Persister) GetSession(ctx context.Context, sid uuid.UUID, expandables session.Expandables) (*session.Session, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetSession")
	defer span.End()

	var s session.Session
	s.Devices = make([]session.Device, 0)
	nid := p.NetworkID(ctx)

	q := p.GetConnection(ctx).Q()
	if len(expandables) > 0 {
		q = q.EagerPreload(expandables.ToEager()...)
	}

	if err := q.Where("id = ? AND nid = ?", sid, nid).First(&s); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	// This is needed because of how identities are fetched from the store (if we use eager not all fields are
	// available!).
	i, err := p.GetIdentity(ctx, s.IdentityID)
	if err != nil {
		return nil, err
	}

	s.Identity = i
	return &s, nil
}

// ListSessionsByIdentity retrieves sessions for an identity from the store.
func (p *Persister) ListSessionsByIdentity(ctx context.Context, iID uuid.UUID, active *bool, page, perPage int, except uuid.UUID) ([]*session.Session, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListSessionsByIdentity")
	defer span.End()

	s := make([]*session.Session, 0)
	nid := p.NetworkID(ctx)

	if err := p.Transaction(ctx, func(ctx context.Context, c *pop.Connection) error {
		q := c.Where("identity_id = ? AND nid = ?", iID, nid).Paginate(page, perPage)
		if except != uuid.Nil {
			q = q.Where("id != ?", except)
		}
		if active != nil {
			q = q.Where("active = ?", *active)
		}
		if err := q.Eager("Devices").All(&s); err != nil {
			return sqlcon.HandleError(err)
		}

		for _, s := range s {
			i, err := p.GetIdentity(ctx, s.IdentityID)
			if err != nil {
				return err
			}

			s.Identity = i
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s, nil
}

// UpsertSession creates a session if not found else updates.
// This operation also inserts Session device records when a session is being created.
// The update operation skips updating Session device records since only one record would need to be updated in this case.
func (p *Persister) UpsertSession(ctx context.Context, s *session.Session) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpsertSession")
	defer span.End()

	s.NID = p.NetworkID(ctx)

	return errors.WithStack(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		if err := sqlcon.HandleError(tx.Where("id = ? AND nid = ?", s.ID, s.NID).First(new(session.Session))); errors.Is(err, sqlcon.ErrNoRows) {
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

				if err := sqlcon.HandleError(tx.Create(device)); err != nil {
					return err
				}

				s.Devices[i] = *device
			}

			return nil
		} else if err != nil {
			return err
		}

		// This must not be eager or identities will be created / updated
		// Only update session and not corresponding session device records
		return tx.Update(s)
	}))
}

func (p *Persister) DeleteSession(ctx context.Context, sid uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteSession")
	defer span.End()

	return p.delete(ctx, new(session.Session), sid)
}

func (p *Persister) DeleteSessionsByIdentity(ctx context.Context, identityID uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteSessionsByIdentity")
	defer span.End()

	// #nosec G201
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE identity_id = ? AND nid = ?",
		"sessions",
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

func (p *Persister) GetSessionByToken(ctx context.Context, token string) (*session.Session, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetSessionByToken")
	defer span.End()

	var s session.Session
	if err := p.GetConnection(ctx).Where("token = ? AND nid = ?",
		token,
		p.NetworkID(ctx),
	).First(&s); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	// This is needed because of how identities are fetched from the store (if we use eager not all fields are
	// available!).
	i, err := p.GetIdentity(ctx, s.IdentityID)
	if err != nil {
		return nil, err
	}
	s.Identity = i
	return &s, nil
}

func (p *Persister) DeleteSessionByToken(ctx context.Context, token string) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteSessionByToken")
	defer span.End()

	// #nosec G201
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE token = ? AND nid = ?",
		"sessions",
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

func (p *Persister) RevokeSessionByToken(ctx context.Context, token string) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSessionByToken")
	defer span.End()

	// #nosec G201
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"UPDATE %s SET active = false WHERE token = ? AND nid = ?",
		"sessions",
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

// RevokeSession revokes a given session. If the session does not exist or was not modified,
// it effectively has been revoked already, and therefore that case does not return an error.
func (p *Persister) RevokeSession(ctx context.Context, iID, sID uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSession")
	defer span.End()

	// #nosec G201
	err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"UPDATE %s SET active = false WHERE id = ? AND identity_id = ? AND nid = ?",
		"sessions",
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
func (p *Persister) RevokeSessionsIdentityExcept(ctx context.Context, iID, sID uuid.UUID) (int, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.RevokeSessionsIdentityExcept")
	defer span.End()

	// #nosec G201
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"UPDATE %s SET active = false WHERE identity_id = ? AND id != ? AND nid = ?",
		"sessions",
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

func (p *Persister) DeleteExpiredSessions(ctx context.Context, expiresAt time.Time, limit int) error {
	err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE id in (SELECT id FROM (SELECT id FROM %s c WHERE expires_at <= ? and nid = ? ORDER BY expires_at ASC LIMIT %d ) AS s )",
		"sessions",
		"sessions",
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
