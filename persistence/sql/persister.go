package sql

import (
	"context"
	"io"

	"github.com/ory/x/pkgerx"

	"github.com/gobuffalo/pop/v5"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
)

var _ persistence.Persister = new(Persister)

var migrations = pkger.Dir("github.com/ory/kratos:/persistence/sql/migrations/sql") // do not remove this!

type (
	persisterDependencies interface {
		IdentityTraitsSchemas(ctx context.Context) schema.Schemas
		identity.ValidationProvider
		x.LoggingProvider
		config.Provider
	}
	Persister struct {
		c        *pop.Connection
		mb       *pkgerx.MigrationBox
		r        persisterDependencies
		isSQLite bool
	}
)

func NewPersister(r persisterDependencies, c *pop.Connection) (*Persister, error) {
	m, err := pkgerx.NewMigrationBox(migrations, c, r.Logger())
	if err != nil {
		return nil, err
	}

	return &Persister{c: c, mb: m, r: r, isSQLite: c.Dialect.Name() == "sqlite3"}, nil
}

func (p *Persister) Connection(ctx context.Context) *pop.Connection {
	return p.c.WithContext(ctx)
}

func (p *Persister) MigrationStatus(ctx context.Context, w io.Writer) error {
	return errors.WithStack(p.mb.Status(w))
}

func (p *Persister) MigrateDown(ctx context.Context, steps int) error {
	return errors.WithStack(p.mb.Down(steps))
}

func (p *Persister) MigrateUp(ctx context.Context) error {
	return errors.WithStack(p.mb.Up())
}

func (p *Persister) Close(ctx context.Context) error {
	return errors.WithStack(p.GetConnection(ctx).Close())
}

func (p *Persister) Ping() error {
	type pinger interface {
		Ping() error
	}

	// This can not be contextualized because of some gobuffalo/pop limitations.
	return errors.WithStack(p.c.Store.(pinger).Ping())
}
