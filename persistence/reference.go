package persistence

import (
	"context"

	"github.com/ory/x/networkx"

	"github.com/gofrs/uuid"

	"github.com/gobuffalo/pop/v6"

	"github.com/ory/x/popx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/session"
)

type Provider interface {
	Persister() Persister
	SetPersister(Persister)
}

type Persister interface {
	continuity.Persister
	identity.PrivilegedPool
	registration.FlowPersister
	login.FlowPersister
	settings.FlowPersister
	courier.Persister
	session.Persister
	errorx.Persister
	verification.FlowPersister
	recovery.FlowPersister
	link.RecoveryTokenPersister
	link.VerificationTokenPersister

	Close(context.Context) error
	Ping() error
	MigrationStatus(c context.Context) (popx.MigrationStatuses, error)
	MigrateDown(c context.Context, steps int) error
	MigrateUp(c context.Context) error
	Migrator() *popx.Migrator
	GetConnection(ctx context.Context) *pop.Connection
	Transaction(ctx context.Context, callback func(ctx context.Context, connection *pop.Connection) error) error
	Networker
}

type Networker interface {
	WithNetworkID(sid uuid.UUID) Persister
	NetworkID() uuid.UUID
	DetermineNetwork(ctx context.Context) (*networkx.Network, error)
}
