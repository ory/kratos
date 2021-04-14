package settings

import (
	"context"

	"github.com/gofrs/uuid"
)

type (
	FlowPersister interface {
		CreateSettingsFlow(context.Context, *Flow) error
		GetSettingsFlow(ctx context.Context, id uuid.UUID) (*Flow, error)
		UpdateSettingsFlow(context.Context, *Flow) error
	}
	FlowPersistenceProvider interface {
		SettingsFlowPersister() FlowPersister
	}
)
