package sql

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/session"
)

var _ session.Persister = new(Persister)

func (p *Persister) GetSession(ctx context.Context, sid uuid.UUID) (*session.Session, error) {
	var s session.Session
	if err := p.GetConnection(ctx).Find(&s, sid); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	i, err := p.GetIdentity(ctx, s.IdentityID)
	if err != nil {
		return nil, err
	}
	s.Identity = i
	return &s, nil
}

func (p *Persister) CreateSession(ctx context.Context, s *session.Session) error {
	return p.GetConnection(ctx).Create(s) // This must not be eager or identities will be created / updated
}

func (p *Persister) DeleteSession(ctx context.Context, sid uuid.UUID) error {
	return p.GetConnection(ctx).Destroy(&session.Session{ID: sid}) // This must not be eager or identities will be created / updated
}

func (p *Persister) DeleteSessionsFor(ctx context.Context, sid uuid.UUID) error {
	if err := p.GetConnection(ctx).RawQuery("DELETE FROM sessions WHERE identity_id =?", sid).Exec(); err != nil {
		return sqlcon.HandleError(err)
	}
	return nil
}
