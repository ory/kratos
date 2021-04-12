package login

import (
	"context"

	"github.com/gofrs/uuid"
)

type (
	FlowPersister interface {
		UpdateLoginFlow(context.Context, *Flow) error
		CreateLoginFlow(context.Context, *Flow) error
		GetLoginFlow(context.Context, uuid.UUID) (*Flow, error)
		ForceLoginFlow(ctx context.Context, id uuid.UUID) error
	}
	FlowPersistenceProvider interface {
		LoginFlowPersister() FlowPersister
	}
)
