// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"embed"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/laher/mergefs"
	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/persistence/sql/devices"
	idpersistence "github.com/ory/kratos/persistence/sql/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/contextx"
	"github.com/ory/x/networkx"
	"github.com/ory/x/popx"
)

var _ persistence.Persister = new(Persister)

//go:embed migrations/sql/*.sql
var migrations embed.FS

type (
	persisterDependencies interface {
		x.LoggingProvider
		config.Provider
		contextx.Provider
		x.TracingProvider
		schema.IdentityTraitsProvider
		identity.ValidationProvider
	}
	Persister struct {
		nid uuid.UUID
		c   *pop.Connection
		mb  *popx.MigrationBox
		mbs popx.MigrationStatuses
		r   persisterDependencies
		p   *networkx.Manager

		identity.PrivilegedPool
		session.DevicePersister
	}
)

func NewPersister(ctx context.Context, r persisterDependencies, c *pop.Connection) (*Persister, error) {
	m, err := popx.NewMigrationBox(mergefs.Merge(migrations, networkx.Migrations), popx.NewMigrator(c, r.Logger(), r.Tracer(ctx), 0))
	if err != nil {
		return nil, err
	}
	m.DumpMigrations = false

	return &Persister{
		c:               c,
		mb:              m,
		r:               r,
		PrivilegedPool:  idpersistence.NewPersister(r, c),
		DevicePersister: devices.NewPersister(r, c),
		p:               networkx.NewManager(c, r.Logger(), r.Tracer(ctx)),
	}, nil
}

func (p *Persister) NetworkID(ctx context.Context) uuid.UUID {
	return p.r.Contextualizer().Network(ctx, p.nid)
}

func (p Persister) WithNetworkID(nid uuid.UUID) persistence.Persister {
	p.nid = nid
	if pp, ok := p.PrivilegedPool.(interface {
		WithNetworkID(uuid.UUID) identity.PrivilegedPool
	}); ok {
		p.PrivilegedPool = pp.WithNetworkID(nid)
	}
	if dp, ok := p.DevicePersister.(interface {
		WithNetworkID(uuid.UUID) session.DevicePersister
	}); ok {
		p.DevicePersister = dp.WithNetworkID(nid)
	}
	return &p
}

func (p *Persister) DetermineNetwork(ctx context.Context) (*networkx.Network, error) {
	return p.p.Determine(ctx)
}

func (p *Persister) Connection(ctx context.Context) *pop.Connection {
	return p.c.WithContext(ctx)
}

func (p *Persister) MigrationStatus(ctx context.Context) (popx.MigrationStatuses, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.MigrationStatus")
	defer span.End()

	if p.mbs != nil {
		return p.mbs, nil
	}

	status, err := p.mb.Status(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if !status.HasPending() {
		p.mbs = status
	}

	return status, nil
}

func (p *Persister) MigrateDown(ctx context.Context, steps int) error {
	return p.mb.Down(ctx, steps)
}

func (p *Persister) MigrateUp(ctx context.Context) error {
	return p.mb.Up(ctx)
}

func (p *Persister) MigrationBox() *popx.MigrationBox {
	return p.mb
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

func (p *Persister) CleanupDatabase(ctx context.Context, wait time.Duration, older time.Duration, batchSize int) error {
	currentTime := time.Now().Add(-older)
	p.r.Logger().Printf("Cleaning up records older than %s\n", currentTime)

	p.r.Logger().Println("Cleaning up expired sessions")
	if err := p.DeleteExpiredSessions(ctx, currentTime, batchSize); err != nil {
		return err
	}
	time.Sleep(wait)

	p.r.Logger().Println("Cleaning up expired continuity containers")
	if err := p.DeleteExpiredContinuitySessions(ctx, currentTime, batchSize); err != nil {
		return err
	}
	time.Sleep(wait)

	p.r.Logger().Println("Cleaning up expired login flows")
	if err := p.DeleteExpiredLoginFlows(ctx, currentTime, batchSize); err != nil {
		return err
	}
	time.Sleep(wait)

	p.r.Logger().Println("Cleaning up expired recovery flows")
	if err := p.DeleteExpiredRecoveryFlows(ctx, currentTime, batchSize); err != nil {
		return err
	}
	time.Sleep(wait)

	p.r.Logger().Println("Cleaning up expired registation flows")
	if err := p.DeleteExpiredRegistrationFlows(ctx, currentTime, batchSize); err != nil {
		return err
	}
	time.Sleep(wait)

	p.r.Logger().Println("Cleaning up expired settings flows")
	if err := p.DeleteExpiredSettingsFlows(ctx, currentTime, batchSize); err != nil {
		return err
	}
	time.Sleep(wait)

	p.r.Logger().Println("Cleaning up expired verification flows")
	if err := p.DeleteExpiredVerificationFlows(ctx, currentTime, batchSize); err != nil {
		return err
	}
	time.Sleep(wait)

	p.r.Logger().Println("Cleaning up expired session token exchangers")
	if err := p.DeleteExpiredExchangers(ctx, currentTime, batchSize); err != nil {
		return err
	}
	time.Sleep(wait)

	p.r.Logger().Println("Successfully cleaned up the latest batch of the SQL database! " +
		"This should be re-run periodically, to be sure that all expired data is purged.")
	return nil
}
