// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type (
	FlowPersister interface {
		CreateSettingsFlow(context.Context, *Flow) error
		GetSettingsFlow(ctx context.Context, id uuid.UUID) (*Flow, error)
		UpdateSettingsFlow(context.Context, *Flow) error
		DeleteExpiredSettingsFlows(context.Context, time.Time, int) error
	}
	FlowPersistenceProvider interface {
		SettingsFlowPersister() FlowPersister
	}
)
