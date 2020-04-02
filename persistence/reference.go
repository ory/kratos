package persistence

import (
	"context"
	"io"

	"github.com/gobuffalo/pop/v5"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verify"
	"github.com/ory/kratos/session"
)

type Provider interface {
	Persister() Persister
}

type Persister interface {
	continuity.Persister
	identity.PrivilegedPool
	registration.RequestPersister
	login.RequestPersister
	settings.RequestPersister
	courier.Persister
	session.Persister
	errorx.Persister
	verify.Persister

	Close(context.Context) error
	Ping(context.Context) error
	MigrationStatus(c context.Context, b io.Writer) error
	MigrateDown(c context.Context, steps int) error
	MigrateUp(c context.Context) error
	GetConnection(ctx context.Context) *pop.Connection
	Transaction(ctx context.Context, callback func(connection *pop.Connection) error) error
}
