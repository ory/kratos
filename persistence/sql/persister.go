package sql

import (
	"context"
	"io"

	"github.com/gobuffalo/packr"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"

	"github.com/ory/kratos/schema"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/x"
)

var _ persistence.Persister = new(Persister)
var migrations = packr.NewBox("../../contrib/sql/migrations")

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

func (p *Persister) MigrationStatus(c context.Context, w io.Writer) error {
	return errors.WithStack(p.mb.Status(w))
}

func (p *Persister) MigrateDown(c context.Context, steps int) error {
	return errors.WithStack(p.mb.Down(steps))
}

func (p *Persister) MigrateUp(c context.Context) error {
	return errors.WithStack(p.mb.Up())
}

func (p *Persister) Close(c context.Context) error {
	return errors.WithStack(p.c.Close())
}

func (p *Persister) Ping(c context.Context) error {
	type pinger interface {
		Ping() error
	}

	return errors.WithStack(p.c.Store.(pinger).Ping())
}
