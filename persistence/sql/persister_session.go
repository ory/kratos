package sql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gobuffalo/pop/v6"

	"github.com/pkg/errors"

	"github.com/ory/kratos/corp"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/session"
)

var _ session.Persister = new(Persister)

func (p *Persister) GetSession(ctx context.Context, sid uuid.UUID) (*session.Session, error) {
	var s session.Session
	nid := corp.ContextualizeNID(ctx, p.nid)
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", sid, nid).First(&s); err != nil {
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
	var s []*session.Session
	nid := corp.ContextualizeNID(ctx, p.nid)

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

			s.Identity = i
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s, nil
}

func (p *Persister) UpsertSession(ctx context.Context, s *session.Session) error {
	s.NID = corp.ContextualizeNID(ctx, p.nid)

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
	return p.delete(ctx, new(session.Session), sid)
}

func (p *Persister) DeleteSessionsByIdentity(ctx context.Context, identityID uuid.UUID) error {
	// #nosec G201
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE identity_id = ? AND nid = ?",
		corp.ContextualizeTableName(ctx, "sessions"),
	),
		identityID,
		corp.ContextualizeNID(ctx, p.nid),
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
	var s session.Session
	if err := p.GetConnection(ctx).Where("token = ? AND nid = ?",
		token,
		corp.ContextualizeNID(ctx, p.nid),
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
	// #nosec G201
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE token = ? AND nid = ?",
		corp.ContextualizeTableName(ctx, "sessions"),
	),
		token,
		corp.ContextualizeNID(ctx, p.nid),
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
	// #nosec G201
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"UPDATE %s SET active = false WHERE token = ? AND nid = ?",
		corp.ContextualizeTableName(ctx, "sessions"),
	),
		token,
		corp.ContextualizeNID(ctx, p.nid),
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
	// #nosec G201
	err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"UPDATE %s SET active = false WHERE id = ? AND identity_id = ? AND nid = ?",
		corp.ContextualizeTableName(ctx, "sessions"),
	),
		sID,
		iID,
		corp.ContextualizeNID(ctx, p.nid),
	).Exec()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	return nil
}

// RevokeSessionsIdentityExcept marks all except the given session of an identity inactive.
func (p *Persister) RevokeSessionsIdentityExcept(ctx context.Context, iID, sID uuid.UUID) (int, error) {
	// #nosec G201
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"UPDATE %s SET active = false WHERE identity_id = ? AND id != ? AND nid = ?",
		corp.ContextualizeTableName(ctx, "sessions"),
	),
		iID,
		sID,
		corp.ContextualizeNID(ctx, p.nid),
	).ExecWithCount()
	if err != nil {
		return 0, sqlcon.HandleError(err)
	}
	return count, nil
}
