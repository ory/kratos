// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package persistence

import (
	"context"
	"time"

	"github.com/ory/kratos/x"

	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/x/networkx"

	"github.com/gofrs/uuid"

	"github.com/ory/pop/v6"

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
	"github.com/ory/kratos/selfservice/strategy/code"
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
	sessiontokenexchange.Persister
	errorx.Persister
	verification.FlowPersister
	recovery.FlowPersister
	link.RecoveryTokenPersister
	link.VerificationTokenPersister
	code.RecoveryCodePersister
	code.VerificationCodePersister
	code.RegistrationCodePersister
	code.LoginCodePersister

	CleanupDatabase(context.Context, time.Duration, time.Duration, int) error
	Close(context.Context) error
	Ping(context.Context) error
	MigrationStatus(context.Context) (popx.MigrationStatuses, error)
	MigrateDown(ctx context.Context, steps int) error
	MigrateUp(context.Context) error
	MigrationBox() *popx.MigrationBox
	GetConnection(context.Context) *pop.Connection
	Connection(ctx context.Context) *pop.Connection
	x.TransactionalPersister
	Networker
}

type Networker interface {
	WithNetworkID(nid uuid.UUID) Persister
	NetworkID(ctx context.Context) uuid.UUID
	DetermineNetwork(ctx context.Context) (*networkx.Network, error)
}
