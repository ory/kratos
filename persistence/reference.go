package persistence

import (
	"context"
	"io"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/verify"
	"github.com/ory/kratos/session"
)

type Provider interface {
	Persister() Persister
}

type Persister interface {
	identity.PrivilegedPool
	registration.RequestPersister
	login.RequestPersister
	profile.RequestPersister
	courier.Persister
	session.Persister
	errorx.Persister
	verify.Persister

	Close(context.Context) error
	Ping(context.Context) error
	MigrationStatus(c context.Context, b io.Writer) error
	MigrateDown(c context.Context, steps int) error
	MigrateUp(c context.Context) error
}
