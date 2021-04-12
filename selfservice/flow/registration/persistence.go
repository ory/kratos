package registration

import (
	"context"
	"github.com/gofrs/uuid"
)

type FlowPersister interface {
	UpdateRegistrationFlow(context.Context, *Flow) error
	CreateRegistrationFlow(context.Context, *Flow) error
	GetRegistrationFlow(context.Context, uuid.UUID) (*Flow, error)
}

type FlowPersistenceProvider interface {
	RegistrationFlowPersister() FlowPersister
}
