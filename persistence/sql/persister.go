package sql

import (
	"context"
	"io"

	"github.com/gobuffalo/pop/v5"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
)

var _ persistence.Persister = new(Persister)

var migrations = pkger.Dir("/persistence/sql/migrations/sql") // do not remove this!

type (
	persisterDependencies interface {
		IdentityTraitsSchemas() schema.Schemas
		identity.ValidationProvider
		x.LoggingProvider
	}
	Persister struct {
		c        *pop.Connection
		mb       x.MigrationPkger
		r        persisterDependencies
		cf       configuration.Provider
		isSQLite bool
	}
)

func NewPersister(r persisterDependencies, conf configuration.Provider, c *pop.Connection) (*Persister, error) {
	m, err := x.NewPkgerMigration(migrations, c, r)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Persister{c: c, mb: m, cf: conf, r: r, isSQLite: c.Dialect.Name() == "sqlite3"}, nil
}

func (p *Persister) Connection() *pop.Connection {
	return p.c
}

func (p *Persister) MigrationStatus(ctx context.Context, w io.Writer) error {
	return errors.WithStack(p.mb.Status(w))
}

func (p *Persister) MigrateDown(ctx context.Context, steps int) error {
	return errors.WithStack(p.mb.Down(steps))
}

func (p *Persister) MigrateUp(ctx context.Context) error {
	if err := p.mb.Up(); err != nil {
		return errors.WithStack(err)
	}

	return identity.GuaranteeCredentialTypes(p.c, p.r.Logger())
}

func (p *Persister) Close(ctx context.Context) error {
	return errors.WithStack(p.GetConnection(ctx).Close())
}

func (p *Persister) Ping(ctx context.Context) error {
	type pinger interface {
		Ping() error
	}

	return errors.WithStack(p.GetConnection(ctx).Store.(pinger).Ping())
}
