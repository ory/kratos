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
