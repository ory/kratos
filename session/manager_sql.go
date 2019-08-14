package session

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"

	"github.com/ory/herodot"
)

var _ Manager = new(ManagerSQL)

type ManagerSQL struct {
	*ManagerHTTP
	db *sqlx.DB
}

type sessionSQL struct {
	SID             string    `db:"sid"`
	ExpiresAt       time.Time `db:"expires_at"`
	AuthenticatedAt time.Time `db:"authenticated_at"`
	IssuedAt        time.Time `db:"issued_at"`
	IdentityPK      uint64    `db:"identity_pk"`
	IdentityID      string    `db:"identity_id"`
}

const sessionSQLTableName = "session"

func NewManagerSQL(c Configuration, r Registry, db *sqlx.DB) *ManagerSQL {
	m := &ManagerSQL{db: db}
	m.ManagerHTTP = NewManagerHTTP(c, r, m)
	return m
}

func (s *ManagerSQL) Get(ctx context.Context, sid string) (*Session, error) {
	var interim sessionSQL
	columns, _ := sqlxx.NamedInsertArguments(interim, "pk")
	query := fmt.Sprintf("SELECT %s, i.id FROM %s JOIN identity as i ON (i.pk = identity_pk) WHERE sid=?", columns, sessionSQLTableName)
	if err := sqlcon.HandleError(s.db.GetContext(ctx, &interim, s.db.Rebind(query), sid)); err != nil {
		if errors.Cause(err) == sqlcon.ErrNoRows {
			return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("%s", err))
		}
		return nil, err
	}

	i, err := s.r.IdentityPool().Get(ctx, interim.IdentityID)
	if err != nil {
		return nil, err
	}

	return &Session{
		SID:             interim.SID,
		ExpiresAt:       interim.ExpiresAt,
		AuthenticatedAt: interim.AuthenticatedAt,
		IssuedAt:        interim.IssuedAt,
		Identity:        i,
	}, nil
}

func (s *ManagerSQL) Create(ctx context.Context, session *Session) error {
	var pk uint64
	if err := sqlcon.HandleError(s.db.GetContext(ctx, &pk, s.db.Rebind("SELECT pk FROM identity WHERE id=?"), session.Identity.ID)); err != nil {
		if errors.Cause(err) == sqlcon.ErrNoRows {
			return errors.WithStack(herodot.ErrNotFound.WithReasonf("%s", err))
		}
		return err
	}

	insert := &sessionSQL{
		SID:             session.SID,
		ExpiresAt:       session.ExpiresAt,
		AuthenticatedAt: session.AuthenticatedAt,
		IssuedAt:        session.IssuedAt,
		IdentityPK:      pk,
	}

	columns, arguments := sqlxx.NamedInsertArguments(insert, "identity_id")
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", sessionSQLTableName, columns, arguments)
	if _, err := s.db.NamedExecContext(ctx, query, insert); err != nil {
		return sqlcon.HandleError(err)
	}

	return nil
}

func (s *ManagerSQL) Delete(ctx context.Context, sid string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE sid = ?", sid)
	_, err := s.db.ExecContext(ctx, s.db.Rebind(query), sid)
	return sqlcon.HandleError(err)
}
