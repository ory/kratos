package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"

	"github.com/pkg/errors"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/session"
)

var _ session.Persister = new(Persister)

func (p *Persister) GetSession(ctx context.Context, sid uuid.UUID) (*session.Session, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetSession")
	defer span.End()

	var s session.Session
	nid := p.NetworkID(ctx)
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", sid, nid).First(&s); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	// This is needed because of how identities are fetched from the store (if we use eager not all fields are
	// available!).
	i, err := p.GetIdentity(ctx, s.IdentityID)
	if err != nil {
		return nil, err
	}

	devices, err := p.GetSessionDevices(ctx, sid)
	if err != nil {
		return nil, err
	}

	s.Identity = i
	s.Devices = devices
	return &s, nil
}

func (p *Persister) GetSessionDevices(ctx context.Context, sid uuid.UUID) ([]session.Device, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetSessionDevices")
	defer span.End()

	devices := make([]session.Device, 0)
	nid := p.NetworkID(ctx)
	if err := p.GetConnection(ctx).Where("session_id = ? AND nid = ?", sid, nid).All(&devices); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return devices, nil
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
		if err := q.All(&s); err != nil {
			return sqlcon.HandleError(err)
		}

		for _, s := range s {
			i, err := p.GetIdentity(ctx, s.IdentityID)
			if err != nil {
				return err
			}

			devices, err := p.GetSessionDevices(ctx, s.ID)
			if err != nil {
				return err
			}

			s.Identity = i
			s.Devices = devices
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s, nil
}

func (p *Persister) UpsertSession(ctx context.Context, s *session.Session) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpsertSession")
	defer span.End()

	s.NID = p.NetworkID(ctx)

	if err := p.Connection(ctx).Find(new(session.Session), s.ID); errors.Is(err, sql.ErrNoRows) {
		// This must not be eager or identities will be created / updated
		return errors.WithStack(p.GetConnection(ctx).Create(s))
	} else if err != nil {
		return errors.WithStack(err)
	}

	// This must not be eager or identities will be created / updated
	return p.GetConnection(ctx).Update(s)
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
