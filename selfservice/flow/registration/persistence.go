// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type FlowPersister interface {
	UpdateRegistrationFlow(context.Context, *Flow) error
	CreateRegistrationFlow(context.Context, *Flow) error
	GetRegistrationFlow(context.Context, uuid.UUID) (*Flow, error)
	DeleteExpiredRegistrationFlows(context.Context, time.Time, int) error
}

type FlowPersistenceProvider interface {
	RegistrationFlowPersister() FlowPersister
}
