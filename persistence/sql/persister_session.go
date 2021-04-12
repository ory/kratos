package sql

import (
	"context"
	"fmt"
	"github.com/ory/kratos/corp"
	"github.com/pkg/errors"

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

func (p *Persister) CreateSession(ctx context.Context, s *session.Session) error {
	s.NID = corp.ContextualizeNID(ctx, p.nid)
	return p.GetConnection(ctx).Create(s) // This must not be eager or identities will be created / updated
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
