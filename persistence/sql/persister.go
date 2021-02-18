package sql

import (
	"context"
	"embed"

	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"

	"github.com/ory/x/popx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
)

var _ persistence.Persister = new(Persister)

//go:embed migrations/sql/*.sql
var migrations embed.FS

type (
	persisterDependencies interface {
		IdentityTraitsSchemas(ctx context.Context) schema.Schemas
		identity.ValidationProvider
		x.LoggingProvider
		config.Provider
		x.TracingProvider
	}
	Persister struct {
		c        *pop.Connection
		mb       *popx.MigrationBox
		r        persisterDependencies
		isSQLite bool
	}
)

func NewPersister(ctx context.Context, r persisterDependencies, c *pop.Connection) (*Persister, error) {
	m, err := popx.NewMigrationBox(migrations, popx.NewMigrator(c, r.Logger(), r.Tracer(ctx), 0))
	if err != nil {
		return nil, err
	}

	return &Persister{c: c, mb: m, r: r, isSQLite: c.Dialect.Name() == "sqlite3"}, nil
}

func (p *Persister) Connection(ctx context.Context) *pop.Connection {
	return p.c.WithContext(ctx)
}

func (p *Persister) MigrationStatus(ctx context.Context) (popx.MigrationStatuses, error) {
	status, err := p.mb.Status(ctx)
	if err != nil {
		return nil, err
	}

	return status, nil
}

func (p *Persister) MigrateDown(ctx context.Context, steps int) error {
	return p.mb.Down(ctx, steps)
}

func (p *Persister) MigrateUp(ctx context.Context) error {
	return p.mb.Up(ctx)
}

func (p *Persister) Migrator() *popx.Migrator {
	return p.mb.Migrator
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
