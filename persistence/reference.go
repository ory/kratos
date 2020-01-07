package persistence

import (
	"context"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
)

type Provider interface {
	Persister() Persister
}

type Persister interface {
	identity.Pool
	schema.Persister
	registration.RequestPersister
	login.RequestPersister
	profile.RequestPersister
	courier.Persister
	session.Persister
	errorx.Persister

	Close(context.Context) error
	Ping(context.Context) error
	MigrationStatus(c context.Context) error
	MigrateDown(c context.Context, steps int) error
	MigrateUp(c context.Context) error
}
