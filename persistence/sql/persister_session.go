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
	if err := p.c.Eager().Find(&s, sid); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	if err := p.injectTraitsSchemaURL(s.Identity); err != nil {
		return nil, err
	}
	return &s, nil
}

func (p *Persister) CreateSession(ctx context.Context, s *session.Session) error {
	return p.c.Create(s) // This must not be eager or identities will be created / updated
}

func (p *Persister) DeleteSession(ctx context.Context, sid uuid.UUID) error {
	return p.c.Destroy(&session.Session{ID: sid}) // This must not be eager or identities will be created / updated
}

func (p *Persister) DeleteSessionsFor(ctx context.Context, sid uuid.UUID) error {
	if err := p.c.RawQuery("DELETE FROM sessions WHERE identity_id =?", sid).Exec(); err != nil {
		return sqlcon.HandleError(err)
	}
	return nil
}
