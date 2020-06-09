package sql

import (
	"context"
	"io"

	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
)

var _ persistence.Persister = new(Persister)
var migrations = packr.New("migrations", "migrations")

type (
	persisterDependencies interface {
		IdentityTraitsSchemas() schema.Schemas
		identity.ValidationProvider
		x.LoggingProvider
	}
	Persister struct {
		c  *pop.Connection
		mb pop.MigrationBox
		r  persisterDependencies
		cf configuration.Provider
	}
)

func NewPersister(r persisterDependencies, conf configuration.Provider, c *pop.Connection) (*Persister, error) {
	m, err := pop.NewMigrationBox(migrations, c)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Persister{c: c, mb: m, cf: conf, r: r}, nil
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
	return errors.WithStack(p.mb.Up())
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
