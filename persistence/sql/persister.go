package sql

import (
	"context"
	"embed"
	"fmt"

	"github.com/ory/kratos/corp"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/pop/v5/columns"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/networkx"
	"github.com/ory/x/sqlcon"

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
		nid      uuid.UUID
		c        *pop.Connection
		mb       *popx.MigrationBox
		r        persisterDependencies
		p        *networkx.Manager
		isSQLite bool
	}
)

func NewPersister(ctx context.Context, r persisterDependencies, c *pop.Connection) (*Persister, error) {
	m, err := popx.NewMigrationBox(migrations, popx.NewMigrator(c, r.Logger(), r.Tracer(ctx), 0))
	if err != nil {
		return nil, err
	}

	return &Persister{
		c: c, mb: m, r: r, isSQLite: c.Dialect.Name() == "sqlite3",
		p: networkx.NewManager(c, r.Logger(), r.Tracer(ctx)),
	}, nil
}

func (p *Persister) NetworkID() uuid.UUID {
	if p.nid == uuid.Nil {
		panic("NetworkID called before initialized")
	}

	return p.nid
}

func (p Persister) WithNetworkID(sid uuid.UUID) persistence.Persister {
	p.nid = sid
	return &p
}

func (p *Persister) DetermineNetwork(ctx context.Context) (*networkx.Network, error) {
	if err := p.p.MigrateUp(ctx); err != nil {
		return nil, err
	}
	return p.p.Determine(ctx)
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
	if err := p.p.MigrateUp(ctx); err != nil {
		return err
	}

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

type quotable interface {
	Quote(key string) string
}

type node interface {
	GetID() uuid.UUID
	GetNID() uuid.UUID
}

func (p *Persister) update(ctx context.Context, v node, columnNames ...string) error {
	c := p.GetConnection(ctx)
	quoter, ok := c.Dialect.(quotable)
	if !ok {
		return errors.Errorf("store is not a quoter: %T", p.c.Store)
	}

	model := pop.NewModel(v, ctx)
	tn := model.TableName()

	cols := columns.Columns{}
	if len(columnNames) > 0 && tn == model.TableName() {
		cols = columns.NewColumnsWithAlias(tn, model.As, model.IDField())
		cols.Add(columnNames...)
	} else {
		cols = columns.ForStructWithAlias(v, tn, model.As, model.IDField())
	}

	stmt := fmt.Sprintf("SELECT COUNT(id) FROM %s AS %s WHERE %s.id = ? AND %s.nid = ?",
		quoter.Quote(model.TableName()),
		model.Alias(),
		model.Alias(),
		model.Alias(),
	)

	var count int
	if err := c.Store.GetContext(ctx, &count, c.Dialect.TranslateSQL(stmt), v.GetID(), v.GetNID()); err != nil {
		return sqlcon.HandleError(err)
	} else if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}

	stmt = fmt.Sprintf("UPDATE %s AS %s SET %s WHERE %s AND %s.nid = :nid",
		quoter.Quote(model.TableName()),
		model.Alias(),
		cols.Writeable().QuotedUpdateString(quoter),
		model.WhereNamedID(),
		model.Alias(),
	)

	if _, err := c.Store.NamedExecContext(ctx, stmt, v); err != nil {
		return sqlcon.HandleError(err)
	}
	return nil
}

func (p *Persister) delete(ctx context.Context, v interface{}, id uuid.UUID) error {
	nid := corp.ContextualizeNID(ctx, p.nid)

	tabler, ok := v.(interface {
		TableName(ctx context.Context) string
	})
	if !ok {
		return errors.Errorf("expected model to have TableName signature but got: %T", v)
	}

	/* #nosec G201 TableName is static */
	count, err := p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE id = ? AND nid = ?", tabler.TableName(ctx)),
		id,
		nid,
	).ExecWithCount()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}
	return nil
}
